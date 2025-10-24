package handlers

import (
	"context"
	"log"

	"github.com/hibiken/asynq"
)

// HandleCheckOverdue processes the periodic overdue notification task
func (h *Handlers) HandleCheckOverdue(ctx context.Context, task *asynq.Task) error {
	log.Println("Processing overdue notification check...")

	// Call the usecase method to process overdue notifications
	err := h.usecase.ProcessOverdueNotifications(ctx)
	if err != nil {
		log.Printf("Error processing overdue notifications: %v", err)
		return err
	}

	log.Println("Overdue notification check completed successfully")
	return nil
}
