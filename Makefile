# Simple Makefile for a Go project

# Build the application
all: build test

build:
	@echo "Building API server..."
	@go build -o bin/api cmd/api/main.go
	@echo "Building worker..."
	@go build -o bin/worker cmd/worker/main.go
	@echo "Build complete!"

build-api:
	@echo "Building API server..."
	@go build -o bin/api cmd/api/main.go

build-worker:
	@echo "Building worker..."
	@go build -o bin/worker cmd/worker/main.go

# Run the application
run:
	@go run cmd/api/main.go

run-worker:
	@go run cmd/worker/main.go
# Start local development infrastructure (DB, Redis, MinIO)
docker-run:
	@echo "Starting local development infrastructure..."
	@if docker compose -f docker-compose.dev.yml up -d 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose -f docker-compose.dev.yml up -d; \
	fi
	@echo "Infrastructure is ready!"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"
	@echo "MinIO: localhost:9000 (console: localhost:9001)"

# Shutdown local development infrastructure
docker-down:
	@echo "Stopping local development infrastructure..."
	@if docker compose -f docker-compose.dev.yml down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose -f docker-compose.dev.yml down; \
	fi

# View logs from development infrastructure
docker-logs:
	@if docker compose -f docker-compose.dev.yml logs -f 2>/dev/null; then \
		: ; \
	else \
		docker-compose -f docker-compose.dev.yml logs -f; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go test -v ./internal/server/... \
	./internal/usecase/...
# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f bin/api bin/worker
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

# Live Reload for Worker
watch-worker:
	@if command -v air > /dev/null; then \
            air -c .air.worker.toml; \
            echo "Watching Worker...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air -c .air.worker.toml; \
                echo "Watching Worker...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

# Debug the application
.PHONY: debug debug-worker

debug:
	@echo "Starting API debugger on :2345..."
	@dlv debug cmd/api/main.go --headless --listen=:2345 --api-version=2 --log

debug-worker:
	@echo "Starting worker debugger on :2346..."
	@dlv debug cmd/worker/main.go --headless --listen=:2346 --api-version=2 --log

.PHONY: all build build-prod run run-worker test clean watch watch-worker docker-run docker-down docker-logs itest debug debug-worker
