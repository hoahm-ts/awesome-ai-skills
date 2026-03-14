# AGENTS.md — Unified Instructions for AI Coding Agents & Contributors

This file is the **canonical reference** for all AI coding agents and human contributors working in this repository. All agent-specific instruction files (`CLAUDE.md`, `CODEX.md`, `JUNIE.md`, `.cursorrules`, `.github/copilot-instructions.md`) should be kept consistent with the guidelines here.

---

## Project Overview

`awesome-ai-skills` is a curated collection of configuration files, instructions, and best practices for working with AI coding agents. The goal is to provide a ready-to-use set of files that can be dropped into any project to give AI tools the context they need to be maximally helpful.

---

## Core Principles

These principles are non-negotiable and apply to every change made in this repository.

- **Modular design**: business logic lives in well-defined, bounded modules. Entry points (CLI, HTTP, workers) are thin wrappers that wire dependencies and delegate to the core.
- **One composition root**: dependency wiring happens in a single, explicit place. Avoid hidden global singletons.
- **Edges vs core**:
  - *Edges*: delivery/transport layers (HTTP handlers, CLI commands) and external integrations (third-party clients, adapters).
  - *Core*: domain modules containing business rules and services.
- **Prefer interfaces/ports** when crossing module boundaries. Do not import concrete implementations from one domain into another.
- **Explicit over implicit**: no magic numbers, no unclear abbreviations, no unexplained side effects.

---

## Architecture & Layering Rules

### Allowed dependency direction

```
entry points  ➜  wiring/DI  ➜  domain modules
handlers      ➜  domain services (via interfaces)
integrations  ➜  external services
domain        ➜  shared ports/types + infrastructure abstractions
```

### Avoid / disallow

- Handlers calling data stores directly (unless explicitly part of a thin read-only endpoint).
- Domain A importing concrete packages from Domain B — use shared ports/interfaces instead.
- Shared/utility packages accumulating business logic (avoid "god packages").
- Introducing global state or singletons without a documented reason.

---

## Implementing a Feature (Standard Path)

### Step 1 — Start from the user-facing surface

- Identify the target entry point: HTTP handler, background worker, CLI command, or a combination.
- If it changes an API contract, update the API spec (e.g. OpenAPI) first.
- Add or modify the handler/command: parse and validate input, call a domain service, return a standard response shape. Keep transport concerns out of the domain.

### Step 2 — Implement domain logic

- Add or modify code in the relevant domain module.
- Keep domain logic free of transport (HTTP, CLI) concerns.
- For cross-domain behaviour, define a port interface and inject it via DI.

### Step 3 — Integrations

- Add or modify external client/adapters in the integrations layer.
- Domain code must depend on an interface (port), not the concrete integration.

### Step 4 — Wiring / DI

- Register new dependencies in the composition root.
- Keep wiring code outside domain packages.

### Step 5 — Persistence & migrations

- If schema changes are required, add a migration with safe, forward-only changes.
- Ensure the migration can be applied in CI and locally before opening a PR.

---

## General Coding Guidelines

