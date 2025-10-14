package handlers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/librarease/librarease/internal/usecase"
)

// Handlers contains all queue task handlers
type Handlers struct {
	usecase usecase.Usecase
}

// NewHandlers creates a new handlers instance
func NewHandlers(uc usecase.Usecase) *Handlers {
	return &Handlers{
		usecase: uc,
	}
}

// TaskPayload represents the standard payload structure for all tasks
type TaskPayload struct {
	JobID   string `json:"job_id"`
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

// HandleExportBorrowings processes export borrowings tasks
// This is a thin wrapper that delegates to the usecase method
func (h *Handlers) HandleExportBorrowings(ctx context.Context, task *asynq.Task) error {
	// Parse task payload to extract job ID
	var payload TaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("[Queue] Failed to parse task payload: %v\n", err)
		return err
	}

	jobID, err := uuid.Parse(payload.JobID)
	if err != nil {
		log.Printf("[Queue] Invalid job ID: %v\n", err)
		return err
	}

	log.Printf("[Queue] Processing export:borrowings job: %s\n", jobID)

	// Delegate to usecase method - all business logic is there
	if err := h.usecase.ProcessExportBorrowingsJob(ctx, jobID); err != nil {
		log.Printf("[Queue] Failed to process job %s: %v\n", jobID, err)
		return err
	}

	log.Printf("[Queue] Successfully completed job: %s\n", jobID)
	return nil
}
