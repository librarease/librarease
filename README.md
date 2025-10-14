# librarease

A library management system built with Go and PostgreSQL.

## Architecture

- **API Server**: HTTP REST API (runs by default)
- **Worker Server**: Background job processor (runs with `-mode worker` flag)
- Both modes use the same binary built from `cmd/api/main.go`

## Quick Start

```bash
# Start infrastructure (PostgreSQL, Redis, MinIO)
make docker-run

# Run API server
make run

# Run worker (in another terminal)
make run-worker
```

## Development

### Build

```bash
make build          # Builds ./main binary
```

### Run

```bash
# API mode (default)
./main

# Worker mode
./main -mode worker

# Or use make commands
make run            # API
make run-worker     # Worker
```

### Live Reload

```bash
make watch          # API with auto-reload
make watch-worker   # Worker with auto-reload
```

### Testing

```bash
make test           # Unit tests
make itest          # Integration tests
make all            # Build + test
```

## Docker

### Development

```bash
make docker-run     # Start PostgreSQL, Redis, MinIO
make docker-down    # Stop infrastructure
make docker-logs    # View logs
```

### Production

```bash
# Build image
docker build -t librarease:latest .

# Run API
docker run -p 8080:8080 librarease:latest

# Run Worker
docker run librarease:latest ./main -mode worker

# Or use docker-compose
docker compose up -d
```

## Environment Variables

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=librarease
DB_USER=postgres
DB_PASSWORD=postgres

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Worker
WORKER_CONCURRENCY=10

# MinIO/S3
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin

# SMTP
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email
SMTP_PASSWORD=your-password
```

See `.env.example` for full configuration.

## Project Structure

```
cmd/api/main.go     # Single entry point with -mode flag
internal/
  server/           # HTTP handlers
  usecase/          # Business logic
  database/         # GORM repositories  
  queue/            # Background jobs
```

## Background Jobs

Worker processes jobs from Redis queue:
- `export:borrowings` - Export borrowing records
- Add more handlers in `internal/queue/handlers/`

## Makefile Commands

```bash
make build          # Build binary
make run            # Run API
make run-worker     # Run worker
make watch          # API with live reload
make watch-worker   # Worker with live reload
make test           # Unit tests
make itest          # Integration tests
make docker-run     # Start infrastructure
make docker-down    # Stop infrastructure
make clean          # Remove binary
```
