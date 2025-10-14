# Background Jobs with Asynq

## Overview

Async job processing using Asynq (Redis-based queue).

**Pattern**: Thin handlers delegate to usecase methods.

```
API → CreateJob → Redis Queue → Worker → Usecase Method
```

## Structure

```
internal/
├── queue/
│   ├── client.go           # Enqueue jobs
│   ├── server.go           # Worker setup
│   └── handlers/           # Thin wrappers (~10 lines)
│       └── export_borrowings.go
├── usecase/
│   └── borrowing.go        # ProcessExportBorrowingsJob()
cmd/
└── worker/main.go          # Worker entry point
```

## Flow

### 1. Create Job (API)
```go
// Handler
job := usecase.CreateJob(ctx, usecase.Job{
    Type:    "export:borrowings",
    Payload: payloadBytes,
})
// Job enqueued to Redis, returns 202 immediately
```

### 2. Process Job (Worker)
```go
// Handler (thin)
func HandleExportBorrowings(ctx, task) error {
    var p TaskPayload
    json.Unmarshal(task.Payload(), &p)
    return usecase.ProcessExportBorrowingsJob(ctx, p.JobID)
}

// Usecase (business logic)
func ProcessExportBorrowings(ctx, jobID) error {
    job := repo.GetJobByID(jobID)
    job.Status = "PROCESSING"
    repo.UpdateJob(job)
    
    // Do work: query data, generate file, upload
    data := repo.ListBorrowings(filters)
    file := generateCSV(data)
    url := storage.Upload(file)
    
    job.Status = "COMPLETED"
    job.Result = url
    repo.UpdateJob(job)
    
    notifications.Send(...)
    return nil
}
```

### 3. Check Status
```
GET /api/v1/jobs/:id
```

## Adding New Jobs

1. Add usecase method: `ProcessXJob(ctx, jobID)`
2. Add handler: `HandleX(ctx, task) → usecase.ProcessXJob()`
3. Register: `srv.RegisterHandler("x:job", h.HandleX)`

## Running

```bash
make docker-run  # Start Redis
make run         # API
make run-worker  # Worker
```

## Job States

- `PENDING` → `PROCESSING` → `COMPLETED` / `FAILED`
