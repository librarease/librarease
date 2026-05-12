package handlers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/librarease/librarease/internal/usecase"
)

// HandleCheckOverdue processes the periodic overdue notification task
func (h *Handlers) HandleCheckOverdue(ctx context.Context, task *asynq.Task) error {
	log.Println("Processing overdue notification check...")

	err := h.usecase.ProcessOverdueNotifications(ctx)
	if err != nil {
		log.Printf("Error processing overdue notifications: %v", err)
		return err
	}

	log.Println("Overdue notification check completed successfully")
	return nil
}

func (h *Handlers) HandleCreateNotification(ctx context.Context, task *asynq.Task) error {
	var notification usecase.Notification
	if err := json.Unmarshal(task.Payload(), &notification); err != nil {
		log.Printf("[Queue] Failed to parse notification payload: %v\n", err)
		return err
	}

	if err := h.usecase.CreateNotification(ctx, notification); err != nil {
		log.Printf("[Queue] Failed to create notification for user %s: %v\n", notification.UserID, err)
		return err
	}

	return nil
}

func (h *Handlers) HandleDeliverNotification(ctx context.Context, task *asynq.Task) error {
	var payload struct {
		NotificationID string `json:"notification_id"`
	}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("[Queue] Failed to parse notification delivery payload: %v\n", err)
		return err
	}

	notificationID, err := uuid.Parse(payload.NotificationID)
	if err != nil {
		log.Printf("[Queue] Invalid notification ID: %v\n", err)
		return err
	}

	if err := h.usecase.DeliverNotification(ctx, notificationID); err != nil {
		log.Printf("[Queue] Failed to deliver notification %s: %v\n", notificationID, err)
		return err
	}

	return nil
}
