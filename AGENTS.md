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

- **Idempotency**: every Activity must be safe to run more than once with the same inputs. Use idempotency keys when calling external APIs.
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

Always configure timeouts explicitly — never rely on Temporal's unlimited defaults in production.

| Option | Purpose | Guideline |
|---|---|---|
| `StartToCloseTimeout` | Max time for a single Activity attempt | Required; set based on p99 latency |
| `ScheduleToCloseTimeout` | Max total time including all retries | Set for Activities with SLA requirements |
| `HeartbeatTimeout` | Max time between heartbeats | Required for long-running Activities |
| `WorkflowExecutionTimeout` | Max lifetime of an entire workflow | Set to prevent unbounded open workflows |
| `WorkflowRunTimeout` | Max lifetime of a single workflow run | Useful for workflows that use `ContinueAsNew` |

```go
ao := workflow.ActivityOptions{
    StartToCloseTimeout:  30 * time.Second,
    ScheduleToCloseTimeout: 5 * time.Minute,
    HeartbeatTimeout:     10 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:    time.Second,
        BackoffCoefficient: 2.0,
        MaximumInterval:    30 * time.Second,
        MaximumAttempts:    5,
        NonRetryableErrorTypes: []string{"PermanentProcessError"},
    },
}
ctx = workflow.WithActivityOptions(ctx, ao)
```

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
- [ ] `StartToCloseTimeout` is set on every `ActivityOptions` block.
- [ ] Non-retryable errors use `temporal.NewNonRetryableApplicationError`.
- [ ] Workflow logic changes use `workflow.GetVersion` to protect in-flight executions.
- [ ] Workflows that process unbounded events use `workflow.NewContinueAsNewError`.
- [ ] Worker registration happens in the composition root, not inside domain packages.
- [ ] Workflow and Activity unit tests use `testsuite.WorkflowTestSuite` and mock all Activities.
