# Architecture

## Structure

```
cmd/
├── api/main.go      → internal/server/server.go
└── worker/main.go   → internal/queue/server.go

internal/
├── server/          # HTTP handlers & setup
├── queue/           # Job queue client & worker
├── usecase/         # Business logic (shared)
├── database/        # Data access (shared)
└── config/          # Configuration
```

## Data Flow

```
HTTP Request              Background Job
     │                         │
     ↓                         ↓
┌─────────┐              ┌─────────┐
│   API   │──Enqueue──→  │  Worker │
└─────────┘              └─────────┘
     │                         │
     └──────────┬──────────────┘
                ↓
         ┌──────────────┐
         │   Usecase    │
         └──────────────┘
                ↓
         ┌──────────────┐
         │  Database    │
         └──────────────┘
```

## Deployment

```
Dockerfile → builds 2 binaries → deploys as 2 services

cmd/api    → bin/api    → api-service (./api)
cmd/worker → bin/worker → worker-service (./worker)
```

## Key Principles

- Entry points handle lifecycle only
- All business logic in `internal/usecase`
- Dependencies point inward
- Scale services independently
