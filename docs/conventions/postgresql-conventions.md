# PostgreSQL & GORM Conventions

> **When writing or reviewing any PostgreSQL/GORM code, read this file first.**
>
> These guidelines apply to all Go code that interacts with PostgreSQL via [GORM](https://gorm.io/).

---

## Table of Contents

- [Model Definition](#model-definition)
- [Primary Key Strategy](#primary-key-strategy)
- [Naming Conventions](#naming-conventions)
- [Connection & Pool Configuration](#connection--pool-configuration)
- [Repository Pattern](#repository-pattern)
- [SQL Injection Prevention](#sql-injection-prevention)
- [Avoiding N+1 Queries](#avoiding-n1-queries)
- [Query Optimization](#query-optimization)
- [Transactions](#transactions)
- [Error Handling](#error-handling)
- [Data Types & Storage Optimization](#data-types--storage-optimization)
- [Migrations](#migrations)
- [Soft Deletes](#soft-deletes)
- [Indexes & Constraints](#indexes--constraints)
- [Data Growth & Capacity Planning](#data-growth--capacity-planning)
- [Partitioning & Sharding](#partitioning--sharding)

---

## Model Definition

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

---

## Primary Key Strategy

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

---

## Naming Conventions

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

---

## Connection & Pool Configuration

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

---

## Repository Pattern

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

---

## SQL Injection Prevention

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

---

## Avoiding N+1 Queries

- Use `Preload` for loading associations when you need all records and their relations in one logical operation:
  ```go
  db.WithContext(ctx).Preload("Orders").Find(&users)
  ```
- Use `Joins` when you need to filter on the association or want a single SQL join:
  ```go
  db.WithContext(ctx).Joins("JOIN orders ON orders.user_id = users.id").Find(&users)
  ```
- Audit any loop that performs a database query inside it — refactor to a batch fetch instead.

---

## Query Optimization

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

---

## Transactions

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

---

## Error Handling

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

---

## Data Types & Storage Optimization

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
- Use `BIGSERIAL` (not `SERIAL`) for auto-increment primary keys on tables that may grow beyond ~2 billion rows.
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

---

## Migrations

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

---

## Soft Deletes

- Use `gorm.DeletedAt` (which sets `DeletedAt` to the current time) for soft-deleting records. GORM automatically adds `WHERE deleted_at IS NULL` to all queries on models with this field.
- Add a database index on the `deleted_at` column for tables with high query volume.
- When you genuinely need to query deleted records, use `db.Unscoped()`:
  ```go
  db.WithContext(ctx).Unscoped().Where("id = ?", id).First(&user)
  ```
- Do not mix soft-delete and hard-delete patterns on the same table.

---

## Indexes & Constraints

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

---

## Data Growth & Capacity Planning

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
- **Integer overflow**: with `SERIAL` (int32), the sequence wraps at ~2.1 billion. Audit high-write tables and migrate to `BIGSERIAL` or UUID before the sequence is exhausted.
- **Autovacuum tuning**: tune `autovacuum_vacuum_scale_factor` and `autovacuum_analyze_scale_factor` per table when the default 20% threshold is too coarse for large tables:
  ```sql
  ALTER TABLE events SET (
    autovacuum_vacuum_scale_factor  = 0.01,
    autovacuum_analyze_scale_factor = 0.01
  );
  ```

---

## Partitioning & Sharding

Use partitioning and sharding when a single table or a single server can no longer sustain the required throughput or capacity.

**Table Partitioning (single server)**

- **Range partitioning** — ideal for time-series data (events, logs, audit trails):
  ```sql
  CREATE TABLE events (
    id         BIGSERIAL,
    created_at TIMESTAMPTZ NOT NULL,
    payload    JSONB
  ) PARTITION BY RANGE (created_at);

  CREATE TABLE events_2025_q1 PARTITION OF events
    FOR VALUES FROM ('2025-01-01') TO ('2025-04-01');
  ```
- **List partitioning** — useful when rows are naturally grouped by a discrete key (e.g. `tenant_id`, `region`):
  ```sql
  CREATE TABLE orders (
    id        BIGSERIAL,
    region    TEXT NOT NULL,
    amount    BIGINT NOT NULL
  ) PARTITION BY LIST (region);

  CREATE TABLE orders_us PARTITION OF orders FOR VALUES IN ('us-east', 'us-west');
  ```
- Always create indexes on the partition key column and on any column used in `WHERE` filters across partitions.
- Automate partition creation in advance using a scheduled job or migration.
- Drop old partitions instead of `DELETE`-ing rows — dropping a partition is an O(1) metadata operation.

**Application-level Sharding (multiple servers)**

- Choose a shard key that distributes writes evenly and avoids cross-shard joins (e.g. `tenant_id`, `user_id`).
- Keep all data for a single shard key on the same node to allow most queries to be executed on a single shard.
- Route queries in the application layer via a shard map (shard key → DSN).
- Avoid cross-shard transactions; redesign data models to keep related entities on the same shard.
- Consider [Citus](https://github.com/citusdata/citus) (PostgreSQL extension) for transparent sharding before implementing a custom shard router.
