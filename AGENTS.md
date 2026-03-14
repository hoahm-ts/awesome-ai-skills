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

### General Rules

- Include request IDs or correlation IDs wherever available.
- Wrap errors with context at boundaries (handlers, integrations).
- Prefer structured logs over ad-hoc strings.
- Do not log secrets, tokens, credentials, or PII — mask or redact them.

### Structured Logging with zerolog

Use [zerolog](https://github.com/rs/zerolog) as the standard logging library. It produces zero-allocation JSON logs and integrates cleanly with DataDog and OpenTelemetry.

#### Setup

Initialise a single `zerolog.Logger` in the composition root and inject it via dependency injection. Never use the global `log` package or a package-level logger in domain code.

```go
import (
    "os"

    "github.com/rs/zerolog"
)

func newLogger(level zerolog.Level) zerolog.Logger {
    return zerolog.New(os.Stdout).
        Level(level).
        With().
        Timestamp().
        Logger()
}
```

#### Log levels

| Level | When to use |
|---|---|
| `Trace` | Very verbose, developer-only diagnostics |
| `Debug` | Detailed flow information useful during development |
| `Info` | Normal operational events (service started, request handled) |
| `Warn` | Recoverable anomalies that do not affect the outcome |
| `Error` | Failures that affect the current operation; always attach `Err(err)` |
| `Fatal` | Unrecoverable startup failures — `main()` only |

#### Structured fields

Always add fields with typed methods rather than embedding data in the message string.

```go
// Bad
log.Info().Msgf("user %s created order %d", userID, orderID)

// Good
log.Info().
    Str("user_id", userID).
    Int("order_id", orderID).
    Msg("order created")
```

Standard field names to use consistently across the service:

| Field | Type | Description |
|---|---|---|
| `request_id` | `string` | Incoming request / trace ID |
| `user_id` | `string` | Authenticated user identifier |
| `service` | `string` | Service name (set once at logger creation) |
| `env` | `string` | Deployment environment (`production`, `staging`, …) |
| `error` | `string` | Error message (use `.Err(err)`) |
| `duration_ms` | `int64` | Elapsed time for an operation |

#### HTTP request logging patterns

Use consistent structured fields when logging HTTP traffic so that log queries and dashboards work uniformly across services.

The `marker` field is a fixed string tag that identifies the log category (`[api]` for inbound, `[request]` for outbound). It enables fast log filtering (e.g. `marker:[api]` in DataDog) without full-text search.

**Incoming API requests** (server-side middleware)

Log at `Info` level after the handler returns. Read the request ID from the `X-Request-ID` header.

```go
requestID := request.Header.Get("X-Request-ID")
zerolog.Ctx(ctx).Info().
    Str("marker", "[api]").
    Str("method", request.Method).
    Str("upstream_host", request.URL.Host).
    Str("path", request.URL.Path).
    Int("status_code", statusCode).
    Int64("duration_ns", runTime.Nanoseconds()).
    Dur("duration", runTime).
    Str("request_id", requestID).
    Msg("api request")
```

**Outbound requests** (HTTP client middleware / transport wrapper)

Log at `Info` level after the response is received. Read the request ID from the context when available.

```go
// requestIDKey is the context key used to store and retrieve the request ID.
// Define it once per package: type contextKey string; const requestIDKey contextKey = "request_id"
requestID, _ := ctx.Value(requestIDKey).(string)
zerolog.Ctx(ctx).Info().
    Str("marker", "[request]").
    Str("method", request.Method).
    Int("status_code", res.StatusCode).
    Str("upstream_host", request.URL.Host).
    Str("path", request.URL.Path).
    Int64("duration_ns", runTime.Nanoseconds()).
    Dur("duration", runTime).
    Str("request_id", requestID).
    Msg("outbound request")
```

#### Logger propagation via context

Attach the logger to `context.Context` at entry points so downstream functions receive it automatically.

```go
// At the handler boundary
ctx = log.With().Str("request_id", requestID).Logger().WithContext(ctx)

// Inside domain / service code
zerolog.Ctx(ctx).Info().Str("order_id", orderID).Msg("processing order")
```

Never pass a `*zerolog.Logger` as a struct field in domain types — always use `zerolog.Ctx(ctx)`.

#### Avoid common mistakes

- Do not call `.Msg("")` with an empty string — use `.Send()` only when the structured fields alone fully describe the event; always prefer a concise, human-readable message otherwise.
- Do not construct log messages with `fmt.Sprintf` — use zerolog's typed field methods.
- Do not log and return an error at the same call site — choose one (see Error handling rules above).
- Do not enable `Debug` or `Trace` in production without a feature-flag-controlled log level.

---

### DataDog Integration

#### Log correlation

To correlate logs with APM traces in DataDog, inject the active trace and span IDs into every log entry. Use the [dd-trace-go](https://github.com/DataDog/dd-trace-go) library.

```go
// v1: gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer
// v2: github.com/DataDog/dd-trace-go/v2/ddtrace/tracer
import (
    "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// ddContextHook enriches a zerolog event with the active DataDog trace context.
func ddContextHook(ctx context.Context, e *zerolog.Event) *zerolog.Event {
    span, ok := tracer.SpanFromContext(ctx)
    if !ok {
        return e
    }
    return e.
        Uint64("dd.trace_id", span.Context().TraceID()).
        Uint64("dd.span_id", span.Context().SpanID())
}
```

Apply the hook when building the logger so all log entries emitted within a traced context carry the correlation IDs automatically.

#### Required tags / attributes

Every log entry must carry the following DataDog reserved attributes so that log management works out of the box:

| zerolog field | DataDog attribute | Notes |
|---|---|---|
| `service` | `service` | Set once at logger creation |
| `env` | `env` | Set once at logger creation |
| `dd.trace_id` | `dd.trace_id` | Injected per-request via hook |
| `dd.span_id` | `dd.span_id` | Injected per-request via hook |

#### APM spans

- Create a child span for every significant unit of work (outbound HTTP call, DB query, background job step).
- Always `defer span.Finish()` immediately after starting a span.
- Set `span.SetTag(ext.Error, err)` on error paths.
- Use [contrib packages](https://pkg.go.dev/gopkg.in/DataDog/dd-trace-go.v1/contrib) for automatic instrumentation of standard libraries (net/http, database/sql, gRPC, Redis, …).

```go
span, ctx := tracer.StartSpanFromContext(ctx, "order.create")
defer span.Finish()

if err := svc.CreateOrder(ctx, order); err != nil {
    span.SetTag(ext.Error, err)
    return fmt.Errorf("create order: %w", err)
}
```

---

### OpenTelemetry Tracing

Use the [OpenTelemetry Go SDK](https://opentelemetry.io/docs/languages/go/) for vendor-neutral distributed tracing. Configure the DataDog exporter (via OTLP) or the DataDog Agent as the collector backend.

#### Tracer initialisation

Initialise a single `trace.TracerProvider` in the composition root and set it as the global provider. Shut it down gracefully on service exit.

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func newTracerProvider(ctx context.Context, serviceName, env string) (*sdktrace.TracerProvider, error) {
    exp, err := otlptracegrpc.New(ctx)
    if err != nil {
        return nil, fmt.Errorf("otlp exporter: %w", err)
    }

    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceName(serviceName),
            semconv.DeploymentEnvironment(env),
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("otel resource: %w", err)
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exp),
        sdktrace.WithResource(res),
    )
    otel.SetTracerProvider(tp)
    return tp, nil
}
```

#### Creating spans

Obtain a tracer from `otel.Tracer("<package-name>")` once per package. Create child spans for every meaningful operation.

```go
var tracer = otel.Tracer("github.com/myorg/myservice/order")

func (s *Service) CreateOrder(ctx context.Context, o Order) error {
    ctx, span := tracer.Start(ctx, "Service.CreateOrder")
    defer span.End()

    span.SetAttributes(
        attribute.String("order.id", o.ID),
        attribute.Int("order.items", len(o.Items)),
    )

    if err := s.repo.Save(ctx, o); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return fmt.Errorf("save order: %w", err)
    }
    return nil
}
```

#### Log-trace correlation with OpenTelemetry

Inject the active OTel span into zerolog entries so logs and traces are correlated in DataDog.

```go
import (
    "go.opentelemetry.io/otel/trace"
)

func otelContextHook(ctx context.Context, e *zerolog.Event) *zerolog.Event {
    sc := trace.SpanFromContext(ctx).SpanContext()
    if !sc.IsValid() {
        return e
    }
    return e.
        Str("trace_id", sc.TraceID().String()).
        Str("span_id", sc.SpanID().String())
}
```

#### Propagation

- Always use `otel.GetTextMapPropagator()` to inject and extract trace context across HTTP and messaging boundaries.
- Use the `W3C TraceContext` + `Baggage` propagators (set as defaults in the composition root).

```go
otel.SetTextMapPropagator(
    propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ),
)
```

#### Naming conventions

- Span names use `TypeName.MethodName` for service methods (e.g. `OrderService.Create`).
- Span names use `http.method http.route` for HTTP server spans (handled automatically by OTel contrib middleware).
- Attribute keys follow [OpenTelemetry Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/).

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

## PostgreSQL & GORM Best Practices (Go)

> Follow these guidelines for all code that interacts with PostgreSQL via [GORM](https://gorm.io/).

### Model Definition

- Embed `gorm.Model` only when you need the built-in `ID`, `CreatedAt`, `UpdatedAt`, and `DeletedAt` fields. Define these fields explicitly when you need different types or naming:
  ```go
  type User struct {
    ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
    Name      string         `gorm:"not null"`
    Email     string         `gorm:"uniqueIndex;not null"`
  }
  ```
- Always annotate model fields with `gorm` struct tags. Declare constraints (`not null`, `uniqueIndex`, `default`) in tags so GORM-generated DDL matches the database schema.
- Add JSON tags alongside GORM tags when the model is also used in API responses; otherwise keep persistence models and transport models separate.

### Primary Key Strategy

Choose the primary key type deliberately — both options have real trade-offs:

| Consideration | Auto-increment (`BIGSERIAL`) | UUID (`uuid`) |
|---|---|---|
| **Storage** | 8 bytes | 16 bytes |
| **Insert performance** | Sequential — minimal index fragmentation | Random v4 — can fragment B-tree indexes; use UUIDv7 or `gen_random_uuid()` (v4 is fine for most loads) |
| **Distributed / merge-safe** | ❌ Conflicts when merging shards or replicas | ✅ Globally unique without coordination |
| **Predictability** | Sequential IDs are enumerable — avoid exposing them in URLs | Opaque — safe to expose in public APIs |
| **Readability** | Easier to use in `WHERE id = 42` during debugging | Harder to read in logs |

Guidelines:
- Use **`BIGSERIAL` / auto-increment** for internal tables that will never be distributed, merged, or exposed in public APIs (e.g. audit log entries, internal job queues).
- Use **UUID** (v4 or v7) for any entity that may be referenced across service boundaries, shared in URLs, or inserted from multiple sources simultaneously.
- If using UUID, prefer **UUIDv7** (time-ordered) when available to reduce B-tree index fragmentation.
- Never use `int32` (`SERIAL`) for tables expected to grow beyond ~2 billion rows; use `BIGSERIAL` or UUID instead.
  ```go
  // Auto-increment example
  type AuditLog struct {
    ID        int64     `gorm:"primaryKey;autoIncrement"`
    CreatedAt time.Time
    Message   string    `gorm:"not null"`
  }

  // UUID example
  type User struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
    CreatedAt time.Time
    Name      string    `gorm:"not null"`
  }
  ```

### Naming Conventions

Consistent naming makes queries, migrations, and code reviews easier to follow.

| Object | Convention | Example |
|---|---|---|
| Table | `snake_case`, plural | `users`, `order_items` |
| Column | `snake_case` | `created_at`, `first_name` |
| Primary key | `id` | `id` |
| Foreign key | `<referenced_table_singular>_id` | `user_id`, `order_id` |
| Index | `idx_<table>_<columns>` | `idx_users_email`, `idx_orders_user_id_status` |
| Unique constraint | `uq_<table>_<columns>` | `uq_users_email` |
| Check constraint | `ck_<table>_<condition>` | `ck_orders_positive_amount` |
| Enum type | `<table>_<column>_enum` | `orders_status_enum` |
| Migration file | `<seq>_<verb>_<object>.{up,down}.sql` | `001_create_users.up.sql` |

Additional rules:
- Never use reserved SQL keywords as identifiers (e.g. `user`, `order`, `group`).
- Avoid abbreviations unless they are universally understood (e.g. `id`, `url`, `ip`).
- Use consistent tense for boolean columns: prefix with `is_`, `has_`, or `can_` (e.g. `is_active`, `has_verified_email`).
- Timestamp columns that record when something happened use past tense: `created_at`, `deleted_at`, `verified_at`.

### Connection & Pool Configuration

- Never hard-code DSN strings; read them from environment variables or a secrets manager at the single composition root.
- Configure the connection pool explicitly after opening a connection:
  ```go
  sqlDB, err := db.DB()
  if err != nil {
    return fmt.Errorf("get underlying sql.DB: %w", err)
  }
  sqlDB.SetMaxOpenConns(25)
  sqlDB.SetMaxIdleConns(5)
  sqlDB.SetConnMaxLifetime(5 * time.Minute)
  ```
- Always pass a `context.Context` to GORM operations using `db.WithContext(ctx)` so that queries respect request deadlines and cancellation signals.

### Repository Pattern

- Wrap all GORM access behind a repository interface defined in the domain package. Domain services must depend on the interface, not on `*gorm.DB` directly:
  ```go
  // domain/user/repository.go
  type Repository interface {
    FindByID(ctx context.Context, id uuid.UUID) (*User, error)
    Save(ctx context.Context, u *User) error
    Delete(ctx context.Context, id uuid.UUID) error
  }

  // infrastructure/postgres/user_repository.go
  type userRepository struct {
    db *gorm.DB
  }

  var _ user.Repository = (*userRepository)(nil) // compile-time check

  func NewUserRepository(db *gorm.DB) user.Repository {
    return &userRepository{db: db}
  }
  ```
- Keep `*gorm.DB` confined to the infrastructure layer. Never let it leak into domain or handler packages.

### SQL Injection Prevention

SQL injection is the most critical database security risk. Defend against it at every layer:

- **Always** use GORM's parameterised methods (`Where`, `Find`, `First`, `Create`, `Save`, `Delete`). These bind values as parameters, never as raw SQL text:
  ```go
  // Bad — vulnerable to injection
  db.Raw("SELECT * FROM users WHERE email = '" + email + "'")
  db.Where("role = " + role)

  // Good — parameterised
  db.WithContext(ctx).Where("email = ?", email).First(&user)
  db.WithContext(ctx).Where("role = ?", role).Find(&users)
  ```
- When `db.Raw` or `db.Exec` is unavoidable (DDL, complex CTEs), always use positional placeholders:
  ```go
  db.WithContext(ctx).Raw("SELECT * FROM users WHERE status = ? AND created_at > ?", status, since).Scan(&result)
  ```
- **Never** build `ORDER BY`, `LIMIT`, table names, or column names by concatenating user input — these cannot be parameterised. Validate both the column name and sort direction against allow-lists in application code:
  ```go
  allowedColumns := map[string]struct{}{"name": {}, "created_at": {}, "email": {}}
  if _, ok := allowedColumns[sortColumn]; !ok {
    return nil, ErrInvalidSortColumn
  }
  allowedDirections := map[string]struct{}{"ASC": {}, "DESC": {}}
  if _, ok := allowedDirections[strings.ToUpper(direction)]; !ok {
    return nil, ErrInvalidSortDirection
  }
  db.WithContext(ctx).Order(sortColumn + " " + direction).Find(&users)
  ```
- Avoid accepting raw filter expressions from external callers. Translate request parameters into typed, validated query predicates inside the repository layer.
- Enable the `pgaudit` extension in PostgreSQL (staging/prod) to log all SQL statements for security auditing.
- Select only the columns you need with `.Select(...)` to avoid over-fetching, especially for large tables:
  ```go
  db.WithContext(ctx).Select("id", "email", "created_at").Find(&users)
  ```

### Avoiding N+1 Queries

- Use `Preload` for loading associations when you need all records and their relations in one logical operation:
  ```go
  db.WithContext(ctx).Preload("Orders").Find(&users)
  ```
- Use `Joins` when you need to filter on the association or want a single SQL join:
  ```go
  db.WithContext(ctx).Joins("JOIN orders ON orders.user_id = users.id").Find(&users)
  ```
- Audit any loop that performs a database query inside it — refactor to a batch fetch instead.

### Query Optimization

- **Index coverage**: ensure every column that appears in `WHERE`, `ORDER BY`, or `JOIN ON` clauses has an index. Verify with `EXPLAIN ANALYZE`:
  ```sql
  EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM orders WHERE user_id = $1 AND status = $2;
  ```
  Look for `Index Scan` or `Index Only Scan`; a `Seq Scan` on a large table is a red flag.
- **Pagination**: never use `OFFSET` for deep pagination on large tables — use keyset (cursor) pagination instead:
  ```go
  // Bad — OFFSET degrades as the page number grows
  db.WithContext(ctx).Offset(page * size).Limit(size).Find(&users)

  // Good — keyset pagination
  db.WithContext(ctx).Where("created_at < ? OR (created_at = ? AND id < ?)", cursor.CreatedAt, cursor.CreatedAt, cursor.ID).
    Order("created_at DESC, id DESC").Limit(size).Find(&users)
  ```
- **Batch large writes**: use `db.CreateInBatches` for bulk inserts to reduce round-trips:
  ```go
  db.WithContext(ctx).CreateInBatches(records, 500)
  ```
- **Read replicas**: route heavy read queries to a read replica by providing a secondary `*gorm.DB` instance configured with the replica DSN. Keep write operations on the primary.
- **Avoid function calls on indexed columns in WHERE**: wrapping a column in a function prevents index use:
  ```sql
  -- Bad
  WHERE LOWER(email) = 'test@example.com'

  -- Good: store the value normalised, or use a functional index
  WHERE email = 'test@example.com'
  ```
- **Vacuum and analyze**: configure `autovacuum` appropriately and run `ANALYZE` after bulk loads so the query planner has up-to-date statistics.

### Transactions

- Use `db.Transaction` for multi-step writes that must succeed or fail together:
  ```go
  err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&order).Error; err != nil {
      return err // triggers rollback
    }
    if err := tx.Save(&inventory).Error; err != nil {
      return err
    }
    return nil // triggers commit
  })
  ```
- Propagate the transaction `*gorm.DB` to repository methods that must participate in the same transaction — accept `*gorm.DB` as an optional argument or use a transaction-aware context key.
- Never start a transaction and forget to commit or roll it back. Prefer the `db.Transaction` callback form over manual `Begin`/`Commit`/`Rollback` to avoid leaks.

### Error Handling

- Always check `result.Error` after GORM operations; never assume a query succeeded silently.
- Distinguish "not found" from other errors using `errors.Is`:
  ```go
  result := db.WithContext(ctx).First(&user, id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, ErrNotFound
  }
  if result.Error != nil {
    return nil, fmt.Errorf("find user %v: %w", id, result.Error)
  }
  ```
- Map GORM/PostgreSQL errors to domain errors at the repository boundary. Domain services must never import `gorm` or `lib/pq` error types.
- Check for duplicate-key violations (PostgreSQL error code `23505`) when creating records where uniqueness is enforced:
  ```go
  // pgconn is from github.com/jackc/pgx/v5/pgconn
  var pgErr *pgconn.PgError
  if errors.As(err, &pgErr) && pgErr.Code == "23505" {
    return nil, ErrAlreadyExists
  }
  ```

### Data Types & Storage Optimization

Choosing the right column type reduces storage, improves index efficiency, and prevents silent data truncation.

**Text**
- Prefer `VARCHAR(n)` over `TEXT` when the maximum length is known and bounded — it communicates intent and allows the database to enforce the constraint:
  ```sql
  -- Good
  email VARCHAR(254) NOT NULL
  -- Acceptable when length is truly unbounded
  description TEXT
  ```
- Avoid `TEXT` or `BYTEA`/`BLOB` for data that will be queried or indexed frequently; store large objects in object storage (e.g. S3) and keep only a reference URL in the database.
- `CHAR(n)` is almost never the right choice in PostgreSQL — use `VARCHAR(n)` or `TEXT` instead.

**Integers**
- Choose the smallest integer type that safely covers the expected range:
  | Type | Range | Bytes | Use when |
  |---|---|---|---|
  | `SMALLINT` | −32 768 … 32 767 | 2 | Status codes, small enumerations |
  | `INTEGER` | −2.1B … 2.1B | 4 | General counters, IDs expected to stay < 2B |
  | `BIGINT` | −9.2 × 10¹⁸ … 9.2 × 10¹⁸ | 8 | Row counts, timestamps-as-int, high-volume IDs |
- Use `BIGSERIAL` (not `SERIAL`) for auto-increment primary keys on tables that may grow beyond ~2 billion rows (see the [Primary Key Strategy](#primary-key-strategy) section for the full trade-off discussion).
- Do **not** use `NUMERIC` or `FLOAT` for monetary values — use `INTEGER` (store cents) or `NUMERIC(19,4)` (exact decimal).

**Boolean & Enumerations**
- Use the native `BOOLEAN` type for true/false columns; avoid `SMALLINT` or `CHAR(1)` as boolean stand-ins.
- Use a PostgreSQL `ENUM` type (or a `VARCHAR` + check constraint) for columns with a small, stable set of values. Prefer `VARCHAR` + check constraint when the set may evolve, since altering a `pg_enum` type is DDL-locked:
  ```sql
  status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'active', 'cancelled'))
  ```

**Dates & Times**
- Use `TIMESTAMPTZ` (timestamp with time zone) for all timestamp columns — it stores UTC and converts on read. Avoid `TIMESTAMP` (without time zone), which discards offset information.
- Use `DATE` only when the time component is genuinely irrelevant (e.g. a birthday).

**JSON**
- Use `JSONB` (not `JSON`) for storing JSON documents — `JSONB` is stored in a decomposed binary form that supports indexing and efficient key access. Use `JSON` only when you need to preserve key order or exact whitespace.
- Index frequently-queried JSONB keys with a GIN index:
  ```sql
  CREATE INDEX idx_events_metadata ON events USING GIN (metadata);
  ```

**In GORM**
```go
type Product struct {
  ID          int64          `gorm:"primaryKey;autoIncrement"`
  Name        string         `gorm:"type:varchar(255);not null"`
  Description string         `gorm:"type:text"`
  Price       int64          `gorm:"not null"` // store in cents
  Stock       int32          `gorm:"not null;default:0"`
  IsActive    bool           `gorm:"not null;default:true"`
  Status      string         `gorm:"type:varchar(20);not null"` // enforced via check constraint in migration
  Metadata    datatypes.JSON `gorm:"type:jsonb"`
  CreatedAt   time.Time
}
```

### Migrations

- **Never** use `db.AutoMigrate` in production. Use a dedicated migration tool (e.g. [golang-migrate](https://github.com/golang-migrate/migrate) or [goose](https://github.com/pressly/goose)) with versioned, sequential SQL files.
- Keep migration files in a `migrations/` directory at the repository root. Name them with a timestamp or sequential integer prefix: `001_create_users.up.sql` / `001_create_users.down.sql`.
- Every migration must have a corresponding rollback (`down`) script.
- Migrations must be backwards-compatible when deployed with zero downtime: add columns as nullable first, backfill, then add constraints in a subsequent migration.
- Run migrations in CI against a real PostgreSQL instance to catch errors before merging.
- `AutoMigrate` is acceptable in local development or test environments only; document this clearly with a build tag or environment check.
- Run golang-migrate via its Go API or CLI in your startup sequence:
  ```go
  import (
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
  )

  func runMigrations(dsn string) error {
    m, err := migrate.New("file://migrations", dsn)
    if err != nil {
      return fmt.Errorf("create migrator: %w", err)
    }
    if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
      return fmt.Errorf("run migrations: %w", err)
    }
    return nil
  }
  ```

### Soft Deletes

- Use `gorm.DeletedAt` (which sets `DeletedAt` to the current time) for soft-deleting records. GORM automatically adds `WHERE deleted_at IS NULL` to all queries on models with this field.
- Add a database index on the `deleted_at` column for tables with high query volume.
- When you genuinely need to query deleted records, use `db.Unscoped()`:
  ```go
  db.WithContext(ctx).Unscoped().Where("id = ?", id).First(&user)
  ```
- Do not mix soft-delete and hard-delete patterns on the same table.

### Indexes & Constraints

- Declare indexes and constraints in GORM struct tags so that schema tooling and migration generators can reflect them:
  ```go
  type Order struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
    UserID    uuid.UUID `gorm:"not null;index"`
    Status    string    `gorm:"not null;index:idx_orders_status_created,priority:1"`
    CreatedAt time.Time `gorm:"index:idx_orders_status_created,priority:2"`
  }
  ```
- Always add an index on foreign-key columns.
- Prefer partial indexes in raw SQL migrations for sparse or enum-like columns to reduce index size.
- Review the `EXPLAIN ANALYZE` output for any query that scans large tables; ensure the query plan uses an index seek, not a sequential scan.

### Data Growth & Capacity Planning

Plan for growth before it becomes a crisis.

- **Estimate row size**: calculate the approximate bytes per row and multiply by the expected row count at 1×, 10×, and 100× current load. Factor in index overhead (typically 30–50% on top of table size):
  ```sql
  -- Inspect current table and index sizes
  SELECT
    relname                                          AS table,
    pg_size_pretty(pg_total_relation_size(oid))     AS total,
    pg_size_pretty(pg_relation_size(oid))            AS table_only,
    pg_size_pretty(pg_indexes_size(oid))             AS indexes
  FROM pg_class
  WHERE relkind = 'r'
  ORDER BY pg_total_relation_size(oid) DESC;
  ```
- **Monitor growth rate**: record table sizes regularly (e.g. nightly) and alert when a table exceeds a configurable threshold or when the 30-day growth rate projects a disk-full event within 90 days.
- **Archive or purge cold data**: define a retention policy for every table. Move records older than the retention window to a separate archive table or object storage. Use scheduled jobs (pg_cron, an external worker) rather than ad-hoc manual deletes.
- **Integer overflow**: with `SERIAL` (int32), the sequence wraps at ~2.1 billion. Audit high-write tables and migrate to `BIGSERIAL` or UUID before the sequence is exhausted — do this as a zero-downtime migration, not an emergency.
- **Autovacuum tuning**: heavily updated or deleted tables accumulate dead tuples. Tune `autovacuum_vacuum_scale_factor` and `autovacuum_analyze_scale_factor` per table when the default 20 % threshold is too coarse for large tables:
  ```sql
  ALTER TABLE events SET (
    autovacuum_vacuum_scale_factor  = 0.01,
    autovacuum_analyze_scale_factor = 0.01
  );
  ```

### Partitioning & Sharding

Use partitioning and sharding when a single table or a single server can no longer sustain the required throughput or capacity.

**Table Partitioning (single server)**

PostgreSQL native declarative partitioning is the first tool to reach for. It keeps data in one logical table while physically splitting it across partitions, enabling partition pruning and faster scans:

- **Range partitioning** — ideal for time-series data (events, logs, audit trails):
  ```sql
  CREATE TABLE events (
    id         BIGSERIAL,
    created_at TIMESTAMPTZ NOT NULL,
    payload    JSONB
  ) PARTITION BY RANGE (created_at);

  CREATE TABLE events_2025_q1 PARTITION OF events
    FOR VALUES FROM ('2025-01-01') TO ('2025-04-01');

  CREATE TABLE events_2025_q2 PARTITION OF events
    FOR VALUES FROM ('2025-04-01') TO ('2025-07-01');
  ```
- **List partitioning** — useful when rows are naturally grouped by a discrete key (e.g. `tenant_id`, `region`):
  ```sql
  CREATE TABLE orders (
    id        BIGSERIAL,
    region    TEXT NOT NULL,
    amount    BIGINT NOT NULL
  ) PARTITION BY LIST (region);

  CREATE TABLE orders_us PARTITION OF orders FOR VALUES IN ('us-east', 'us-west');
  CREATE TABLE orders_eu PARTITION OF orders FOR VALUES IN ('eu-west', 'eu-central');
  ```
- Always create indexes on the partition key column and on any column used in `WHERE` filters across partitions.
- Automate partition creation in advance (e.g. create next month's partition at the start of the current month) using a scheduled job or migration.
- Drop old partitions instead of `DELETE`-ing rows — dropping a partition is an O(1) metadata operation, while `DELETE` triggers autovacuum work.

**Application-level Sharding (multiple servers)**

Sharding distributes data across multiple independent PostgreSQL instances. Introduce it only when a single instance can no longer handle the write load or storage requirements.

- Choose a shard key that distributes writes evenly and avoids cross-shard joins (e.g. `tenant_id`, `user_id`).
- Keep all data for a single shard key on the same node to allow most queries to be executed on a single shard.
- Route queries in the application layer: maintain a shard map (shard key → DSN) and resolve the correct `*gorm.DB` instance before issuing any query:
  ```go
  func (r *shardedRepo) FindOrder(ctx context.Context, tenantID, orderID uuid.UUID) (*Order, error) {
    db := r.shardMap.DBFor(tenantID)
    var order Order
    if err := db.WithContext(ctx).Where("id = ?", orderID).First(&order).Error; err != nil {
      return nil, fmt.Errorf("find order: %w", err)
    }
    return &order, nil
  }
  ```
- Avoid cross-shard transactions; redesign data models to keep related entities on the same shard.
- Consider [Citus](https://github.com/citusdata/citus) (PostgreSQL extension) for transparent sharding before implementing a custom shard router.
## Kafka Best Practices (Go)

> These guidelines apply to all Go services that produce or consume Kafka messages.
> Follow them alongside the Go Style Guidelines above.

### Topic Naming Conventions

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

### Message Keys & Partitioning

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

### Consumer

- **Consumer Groups**: assign a unique `group.id` per logical consumer (one consumer group per independent processing pipeline). Never share a group ID between unrelated services — this creates split-brain processing where each service receives only a fraction of messages.
- Commit offsets **only after** processing succeeds to guarantee at-least-once delivery; never auto-commit.
- Handle rebalances explicitly: flush or checkpoint in-flight work inside `ConsumerGroupHandler.Cleanup` before returning, so no messages are lost mid-batch.
- Set `max.poll.interval.ms` long enough to cover the worst-case processing time; breach causes the consumer to leave the group and trigger rebalance.
- **Batch Fetching**: tune `fetch.min.bytes` and `fetch.max.wait.ms` to control the trade-off between latency and throughput. Higher `fetch.min.bytes` reduces the number of fetch requests but increases end-to-end latency; set `fetch.max.wait.ms` as an upper bound:
  ```go
  kgo.FetchMinBytes(1<<10),        // 1 KiB minimum before returning
  kgo.FetchMaxWait(500*time.Millisecond),
  ```
- **Multi-threaded Consumers**: process messages in parallel using a fixed goroutine pool fed by a bounded channel. Keep the poller single-threaded (Kafka partition assignment is single-threaded) and fan out to workers:
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
- **Context Usage**: pass a `context.Context` derived from the application lifecycle into every poll and process call. Use `context.WithTimeout` for individual message processing to enforce per-message deadlines and prevent slow consumers from blocking the entire group:
  ```go
  processCtx, cancel := context.WithTimeout(ctx, cfg.ProcessingTimeout)
  defer cancel()
  if err := handler.Handle(processCtx, rec); err != nil { ... }
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
## GoCraft/Work Best Practices (Go)