- Write clean, readable, and well-structured code.
- Prefer explicit over implicit — avoid magic numbers, unclear abbreviations, and unexplained side effects.
- Follow the conventions already established in the codebase.
- Keep functions and modules small and focused on a single responsibility.
- Write meaningful commit messages using the [Conventional Commits](https://www.conventionalcommits.org/) format (e.g. `feat:`, `fix:`, `docs:`, `chore:`).

---

## Testing Expectations

- **Domain logic**: unit tests inside the domain package. Aim for high coverage of business rules.
- **Handler/command logic**: prefer table-driven tests for input validation and response mapping; use fakes or mocks for domain services via interfaces.
- **Integration adapters**: unit-test adapters with mocked HTTP/SDK responses. Do not call real external services in automated tests.
- Avoid flaky tests; tests must be deterministic and independent of external state.

---

## Code Quality Checklist

Before requesting a review, verify:

- [ ] No circular dependencies between modules.
- [ ] Shared/utility packages contain only ports, types, and utilities — no domain behaviour.
- [ ] Handlers do not embed business rules.
- [ ] Integration clients are isolated behind interfaces.
- [ ] Configuration is read from a single, explicit source (no scattered environment reads).
- [ ] DI/wiring compiles cleanly.
- [ ] Tests are added or updated for any new or modified logic.
- [ ] A migration is included if the schema changed, and it is backwards-compatible if needed.

---

## Logging, Errors & Observability

- Include request IDs or correlation IDs wherever available.
- Wrap errors with context at boundaries (handlers, integrations).
- Prefer structured logs over ad-hoc strings.
- Do not log secrets, tokens, credentials, or PII — mask or redact them.

---

## Security & Data Handling

- Treat user and customer data as sensitive by default.
- Never commit secrets; use a secrets/config management solution.
- Validate and authorise in the handler layer; enforce invariants in domain services where appropriate.
- Keep auth and permission logic centralised rather than scattered across handlers.

---

## Documentation Requirements

When adding or altering a capability:

- Update `docs/` for changes that affect system design or architecture.
- Update the API spec if the API contract changes.
- Add inline package/module documentation for non-obvious modules or decisions.
- Keep the `README.md` directory structure table up to date if new files are added.

---

## File & Directory Conventions

- Markdown files use [GitHub Flavored Markdown](https://github.github.com/gfm/).
- JSON files must be valid and formatted with 2-space indentation.
- Keep all AI configuration files at the repository root or inside their designated directories (`.claude/`, `.github/`).

---

## Contribution Guidelines

1. Fork the repository and create a feature branch from `main`.
2. Make the smallest possible changes that fully address the issue.
3. Verify your changes before opening a pull request.
4. Ensure the `README.md` directory structure table stays up to date if new files are added.
5. Open a pull request with a clear title and description (see PR etiquette below).

---

## Git Workflow & Branching Rules

Branch format: `<type>/<ticket>` — ticket format is `JIRA-<number>`. No description suffix. Examples: `spec/JIRA-1`, `feat/JIRA-1`, `fix/JIRA-1`.

---

## Pull Request Guidelines

**Title format:** `<TICKET_NUMBER>: <description>` — e.g. `JIRA-29: init the project structure`

**Description:** Always follow [`.github/PULL_REQUEST_TEMPLATE.md`](.github/PULL_REQUEST_TEMPLATE.md). Fill in all fields.

---

## AI Agent Behaviour

### Always do first

1. **Read this file** before making any changes.
2. Identify the target entry point(s): HTTP handler, worker, CLI command, or migration.
3. Identify the domain module(s) impacted.
4. Identify boundary changes: HTTP contract, external integration, workflow/state, or schema.

### During implementation

- Prefer small, incremental diffs.
- Keep changes localised to one domain when possible.
- Introduce interfaces (ports) when crossing domains or calling integrations.
- Register new dependencies in the composition root / DI wiring.
- Follow existing patterns in adjacent code; do not invent new frameworks.
- Do not modify files unrelated to the task.
- If unsure about a convention, look for existing examples before inventing one.

### Before finishing

- Run existing linters and tests; ensure compile-level correctness at minimum.
- Confirm DI/wiring compiles cleanly if applicable.
- Do not commit secrets, credentials, or sensitive data.
- Follow the directory structure documented in `README.md`.

---

## PR / Change Etiquette

- Keep PRs focused; avoid drive-by refactors unless they are directly necessary.
- Every PR description must include:
  - **Motivation / problem statement** — why this change is needed.
  - **What changed** — a concise summary of what was added, modified, or removed.
  - **How to test** — steps to validate the change locally or in CI.
  - **Migration / rollout notes** — any schema changes, feature flags, or deployment considerations.
- If behaviour changes, state explicitly what the old and new behaviour are.

---

## Go Style Guidelines (Uber Style)

> This section consolidates the key rules from the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md).
> Follow these guidelines for all Go code in this repository.

### Tooling & Linting

- Run `goimports` on save to format code and manage imports.
- Run `golint` / `revive` and `go vet` to check for errors before committing.
- Use [golangci-lint](https://github.com/golangci/golangci-lint) as the lint runner with at minimum: `errcheck`, `goimports`, `revive`, `govet`, and `staticcheck`.
- All code must be error-free under `golint` and `go vet`.

### Interfaces & Types

- **Pointers to interfaces**: almost never use a pointer to an interface; pass interfaces as values.
- **Verify interface compliance** at compile time using blank identifier assignments:
  ```go
  var _ http.Handler = (*Handler)(nil)
  ```
- **Receivers and interfaces**: use value receivers when the method does not modify state; use pointer receivers when it does. Be consistent within a type.
- **Zero-value mutexes**: `sync.Mutex` and `sync.RWMutex` zero values are valid — never use a pointer to a mutex. Do not embed mutexes anonymously; always use a named field instead:
  ```go
  // Good
  type SMap struct {
    mu   sync.Mutex
    data map[string]string
  }
  ```
- **Avoid embedding types in public structs**: embedding leaks implementation details. Prefer explicit delegation methods instead.
- **Avoid using built-in names** (`error`, `string`, `len`, etc.) as variable or field names.

### Slices & Maps

- **Copy at boundaries**: copy slices and maps received as arguments before storing them; return copies rather than internal references to prevent unintended mutation.
- **nil is a valid slice**: return `nil` instead of an empty slice literal (`[]T{}`). Check emptiness with `len(s) == 0` — this works correctly for both `nil` and allocated-but-empty slices, so prefer it over `s == nil`.
- **Initializing maps**: use `make(map[K]V)` for programmatically populated maps; use map literals for fixed sets of elements. Provide a size hint when the size is known:
  ```go
  m := make(map[string]int, len(items))
  ```

### Errors

- **Error types — choose based on need**:
  | Caller needs to match? | Message type | Use |
  |---|---|---|
  | No | static | `errors.New` |
  | No | dynamic | `fmt.Errorf` |
  | Yes | static | exported `var` with `errors.New` |
  | Yes | dynamic | custom `error` type |
- **Error wrapping**: use `%w` when callers should be able to match the underlying error via `errors.Is`/`errors.As`; use `%v` to obfuscate it. Keep context succinct — avoid prefixes like "failed to":
  ```go
  // Bad:  fmt.Errorf("failed to create new store: %w", err)
  // Good: fmt.Errorf("new store: %w", err)
  ```
- **Error naming**: prefix exported error variables with `Err` and unexported ones with `err`; suffix custom error types with `Error` (e.g. `NotFoundError`).
- **Handle errors once**: either log the error or return it — never both. Logging and then returning causes duplicate noise up the call stack.
- **Type assertion failures**: always use the comma-ok form:
  ```go
  t, ok := i.(string)
  if !ok { /* handle */ }
  ```
- **Don't panic**: production code must not panic. Return errors and let callers decide. Use `t.Fatal`/`t.FailNow` in tests instead of `panic`.

### Goroutines

- **No fire-and-forget goroutines**: every goroutine must have a predictable stop condition. Provide a mechanism to signal stop and wait for exit:
  ```go
  stop := make(chan struct{})
  done := make(chan struct{})
  go func() {
    defer close(done)
    for {
      select {
      case <-ticker.C:
        flush()
      case <-stop:
        return
      }
    }
  }()
  close(stop)
  <-done
  ```
- **Wait for goroutines**: use `sync.WaitGroup` for multiple goroutines; use a `done chan struct{}` for a single goroutine.
- **No goroutines in `init()`**: packages that need background goroutines must expose an object with a lifecycle method (`Close`, `Stop`, `Shutdown`) that signals the goroutine to stop and waits for it to exit.
- Use [go.uber.org/goleak](https://pkg.go.dev/go.uber.org/goleak) to test for goroutine leaks in packages that spawn goroutines.

### Concurrency & Atomics

- Use [go.uber.org/atomic](https://pkg.go.dev/go.uber.org/atomic) instead of `sync/atomic` raw types to get compile-time type safety and convenient types like `atomic.Bool`.
- **Avoid mutable globals**: use dependency injection instead of mutating package-level variables.

### Resource Management

- **Use `defer` to clean up** resources (files, locks, etc.) immediately after acquiring them. The readability benefit outweighs the negligible overhead.
- **Channel size is one or none**: channels should be unbuffered or have a size of one. Any other size requires explicit justification.

### Enums & Constants

- **Start enums at one** with `iota + 1` unless the zero value is a meaningful default:
  ```go
  type Operation int
  const (
    Add Operation = iota + 1
    Subtract
    Multiply
  )
  ```
- **Prefix unexported package-level `var`s and `const`s with `_`** to make their global scope obvious (exception: unexported error values use `err` prefix without underscore).
- **Format strings**: declare `Printf`-style format strings as `const`. Name custom `Printf`-style functions with a trailing `f` suffix (e.g. `Wrapf`).

### Time

- Always use the `"time"` package for time values. Use `time.Time` for instants and `time.Duration` for periods. Avoid raw `int`/`float64` for time unless the unit is encoded in the field name (e.g. `IntervalMillis`).
- Use `Time.AddDate` to advance a calendar day; use `Time.Add` for exact durations.

### Serialization

- Always annotate marshaled struct fields with tags:
  ```go
  type Stock struct {
    Price int    `json:"price"`
    Name  string `json:"name"`
  }
  ```

### Program Exit

- Call `os.Exit` or `log.Fatal*` **only in `main()`**. All other functions must return errors.
- Call `os.Exit`/`log.Fatal` **at most once** in `main()`. Encapsulate startup logic in a `run()` function that returns an error:
  ```go
  func main() {
    if err := run(); err != nil {
      fmt.Fprintln(os.Stderr, err)
      os.Exit(1)
    }
  }
  ```

### `init()` Functions

- Avoid `init()` where possible. When unavoidable, `init()` must be fully deterministic, must not depend on other `init()` functions, must not access global/environment state, and must not perform I/O.

### Performance (Hot Path Only)

- **Prefer `strconv` over `fmt`** for primitive-to-string conversions.
- **Avoid repeated `[]byte` conversions** from fixed strings — perform the conversion once and reuse the result.
- **Specify container capacity** when known: `make([]T, 0, n)` for slices and `make(map[K]V, n)` for maps.

### Style

- **Line length**: soft limit of 99 characters; wrap before hitting it.
- **Be consistent**: follow the style already established in the file being edited.
- **Group similar declarations** using `import (...)`, `const (...)`, `var (...)`, `type (...)` blocks. Only group related items together.
- **Import group ordering** (two groups, separated by a blank line):
  1. Standard library
  2. Everything else (apply `goimports` to manage automatically)
- **Import aliasing**: use an alias only when the package name does not match the last element of the import path, or to resolve a direct conflict. Avoid aliases otherwise.
- **Package names**: all lowercase, no underscores or capitals, short and singular (e.g. `url`, not `urls`). Avoid generic names like `util`, `common`, or `lib`.
- **Function names**: use `MixedCaps`. Test functions may use underscores for grouping: `TestMyFunc_WhatIsBeingTested`.
- **Function grouping and ordering**: sort functions in rough call order; group by receiver. Exported functions appear first after type/const/var definitions; `newXYZ`/`NewXYZ` follows the type definition.
- **Reduce nesting**: handle error cases and special conditions first, return or `continue` early to flatten indentation.
- **Unnecessary else**: if both branches set the same variable, eliminate the else:
  ```go
  // Good
  a := 10
  if b {
    a = 100
  }
  ```
- **Top-level variable declarations**: use `var` without an explicit type unless the expression type differs from the desired type.
- **Local variable declarations**: use `:=` when setting an explicit value; use `var` when the zero value is the intent (e.g. `var filtered []int`).
- **Struct initialization**:
  - Always use field names when initializing structs (enforced by `go vet`).
  - Omit zero-value fields unless they provide meaningful context.
  - Use `var user User` for all-zero structs instead of `user := User{}`.
  - Use `&T{Name: "foo"}` instead of `new(T)` for struct references.
- **Struct embedding**: embedded types must appear at the top of the field list with a blank line separating them from regular fields. Embed only when it provides a tangible, semantically appropriate benefit.
- **Naked parameters**: add inline `/* name */` comments for non-obvious boolean or numeric arguments, or better, use custom types instead of raw `bool`/`int`.
- **Raw string literals**: prefer backtick raw strings over escaped strings for readability.
- **Reduce variable scope**: declare variables as close to their use as possible using `if err := ...; err != nil { }` forms where appropriate.

### Patterns

- **Test tables**: use table-driven tests with `t.Run` subtests for repeated logic. Name the slice `tests` and each case `tt`. Use `give` / `want` prefixes for input/output fields. Avoid complex conditional logic or branching inside table test loops — split into separate `Test...` functions instead.
- **Functional options**: for constructors/APIs with three or more optional arguments (or those expected to grow), use the functional options pattern with an `Option` interface and unexported `options` struct rather than long parameter lists or boolean flags.

---

## Kafka Best Practices (Go)

> These guidelines apply to all Go services that produce or consume Kafka messages.
> Follow them alongside the Go Style Guidelines above.

### Client Library

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

### Producer

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

### Consumer

- Commit offsets **only after** processing succeeds to guarantee at-least-once delivery; never auto-commit.
- Assign a unique `group.id` per logical consumer; never share a group ID between unrelated services.
- Handle rebalances explicitly: flush or checkpoint in-flight work inside `ConsumerGroupHandler.Cleanup` before returning, so no messages are lost mid-batch.
- Set `max.poll.interval.ms` long enough to cover the worst-case processing time; breach causes the consumer to leave the group and trigger rebalance.
- Apply back-pressure: pass messages over a bounded channel to the processing goroutine pool; do not accumulate unbounded in memory:
  ```go
  work := make(chan *Message, cfg.WorkerBufferSize) // bounded
  ```
- Always perform a **graceful shutdown**: stop polling new messages → drain the work channel → commit offsets → close the client (see [Graceful Shutdown](#graceful-shutdown) below).

### Error Handling

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

### Schema & Serialization

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

### Observability

- Emit consumer lag per topic-partition as a gauge metric (label: `topic`, `partition`, `consumer_group`). Alert when lag exceeds a defined threshold.
- Record processing latency as a histogram for every consumed message (label: `topic`, `consumer_group`, `status`).
- Propagate trace context via message headers using the [W3C Trace Context](https://www.w3.org/TR/trace-context/) format; extract it at the consumer and start a child span for each message:
  ```go
  ctx = otel.GetTextMapPropagator().Extract(ctx, kafkaHeaderCarrier(msg.Headers))
  ctx, span := tracer.Start(ctx, "consume "+msg.Topic)
  defer span.End()
  ```
- Log at least: `topic`, `partition`, `offset`, `key`, `consumer_group`, `trace_id`, and processing latency at the DEBUG level on success and ERROR level on failure.

### Testing

- Unit-test domain logic and adapter code against the `Producer`/`Consumer` interfaces using fakes or mocks — never connect to a real broker in unit tests.
- Use [`kfake`](https://pkg.go.dev/github.com/twmb/franz-go/pkg/kfake) (franz-go's in-process broker) or [`testcontainers-go`](https://github.com/testcontainers/testcontainers-go) for integration tests that require a real Kafka wire protocol.
- In integration tests, assert:
  - Messages are produced to the correct topic and partition key.
  - Offsets are committed only after successful processing.
  - Fatal messages are routed to the DLT with the expected error headers.
- Never use a shared broker state between test cases — reset or re-create topics between runs to keep tests deterministic.

### Configuration

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

### Graceful Shutdown

- On receiving a termination signal, follow this sequence to avoid message loss:
  1. Cancel the consumer context to stop polling new messages.
  2. Wait for in-flight processing goroutines to finish (use `sync.WaitGroup`).
  3. Commit the final batch of offsets.
  4. Close the Kafka client.
- Encode this sequence using the goroutine lifecycle pattern from the [Goroutines](#goroutines) section:
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
