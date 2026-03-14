# Architecture

This document describes the high-level architecture of awesome-ai-skills.

## Layers

```
cmd/              Entry points (HTTP server, background worker, migration runner)
  api/            API server binary
  worker/         Temporal worker binary
  migration/      Database migration binary

internal/         Application core — not importable outside this module
  <domain>/       Domain module: entities, services, repository ports
  handler/        HTTP route handlers (delivery layer)
  integration/    External service adapters (infra layer)
  lifecycle/      Startup and graceful shutdown management
  timeline/       Temporal workflow and activity definitions
  event/          Kafka event publishers and consumers
  shared/         Shared kernel: sentinel errors, common value objects, pagination types
  wire/           Single composition root — all dependency wiring lives here

pkg/              Reusable helper libraries with no business logic
  config/         Application configuration loading
  logger/         zerolog logger setup and context helpers
  middleware/      HTTP middleware (request logging, etc.)
  response/       Standard JSON HTTP response helpers
  utils/          Small general-purpose helpers
  security/       Password hashing and verification
```

## Dependency Direction

```
cmd  →  wire  →  internal/<domain>  →  internal/shared
                 internal/handler
                 internal/integration
                 internal/lifecycle
                 internal/timeline
                 internal/event
             ↘  pkg/*
```

- `cmd` wires everything together via `internal/wire`.
- `internal/<domain>` packages define ports (interfaces) that `internal/integration` implements.
- `internal/handler` depends on domain service interfaces, not concrete types.
- `pkg` packages have no dependencies on `internal` packages.

## Key Design Decisions

- **One composition root** (`internal/wire`): all `*gorm.DB`, Redis, Temporal, and zerolog instances are created once and injected.
- **Repository pattern**: every database table has a repository interface defined in the domain package and a concrete implementation in `internal/integration` (or a dedicated infra sub-package).
- **No global state**: package-level variables are avoided. Dependencies are explicit constructor arguments.
- **Cursor-based pagination**: keyset pagination is used instead of OFFSET to support large data sets efficiently.

## Tech Stack

| Concern            | Technology                          |
|--------------------|-------------------------------------|
| HTTP framework     | [chi](https://github.com/go-chi/chi) |
| Database           | PostgreSQL + [GORM](https://gorm.io) |
| Cache              | Redis                               |
| Messaging          | Kafka                               |
| Workflow engine    | [Temporal](https://temporal.io)     |
| Logging            | [zerolog](https://github.com/rs/zerolog) |
| Tracing            | OpenTelemetry → Datadog             |
| Dependency injection | [Google Wire](https://github.com/google/wire) |
| Migrations         | [golang-migrate](https://github.com/golang-migrate/migrate) |
