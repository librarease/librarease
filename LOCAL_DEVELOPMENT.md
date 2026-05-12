# Local Development

This guide is for engineers running the LibrarEase backend on a local machine.
After reading it, you should be able to start the API server, worker, and scheduler with local infrastructure and run the test suite.

## Overview

The backend runs as three separate Go processes:

- API server: HTTP REST API.
- Worker: background job processor.
- Scheduler: periodic job producer.

All three processes read configuration from environment variables. In local development, use a `.env` file and run each process from the repository root.

## Prerequisites

Install:

- Go 1.25 or newer.
- PostgreSQL 16 or newer.
- Redis.
- MinIO or another S3-compatible object store.
- A Firebase service account JSON file.
- Docker, if you want to run infrastructure in containers.

Optional tools:

- Air for live reload.
- Delve for debugging.

## Environment Setup

Create a local environment file:

```bash
cp .env.example .env
```

Set at least these values:

```bash
PORT=8080
APP_ENV=development
LOG_LEVEL=DEBUG
DB_AUTO_MIGRATE=true

DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=librarease
DB_USER=postgres
DB_PASSWORD=postgres

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

MINIO_BUCKET_NAME=librarease
MINIO_BUCKET_TEMP_PATH=temp
MINIO_BUCKET_PUBLIC_PATH=public
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin

FIREBASE_SERVICE_ACCOUNT_KEY_PATH=./firebase-service-account.json

SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USERNAME=
SMTP_PASSWORD=

CLIENT_ID=local-client
SESSION_COOKIE_NAME=session
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
WORKER_CONCURRENCY=10
```

The app initializes Firebase during startup, so `FIREBASE_SERVICE_ACCOUNT_KEY_PATH` must point to a readable service account JSON file.

OpenTelemetry variables can be left empty for local development unless you are testing telemetry export.

## Start Infrastructure

You need PostgreSQL, Redis, and MinIO running before starting the backend processes.

The Makefile has `docker-run`, `docker-down`, and `docker-logs` targets for a local Docker Compose file named `docker-compose.dev.yml`. If that file exists in your checkout, use:

```bash
make docker-run
```

If you do not have that local compose file, start the services yourself. One simple Docker-based setup is:

```bash
docker run --name librarease-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=librarease \
  -p 5432:5432 \
  -d postgres:16-alpine

docker run --name librarease-redis \
  -p 6379:6379 \
  -d redis:alpine

docker run --name librarease-minio \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  -p 9000:9000 \
  -p 9001:9001 \
  -d minio/minio server /data --console-address ":9001"
```

Create the MinIO bucket named by `MINIO_BUCKET_NAME` before using upload or import flows.

## Run The Backend

Run the API server:

```bash
make run
```

Run the worker in another terminal:

```bash
make run-worker
```

Run the scheduler in a third terminal:

```bash
make run-scheduler
```

The API listens on the `PORT` from `.env`. With the example above, it is available at:

```text
http://localhost:8080
```

## Live Reload

Run the API with live reload:

```bash
make watch
```

Run the worker with live reload:

```bash
make watch-worker
```

Run the scheduler with live reload:

```bash
make watch-scheduler
```

If Air is not installed, these targets prompt to install it.

## Build

Build all backend binaries:

```bash
make build
```

Build one role:

```bash
make build-api
make build-worker
make build-scheduler
```

Generated binaries are written under `bin/`.

## Tests

Run the standard unit test target:

```bash
make test
```

Run database integration tests:

```bash
make itest
```

Integration tests expect a reachable PostgreSQL instance and the database configuration from your local environment.

## Debugging

Start Delve for the API server:

```bash
make debug
```

Start Delve for the worker:

```bash
make debug-worker
```

Start Delve for the scheduler:

```bash
make debug-scheduler
```

Default debugger ports:

- API: `2345`
- Worker: `2346`
- Scheduler: `2347`

## Common Issues

**Firebase fails during startup**

Check that `FIREBASE_SERVICE_ACCOUNT_KEY_PATH` points to an existing JSON file. The backend initializes Firebase before serving requests.

**Database tables are missing**

Set `DB_AUTO_MIGRATE=true` locally, then restart the API or worker. Migrations run during repository initialization.

**Redis connection fails**

Check `REDIS_HOST`, `REDIS_PORT`, and `REDIS_PASSWORD`. If your local Redis has no password, leave `REDIS_PASSWORD` empty.

**Object storage requests fail**

Check the MinIO endpoint, credentials, and bucket name. The bucket must exist before file upload flows work.

**`make docker-run` cannot find `docker-compose.dev.yml`**

That target expects a local development compose file. Start PostgreSQL, Redis, and MinIO manually, or create a local compose file for those services.

## Useful Commands

```bash
make build
make run
make run-worker
make run-scheduler
make watch
make watch-worker
make watch-scheduler
make test
make itest
make clean
```
