# syntax=docker/dockerfile:1

# ── Build stage ──────────────────────────────────────────────────────────────────
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS base

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -trimpath -o bin/api ./cmd/api

# ── Runtime stage ─────────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot AS api

WORKDIR /app

COPY --from=base /src/bin/api .
COPY --from=base /src/etc/config/ /app/etc/config/
COPY --from=powerman/dockerize:0.19.0 /usr/local/bin/dockerize /usr/local/bin/dockerize

EXPOSE 8080

ENTRYPOINT ["dockerize", "-template", "etc/config/template.yml:etc/config/app_config.yml"]
CMD ["/app/api"]
