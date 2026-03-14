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

EXPOSE 8080

ENTRYPOINT ["/app/api"]
