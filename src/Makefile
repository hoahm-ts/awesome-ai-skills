# ── Build ───────────────────────────────────────────────────────────────────────
.PHONY: build build-api build-worker build-migration

build: build-api build-worker build-migration

build-api:
	go build -o bin/api ./cmd/api

build-worker:
	go build -o bin/worker ./cmd/worker

build-migration:
	go build -o bin/migration ./cmd/migration

# ── Run ─────────────────────────────────────────────────────────────────────────
.PHONY: run-api run-worker

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

# ── Test ─────────────────────────────────────────────────────────────────────────
.PHONY: test test-cover

test:
	go test ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# ── Lint ─────────────────────────────────────────────────────────────────────────
.PHONY: lint vet fmt

lint:
	golangci-lint run ./...

vet:
	go vet ./...

fmt:
	goimports -w .

# ── Code Generation ──────────────────────────────────────────────────────────────
.PHONY: wire generate

wire:
	wire gen ./internal/wire/...

generate:
	go generate ./...

# ── Database ─────────────────────────────────────────────────────────────────────
.PHONY: migrate-up migrate-down migrate-create

migrate-up:
	go run ./cmd/migration

migrate-down:
	migrate -database "$(DATABASE_DSN)" -path migrations down 1

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

# ── Tidy ─────────────────────────────────────────────────────────────────────────
.PHONY: tidy

tidy:
	go mod tidy

# ── Clean ────────────────────────────────────────────────────────────────────────
.PHONY: clean

clean:
	rm -rf bin/ coverage.out coverage.html