> These guidelines apply to all background job processing that uses [GoCraft/Work](https://github.com/gocraft/work) as the job queue library.

### Worker Pool Lifecycle

- Create the worker pool in the composition root and pass it as a dependency — never create a pool inside a domain package.
- Always call `pool.Stop()` on application shutdown to drain in-flight jobs before the process exits:
  ```go
  pool := work.NewWorkerPool(AppContext{}, concurrency, namespace, redisPool)
  // ... register jobs ...
  pool.Start()

  quit := make(chan os.Signal, 1)
  signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
  <-quit

  pool.Stop() // blocks until all in-flight jobs finish
  ```
- Use a single `work.Pool` (Redis connection pool) per process; share it between the enqueuer and the worker pool.

### Job Registration

- Register all job handlers before calling `pool.Start()`.
- Group job registrations in a single, dedicated function (e.g. `RegisterJobs(pool *work.WorkerPool)`) inside the wiring layer — not scattered across domain packages.
- Use `JobOptions` to specify concurrency, max retries, and backoff at registration time rather than hard-coding them inline:
  ```go
  pool.JobWithOptions("send_email", work.JobOptions{
      MaxFails:    5,
      Concurrency: 10,
  }, handler.SendEmail)
  ```

### Handler Signatures and Argument Passing

- All job handlers must match the signature `func(job *work.Job) error`.
- Deserialise job arguments into a strongly typed struct at the top of the handler; return an error immediately if deserialisation fails:
  ```go
  type SendEmailArgs struct {
      UserID  int64  `json:"user_id"`
      Subject string `json:"subject"`
  }

  func (h *EmailHandler) SendEmail(job *work.Job) error {
      var args SendEmailArgs
      if err := job.UnmarshalPayload(&args); err != nil {
          return fmt.Errorf("unmarshal payload: %w", err)
      }
      // ...
  }
  ```
- Never read raw `job.Args` map entries directly in business logic — centralise argument extraction in a helper or unmarshal into a typed struct.

### Enqueuing Jobs

- Wrap `work.Enqueuer` behind an interface so callers depend on the interface rather than the concrete type:
  ```go
  type JobEnqueuer interface {
      Enqueue(jobName string, args work.Q) (*work.Job, error)
      EnqueueIn(jobName string, secondsFromNow int64, args work.Q) (*work.ScheduledJob, error)
  }
  ```
- Use `work.Q` (a `map[string]interface{}`) only at the enqueuing boundary; convert to typed structs as soon as the job is consumed.
- Prefer `EnqueueUnique` when duplicate jobs must be suppressed (e.g. processing the same resource multiple times within a short window).

### Error Handling and Retries

- Return a non-nil error from a handler to signal that the job should be retried according to its `MaxFails` / backoff policy.
- Return `nil` to mark the job as successfully completed — even if the operation was a no-op.
- Use sentinel errors (exported `var Err...`) when callers or middleware need to distinguish failure modes:
  ```go
  var ErrSkipRetry = errors.New("skip retry")
  ```
- Do not `panic` inside a handler; rely on the middleware panic-recovery layer instead.

### Middleware

- Register middleware on the pool (not on individual handlers) for cross-cutting concerns: structured logging, panic recovery, metrics, and distributed tracing.
- Keep each middleware focused on a single concern. Chain middleware in this order:
  1. Panic recovery (outermost — must always run)
  2. Distributed trace / request-ID injection
  3. Structured logging (log job name, duration, and outcome)
  4. Metrics / instrumentation
- Access the job's name and arguments via `job.Name` and `job.Args` inside middleware — do not re-parse the payload.

### Concurrency and Throttling

- Set per-job `Concurrency` at registration time based on downstream capacity (DB connections, external API rate limits).
- The pool-level `concurrency` argument is the global cap; per-job concurrency is an additional constraint.
- Avoid setting global concurrency higher than the Redis connection pool size to prevent connection starvation.

### Observability

- Log the job name, enqueue time, attempt number (`job.Fails`), and final outcome (success or error) at the middleware layer.
- Emit a counter metric for `job.completed` and `job.failed` tagged with the job name.
- Include a correlation/request ID in job arguments so logs across enqueue and execute can be correlated.

### Testing Job Handlers

- Unit-test handlers by constructing a `*work.Job` directly — no Redis required:
  ```go
  func TestSendEmail(t *testing.T) {
      job := &work.Job{}
      job.SetArg("user_id", int64(42))
      job.SetArg("subject", "Hello")

      h := &EmailHandler{mailer: &fakeMailer{}}
      err := h.SendEmail(job)
      require.NoError(t, err)
  }
  ```
- Use fakes or mocks for all external dependencies (mailers, DB, HTTP clients) injected into the handler struct.
- Test retry behaviour by asserting that the handler returns a non-nil error when a dependency fails.
- Integration tests that require Redis should use a dedicated test Redis instance and clean up all keys with a `t.Cleanup` function.
## Temporal Best Practices (Go)

> This section covers best practices for building reliable, maintainable workflows with [Temporal](https://docs.temporal.io/) in Go.

### Workflow Design — Determinism

Temporal replays workflow history to reconstruct state. Any non-deterministic operation in a workflow function will cause a non-deterministic error during replay.

**Rules:**
- Never call `time.Now()` or `time.Sleep()` directly — use `workflow.Now(ctx)` and `workflow.Sleep(ctx, d)`.
- Never use `math/rand` directly — use `workflow.NewRandom(ctx)` for random values.
- Never read from environment variables, files, or external state inside a workflow function.
- Never spawn raw goroutines (`go func() {...}`) — use `workflow.Go(ctx, func(ctx workflow.Context) {...})`.
- Never use `context.Context` as workflow parameters — use `workflow.Context` instead.
- Never use `select` with real channels — use `workflow.NewSelector(ctx)` instead.
- All side-effecting or non-deterministic logic must be pushed into Activities.

```go
// Bad — not deterministic
func MyWorkflow(ctx workflow.Context) error {
    t := time.Now()         // non-deterministic
    _ = os.Getenv("KEY")    // non-deterministic
    return nil
}

// Good
func MyWorkflow(ctx workflow.Context) error {
    t := workflow.Now(ctx)  // deterministic replay-safe clock
    _ = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
        return os.Getenv("KEY") // side effect captured once
    })
    return nil
}
```

### Activity Design

Activities are the unit of work that performs all I/O and non-deterministic operations. They must be designed to be safe to retry.

- **Idempotency**: every Activity must be safe to run more than once with the same inputs. Use idempotency keys when calling external APIs. For non-idempotent operations (e.g. charging a card), set `MaximumAttempts: 1` and apply exactly-once execution patterns (unique identifiers or deduplication strategies) to prevent duplicate side effects.
- **Heartbeating**: long-running Activities must call `activity.RecordHeartbeat(ctx, details...)` periodically so the worker can detect failures and resume from the last checkpoint.
- **Context propagation**: always accept `context.Context` as the first parameter so that Temporal can cancel the Activity on timeout.
- **Return meaningful errors**: return non-retryable errors with `temporal.NewNonRetryableApplicationError` for permanent failures; return retryable errors (or plain `error`) for transient failures.

```go
func ProcessOrderActivity(ctx context.Context, orderID string) error {
    // Heartbeat for long-running work
    activity.RecordHeartbeat(ctx, "starting")

    if err := externalClient.Process(ctx, orderID); err != nil {
        if isPermanentFailure(err) {
            // Temporal will NOT retry this
            return temporal.NewNonRetryableApplicationError(
                "permanent failure processing order",
                "PermanentProcessError",
                err,
            )
        }
        return err // Temporal will retry
    }
    return nil
}
```

### Worker Registration

- Register all Workflows and Activities in the composition root / `main()` — never inside domain packages.
- Group registrations by task queue; each `worker.Worker` instance must correspond to exactly one task queue.
- Verify that every Workflow and Activity used in production is registered on at least one worker.

```go
w := worker.New(temporalClient, "order-processing", worker.Options{})
w.RegisterWorkflow(OrderWorkflow)
w.RegisterActivity(ProcessOrderActivity)
w.RegisterActivity(NotifyCustomerActivity)
```

### Error Handling

Use Temporal's typed error hierarchy to distinguish between retryable and non-retryable failures:

| Error type | Retried by default | Use when |
|---|---|---|
| `temporal.ApplicationError` (retryable) | Yes | Transient failures; caller does not need to match the type |
| `temporal.ApplicationError` (non-retryable) | No | Permanent / business-rule failures |
| `temporal.CanceledError` | No | Activity/workflow was cancelled |
| `temporal.TimeoutError` | No | Execution exceeded a deadline |

Unwrap Temporal errors with `errors.As` to inspect the type at the calling layer:

```go
var appErr *temporal.ApplicationError
if errors.As(err, &appErr) {
    log.Error("activity failed", zap.String("type", appErr.Type()), zap.Error(appErr))
}
```

### Workflow Versioning

When changing workflow logic that may affect in-flight executions, use `workflow.GetVersion` to maintain backward compatibility. Never remove a `GetVersion` check until all workflows that used the old branch have completed.

```go
const (
    _featureAddNotification = "add-notification"
    _defaultVersion         = workflow.DefaultVersion
    _versionOne             = 1
)

func OrderWorkflow(ctx workflow.Context, order Order) error {
    v := workflow.GetVersion(ctx, _featureAddNotification, _defaultVersion, _versionOne)

    ao := workflow.ActivityOptions{StartToCloseTimeout: 30 * time.Second}
    ctx = workflow.WithActivityOptions(ctx, ao)

    if err := workflow.ExecuteActivity(ctx, ProcessOrderActivity, order.ID).Get(ctx, nil); err != nil {
        return fmt.Errorf("process order: %w", err)
    }

    if v >= _versionOne {
        if err := workflow.ExecuteActivity(ctx, NotifyCustomerActivity, order.ID).Get(ctx, nil); err != nil {
            return fmt.Errorf("notify customer: %w", err)
        }
    }
    return nil
}
```

### Signals, Queries, and Updates

- **Signals** are fire-and-forget messages delivered to a running workflow. Handle them with `workflow.GetSignalChannel(ctx, name).Receive(ctx, &val)` or via a `workflow.NewSelector`.
- **Queries** must be read-only and must not modify workflow state. Register with `workflow.SetQueryHandler`.
- **Updates** (Temporal ≥ 1.21) combine signal delivery and synchronous validation. Use `workflow.SetUpdateHandlerWithOptions` and provide a validator function to reject invalid inputs before they are written to history.
- Name signals, queries, and updates as lowercase kebab-case strings (e.g. `"cancel-order"`, `"get-status"`).

```go
const _signalCancelOrder = "cancel-order"

func OrderWorkflow(ctx workflow.Context, order Order) error {
    cancelCh := workflow.GetSignalChannel(ctx, _signalCancelOrder)

    ao := workflow.ActivityOptions{StartToCloseTimeout: 30 * time.Second}
    ctx = workflow.WithActivityOptions(ctx, ao)

    sel := workflow.NewSelector(ctx)
    var processErr error
    actFuture := workflow.ExecuteActivity(ctx, ProcessOrderActivity, order.ID)
    sel.AddFuture(actFuture, func(f workflow.Future) {
        processErr = f.Get(ctx, nil)
    })
    sel.AddReceive(cancelCh, func(c workflow.ReceiveChannel, _ bool) {
        var reason string
        c.Receive(ctx, &reason)
        processErr = temporal.NewCanceledError(reason)
    })
    sel.Select(ctx)

    return processErr
}
```

### Retries and Timeouts

Always configure timeouts and retry policies explicitly — never rely on Temporal's unlimited defaults in production.

**Timeout options:**

| Option | Purpose | Guideline |
|---|---|---|
| `StartToCloseTimeout` | Max time for a single Activity attempt | Required; must be greater than the upstream service's own request timeout |
| `ScheduleToCloseTimeout` | Max total time including all retries | Must accommodate the worst-case cumulative retry duration; set for Activities with SLA requirements |
| `HeartbeatTimeout` | Max time between heartbeats | Required for long-running Activities |
| `WorkflowExecutionTimeout` | Max lifetime of an entire workflow | Set to prevent unbounded open workflows |
| `WorkflowRunTimeout` | Max lifetime of a single workflow run | Useful for workflows that use `ContinueAsNew` |

**Timeout rules:**
- `StartToCloseTimeout` must be greater than the upstream service's own request timeout to avoid races where Temporal cancels the Activity before the service has a chance to respond.
- `ScheduleToCloseTimeout` must cover the worst-case cumulative retry scenario: all retry attempts plus the wait periods between them. Use the [Temporal retry simulator](https://temporal-time.netlify.app/) or the [Temporal Activity Retry Simulator](https://docs.temporal.io/develop/activity-retry-simulator) to calculate a safe value before setting this in production.

**Exponential backoff strategy:**
- Use a `BackoffCoefficient` between **1.5 and 2.0** to balance retry frequency against resource consumption.
- Cap `MaximumInterval` based on the nature of the upstream:
  - HTTP / lightweight internal APIs: **10–30 seconds**
  - Slow or unstable third-party services: **30–60 seconds**
- For **fast-responding upstreams**: use a short `InitialInterval` (1–2 s) with a coefficient of 2.0.
- For **slow or unstable upstreams**: use a longer `InitialInterval` (2–5 s) and a larger `MaximumInterval` (30–60 s).

**Retry limits:**
- Avoid unlimited retries (`MaximumAttempts: 0`) for critical external calls — this risks resource exhaustion and cascading failures.
- Set a finite `MaximumAttempts` or a `ScheduleToCloseTimeout` with a compensation path for failure scenarios.
- `MaximumAttempts` **includes the initial attempt**:
  - `1` → no retries (execute once and stop).
  - `0` → unlimited retries (use only for non-critical, fully idempotent operations).
- For **single-shot operations** (e.g. charging a payment), set `MaximumAttempts: 1` and ensure strong alerting for immediate failure notification.

```go
// Fast upstream (e.g. internal HTTP service): short intervals, higher coefficient
fastAO := workflow.ActivityOptions{
    StartToCloseTimeout:    10 * time.Second,
    ScheduleToCloseTimeout: 2 * time.Minute,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:        time.Second,
        BackoffCoefficient:     2.0,
        MaximumInterval:        30 * time.Second,
        MaximumAttempts:        5,
        NonRetryableErrorTypes: []string{"PermanentProcessError"},
    },
}

// Slow / unstable upstream (e.g. third-party payment provider): longer intervals, lower coefficient
slowAO := workflow.ActivityOptions{
    StartToCloseTimeout:    60 * time.Second,
    ScheduleToCloseTimeout: 10 * time.Minute,
    HeartbeatTimeout:       15 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:        3 * time.Second,
        BackoffCoefficient:     1.5,
        MaximumInterval:        60 * time.Second,
        MaximumAttempts:        4,
        NonRetryableErrorTypes: []string{"PermanentProcessError"},
    },
}

// Single-shot operation (e.g. charge a payment — must not retry)
singleShotAO := workflow.ActivityOptions{
    StartToCloseTimeout: 30 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        MaximumAttempts: 1,
    },
}
```

### Activity Retry Patterns

Choose a retry pattern based on the nature of the operation and the acceptable failure mode.

| Pattern | When to use | Key config |
|---|---|---|
| **Standard retry** | Transient failures on idempotent calls (HTTP timeouts, rate limits, temporary unavailability) | `MaximumAttempts > 1`, exponential backoff |
| **Single-shot (no retry)** | Non-idempotent operations where a duplicate execution causes harm (e.g. charge a card, send an SMS) | `MaximumAttempts: 1`, strong alerting |
| **Retry with compensation** | Retries are exhausted and a partial side effect must be rolled back (saga pattern) | `MaximumAttempts > 1` + a compensation Activity in the failure path |
| **Heartbeat-based resume** | Long-running operations that can checkpoint progress and resume from the last known state instead of restarting from scratch | `HeartbeatTimeout` set, checkpoint stored in heartbeat details |
| **Manual retry via signal** | Human approval or intervention is required before a retry (e.g. fraud review, operator gate) | `MaximumAttempts: 1`, workflow waits for a signal before re-executing |
| **Polling with delay** | External resource is not yet ready and must be re-checked after a fixed wait (e.g. async job status, eventual-consistency read) | `workflow.Sleep` between attempts inside the workflow loop |
| **Custom retry logic in Activity** | Built-in retry policy is too coarse; different error types need different back-off or fallback strategies | Manual loop with typed error inspection inside the Activity |

> **Tip:** See [temporalio/samples-go](https://github.com/temporalio/samples-go) for runnable examples of these and other patterns in Go.

#### Standard Retry

Use for any idempotent call that may experience transient failures. Temporal retries the Activity automatically using the configured policy.

```go
// Use case: calling an internal inventory service that may be temporarily overloaded.
ao := workflow.ActivityOptions{
    StartToCloseTimeout:    10 * time.Second,
    ScheduleToCloseTimeout: 2 * time.Minute,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:        time.Second,
        BackoffCoefficient:     2.0,
        MaximumInterval:        30 * time.Second,
        MaximumAttempts:        5,
        NonRetryableErrorTypes: []string{"ItemNotFound"},
    },
}
ctx = workflow.WithActivityOptions(ctx, ao)
if err := workflow.ExecuteActivity(ctx, ReserveInventoryActivity, orderID).Get(ctx, nil); err != nil {
    return fmt.Errorf("reserve inventory: %w", err)
}
```

#### Single-Shot (No Retry)

Use for non-idempotent operations. Duplicate execution would cause real harm (double charge, duplicate notification). Pair with strong alerting so failures surface immediately.

```go
// Use case: charging a customer's payment method — must never execute twice.
ao := workflow.ActivityOptions{
    StartToCloseTimeout: 30 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        MaximumAttempts: 1, // no retries
    },
}
ctx = workflow.WithActivityOptions(ctx, ao)
if err := workflow.ExecuteActivity(ctx, ChargePaymentActivity, order).Get(ctx, nil); err != nil {
    // Handle the failure explicitly — do not silently swallow it.
    return fmt.Errorf("charge payment: %w", err)
}
```

#### Retry with Compensation (Saga)

Use when retries are exhausted and partial state must be rolled back. Execute a compensation Activity to undo any side effects committed before the failure.

```go
// Use case: reserve inventory then charge payment; roll back the reservation if payment fails.
reserveCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    StartToCloseTimeout: 10 * time.Second,
    RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 3},
})
if err := workflow.ExecuteActivity(reserveCtx, ReserveInventoryActivity, orderID).Get(ctx, nil); err != nil {
    return fmt.Errorf("reserve inventory: %w", err)
}

chargeCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    StartToCloseTimeout: 30 * time.Second,
    RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 1},
})
if err := workflow.ExecuteActivity(chargeCtx, ChargePaymentActivity, order).Get(ctx, nil); err != nil {
    // Compensate: release the reservation that was already committed.
    compensateCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Second,
        RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 5},
    })
    _ = workflow.ExecuteActivity(compensateCtx, ReleaseInventoryActivity, orderID).Get(ctx, nil)
    return fmt.Errorf("charge payment: %w", err)
}
```

#### Heartbeat-Based Resume

Use for long-running Activities (file processing, batch jobs) that can record checkpoints. On retry, resume from the last checkpoint instead of restarting from scratch.

```go
// Use case: processing a large file page by page; resume from the last committed page on retry.
func ProcessFileActivity(ctx context.Context, fileID string) error {
    // Recover the last checkpoint from a previous attempt, if any.
    startPage := 0
    if details := activity.GetInfo(ctx).HeartbeatDetails; details != nil {
        _ = converter.GetDefaultDataConverter().FromPayloads(details, &startPage)
    }

    for page := startPage; ; page++ {
        if err := ctx.Err(); err != nil {
            return err // respect cancellation / timeout
        }
        done, err := processPage(ctx, fileID, page)
        if err != nil {
            return err
        }
        // Checkpoint progress so the next attempt can resume here.
        activity.RecordHeartbeat(ctx, page+1)
        if done {
            return nil
        }
    }
}

// Activity options: heartbeat timeout shorter than the per-page processing time.
ao := workflow.ActivityOptions{
    StartToCloseTimeout:    30 * time.Minute,
    ScheduleToCloseTimeout: 2 * time.Hour,
    HeartbeatTimeout:       30 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:    5 * time.Second,
        BackoffCoefficient: 2.0,
        MaximumInterval:    60 * time.Second,
        MaximumAttempts:    10,
    },
}
```

#### Manual Retry via Signal

Use when a human or external system must review a failure before the operation is retried. The workflow pauses and waits for an approval signal rather than retrying automatically.

```go
const (
    _signalRetryApproved = "retry-approved"
    _signalRetryAborted  = "retry-aborted"
)

// Use case: a fraud-check Activity failed; a human must approve before retrying.
func OrderWorkflow(ctx workflow.Context, order Order) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Second,
        RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 1},
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    if err := workflow.ExecuteActivity(ctx, FraudCheckActivity, order).Get(ctx, nil); err != nil {
        // Notify an operator and wait for a manual decision.
        _ = workflow.ExecuteActivity(ctx, NotifyOperatorActivity, order.ID, err.Error()).Get(ctx, nil)

        approvedCh := workflow.GetSignalChannel(ctx, _signalRetryApproved)
        abortedCh := workflow.GetSignalChannel(ctx, _signalRetryAborted)

        sel := workflow.NewSelector(ctx)
        var approved bool
        sel.AddReceive(approvedCh, func(c workflow.ReceiveChannel, _ bool) {
            c.Receive(ctx, nil)
            approved = true
        })
        sel.AddReceive(abortedCh, func(c workflow.ReceiveChannel, _ bool) {
            c.Receive(ctx, nil)
        })
        sel.Select(ctx)

        if !approved {
            return temporal.NewNonRetryableApplicationError("fraud check aborted by operator", "FraudCheckAborted", nil)
        }
        // Retry once after approval.
        if err := workflow.ExecuteActivity(ctx, FraudCheckActivity, order).Get(ctx, nil); err != nil {
            return fmt.Errorf("fraud check after approval: %w", err)
        }
    }
    return nil
}
```

#### Polling with Delay

Use when an external resource transitions through states asynchronously and the workflow must wait until a condition is met (e.g. a batch job completes, a record becomes consistent). Poll inside the workflow using `workflow.Sleep` so Temporal's deterministic clock is respected and the wait is durable across worker restarts.

```go
// Use case: wait for an async export job to finish before proceeding.
func ExportWorkflow(ctx workflow.Context, jobID string) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Second,
        RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 3},
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    const (
        _maxPolls    = 20
        _pollInterval = 15 * time.Second
    )
    for i := 0; i < _maxPolls; i++ {
        var done bool
        if err := workflow.ExecuteActivity(ctx, CheckExportStatusActivity, jobID).Get(ctx, &done); err != nil {
            return fmt.Errorf("check export status: %w", err)
        }
        if done {
            return nil
        }
        // Sleep deterministically — never use time.Sleep inside a workflow.
        if err := workflow.Sleep(ctx, _pollInterval); err != nil {
            return err // workflow was cancelled
        }
    }
    return temporal.NewNonRetryableApplicationError("export job timed out", "ExportTimeout", nil)
}
```

#### Custom Retry Logic in Activity

Use when the built-in retry policy applies the same back-off to every error, but different error types need different treatment — for example, throttling errors should back off longer while transient network errors should retry immediately. Implement a manual retry loop inside the Activity and classify errors explicitly.

```go
// Use case: calling a third-party API that returns both throttle errors (need longer back-off)
// and transient errors (can retry quickly).
func CallExternalAPIActivity(ctx context.Context, req Request) (*Response, error) {
    const _maxAttempts = 5
    backoff := time.Second

    for attempt := 1; attempt <= _maxAttempts; attempt++ {
        resp, err := externalClient.Call(ctx, req)
        if err == nil {
            return resp, nil
        }

        var throttleErr *ThrottleError
        var networkErr *NetworkError
        switch {
        case errors.As(err, &throttleErr):
            // Throttle errors require a longer wait; honour the Retry-After hint if available.
            wait := throttleErr.RetryAfter
            if wait == 0 {
                wait = 30 * time.Second
            }
            activity.RecordHeartbeat(ctx, fmt.Sprintf("throttled, waiting %s", wait))
            select {
            case <-ctx.Done():
                return nil, ctx.Err()
            case <-time.After(wait):
            }
        case errors.As(err, &networkErr):
            // Network errors retry quickly with exponential back-off.
            activity.RecordHeartbeat(ctx, fmt.Sprintf("network error attempt %d", attempt))
            select {
            case <-ctx.Done():
                return nil, ctx.Err()
            case <-time.After(backoff):
            }
            backoff *= 2
            if backoff > 30*time.Second {
                backoff = 30 * time.Second
            }
        default:
            // Unknown error — non-retryable.
            return nil, temporal.NewNonRetryableApplicationError(
                "unexpected API error",
                "UnexpectedAPIError",
                err,
            )
        }
    }
    return nil, fmt.Errorf("external API call failed after %d attempts", _maxAttempts)
}
```

> **Note:** Disable Temporal's built-in retry when using a manual loop (`MaximumAttempts: 1`) to avoid double-retrying on the same error.

### ContinueAsNew

Use `workflow.NewContinueAsNewError` to restart a workflow with fresh history when it processes unbounded event streams or long-running loops. This prevents history size from growing indefinitely.

```go
const _maxIterations = 1000

func PollingWorkflow(ctx workflow.Context, state State) error {
    for i := 0; i < _maxIterations; i++ {
        // ... process one iteration ...
    }
    // Hand off to a new run with updated state
    return workflow.NewContinueAsNewError(ctx, PollingWorkflow, state)
}
```

### Testing

Use `testsuite.WorkflowTestSuite` for unit-testing workflows and activities in a deterministic, in-process environment. Do not call a real Temporal server in unit tests.

- Mock Activities with `env.OnActivity(...)` to control return values and assert call counts.
- Use `env.RegisterDelayedCallback` to simulate signals and timers.
- Test Activities independently with plain Go unit tests — they are ordinary functions.

```go
func TestOrderWorkflow_Success(t *testing.T) {
    suite.Run(t, new(OrderWorkflowTestSuite))
}

type OrderWorkflowTestSuite struct {
    suite.Suite
    testsuite.WorkflowTestSuite
}

func (s *OrderWorkflowTestSuite) Test_ProcessOrder() {
    env := s.NewTestWorkflowEnvironment()
    env.OnActivity(ProcessOrderActivity, mock.Anything, "order-123").Return(nil)
    env.OnActivity(NotifyCustomerActivity, mock.Anything, "order-123").Return(nil)

    env.ExecuteWorkflow(OrderWorkflow, Order{ID: "order-123"})

    s.True(env.IsWorkflowCompleted())
    s.NoError(env.GetWorkflowError())
    env.AssertExpectations(s.T())
}
```

### Naming Conventions

| Artifact | Convention | Example |
|---|---|---|
| Workflow function | `MixedCaps` + `Workflow` suffix | `OrderWorkflow` |
| Activity function | `MixedCaps` + `Activity` suffix | `ProcessOrderActivity` |
| Task queue name | kebab-case string constant | `"order-processing"` |
| Signal / query / update name | kebab-case string constant | `"cancel-order"` |
| Workflow ID | Unique, human-readable, business-scoped | `"order-" + orderID` |

### Code Quality Checklist — Temporal

Before requesting a review on Temporal-related changes, verify:

- [ ] Workflow functions contain no non-deterministic code (no `time.Now`, no I/O, no raw goroutines).
- [ ] Every Activity is idempotent and accepts `context.Context` as the first parameter.
- [ ] Long-running Activities call `activity.RecordHeartbeat` with a `HeartbeatTimeout` set.
- [ ] `StartToCloseTimeout` is set on every `ActivityOptions` block and is greater than the upstream service timeout.
- [ ] `ScheduleToCloseTimeout` covers the worst-case cumulative retry duration (validated with the retry simulator).
- [ ] `MaximumAttempts` is explicitly set; unlimited retries (`0`) are avoided for critical external calls.
- [ ] Non-idempotent Activities use `MaximumAttempts: 1` with alerting, or an exactly-once deduplication strategy.
- [ ] Non-retryable errors use `temporal.NewNonRetryableApplicationError`.
- [ ] Workflow logic changes use `workflow.GetVersion` to protect in-flight executions.
- [ ] Workflows that process unbounded events use `workflow.NewContinueAsNewError`.
- [ ] Worker registration happens in the composition root, not inside domain packages.
- [ ] Workflow and Activity unit tests use `testsuite.WorkflowTestSuite` and mock all Activities.
## Redis & Go Best Practices

> Use [github.com/redis/go-redis/v9](https://github.com/redis/go-redis) as the standard client.
> All rules below apply to every Redis interaction in this repository.

### Client Setup & Connection Pooling

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

### Interface / Port Pattern

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

### Key Naming Conventions

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

### Context Propagation

- Pass `context.Context` as the **first argument** to every Redis call. Never use `context.Background()` inside a handler — propagate the request context so timeouts and cancellations are respected:
  ```go
  val, err := rdb.Get(ctx, key).Result()
  ```

### Data Types

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

### Error Handling

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

### TTL / Expiration Discipline

- **Every** key written to Redis must have an explicit TTL. Omitting a TTL risks unbounded memory growth:
  ```go
  rdb.Set(ctx, key, value, 24*time.Hour)
  ```
- Define TTL values as named constants or configuration, not magic numbers.
- Use `EXPIREAT` / `ExpireAt` when the expiry must align with a wall-clock boundary (e.g. end of day).
- Never rely on Redis eviction policies as a substitute for intentional TTLs.

### Cache Invalidation

Choose an invalidation strategy explicitly; do not leave it up to TTL alone:

- **TTL-based expiration** (passive): set an appropriate TTL on every write and let Redis expire the key automatically. Use for read-heavy data that tolerates brief staleness.
- **Active invalidation on write**: delete or update the cache key immediately after the source-of-truth record changes. Keeps the cache consistent but couples write paths to the cache layer.
- **Event-driven invalidation**: consume a change-data-capture or domain event stream to invalidate keys asynchronously. Decouples the write path but introduces eventual consistency.
- **Cache-aside (lazy loading)**: read from the source on a miss and populate the cache. Combine with a short TTL to bound staleness.
- Document the chosen strategy in a comment near the adapter so future readers understand the consistency trade-off.

### Pipelining & Batching

- Use pipelining when issuing **3 or more** independent commands to reduce round-trip overhead:
  ```go
  pipe := rdb.Pipeline()
  pipe.Set(ctx, "key1", "v1", ttl)
  pipe.Set(ctx, "key2", "v2", ttl)
  _, err := pipe.Exec(ctx)
  ```
- Prefer `MGet` / `MSet` over N sequential `Get` / `Set` calls.
- Keep pipeline batches bounded; limit individual batches to a configurable maximum (e.g. 100–500 commands) and flush in chunks to avoid blocking the event loop.

### Transactions (MULTI/EXEC) and Optimistic Locking

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

### Lua Scripting

- Keep Lua scripts short and single-purpose; store them as `const` strings or embed them from `.lua` files via `//go:embed`.
- Load scripts once at startup using `ScriptLoad` and call them by SHA with `EvalSha`; fall back to `Eval` only when the script may not be loaded yet.
- Document the expected `KEYS` and `ARGV` arguments with a comment above each script.

### Pub/Sub

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

### Testing

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

### Observability

- Instrument Redis operations with a hook (`rdb.AddHook(...)`) to emit latency metrics and trace spans — do not add ad-hoc timing code around individual calls.
- Log slow commands (> configured threshold) at `WARN` level with the key name (but never the value if it may contain PII).
- Include the Redis command name and key pattern (not full key) in error context strings.

### Memory Management

- **Always estimate memory before deploying** a new key space. A simple model:
  ```
  total memory ≈ (key_size + value_size + SDS_overhead + per-key_metadata) × key_count
  ```
  Where `SDS_overhead` is ~45 bytes per key (Redis 7.x) for Redis's Simple Dynamic String header plus `dictEntry` and pointer alignment — verify against your specific Redis version using `MEMORY USAGE <key>`. Use `redis-cli --bigkeys` and `MEMORY USAGE <key>` to validate estimates against a real data sample.
- Avoid keys longer than 64 bytes — long key names inflate the key-space memory with no benefit. Prefer compact prefixes and document the abbreviations.
- Use `OBJECT ENCODING <key>` to confirm Redis chose the most compact internal encoding (e.g. `listpack` for small hashes/lists vs `hashtable`/`quicklist`). Tune `hash-max-listpack-entries` and related config thresholds to keep small objects in compact form.
- Set a `maxmemory` limit and an appropriate `maxmemory-policy` (e.g. `allkeys-lru` or `volatile-lru`) in every environment. Monitor the `used_memory_rss` vs `used_memory` ratio; a high ratio indicates memory fragmentation.
- **Close the client** when the application shuts down — call `rdb.Close()` to return pooled connections and prevent goroutine/file-descriptor leaks:
  ```go
  func run(cfg Config) error {
    rdb := newRedisClient(cfg)
    defer rdb.Close()
    // ...
  }
  ```

---

## Dockerfile & Containerization Best Practices (Go)

> Follow these guidelines whenever you dockerize or containerize a Go service.
> The goal is a minimal, secure, reproducible image that starts fast and has a small attack surface.

### Multi-stage Builds

Always use a multi-stage `Dockerfile`. This separates the build environment (which needs the full Go toolchain, source code, and module cache) from the final runtime image (which needs only the compiled binary and its runtime dependencies).

```dockerfile
# syntax=docker/dockerfile:1

# ── Stage 1: build ──────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency manifests first so Docker can cache this layer independently
# of source changes.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source and build a statically linked binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /app/server ./cmd/server

# ── Stage 2: runtime ─────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app/server /server

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
```

Key decisions:
- **`golang:X.Y-alpine` as builder** — smaller than the full `golang` image; `alpine` provides `apk` if you need extra build tools.
- **`CGO_ENABLED=0`** — produces a fully static binary with no libc dependency, which is required for distroless/scratch base images.
- **`-trimpath`** — removes local file-system paths from the binary, improving reproducibility and avoiding information leakage.
- **`-ldflags="-s -w"`** — strips the symbol table and DWARF debug info, reducing binary size by ~30 %.
- **`COPY go.mod go.sum ./` before `COPY . .`** — exploits Docker's layer cache: the `go mod download` layer is only invalidated when dependencies change, not on every source edit.

### Distroless Base Images

[Distroless images](https://github.com/GoogleContainerTools/distroless) contain only the application binary and its minimal runtime dependencies — no shell, package manager, or OS utilities.

| Base image | Use when |
|---|---|
| `gcr.io/distroless/static-debian12` | CGO disabled; pure Go binary |
| `gcr.io/distroless/base-debian12` | CGO enabled; needs glibc |
| `gcr.io/distroless/static-debian12:nonroot` | Same as above but runs as UID 65532 (preferred) |
| `scratch` | Absolute minimal; no CA certificates or timezone data |

Guidelines:
- Prefer `distroless/static-debian12:nonroot` over `scratch` — it ships CA certificates and timezone data, which most services need, and the `:nonroot` tag ensures the container never runs as root.
- Use `distroless/base-debian12` only when your binary requires CGO (e.g. `mattn/go-sqlite3`). Avoid CGO in production services whenever possible.
- Pin the digest, not just the tag, for reproducible builds in production (`FROM gcr.io/distroless/static-debian12@sha256:<digest>`).
- Copy CA certificates explicitly when using `scratch`:
  ```dockerfile
  COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
  ```

### Security Hardening

- **Run as non-root**: always add `USER nonroot:nonroot` (distroless) or create a dedicated user in the final stage:
  ```dockerfile
  # When using a non-distroless base image
  RUN addgroup -S appgroup && adduser -S appuser -G appgroup
  USER appuser
  ```
- **Read-only root filesystem**: set `readOnlyRootFilesystem: true` in the Kubernetes `securityContext` (or pass `--read-only` to `docker run`). Mount explicit `emptyDir` volumes only for paths that genuinely need writes.
- **No new privileges**: set `allowPrivilegeEscalation: false` and `NO_NEW_PRIVILEGES=true`.
- **Minimal capabilities**: drop all Linux capabilities with `--cap-drop=ALL` and add back only what is required.
- **Scan the image**: integrate a vulnerability scanner (e.g. `trivy`, `grype`, or `docker scout`) into CI and fail the build on critical/high CVEs.

### Image Tagging & Versioning

- Never use the `latest` tag in production. Tag images with an immutable identifier:
  ```
  ghcr.io/myorg/myservice:<git-sha>
  ghcr.io/myorg/myservice:<semver>      # e.g. v1.4.2
  ```
- Push both the SHA tag (immutable, for rollback) and the semver tag (human-readable, for deploys).
- Use Docker BuildKit (`DOCKER_BUILDKIT=1`) or `docker buildx` for all builds to get caching, multi-platform support, and inline cache metadata.

### Build Arguments & Secrets

- Pass build-time metadata via `ARG`, not `ENV`, so values do not persist in the final image:
  ```dockerfile
  ARG VERSION=dev
  ARG COMMIT_SHA=unknown
  ARG BUILD_TIME=unknown

  RUN go build -trimpath \
      -ldflags="-s -w \
        -X main.version=${VERSION} \
        -X main.commitSHA=${COMMIT_SHA} \
        -X main.buildTime=${BUILD_TIME}" \
      -o /app/server ./cmd/server
  ```
- Never bake secrets (API keys, certificates, passwords) into the image. Use Docker BuildKit's `--secret` mount for build-time secrets (e.g. private module proxy credentials):
  ```dockerfile
  RUN --mount=type=secret,id=netrc,dst=/root/.netrc \
      go mod download
  ```
- Provide secrets at runtime via environment variables or a secrets manager (Vault, AWS Secrets Manager, GCP Secret Manager) — never via `ENV` instructions in the Dockerfile.

### `.dockerignore`

Always include a `.dockerignore` file to exclude files and directories that must not be sent to the build context. This speeds up builds and prevents leaking sensitive files into the image:

```dockerignore
# Version control
.git
.gitignore

# Local development & IDE
.env
.env.*
*.local
.idea/
.vscode/

# Build artifacts
dist/
bin/
*.exe
*.test

# Documentation
docs/
*.md

# Test files (optional — include if you want them in the image)
**/*_test.go

# CI/CD
.github/
.gitlab-ci.yml
Makefile
```

### Health Checks

Add a `HEALTHCHECK` instruction so the Docker daemon and orchestrators can monitor the container's liveness:

```dockerfile
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/server", "-healthcheck"]
```

Alternatively, expose a `/healthz` HTTP endpoint and use `wget` or `curl` — but avoid installing these tools in distroless images. A self-contained health-check flag in your binary (as above) is the preferred approach for distroless containers.

### Example: Complete Production Dockerfile

```dockerfile
# syntax=docker/dockerfile:1
# Build: docker buildx build --build-arg VERSION=$(git describe --tags) \
#                            --build-arg COMMIT_SHA=$(git rev-parse --short HEAD) \
#                            --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
#                            -t ghcr.io/myorg/myservice:latest .

# ── Stage 1: dependency cache ─────────────────────────────────────────────────
FROM golang:1.24-alpine AS deps
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# ── Stage 2: build ────────────────────────────────────────────────────────────
FROM deps AS builder

ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG BUILD_TIME=unknown

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath \
    -ldflags="-s -w \
      -X main.version=${VERSION} \
      -X main.commitSHA=${COMMIT_SHA} \
      -X main.buildTime=${BUILD_TIME}" \
    -o /app/server ./cmd/server

# ── Stage 3: runtime ──────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app/server /server

EXPOSE 8080
USER nonroot:nonroot
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/server", "-healthcheck"]
ENTRYPOINT ["/server"]
```

### Docker Compose for Local Development

Use Docker Compose to wire the service with its dependencies during local development and CI:

```yaml
# docker-compose.yml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
      target: builder           # Use the builder stage for dev (has shell, debugger)
    environment:
      - DATABASE_URL=postgres://user:pass@db:5432/mydb?sslmode=disable
      - REDIS_ADDR=redis:6379
    ports:
      - "8080:8080"
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - .:/app                  # Hot reload source when using air or similar

  db:
    image: postgres:17-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
      POSTGRES_DB: mydb
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user -d mydb"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
```

- Use `target: builder` in Compose for the development service so you have access to the full toolchain (e.g. `dlv` debugger, `air` hot-reloader).
- Use `target: <final-stage>` (or omit `target`) for the production image build in CI.
- Never commit `.env` files with real credentials; use `.env.example` as a template and load secrets from a secrets manager in production.
