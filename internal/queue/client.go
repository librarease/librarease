package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/librarease/librarease/internal/usecase"
)

const (
	TaskExportBorrowings         = "export:borrowings"
	TaskImportBooks              = "import:books"
	TaskNotificationCheckOverdue = "notification:check-overdue"
	TaskNotificationCreate       = "notification:create"
	TaskNotificationDeliver      = "notification:deliver"
	QueueDefault                 = "default"
	QueueNotifications           = "notifications"
	QueueExports                 = "exports"
	QueueImports                 = "imports"
)

// Client wraps asynq.Client for enqueuing tasks
type Client struct {
	client *asynq.Client
}

// NewClient creates a new queue client
func NewClient(redisAddr string, redisPassword string) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
	})

	return &Client{
		client: client,
	}
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.client.Close()
}

// EnqueueJob enqueues a job task to the queue
func (c *Client) EnqueueJob(ctx context.Context, jobID uuid.UUID, jobType string, payload []byte) error {
	// Create task payload
	taskPayload := map[string]any{
		"job_id":  jobID.String(),
		"type":    jobType,
		"payload": string(payload),
	}

	payloadBytes, err := json.Marshal(taskPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	// Create asynq task
	task := asynq.NewTask(jobType, payloadBytes)
	opts := []asynq.Option{asynq.Queue(QueueDefault)}
	switch jobType {
	case TaskExportBorrowings:
		opts = []asynq.Option{asynq.Queue(QueueExports)}
	case TaskImportBooks:
		opts = []asynq.Option{asynq.Queue(QueueImports)}
	}

	// Enqueue the task
	info, err := c.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	fmt.Printf("[Queue] Enqueued task: id=%s queue=%s\n", info.ID, info.Queue)
	return nil
}

// EnqueueNotification creates a durable notification asynchronously.
func (c *Client) EnqueueNotification(ctx context.Context, n usecase.Notification) error {
	payloadBytes, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("failed to marshal notification payload: %w", err)
	}

	task := asynq.NewTask(TaskNotificationCreate, payloadBytes)
	info, err := c.client.EnqueueContext(ctx, task, asynq.Queue(QueueNotifications))
	if err != nil {
		return fmt.Errorf("failed to enqueue notification create task: %w", err)
	}

	fmt.Printf("[Queue] Enqueued notification create task: id=%s queue=%s\n", info.ID, info.Queue)
	return nil
}

// EnqueueNotificationDelivery sends push delivery for an existing notification asynchronously.
func (c *Client) EnqueueNotificationDelivery(ctx context.Context, notificationID uuid.UUID) error {
	taskPayload := struct {
		NotificationID string `json:"notification_id"`
	}{
		NotificationID: notificationID.String(),
	}
	payloadBytes, err := json.Marshal(taskPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal notification delivery payload: %w", err)
	}

	task := asynq.NewTask(TaskNotificationDeliver, payloadBytes)
	info, err := c.client.EnqueueContext(ctx, task, asynq.Queue(QueueNotifications))
	if err != nil {
		return fmt.Errorf("failed to enqueue notification delivery task: %w", err)
	}

	fmt.Printf("[Queue] Enqueued notification delivery task: id=%s queue=%s\n", info.ID, info.Queue)
	return nil
}
