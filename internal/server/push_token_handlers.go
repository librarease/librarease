package server

import (
	"github.com/labstack/echo/v4"
	"github.com/librarease/librarease/internal/usecase"
)

type PushToken struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Token     string  `json:"token"`
	Provider  string  `json:"provider"`
	LastSeen  string  `json:"last_seen,omitempty"`
	CreatedAt string  `json:"created_at,omitempty"`
	UpdatedAt string  `json:"updated_at,omitempty"`
	DeletedAt *string `json:"deleted_at,omitempty"`
}

type SavePushTokenRequest struct {
	Token    string `json:"token" validate:"required"`
	Provider string `json:"provider" validate:"required,oneof=fcm apns webpush"`
}

func (s *Server) SavePushToken(ctx echo.Context) error {
	var req SavePushTokenRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	provider, _ := usecase.ParsePushProvider(req.Provider)

	err := s.server.SavePushToken(ctx.Request().Context(), req.Token, provider)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.NoContent(204)
}
