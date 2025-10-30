package handlers

import "github.com/librarease/librarease/internal/usecase"

type Handlers struct {
	usecase usecase.Usecase
}

func NewHandlers(uc usecase.Usecase) *Handlers {
	return &Handlers{
		usecase: uc,
	}
}

type TaskPayload struct {
	JobID   string `json:"job_id"`
	Type    string `json:"type"`
	Payload string `json:"payload"`
}
