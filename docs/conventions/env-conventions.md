# Environment & Containerization Conventions

> **When writing or reviewing environment variable definitions, Dockerfiles, or Docker Compose files, read this file first.**

---

## Table of Contents

- [Environment Variable Naming Conventions](#environment-variable-naming-conventions)
- [Multi-stage Builds](#multi-stage-builds)
- [Distroless Base Images](#distroless-base-images)
- [Security Hardening](#security-hardening)
- [Image Tagging & Versioning](#image-tagging--versioning)
- [Build Arguments & Secrets](#build-arguments--secrets)
- [`.dockerignore`](#dockerignore)
- [Health Checks](#health-checks)
- [Example: Complete Production Dockerfile](#example-complete-production-dockerfile)
- [Docker Compose for Local Development](#docker-compose-for-local-development)

---

## Environment Variable Naming Conventions

Environment variables are the primary runtime configuration mechanism for containerized services. Follow these conventions for all `ENV` declarations in Dockerfiles, Docker Compose files, and Kubernetes manifests.

### General rules

- Use **UPPER_SNAKE_CASE** for all variable names.
- Group related variables by a consistent service or domain prefix (e.g. `DB_`, `REDIS_`, `KAFKA_`, `HTTP_`).
- Prefer explicit, descriptive names over terse abbreviations.

```dockerfile
# Good — grouped by domain, uppercase, underscore-separated
ENV DB_HOST=localhost \
    DB_PORT=5432 \
    DB_NAME=mydb \
    HTTP_PORT=8080 \
    HTTP_READ_TIMEOUT_DURATION=30s

# Bad — no grouping, mixed case, ambiguous names
ENV dbhost=localhost port=8080 t=30
```

### Boolean variables

- Prefix with `IS_` for state: `IS_ACTIVE`, `IS_VERIFIED`.
- Prefix with `HAS_` for capability: `HAS_FEATURE_X`, `HAS_PREMIUM_ACCESS`.
- Prefix with `ENABLE_` for feature flags: `ENABLE_DARK_MODE`, `ENABLE_RATE_LIMITING`.
- **Never** use negatives (`DISABLE_`, `INACTIVE_`, `NO_`) — they force double negation in code (`if !disableFeature`).

```dockerfile
# Good
ENV ENABLE_METRICS=true \
    ENABLE_TRACING=false \
    IS_READ_ONLY=false \
    HAS_LEGACY_SUPPORT=false

# Bad
ENV DISABLE_METRICS=false \
    INACTIVE_TRACING=true
```

### Time / duration variables

Use the `_DURATION` suffix for values that are parsed as `time.Duration` (Go-style strings such as `30s`, `5m`, `1h`). When the unit is fixed and the value is a plain integer, encode the unit in the suffix:

| Suffix | Example | Parsed as |
|---|---|---|
| `_DURATION` | `HTTP_READ_TIMEOUT_DURATION=30s` | `time.Duration` |
| `_DURATION_SECONDS` | `SESSION_TTL_DURATION_SECONDS=3600` | `int` seconds |
| `_DURATION_MINUTES` | `CACHE_TTL_DURATION_MINUTES=10` | `int` minutes |
| `_DURATION_HOURS` | `TOKEN_EXPIRY_DURATION_HOURS=24` | `int` hours |

```dockerfile
ENV HTTP_READ_TIMEOUT_DURATION=30s \
    HTTP_WRITE_TIMEOUT_DURATION=30s \
    SESSION_TTL_DURATION_SECONDS=3600 \
    CACHE_TTL_DURATION_MINUTES=10
```

### Numeric variables

Use a descriptive unit suffix to avoid ambiguity:

| Suffix | Example |
|---|---|
| `_AMOUNT` | `MAX_RETRY_AMOUNT=3` |
| `_LIMIT` | `DB_CONNECTION_LIMIT=25` |
| `_COUNT` | `WORKER_COUNT=4` |
| `_RATE` | `PUBLISH_RATE=100` |
| `_LEVEL` | `LOG_LEVEL=1` |
| `_PERCENTAGE` | `CACHE_EVICTION_PERCENTAGE=20` |

```dockerfile
ENV DB_CONNECTION_LIMIT=25 \
    DB_IDLE_CONNECTION_LIMIT=5 \
    WORKER_COUNT=4 \
    MAX_RETRY_AMOUNT=3 \
    RATE_LIMIT_RATE=1000 \
    RATE_LIMIT_BURST_LIMIT=100
```

### Complete example

```dockerfile
ENV \
    # Server
    HTTP_HOST=0.0.0.0 \
    HTTP_PORT=8080 \
    HTTP_READ_TIMEOUT_DURATION=30s \
    HTTP_WRITE_TIMEOUT_DURATION=30s \
    # Database
    DB_HOST=localhost \
    DB_PORT=5432 \
    DB_NAME=mydb \
    DB_CONNECTION_LIMIT=25 \
    DB_IDLE_CONNECTION_LIMIT=5 \
    DB_CONNECTION_MAX_LIFETIME_DURATION_MINUTES=5 \
    # Feature flags
    ENABLE_METRICS=true \
    ENABLE_TRACING=true \
    ENABLE_RATE_LIMITING=false \
    # Rate limiting
    RATE_LIMIT_RATE=1000 \
    RATE_LIMIT_BURST_LIMIT=100 \
    # Workers
    WORKER_COUNT=4 \
    MAX_RETRY_AMOUNT=3 \
    # Observability backend — "gcp" (default) or "datadog"
    OTEL_EXPORTER_BACKEND=gcp
```

---

## Multi-stage Builds

Always use a multi-stage `Dockerfile`. This separates the build environment from the final runtime image.

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
- **`golang:X.Y-alpine` as builder** — smaller than the full `golang` image.
- **`CGO_ENABLED=0`** — produces a fully static binary with no libc dependency, required for distroless/scratch base images.
- **`-trimpath`** — removes local file-system paths from the binary, improving reproducibility and avoiding information leakage.
- **`-ldflags="-s -w"`** — strips the symbol table and DWARF debug info, reducing binary size by ~30%.
- **`COPY go.mod go.sum ./` before `COPY . .`** — exploits Docker's layer cache.

---

## Distroless Base Images

[Distroless images](https://github.com/GoogleContainerTools/distroless) contain only the application binary and its minimal runtime dependencies — no shell, package manager, or OS utilities.

| Base image | Use when |
|---|---|
| `gcr.io/distroless/static-debian12` | CGO disabled; pure Go binary |
| `gcr.io/distroless/base-debian12` | CGO enabled; needs glibc |
| `gcr.io/distroless/static-debian12:nonroot` | Same as above but runs as UID 65532 (preferred) |
| `scratch` | Absolute minimal; no CA certificates or timezone data |

Guidelines:
- Prefer `distroless/static-debian12:nonroot` over `scratch` — it ships CA certificates and timezone data, and the `:nonroot` tag ensures the container never runs as root.
- Use `distroless/base-debian12` only when your binary requires CGO. Avoid CGO in production services whenever possible.
- Pin the digest, not just the tag, for reproducible builds in production (`FROM gcr.io/distroless/static-debian12@sha256:<digest>`).

---

## Security Hardening

- **Run as non-root**: always add `USER nonroot:nonroot` (distroless) or create a dedicated user in the final stage:
  ```dockerfile
  # When using a non-distroless base image
  RUN addgroup -S appgroup && adduser -S appuser -G appgroup
  USER appuser
  ```
- **Read-only root filesystem**: set `readOnlyRootFilesystem: true` in the Kubernetes `securityContext`. Mount explicit `emptyDir` volumes only for paths that genuinely need writes.
- **No new privileges**: set `allowPrivilegeEscalation: false` and `NO_NEW_PRIVILEGES=true`.
- **Minimal capabilities**: drop all Linux capabilities with `--cap-drop=ALL` and add back only what is required.
- **Scan the image**: integrate a vulnerability scanner (e.g. `trivy`, `grype`, or `docker scout`) into CI and fail the build on critical/high CVEs.

---

## Image Tagging & Versioning

- Never use the `latest` tag in production. Tag images with an immutable identifier:
  ```
  ghcr.io/myorg/myservice:<git-sha>
  ghcr.io/myorg/myservice:<semver>      # e.g. v1.4.2
  ```
- Push both the SHA tag (immutable, for rollback) and the semver tag (human-readable, for deploys).
- Use Docker BuildKit (`DOCKER_BUILDKIT=1`) or `docker buildx` for all builds to get caching, multi-platform support, and inline cache metadata.

---

## Build Arguments & Secrets

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
- Never bake secrets (API keys, certificates, passwords) into the image. Use Docker BuildKit's `--secret` mount for build-time secrets:
  ```dockerfile
  RUN --mount=type=secret,id=netrc,dst=/root/.netrc \
      go mod download
  ```
- Provide secrets at runtime via environment variables or a secrets manager (Vault, AWS Secrets Manager, GCP Secret Manager) — never via `ENV` instructions in the Dockerfile.

---

## `.dockerignore`

Always include a `.dockerignore` file to exclude files and directories that must not be sent to the build context:

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

---

## Health Checks

Add a `HEALTHCHECK` instruction so the Docker daemon and orchestrators can monitor the container's liveness:

```dockerfile
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/server", "-healthcheck"]
```

A self-contained health-check flag in your binary is the preferred approach for distroless containers.

---

## Example: Complete Production Dockerfile

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

---

## Docker Compose for Local Development

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

- Use `target: builder` in Compose for the development service so you have access to the full toolchain.
- Use `target: <final-stage>` (or omit `target`) for the production image build in CI.
- Never commit `.env` files with real credentials; use `.env.example` as a template and load secrets from a secrets manager in production.
