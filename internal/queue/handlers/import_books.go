package handlers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

func (h *Handlers) HandleImportBooks(ctx context.Context, task *asynq.Task) error {

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

	log.Printf("[Queue] Processing import:books job: %s\n", jobID)

	if err := h.usecase.ProcessImportBooksJob(ctx, jobID); err != nil {
		log.Printf("[Queue] Failed to process job %s: %v\n", jobID, err)
		return err
	}

	log.Printf("[Queue] Successfully completed job: %s\n", jobID)
	return nil
}
