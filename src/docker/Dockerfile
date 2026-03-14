# ── Build stage ──────────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o bin/api ./cmd/api

# ── Runtime stage ─────────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot AS api

WORKDIR /app

COPY --from=builder /app/bin/api .

# Non-sensitive defaults — override at runtime via environment variables or Docker secrets.
# DATABASE_DSN and REDIS_PASSWORD must always be injected at runtime; never set them here.
ENV APP_NAME="awesome-ai-skills" \
    APP_ENV="production" \
    APP_PORT="8080" \
    REDIS_ADDR="localhost:6379" \
    KAFKA_BROKER="localhost:9092" \
    KAFKA_GROUP_ID="awesome-ai-skills" \
    TEMPORAL_HOST_PORT="localhost:7233" \
    TEMPORAL_NAMESPACE="default" \
    TEMPORAL_TASK_QUEUE="default" \
    DD_SERVICE="awesome-ai-skills" \
    DD_ENV="production" \
    DD_AGENT_HOST="localhost"

EXPOSE 8080

ENTRYPOINT ["/app/api"]
