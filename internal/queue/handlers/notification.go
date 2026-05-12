package handlers

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/librarease/librarease/internal/telemetry"
	"github.com/librarease/librarease/internal/usecase"
	"go.opentelemetry.io/otel/attribute"
)

// HandleCheckOverdue processes the periodic overdue notification task
func (h *Handlers) HandleCheckOverdue(ctx context.Context, task *asynq.Task) (err error) {
	ctx, span := telemetry.StartSpan(ctx, "queue.handle notification:check-overdue",
		attribute.String("messaging.system", "asynq"),
		attribute.String("messaging.operation", "process"),
		attribute.String("messaging.destination.name", "notifications"),
		attribute.String("messaging.message.type", task.Type()),
	)
	defer func() { telemetry.EndSpan(span, err) }()

	h.logger.InfoContext(ctx, "processing overdue notification check")

	err = h.usecase.ProcessOverdueNotifications(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to process overdue notifications", slog.String("err", err.Error()))
		return err
	}

	h.logger.InfoContext(ctx, "completed overdue notification check")
	return nil
}

func (h *Handlers) HandleCreateNotification(ctx context.Context, task *asynq.Task) (err error) {
	ctx, span := telemetry.StartSpan(ctx, "queue.handle notification:create",
		attribute.String("messaging.system", "asynq"),
		attribute.String("messaging.operation", "process"),
		attribute.String("messaging.destination.name", "notifications"),
		attribute.String("messaging.message.type", task.Type()),
	)
	defer func() { telemetry.EndSpan(span, err) }()

	var notification usecase.Notification
	if err := json.Unmarshal(task.Payload(), &notification); err != nil {
		h.logger.ErrorContext(ctx, "failed to parse notification payload", slog.String("err", err.Error()))
		return err
	}
	span.SetAttributes(
		attribute.String("notification.user_id", notification.UserID.String()),
		attribute.String("notification.reference_type", notification.ReferenceType),
	)

	if err := h.usecase.CreateNotification(ctx, notification); err != nil {
		h.logger.ErrorContext(ctx, "failed to create notification", slog.String("user_id", notification.UserID.String()), slog.String("err", err.Error()))
		return err
	}

	return nil
}

func (h *Handlers) HandleDeliverNotification(ctx context.Context, task *asynq.Task) (err error) {
	ctx, span := telemetry.StartSpan(ctx, "queue.handle notification:deliver",
		attribute.String("messaging.system", "asynq"),
		attribute.String("messaging.operation", "process"),
		attribute.String("messaging.destination.name", "notifications"),
		attribute.String("messaging.message.type", task.Type()),
	)
	defer func() { telemetry.EndSpan(span, err) }()

	var payload struct {
		NotificationID string `json:"notification_id"`
	}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		h.logger.ErrorContext(ctx, "failed to parse notification delivery payload", slog.String("err", err.Error()))
		return err
	}

	notificationID, err := uuid.Parse(payload.NotificationID)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid notification delivery id", slog.String("notification_id", payload.NotificationID), slog.String("err", err.Error()))
		return err
	}
	span.SetAttributes(attribute.String("notification.id", notificationID.String()))

	if err := h.usecase.DeliverNotification(ctx, notificationID); err != nil {
		h.logger.ErrorContext(ctx, "failed to deliver notification", slog.String("notification_id", notificationID.String()), slog.String("err", err.Error()))
		return err
	}

	return nil
}
