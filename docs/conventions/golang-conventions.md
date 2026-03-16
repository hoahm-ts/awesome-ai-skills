# Go Conventions

> **When writing or reviewing any Go code, read this file first.**
>
> This consolidates Go-specific style, testing, architecture, and coding guidelines.
> These rules apply to all Go code in this repository.

---

## Table of Contents

- [General Coding Guidelines](#general-coding-guidelines)
- [Architecture & Layering Rules](#architecture--layering-rules)
- [Implementing a Feature (Standard Path)](#implementing-a-feature-standard-path)
- [Testing Expectations](#testing-expectations)
- [Code Quality Checklist](#code-quality-checklist)
- [Go Style Guidelines (Uber Style)](#go-style-guidelines-uber-style)

---

## General Coding Guidelines

- Write clean, readable, and well-structured code.
- Prefer explicit over implicit — avoid magic numbers, unclear abbreviations, and unexplained side effects.
- Follow the conventions already established in the codebase.
- Keep functions and modules small and focused on a single responsibility.
- Write meaningful commit messages using the [Conventional Commits](https://www.conventionalcommits.org/) format (e.g. `feat:`, `fix:`, `docs:`, `chore:`).

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

## Testing Expectations

- **Domain logic**: unit tests inside the domain package. Aim for high coverage of business rules.
- **Handler/command logic**: prefer table-driven tests for input validation and response mapping; use fakes or mocks for domain services via interfaces.
- **Integration adapters**: unit-test adapters with mocked HTTP/SDK responses. Do not call real external services in automated tests.
- Avoid flaky tests; tests must be deterministic and independent of external state.
- **Multi-tenancy**: the product supports multi-tenancy. Unit tests must use general, tenant-agnostic seed data. Features and test scenarios must not depend on or assume any specific client or tenant. If a requirement is client-specific, the implementation must remain general and the tests must validate behaviour using generic data only.
- **Parallel execution**: mark independent unit tests with `t.Parallel()` so the test suite runs as fast as possible. Sub-tests inside a table-driven test should also call `t.Parallel()` where safe:
  ```go
  func TestFoo(t *testing.T) {
    t.Parallel()
    tests := []struct{ ... }{ ... }
    for _, tt := range tests {
      // Do NOT add "tt := tt" — Go 1.22+ captures loop variables by reference automatically.
      t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        // ...
      })
    }
  }
  ```
- **High coverage**: generated tests must cover all meaningful code paths. Focus on high-risk areas and edge cases (boundary values, error paths, concurrent access, nil/empty inputs) rather than chasing an arbitrary percentage such as 100 %. A well-chosen set of targeted tests is more valuable than exhaustive but shallow ones.
- **Readability and maintainability**: AI-generated tests must be clean and structured, resembling human-written code. Use descriptive test names (`TestOrderService_CreateOrder_ReturnsErrNotFound`), keep each test case focused on a single behaviour, and avoid complex conditional logic or branching inside the test body — split into separate `Test...` functions instead.
- **Table-driven testing**: use table-driven tests with `t.Run` subtests whenever multiple inputs exercise the same logic. Name the slice `tests` and each entry `tt`. Use `give`/`want` prefixes for input/output fields:
  ```go
  tests := []struct {
    name    string
    give    Input
    want    Output
    wantErr error
  }{
    {name: "valid input", give: ..., want: ..., wantErr: nil},
    {name: "empty name returns error", give: ..., want: ..., wantErr: ErrInvalidInput},
  }
  for _, tt := range tests {
    // Do NOT add "tt := tt" — Go 1.22+ captures loop variables by reference automatically.
    t.Run(tt.name, func(t *testing.T) {
      t.Parallel()
      got, err := MyFunc(tt.give)
      require.ErrorIs(t, err, tt.wantErr)
      require.Equal(t, tt.want, got)
    })
  }
  ```

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
