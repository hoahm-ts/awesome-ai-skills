# Kafka Conventions

> **When writing or reviewing any Kafka producer/consumer code, read this file first.**
>
> These guidelines apply to all Go services that produce or consume Kafka messages.

---

## Table of Contents

- [Topic Naming Conventions](#topic-naming-conventions)
- [Message Keys & Partitioning](#message-keys--partitioning)
- [Client Library](#client-library)
- [Producer](#producer)
- [Consumer](#consumer)
- [Error Handling](#error-handling)
- [Schema & Serialization](#schema--serialization)
- [Observability](#observability)
- [Testing](#testing)
- [Configuration](#configuration)
- [Graceful Shutdown](#graceful-shutdown)

---

## Topic Naming Conventions

- Use a dot-separated, lowercase hierarchy that encodes ownership and intent:
  ```
  <domain>.<entity>.<event>
  ```
  Examples: `payments.order.created`, `inventory.product.updated`, `auth.user.deleted`.
- Use only lowercase letters, digits, hyphens (`-`), and dots (`.`). Never use underscores within a segment or uppercase — some tooling treats `.` and `_` differently, and mixed case causes confusion across teams. The sole exception is a leading `_` on internal topics (see below).
- Keep each segment short and singular: `order` not `orders`, `product` not `products`. The domain segment may be plural when the service name is conventionally plural (e.g. `payments`).
- Dead-letter topics mirror their source topic with a `.dlt` suffix:
  ```
  payments.order.created.dlt
  ```
- Retry topics (when using staged retries) append `.retry.<n>`:
  ```
  payments.order.created.retry.1
  payments.order.created.retry.2
  ```
- Internal or compacted state topics (e.g. changelog topics for Kafka Streams) use a `_` prefix to signal they are infrastructure-level and not consumed directly by application code:
  ```
  _payments.order.state
  ```
- Declare topic names as typed constants in a shared `ports` or `topics` package; never hard-code raw strings across services:
  ```go
  // ports/topics.go
  const (
    TopicOrderCreated    = "payments.order.created"
    TopicOrderCreatedDLT = "payments.order.created.dlt"
  )
  ```
- Document each topic in the codebase with an inline comment: the event schema it carries, the producing service, and the expected consumers.

---

## Message Keys & Partitioning

- Use a stable, deterministic business key (e.g. entity ID or a composite key) as the message key to route all events for the same entity to the same partition, preserving total order for that entity:
  ```go
  // Stable composite key — always routes to the same partition.
  key := fmt.Sprintf("%s:%s", orderID, customerID)
  ```
- Keep key cardinality proportional to the partition count: too few unique keys leave partitions underutilised; too many unique keys cannot improve ordering and may create hot partitions.
- Use `null` keys only for purely parallel, order-independent workloads (e.g. fire-and-forget notifications). Messages with `null` keys are distributed round-robin across partitions.
- Never use random or time-based values (e.g. UUIDs, timestamps) as keys — they destroy ordering guarantees and create uneven partition load over time.
- Encode keys as UTF-8 strings (entity ID, or fields joined with `:`) rather than opaque bytes to aid debugging and re-processing.
- Scale consumer group instances up to — but never beyond — the partition count; extra instances remain idle and waste resources.
- To achieve parallel processing within a topic, increase the partition count at topic creation time or use `null` keys where ordering is not required. Partition counts cannot be reduced after creation.

---

## Client Library

- Pick one library and use it consistently across all services: [franz-go](https://github.com/twmb/franz-go) (recommended for its idiomatic Go API and first-class context support) or [confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go).
- Wrap the Kafka client behind a port interface so it can be replaced or mocked without touching domain code:
  ```go
  // ports/kafka.go
  type Producer interface {
    Produce(ctx context.Context, msg *Message) error
    Close() error
  }

  type Consumer interface {
    Subscribe(topics []string) error
    Poll(ctx context.Context) (*Message, error)
    CommitOffsets(ctx context.Context) error
    Close() error
  }
  ```
- Register concrete client implementations in the composition root; never instantiate them inside domain packages.

---

## Producer

- Use `acks=all` (or `RequiredAcks: kgo.AllISRAcks()`) to ensure durable writes to all in-sync replicas.
- Enable idempotent producers to prevent duplicate messages on retry (`idempotent: true` / `kgo.ProducerIdempotent()`).
- Never fire-and-forget produce calls — always check the delivery error before advancing:
  ```go
  if err := producer.Produce(ctx, msg); err != nil {
    return fmt.Errorf("produce to %s: %w", msg.Topic, err)
  }
  ```
- Key messages consistently: use a stable business key (e.g. entity ID) to preserve ordering within a partition.
- Tune `linger.ms` and `batch.size` deliberately: lower values favour latency; higher values favour throughput. Document your chosen values and the reason.
- Enable compression at the producer level to reduce broker storage and network I/O. Prefer `snappy` for balanced CPU/ratio; use `zstd` for maximum compression on high-volume topics:
  ```go
  kgo.ProducerBatchCompression(kgo.SnappyCompression())
  ```
- Pass `context.Context` to every produce call to support timeouts and cancellation:
  ```go
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  if err := producer.Produce(ctx, msg); err != nil { ... }
  ```

---

## Consumer

- **Consumer Groups**: assign a unique `group.id` per logical consumer (one consumer group per independent processing pipeline). Never share a group ID between unrelated services — this creates split-brain processing where each service receives only a fraction of messages.
- Commit offsets **only after** processing succeeds to guarantee at-least-once delivery; never auto-commit.
- Handle rebalances explicitly: flush or checkpoint in-flight work inside `ConsumerGroupHandler.Cleanup` before returning, so no messages are lost mid-batch.
- Set `max.poll.interval.ms` long enough to cover the worst-case processing time; breach causes the consumer to leave the group and trigger rebalance.
- **Batch Fetching**: tune `fetch.min.bytes` and `fetch.max.wait.ms` to control the trade-off between latency and throughput:
  ```go
  kgo.FetchMinBytes(1<<10),        // 1 KiB minimum before returning
  kgo.FetchMaxWait(500*time.Millisecond),
  ```
- **Multi-threaded Consumers**: process messages in parallel using a fixed goroutine pool fed by a bounded channel. Keep the poller single-threaded and fan out to workers:
  ```go
  work := make(chan *kgo.Record, cfg.WorkerBuffer)

  go poll(ctx, client, work) // single poller goroutine

  var wg sync.WaitGroup
  for i := 0; i < cfg.Workers; i++ {
    wg.Add(1)
    go func() {
      defer wg.Done()
      for rec := range work {
        process(ctx, rec)
      }
    }()
  }
  ```
- **Context Usage**: pass a `context.Context` derived from the application lifecycle into every poll and process call. Use `context.WithTimeout` for individual message processing to enforce per-message deadlines:
  ```go
  processCtx, cancel := context.WithTimeout(ctx, cfg.ProcessingTimeout)
  defer cancel()
  if err := handler.Handle(processCtx, rec); err != nil { ... }
  ```
- Always perform a **graceful shutdown**: stop polling new messages → drain the work channel → commit offsets → close the client.

---

## Error Handling

- Distinguish transient errors (network timeouts, leader elections) from fatal errors (schema decode failure, business rule violation):
  - **Transient**: retry with exponential back-off.
  - **Fatal**: route the raw message to a dead-letter topic (DLT) with headers describing the failure; do not retry.
- Never silently drop messages — every failure must be either retried, sent to the DLT, or result in an explicit alert.
- Wrap Kafka errors with context before returning:
  ```go
  // Bad:  return err
  // Good: return fmt.Errorf("consume %s partition %d offset %d: %w", topic, partition, offset, err)
  ```
- Include structured fields on every error log: `topic`, `partition`, `offset`, `consumer_group`, and `error`.

---

## Schema & Serialization

- Use a schema registry (Confluent Schema Registry or Apicurio) with Avro, Protobuf, or JSON Schema for all production topics.
- Design schemas for **backwards and forwards compatibility**: add optional fields; never remove or rename existing fields without a migration plan.
- Annotate Go structs with serialization tags that match the schema field names exactly:
  ```go
  type OrderCreated struct {
    OrderID   string `json:"order_id"   avro:"order_id"`
    CreatedAt int64  `json:"created_at" avro:"created_at"` // Unix millis
  }
  ```
- Keep schema version metadata (subject, version, schema ID) in the message headers rather than inside the payload.

---

## Observability

- Emit consumer lag per topic-partition as a gauge metric (label: `topic`, `partition`, `consumer_group`). Alert when lag exceeds a defined threshold.
- Record processing latency as a histogram for every consumed message (label: `topic`, `consumer_group`, `status`).
- Propagate trace context via message headers using the [W3C Trace Context](https://www.w3.org/TR/trace-context/) format; extract it at the consumer and start a child span for each message:
  ```go
  ctx = otel.GetTextMapPropagator().Extract(ctx, kafkaHeaderCarrier(msg.Headers))
  ctx, span := tracer.Start(ctx, "consume "+msg.Topic)
  defer span.End()
  ```
- Log at least: `topic`, `partition`, `offset`, `key`, `consumer_group`, `trace_id`, and processing latency at the DEBUG level on success and ERROR level on failure.

---

## Testing

- Unit-test domain logic and adapter code against the `Producer`/`Consumer` interfaces using fakes or mocks — never connect to a real broker in unit tests.
- Use [`kfake`](https://pkg.go.dev/github.com/twmb/franz-go/pkg/kfake) (franz-go's in-process broker) or [`testcontainers-go`](https://github.com/testcontainers/testcontainers-go) for integration tests that require a real Kafka wire protocol.
- In integration tests, assert:
  - Messages are produced to the correct topic and partition key.
  - Offsets are committed only after successful processing.
  - Fatal messages are routed to the DLT with the expected error headers.
- Never use a shared broker state between test cases — reset or re-create topics between runs to keep tests deterministic.

---

## Configuration

- Consolidate all Kafka settings in a single, explicit config struct; avoid scattered `os.Getenv` calls across packages:
  ```go
  type KafkaConfig struct {
    Brokers         []string      `yaml:"brokers"`
    GroupID         string        `yaml:"group_id"`
    Topics          []string      `yaml:"topics"`
    DeadLetterTopic string        `yaml:"dead_letter_topic"`
    TLSEnabled      bool          `yaml:"tls_enabled"`
    SASLMechanism   string        `yaml:"sasl_mechanism"`
    SessionTimeout  time.Duration `yaml:"session_timeout"`
    Workers         int           `yaml:"workers"`
    WorkerBuffer    int           `yaml:"worker_buffer"`
  }
  ```
- Enable TLS and a SASL mechanism (`SCRAM-SHA-512` or `OAUTHBEARER`) in all non-local environments; never use plaintext in staging or production.
- Document every configuration knob with an inline comment describing the default, valid range, and impact.

---

## Graceful Shutdown

- On receiving a termination signal, follow this sequence to avoid message loss:
  1. Cancel the consumer context to stop polling new messages.
  2. Wait for in-flight processing goroutines to finish (use `sync.WaitGroup`).
  3. Commit the final batch of offsets.
  4. Close the Kafka client.
- Encode this sequence using the goroutine lifecycle pattern:
  ```go
  func (c *Consumer) Run(ctx context.Context) error {
    var wg sync.WaitGroup
    work := make(chan *Message, c.cfg.WorkerBuffer)

    wg.Add(1)
    go func() {
      defer wg.Done()
      c.poll(ctx, work) // stops when ctx is cancelled
    }()

    for i := 0; i < c.cfg.Workers; i++ {
      wg.Add(1)
      go func() {
        defer wg.Done()
        c.process(work)
      }()
    }

    <-ctx.Done()     // wait for external cancel
    close(work)      // signal workers to drain and exit
    wg.Wait()        // wait for all goroutines to finish
    return c.client.CommitOffsets(context.Background())
  }
  ```
- Set a shutdown deadline (e.g. 30 s) by passing a timeout context to `CommitOffsets`; log an error and exit if the deadline is exceeded.
