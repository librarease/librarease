package server

import (
	"net/http"
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Lost struct {
	ID          string    `json:"id"`
	BorrowingID string    `json:"borrowing_id"`
	StaffID     string    `json:"staff_id"`
	ReportedAt  time.Time `json:"reported_at"`
	Fine        int       `json:"fine"`
	Note        string    `json:"note"`
	CreatedAt   string    `json:"created_at,omitempty"`
	UpdatedAt   string    `json:"updated_at,omitempty"`
	DeletedAt   *string   `json:"deleted_at,omitempty"`

	Borrowing *Borrowing `json:"borrowing,omitempty"`
	Staff     *Staff     `json:"staff,omitempty"`
}

type LostBorrowingRequest struct {
	BorrowingID string     `param:"id" validate:"required,uuid"`
	StaffID     string     `json:"staff_id" validate:"omitempty,uuid"`
	ReportedAt  *time.Time `json:"reported_at" validate:"omitempty"`
	Fine        int        `json:"fine"`
	Note        string     `json:"note" validate:"required"`
}

func (s *Server) LostBorrowing(ctx echo.Context) error {
	var req LostBorrowingRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	borrowingID, _ := uuid.Parse(req.BorrowingID)
	staffID, _ := uuid.Parse(req.StaffID)

	// default to now if not provided
	var reportedAt = time.Now()
	if req.ReportedAt != nil {
		reportedAt = *req.ReportedAt
	}

	l, err := s.server.LostBorrowing(ctx.Request().Context(), borrowingID, usecase.Lost{
		StaffID:    staffID,
		ReportedAt: reportedAt,
		Fine:       req.Fine,
		Note:       req.Note,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(200, Res{Data: Lost{
		ID:         l.ID.String(),
		StaffID:    l.StaffID.String(),
		ReportedAt: l.ReportedAt,
		CreatedAt:  l.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  l.UpdatedAt.Format(time.RFC3339),
	}})
}

type DeleteLostRequest struct {
	BorrowingID string `param:"id" validate:"required,uuid"`
}

func (s *Server) DeleteLost(ctx echo.Context) error {
	var req DeleteLostRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}
	borrowingID, _ := uuid.Parse(req.BorrowingID)

	if err := s.server.DeleteLost(ctx.Request().Context(), borrowingID); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, Res{Message: "successfully deleted lost"})
}
