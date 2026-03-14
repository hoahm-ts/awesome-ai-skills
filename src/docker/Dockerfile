# ── Build stage ──────────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS base

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o bin/api ./cmd/api

# ── Runtime stage ─────────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot AS api

WORKDIR /app

COPY --from=base /src/bin/api .
COPY --from=base /src/etc/config/ /app/etc/config/
COPY --from=powerman/dockerize:0.19.0 /usr/local/bin/dockerize /usr/local/bin/dockerize

# Non-sensitive defaults — override at runtime via environment variables or Docker secrets.
# DATABASE_DSN and REDIS_PASSWORD must always be injected at runtime; never set them here.
ENV APP_NAME="awesome-ai-skills" \
    APP_ENV="production" \
    APP_PORT="8080" \
    APP_LOG_LEVEL="info" \
    APP_READ_TIMEOUT_SECONDS="15" \
    APP_WRITE_TIMEOUT_SECONDS="15" \
    APP_IDLE_TIMEOUT_SECONDS="60" \
    APP_SHUTDOWN_TIMEOUT_SECONDS="30" \
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

ENTRYPOINT ["dockerize", "-template", "etc/config/template.yml:etc/config/app_config.yml"]
CMD ["/app/api"]
