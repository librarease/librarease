# librarease

A library management system built with Go and PostgreSQL.

## License

This project is source-available under the [PolyForm Noncommercial License 1.0.0](LICENSE).  
**Commercial use is not permitted.**

For educational, personal, and noncommercial use only. If you're interested in commercial use, contact: solidifyarmor@gmail.com

## System Components

- **API Server** (`cmd/api/main.go`) - HTTP REST API
- **Worker** (`cmd/worker/main.go`) - Background job processor
- **Scheduler** (`cmd/worker/main.go -mode scheduler`) - Periodic task scheduler

## Prerequisites

- Go 1.25+
- PostgreSQL 16+
- Redis
- MinIO or S3
- Firebase service account

## Quick Setup

### 1. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` with required values. Reference `docker-compose.example.yml` (excluding `frontend-service`) for all environment variables:

- **Database**: `DB_HOST`, `DB_PORT`, `DB_DATABASE`, `DB_USER`, `DB_PASSWORD`
- **Redis**: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`
- **MinIO/S3**: `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `MINIO_BUCKET_NAME`
- **Firebase**: `FIREBASE_SERVICE_ACCOUNT_KEY_PATH`
- **SMTP**: `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`

### 2. Start Infrastructure

Run PostgreSQL, and Redis locally.

### 3. Run with Live Reload

```bash
make watch
```

This installs [air](https://github.com/air-verse/air) if needed and runs all components with hot reload.

## Development Commands

```bash
make build           # Build binaries
make run             # Run API server
make run-worker      # Run worker
make run-scheduler   # Run scheduler
make watch           # API with live reload
make watch-worker    # Worker with live reload
make test            # Unit tests
make itest           # Integration tests
make clean           # Remove binaries
```

## Project Structure

```
cmd/
  api/main.go       # API server entry point
  worker/main.go    # Worker/scheduler entry point
internal/
  server/           # HTTP handlers (Echo)
  usecase/          # Business logic
  database/         # GORM repositories
  queue/            # Background job handlers
  config/           # Environment constants
```

## Architecture

Clean architecture with three layers:
- **Handlers** - HTTP API endpoints
- **Usecases** - Business logic with interface injection
- **Database** - GORM repository implementations

All requests follow a bind → validate → usecase pattern with structured JSON responses.
