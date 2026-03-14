# System Architecture

This document describes the high-level architecture, layering rules, and component interactions for services built with the conventions defined in `AGENTS.md`.

## Table of Contents

- [Overview](#overview)
- [Layered Architecture](#layered-architecture)
  - [Layer Responsibilities](#layer-responsibilities)
  - [Allowed Dependency Direction](#allowed-dependency-direction)
- [Component Diagram](#component-diagram)
- [Tech Stack](#tech-stack)
- [Key Design Decisions](#key-design-decisions)
  - [Dependency Injection](#dependency-injection)
  - [Repository Pattern](#repository-pattern)
  - [Error Handling](#error-handling)
  - [Observability](#observability)

---

## Overview

Services follow a **clean, layered architecture** where each layer has a single, well-defined responsibility. Business logic is isolated in the domain layer and is never coupled to transport, persistence, or external integration concerns.

```
┌──────────────────────────────────────────────────────────┐
│                      Entry Points                        │
│            (HTTP handlers · CLI · Workers)               │
└───────────────────────┬──────────────────────────────────┘
                        │ calls via interfaces
┌───────────────────────▼──────────────────────────────────┐
│                    Domain Layer                          │
│         (services · business rules · ports)              │
└────────┬──────────────────────────────────┬──────────────┘
         │ implements                        │ calls
┌────────▼──────────┐             ┌─────────▼──────────────┐
│  Infrastructure   │             │     Integrations        │
│ (PostgreSQL/GORM  │             │  (Kafka · Temporal ·    │
│  Redis · …)       │             │   external HTTP APIs)   │
└───────────────────┘             └────────────────────────┘
```

---

## Layered Architecture

### Layer Responsibilities

| Layer | Package path | Responsibilities |
|---|---|---|
| **Entry points** | `cmd/`, `internal/handler/`, `internal/worker/` | Parse and validate input; call domain services; return transport-formatted responses. No business logic. |
| **Domain** | `internal/<domain>/` | Business rules, domain services, aggregate types, and port interfaces. Free of transport and infrastructure imports. |
| **Infrastructure** | `internal/infrastructure/` | Concrete implementations of domain ports: GORM repositories, Redis clients, in-process caches. |
| **Integrations** | `internal/integration/` | External-service adapters: Kafka producers/consumers, Temporal workflow starters, third-party HTTP clients. |
| **Wiring / DI** | `internal/wire/`, `cmd/main.go` | Single composition root. Constructs all concrete implementations and injects them through domain interfaces. |

### Allowed Dependency Direction

```
entry points  ──►  domain services (via interfaces)
domain        ──►  shared ports / types
infrastructure ──►  domain ports (implements)
integrations  ──►  domain ports (implements)
wiring        ──►  all layers (constructs & injects)
```

**Forbidden** dependencies:

- Domain → infrastructure or integration concrete types.
- Domain A → Domain B concrete packages (use shared port interfaces).
- Handlers → repositories directly (must go through a domain service).
- Any layer → `cmd/` or wiring packages.

---

## Component Diagram

```
                         ┌──────────────────┐
  HTTP client ──────────►│  chi HTTP router  │
                         └────────┬─────────┘
                                  │
                         ┌────────▼─────────┐
                         │   HTTP Handler   │◄── request validation
                         └────────┬─────────┘
                                  │ domain.Service interface
                         ┌────────▼─────────┐
                         │  Domain Service  │◄── business rules
                         └──┬──────────┬────┘
                            │          │
              ┌─────────────▼──┐  ┌────▼──────────────────┐
              │  Repository    │  │  Integration Adapter   │
              │  (PostgreSQL)  │  │  (Kafka / Temporal /   │
              └───────┬────────┘  │   external HTTP)       │
                      │           └───────────────────────┬┘
              ┌───────▼────────┐                          │
              │   PostgreSQL   │              ┌───────────▼────────┐
              └────────────────┘              │  External Services │
                                              └────────────────────┘
              ┌────────────────┐
              │     Redis      │◄── cache layer (optional per domain)
              └────────────────┘
```

---

## Tech Stack

| Category | Technology | Notes |
|---|---|---|
| Language | [Go](https://go.dev/) 1.24+ | All services written in Go |
| HTTP framework | [chi](https://github.com/go-chi/chi) | Lightweight, idiomatic router |
| Database | [PostgreSQL](https://www.postgresql.org/) via [GORM](https://gorm.io/) | Primary datastore |
| Cache | [Redis](https://redis.io/) | Session data, rate limiting, short-lived state |
| Messaging | [Kafka](https://kafka.apache.org/) | Async event streaming between services |
| Workflow engine | [Temporal](https://temporal.io/) | Long-running, durable workflows |
| Tracing | [Datadog](https://www.datadoghq.com/) / [OpenTelemetry](https://opentelemetry.io/) | APM and distributed tracing |
| Dependency injection | [Google Wire](https://github.com/google/wire) | Compile-time DI code generation |
| Logging | [zerolog](https://github.com/rs/zerolog) | Structured, zero-allocation JSON logs |

---

## Key Design Decisions

### Dependency Injection

All dependencies are wired in a single **composition root** — typically `cmd/main.go` or a `wire.go` file generated by Google Wire. No package-level globals or hidden singletons.

```
main.go
  └── wire.go  (generated)
        ├── NewHTTPServer(handler, …)
        ├── NewHandler(service)
        ├── NewService(repo, kafkaProducer, …)
        ├── NewRepository(db)
        └── NewDB(config)
```

### Repository Pattern

Every domain that touches persistent storage defines a **port interface** in the domain package. The concrete GORM implementation lives in `internal/infrastructure/` and depends on that interface — never the reverse.

```go
// domain/order/repository.go  — port (domain package)
type Repository interface {
    FindByID(ctx context.Context, id uuid.UUID) (*Order, error)
    Save(ctx context.Context, o *Order) error
}

// infrastructure/postgres/order_repository.go — adapter
type orderRepository struct{ db *gorm.DB }

var _ order.Repository = (*orderRepository)(nil) // compile-time check
```

### Error Handling

- Errors are wrapped with context at every boundary using `fmt.Errorf("…: %w", err)`.
- Domain services return typed domain errors (e.g. `ErrNotFound`, `ErrAlreadyExists`).
- Handlers translate domain errors to HTTP status codes; they never expose raw `gorm` or database errors.
- The rule "handle errors once" is strictly followed — no log-and-return duplication.

### Observability

All services emit:

- **Structured logs** via zerolog with `request_id`, `user_id`, and DataDog trace correlation fields on every log entry.
- **Distributed traces** via OpenTelemetry with spans created for every significant operation (HTTP handler, DB query, outbound call).
- **Metrics** forwarded to Datadog via the OTLP exporter or the Datadog Agent.

See the [Logging, Errors & Observability](../../AGENTS.md#logging-errors--observability) section in `AGENTS.md` for detailed configuration patterns.
