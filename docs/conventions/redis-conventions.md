# Redis Conventions

> **When writing or reviewing any Redis code, read this file first.**
>
> Use [github.com/redis/go-redis/v9](https://github.com/redis/go-redis) as the standard client.
> All rules below apply to every Redis interaction in this repository.

---

## Table of Contents

- [Client Setup & Connection Pooling](#client-setup--connection-pooling)
- [Interface / Port Pattern](#interface--port-pattern)
- [Key Naming Conventions](#key-naming-conventions)
- [Context Propagation](#context-propagation)
- [Data Types](#data-types)
- [Error Handling](#error-handling)
- [TTL / Expiration Discipline](#ttl--expiration-discipline)
- [Cache Invalidation](#cache-invalidation)
- [Pipelining & Batching](#pipelining--batching)
- [Transactions (MULTI/EXEC) and Optimistic Locking](#transactions-multiexec-and-optimistic-locking)
- [Lua Scripting](#lua-scripting)
- [Pub/Sub](#pubsub)
- [Testing](#testing)
- [Observability](#observability)
- [Memory Management](#memory-management)

---

## Client Setup & Connection Pooling

- Instantiate **one** `redis.Client` (or `redis.ClusterClient`) per application and share it via dependency injection — never create a new client per request.
- Configure the pool explicitly; do not rely on defaults for production workloads:
  ```go
  rdb := redis.NewClient(&redis.Options{
    Addr:         cfg.RedisAddr,
    PoolSize:     10,              // total max connections in pool
    MinIdleConns: 2,
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
  })
  ```
- Call `rdb.Close()` in the application shutdown path (e.g. inside `run()` via `defer`).
- Verify connectivity at startup with `rdb.Ping(ctx)` and fail fast if Redis is unreachable.

---

## Interface / Port Pattern

- Never import `*redis.Client` directly into domain packages. Define a narrow port interface and inject it:
  ```go
  // port — lives in the domain or shared/ports package
  type Cache interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    Del(ctx context.Context, keys ...string) error
  }
  ```
- The concrete adapter (wrapping `*redis.Client`) belongs in the integrations layer.
- This pattern keeps domain tests free of Redis and lets you swap implementations (e.g. `miniredis` in tests).

---

## Key Naming Conventions

- Keys must be **short, descriptive, and use alphanumeric characters only** — no special characters other than `:` and `_`.
- Use **colon `:` as the namespace/hierarchy separator** and **underscore `_` to separate words** within a segment.
- Use a **hierarchical structure** that includes environment, namespace/module, entity, and ID to avoid cross-service or cross-environment collisions:
  ```
  <env>:<namespace>:<entity>:<id>[:<qualifier>]

  staging:dop:dj_api:app:123
  prod:auth:session:u_456
  dev:rate_limit:ip:192_0_2_1
  ```
- Define key templates as typed constants or constructor functions — never build keys with ad-hoc `fmt.Sprintf` calls scattered across handlers:
  ```go
  func sessionKey(env, userID string) string {
    return fmt.Sprintf("%s:auth:session:u_%s", env, userID)
  }
  ```
- Avoid overly broad keys (e.g. `cache:*`) that make scanning and eviction unpredictable.

---

## Context Propagation

- Pass `context.Context` as the **first argument** to every Redis call. Never use `context.Background()` inside a handler — propagate the request context so timeouts and cancellations are respected:
  ```go
  val, err := rdb.Get(ctx, key).Result()
  ```

---

## Data Types

Choose the most memory-efficient Redis data type for the task; avoid storing everything as a plain `String`:

| Use case | Recommended type |
|---|---|
| Single scalar value / token | `String` |
| Structured object with named fields | `Hash` |
| Ordered collection, leaderboard | `Sorted Set (ZSet)` |
| Unique membership test | `Set` |
| Queue / stack | `List` |
| Approximate counting / membership | `HyperLogLog` / `Bloom filter` |

- Prefer `Hash` over multiple `String` keys for the same object — it reduces key-count overhead and allows partial field updates with `HSET`.
- Never store large blobs (> a few KB) without measuring the impact on memory and serialization time.

---

## Error Handling

- Distinguish `redis.Nil` (key not found — an expected condition) from real errors:
  ```go
  val, err := rdb.Get(ctx, key).Result()
  if errors.Is(err, redis.Nil) {
    // cache miss — handle gracefully
    return "", ErrNotFound
  }
  if err != nil {
    return "", fmt.Errorf("cache get %q: %w", key, err)
  }
  ```
- Never silently discard Redis errors; log or propagate them with context.
- Wrap errors at the adapter boundary using `%w` so callers can use `errors.Is`/`errors.As`.

---

## TTL / Expiration Discipline

- **Every** key written to Redis must have an explicit TTL. Omitting a TTL risks unbounded memory growth:
  ```go
  rdb.Set(ctx, key, value, 24*time.Hour)
  ```
- Define TTL values as named constants or configuration, not magic numbers.
- Use `EXPIREAT` / `ExpireAt` when the expiry must align with a wall-clock boundary (e.g. end of day).
- Never rely on Redis eviction policies as a substitute for intentional TTLs.

---

## Cache Invalidation

Choose an invalidation strategy explicitly; do not leave it up to TTL alone:

- **TTL-based expiration** (passive): set an appropriate TTL on every write and let Redis expire the key automatically. Use for read-heavy data that tolerates brief staleness.
- **Active invalidation on write**: delete or update the cache key immediately after the source-of-truth record changes. Keeps the cache consistent but couples write paths to the cache layer.
- **Event-driven invalidation**: consume a change-data-capture or domain event stream to invalidate keys asynchronously. Decouples the write path but introduces eventual consistency.
- **Cache-aside (lazy loading)**: read from the source on a miss and populate the cache. Combine with a short TTL to bound staleness.
- Document the chosen strategy in a comment near the adapter so future readers understand the consistency trade-off.

---

## Pipelining & Batching

- Use pipelining when issuing **3 or more** independent commands to reduce round-trip overhead:
  ```go
  pipe := rdb.Pipeline()
  pipe.Set(ctx, "key1", "v1", ttl)
  pipe.Set(ctx, "key2", "v2", ttl)
  _, err := pipe.Exec(ctx)
  ```
- Prefer `MGet` / `MSet` over N sequential `Get` / `Set` calls.
- Keep pipeline batches bounded; limit individual batches to a configurable maximum (e.g. 100–500 commands) and flush in chunks to avoid blocking the event loop.

---

## Transactions (MULTI/EXEC) and Optimistic Locking

- Use `TxPipelined` for atomic multi-key writes that do not require read-your-write consistency:
  ```go
  _, err := rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
    pipe.Set(ctx, "a", 1, ttl)
    pipe.Set(ctx, "b", 2, ttl)
    return nil
  })
  ```
- Use `Watch` + `TxPipelined` for optimistic locking (check-and-set patterns); always handle `redis.TxFailedErr` by retrying or returning a conflict error to the caller.
- For complex atomic logic, prefer a **Lua script** loaded with `ScriptLoad` / `EvalSha` over multiple round-trips.

---

## Lua Scripting

- Keep Lua scripts short and single-purpose; store them as `const` strings or embed them from `.lua` files via `//go:embed`.
- Load scripts once at startup using `ScriptLoad` and call them by SHA with `EvalSha`; fall back to `Eval` only when the script may not be loaded yet.
- Document the expected `KEYS` and `ARGV` arguments with a comment above each script.

---

## Pub/Sub

- Always run `Subscribe`/`PSubscribe` on a **dedicated connection** (not the shared pool client). Use `rdb.Subscribe(ctx, channels...)` which returns a `*redis.PubSub`.
- Read messages in a separate goroutine; respect context cancellation:
  ```go
  sub := rdb.Subscribe(ctx, "events")
  defer sub.Close()
  ch := sub.Channel()
  for {
    select {
    case msg, ok := <-ch:
      if !ok {
        return
      }
      handle(msg.Payload)
    case <-ctx.Done():
      return
    }
  }
  ```
- Unsubscribe and close `PubSub` in the goroutine's cleanup path to avoid connection leaks.

---

## Testing

- Use [github.com/alicebob/miniredis/v2](https://github.com/alicebob/miniredis) for unit and integration tests — never connect to a real Redis instance in automated tests:
  ```go
  func TestCache(t *testing.T) {
    mr := miniredis.RunT(t)
    rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
    // exercise the adapter under test
  }
  ```
- Advance `miniredis` time with `mr.FastForward(d)` to test TTL expiration without real sleeps.
- Test `redis.Nil` paths explicitly; do not assume a cache miss is always correct behaviour.

---

## Observability

- Instrument Redis operations with a hook (`rdb.AddHook(...)`) to emit latency metrics and trace spans — do not add ad-hoc timing code around individual calls.
- Log slow commands (> configured threshold) at `WARN` level with the key name (but never the value if it may contain PII).
- Include the Redis command name and key pattern (not full key) in error context strings.

---

## Memory Management

- **Always estimate memory before deploying** a new key space. A simple model:
  ```
  total memory ≈ (key_size + value_size + SDS_overhead + per-key_metadata) × key_count
  ```
  Where `SDS_overhead` is ~45 bytes per key (Redis 7.x). Use `redis-cli --bigkeys` and `MEMORY USAGE <key>` to validate estimates against a real data sample.
- Avoid keys longer than 64 bytes — long key names inflate the key-space memory with no benefit. Prefer compact prefixes and document the abbreviations.
- Use `OBJECT ENCODING <key>` to confirm Redis chose the most compact internal encoding. Tune `hash-max-listpack-entries` and related config thresholds to keep small objects in compact form.
- Set a `maxmemory` limit and an appropriate `maxmemory-policy` (e.g. `allkeys-lru` or `volatile-lru`) in every environment. Monitor the `used_memory_rss` vs `used_memory` ratio; a high ratio indicates memory fragmentation.
- **Close the client** when the application shuts down — call `rdb.Close()` to return pooled connections and prevent goroutine/file-descriptor leaks:
  ```go
  func run(cfg Config) error {
    rdb := newRedisClient(cfg)
    defer rdb.Close()
    // ...
  }
  ```
