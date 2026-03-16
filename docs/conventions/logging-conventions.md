# Logging, Errors & Observability Conventions

> **When writing or reviewing any logging, error handling, or observability code, read this file first.**

---

## Table of Contents

- [General Rules](#general-rules)
- [Structured Logging with zerolog](#structured-logging-with-zerolog)
- [Observability Backend Overview](#observability-backend-overview)
- [GCP Cloud Trace / Cloud Monitoring / Cloud Logging (Default)](#gcp-cloud-trace--cloud-monitoring--cloud-logging-default)
- [OpenTelemetry Tracing](#opentelemetry-tracing)
- [DataDog Integration (Alternative)](#datadog-integration-alternative)
- [Switching Backends via OTEL_EXPORTER_BACKEND](#switching-backends-via-otel_exporter_backend)

---

## General Rules

- Include request IDs or correlation IDs wherever available.
- Wrap errors with context at boundaries (handlers, integrations).
- Prefer structured logs over ad-hoc strings.
- Do not log secrets, tokens, credentials, or PII — mask or redact them.

---

## Structured Logging with zerolog

Use [zerolog](https://github.com/rs/zerolog) as the standard logging library. It produces zero-allocation JSON logs and integrates cleanly with GCP Cloud Logging, DataDog, and OpenTelemetry.

### Setup

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

### Log levels

| Level | When to use |
|---|---|
| `Trace` | Very verbose, developer-only diagnostics |
| `Debug` | Detailed flow information useful during development |
| `Info` | Normal operational events (service started, request handled) |
| `Warn` | Recoverable anomalies that do not affect the outcome |
| `Error` | Failures that affect the current operation; always attach `Err(err)` |
| `Fatal` | Unrecoverable startup failures — `main()` only |

### Structured fields

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

### HTTP request logging patterns

Use consistent structured fields when logging HTTP traffic so that log queries and dashboards work uniformly across services.

The `marker` field is a fixed string tag that identifies the log category (`[api]` for inbound, `[request]` for outbound). It enables fast log filtering (e.g. `marker:[api]` in GCP Log Explorer or DataDog) without full-text search.

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

### Logger propagation via context

Attach the logger to `context.Context` at entry points so downstream functions receive it automatically.

```go
// At the handler boundary
ctx = log.With().Str("request_id", requestID).Logger().WithContext(ctx)

// Inside domain / service code
zerolog.Ctx(ctx).Info().Str("order_id", orderID).Msg("processing order")
```

Never pass a `*zerolog.Logger` as a struct field in domain types — always use `zerolog.Ctx(ctx)`.

### Avoid common mistakes

- Do not call `.Msg("")` with an empty string — use `.Send()` only when the structured fields alone fully describe the event; always prefer a concise, human-readable message otherwise.
- Do not construct log messages with `fmt.Sprintf` — use zerolog's typed field methods.
- Do not log and return an error at the same call site — choose one (see Error handling rules above).
- Do not enable `Debug` or `Trace` in production without a feature-flag-controlled log level.

---

## Observability Backend Overview

The default observability stack is **GCP Cloud Trace / Cloud Monitoring / Cloud Logging**. DataDog is supported as a fully configurable alternative. The active backend is selected at runtime via the `OTEL_EXPORTER_BACKEND` environment variable.

| Backend | `OTEL_EXPORTER_BACKEND` value | When to use |
|---|---|---|
| GCP Cloud Trace + Cloud Monitoring + Cloud Logging | `gcp` *(default)* | Services deployed on Google Cloud Platform |
| DataDog (via OTLP or DataDog Agent) | `datadog` | Services deployed with DataDog APM/Logs |

> **Default:** if `OTEL_EXPORTER_BACKEND` is absent or empty, the service initialises the GCP backend.

---

## GCP Cloud Trace / Cloud Monitoring / Cloud Logging (Default)

### Tracer initialisation for GCP

Use the [OpenTelemetry Go SDK](https://opentelemetry.io/docs/languages/go/) with the [GCP Cloud Trace exporter](https://pkg.go.dev/github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace).

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
    cloudtrace "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
)

func newGCPTracerProvider(ctx context.Context, projectID, serviceName, env string) (*sdktrace.TracerProvider, error) {
    exp, err := cloudtrace.New(cloudtrace.WithProjectID(projectID))
    if err != nil {
        return nil, fmt.Errorf("gcp cloud trace exporter: %w", err)
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

### Log correlation for GCP Cloud Logging

GCP Cloud Logging correlates logs with Cloud Trace spans using the `logging.googleapis.com/trace` and `logging.googleapis.com/spanId` fields. Inject them into every zerolog entry via a context hook.

```go
import (
    "fmt"

    "go.opentelemetry.io/otel/trace"
)

func gcpContextHook(projectID string) func(context.Context, *zerolog.Event) *zerolog.Event {
    return func(ctx context.Context, e *zerolog.Event) *zerolog.Event {
        sc := trace.SpanFromContext(ctx).SpanContext()
        if !sc.IsValid() {
            return e
        }
        return e.
            Str("logging.googleapis.com/trace", fmt.Sprintf("projects/%s/traces/%s", projectID, sc.TraceID().String())).
            Str("logging.googleapis.com/spanId", sc.SpanID().String()).
            Bool("logging.googleapis.com/trace_sampled", sc.IsSampled())
    }
}
```

### Required fields for GCP Cloud Logging

Every log entry must carry the following fields for GCP log management and correlation to work out of the box:

| zerolog field | GCP attribute | Notes |
|---|---|---|
| `service` | `service` | Set once at logger creation |
| `env` | `env` | Set once at logger creation |
| `logging.googleapis.com/trace` | Trace resource name | Injected per-request via hook; format `projects/{project}/traces/{trace_id}` |
| `logging.googleapis.com/spanId` | Span ID | Injected per-request via hook |
| `logging.googleapis.com/trace_sampled` | Sampling flag | Injected per-request via hook |

---

## OpenTelemetry Tracing

Use the [OpenTelemetry Go SDK](https://opentelemetry.io/docs/languages/go/) for vendor-neutral distributed tracing. The tracer provider is initialised in the composition root based on `OTEL_EXPORTER_BACKEND` (see [Switching Backends](#switching-backends-via-otel_exporter_backend)).

### Tracer initialisation

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

### Creating spans

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

### Log-trace correlation with OpenTelemetry

Inject the active OTel span into zerolog entries so logs and traces are correlated. When running on GCP, use the `gcpContextHook` (see [GCP section](#gcp-cloud-trace--cloud-monitoring--cloud-logging-default)). For backend-agnostic correlation, a generic hook can be used:

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

### Propagation

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

### Naming conventions

- Span names use `TypeName.MethodName` for service methods (e.g. `OrderService.Create`).
- Span names use `http.method http.route` for HTTP server spans (handled automatically by OTel contrib middleware).
- Attribute keys follow [OpenTelemetry Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/).

---

## DataDog Integration (Alternative)

DataDog is a configurable alternative to GCP. Enable it by setting `OTEL_EXPORTER_BACKEND=datadog` (see [Switching Backends](#switching-backends-via-otel_exporter_backend)).

### Log correlation

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

### Required tags / attributes

Every log entry must carry the following DataDog reserved attributes so that log management works out of the box:

| zerolog field | DataDog attribute | Notes |
|---|---|---|
| `service` | `service` | Set once at logger creation |
| `env` | `env` | Set once at logger creation |
| `dd.trace_id` | `dd.trace_id` | Injected per-request via hook |
| `dd.span_id` | `dd.span_id` | Injected per-request via hook |

### APM spans

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

## Switching Backends via OTEL_EXPORTER_BACKEND

The observability backend is selected at startup by reading the `OTEL_EXPORTER_BACKEND` environment variable in the composition root. The GCP backend is the default.

| Value | Backend initialised |
|---|---|
| `gcp` or *(unset)* | GCP Cloud Trace + Cloud Monitoring + Cloud Logging |
| `datadog` | DataDog APM + DataDog Log Management |

### Composition root wiring example

```go
func initObservability(ctx context.Context, cfg *config.Config) (*sdktrace.TracerProvider, error) {
    switch cfg.OTELExporterBackend {
    case "datadog":
        return newDataDogTracerProvider(ctx, cfg.ServiceName, cfg.Env)
    default: // "gcp" or unset — GCP is the default
        return newGCPTracerProvider(ctx, cfg.GCPProjectID, cfg.ServiceName, cfg.Env)
    }
}
```

### Environment variable reference

| Variable | Required | Default | Description |
|---|---|---|---|
| `OTEL_EXPORTER_BACKEND` | No | `gcp` | Observability backend: `gcp` or `datadog` |
| `GCP_PROJECT_ID` | When backend is `gcp` | — | GCP project ID used by Cloud Trace exporter |
| `DD_AGENT_HOST` | When backend is `datadog` | `localhost` | Hostname of the DataDog Agent |
| `DD_AGENT_PORT` | When backend is `datadog` | `8126` | Port of the DataDog Agent |

