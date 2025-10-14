package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
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

	// Enqueue the task
	info, err := c.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	fmt.Printf("[Queue] Enqueued task: id=%s queue=%s\n", info.ID, info.Queue)
	return nil
}
