package handlers

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/librarease/librarease/internal/telemetry"
	"go.opentelemetry.io/otel/attribute"
)

func (h *Handlers) HandleImportBooks(ctx context.Context, task *asynq.Task) (err error) {
	ctx, span := telemetry.StartSpan(ctx, "queue.handle import:books",
		attribute.String("messaging.system", "asynq"),
		attribute.String("messaging.operation", "process"),
		attribute.String("messaging.destination.name", "imports"),
		attribute.String("messaging.message.type", task.Type()),
	)
	defer func() { telemetry.EndSpan(span, err) }()

	var payload TaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		h.logger.ErrorContext(ctx, "failed to parse import books task payload", slog.String("err", err.Error()))
		return err
	}
	span.SetAttributes(attribute.String("job.type", payload.Type))

	jobID, err := uuid.Parse(payload.JobID)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid import books job id", slog.String("job_id", payload.JobID), slog.String("err", err.Error()))
		return err
	}
	span.SetAttributes(attribute.String("job.id", jobID.String()))

	h.logger.InfoContext(ctx, "processing import books job", slog.String("job_id", jobID.String()))

	if err := h.usecase.ProcessImportBooksJob(ctx, jobID); err != nil {
		h.logger.ErrorContext(ctx, "failed to process import books job", slog.String("job_id", jobID.String()), slog.String("err", err.Error()))
		return err
	}

	h.logger.InfoContext(ctx, "completed import books job", slog.String("job_id", jobID.String()))
	return nil
}
