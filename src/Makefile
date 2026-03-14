.PHONY: fmt lint lint-docker docker-build test test-cover run build generate wire clean install dev-up dev-down migrate-up migrate-down migrate-create tidy vet test-ci

# ── Format ──────────────────────────────────────────────────────────────────────

fmt:
	gofumpt -l -w .
	gci write . --section Standard --section Default --section "Prefix(github.com/hoahm-ts/awesome-ai-skills)"

# ── Lint ─────────────────────────────────────────────────────────────────────────

lint:
	golangci-lint --config .golangci.yml run ./...

vet:
	go vet ./...

# Lint Dockerfile (requires hadolint: brew install hadolint)
lint-docker:
	hadolint docker/Dockerfile

# ── Build ───────────────────────────────────────────────────────────────────────

# Build production Docker image (build context is this directory)
docker-build:
	docker build -f docker/Dockerfile -t awesome-ai-skills:local .

build:
	go build -o bin/api ./cmd/api/
	go build -o bin/worker ./cmd/worker/

# ── Run ─────────────────────────────────────────────────────────────────────────

run:
	go run ./cmd/api/

# ── Test ─────────────────────────────────────────────────────────────────────────

test:
	go vet ./...
	go test -coverprofile=coverage.out ./internal/... ./pkg/...
	go tool cover -func=coverage.out | grep '^total:'

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run tests using the CI docker-compose stack (PostgreSQL + Redis via tmpfs)
test-ci:
	docker compose -f docker/docker-compose.test.yml run --rm test
	docker compose -f docker/docker-compose.test.yml down -v

# ── Code Generation ──────────────────────────────────────────────────────────────

generate:
	go generate ./...

# Regenerate Wire dependency injection
wire:
	wire ./...

# ── Install dev tools ────────────────────────────────────────────────────────────

install:
	go install mvdan.cc/gofumpt@latest
	go install github.com/daixiang0/gci@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# ── Local development environment ────────────────────────────────────────────────

# Start local dependencies (PostgreSQL, Redis, Kafka, Temporal)
dev-up:
	docker compose -f docker/docker-compose.dev.yml up -d

# Stop local dependencies
dev-down:
	docker compose -f docker/docker-compose.dev.yml down -v

# ── Database ─────────────────────────────────────────────────────────────────────

# Apply all pending migrations
migrate-up:
	docker compose -f docker/docker-compose.dev.yml run --rm app go run ./cmd/migration --direction=up

# Roll back the most recent migration
migrate-down:
	docker compose -f docker/docker-compose.dev.yml run --rm app go run ./cmd/migration --direction=down --steps=1

# Create a new empty migration pair: make migrate-create NAME=create_users
migrate-create:
	@[ "$(NAME)" ] || (echo "Usage: make migrate-create NAME=<migration_name>" && exit 1)
	migrate create -ext sql -dir migrations -seq $(NAME)

# ── Tidy ─────────────────────────────────────────────────────────────────────────

tidy:
	go mod tidy

# ── Clean ────────────────────────────────────────────────────────────────────────

clean:
	rm -rf bin/ coverage.out coverage.html
