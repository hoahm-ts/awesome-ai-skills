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
