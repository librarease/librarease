package handlers

import (
	"log/slog"

	"github.com/librarease/librarease/internal/usecase"
)

type Handlers struct {
	usecase usecase.Usecase
	logger  *slog.Logger
}

func NewHandlers(uc usecase.Usecase, logger *slog.Logger) *Handlers {
	return &Handlers{
		usecase: uc,
		logger:  logger.With(slog.String("component", "queue.handlers")),
	}
}

type TaskPayload struct {
	JobID   string `json:"job_id"`
	Type    string `json:"type"`
	Payload string `json:"payload"`
}
