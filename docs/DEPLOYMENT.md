# Deployment Architecture

## Overview

The librarease backend uses a **two-binary architecture** where the API server and worker are separate executables sharing common internal packages.

## Structure

```
cmd/
├── api/          # API server entry point
│   └── main.go   # HTTP server with Echo framework
└── worker/       # Worker entry point
    └── main.go   # Asynq worker for background jobs
```

## Docker Build

A single `Dockerfile` builds **both binaries**:

```dockerfile
# Build stage - builds both binaries
RUN go build -o bin/api cmd/api/main.go
RUN go build -o bin/worker cmd/worker/main.go

# Production stage - copies both binaries
COPY --from=build /app/bin/api /app/api
COPY --from=build /app/bin/worker /app/worker
```

## Docker Compose Services

Two separate services run from the **same Docker image**:

### API Service
```yaml
api-service:
  image: <registry>/backend
  command: ["./api"]
  ports:
    - 8080:8080
  environment:
    - PORT=8080
    - DB_HOST=db
    - REDIS_HOST=redis
    # ... other env vars
```

### Worker Service
```yaml
worker-service:
  image: <registry>/backend
  command: ["./worker"]
  environment:
    - DB_HOST=db
    - REDIS_HOST=redis
    - WORKER_CONCURRENCY=10
    # ... other env vars
```

## Local Development

### Build binaries
```bash
make build          # Build both API and worker
make build-api      # Build API only
make build-worker   # Build worker only
```

### Run services
```bash
make run            # Run API server
make run-worker     # Run worker
```

### Live reload
```bash
make watch          # Watch API server
make watch-worker   # Watch worker
```

### Infrastructure
```bash
make docker-run     # Start PostgreSQL, Redis, MinIO
make docker-down    # Stop infrastructure
make docker-logs    # View logs
```

## Scaling

- **API service**: Scale horizontally for handling HTTP requests
- **Worker service**: Scale horizontally for processing background jobs

Each service can be scaled independently:
```bash
docker compose up -d --scale api-service=3 --scale worker-service=2
```

## Shared Logic

All business logic is shared via `internal/` packages:
- `internal/server` - HTTP handlers (API only)
- `internal/queue` - Asynq client/server setup
- `internal/usecase` - Business logic (shared)
- `internal/database` - Data access (shared)
- `internal/config` - Configuration (shared)
- `internal/email` - Email provider (shared)
- `internal/filestorage` - File storage (shared)
- `internal/firebase` - Firebase auth (shared)
- `internal/push` - Push notifications (shared)

## Environment Variables

### Common (both services)
- `DB_HOST`, `DB_PORT`, `DB_DATABASE`, `DB_USER`, `DB_PASSWORD`
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`
- `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`

### API-specific
- `PORT` - HTTP server port (default: 8080)

### Worker-specific
- `WORKER_CONCURRENCY` - Number of concurrent workers (default: 10)

## Deployment Workflow

1. **Build**: Single Dockerfile builds both binaries
2. **Push**: Push image to registry once
3. **Deploy**: Use same image for both services
4. **Scale**: Scale services independently

## Benefits

✅ **Code Reuse**: All logic in `internal/` packages is shared  
✅ **Single Build**: One Docker image for both services  
✅ **Independent Scaling**: Scale API and worker separately  
✅ **Clean Separation**: Clear entry points for different concerns  
✅ **Easy Testing**: Test shared logic once, use in both services  
