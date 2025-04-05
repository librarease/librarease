package server

import (
	"net/http"
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Returning struct {
	ID          string    `json:"id"`
	BorrowingID string    `json:"borrowing_id"`
	StaffID     string    `json:"staff_id"`
	ReturnedAt  time.Time `json:"returned_at"`
	Fine        int       `json:"fine"`
	CreatedAt   string    `json:"created_at,omitempty"`
	UpdatedAt   string    `json:"updated_at,omitempty"`
	DeletedAt   *string   `json:"deleted_at,omitempty"`

	Borrowing *Borrowing `json:"borrowing,omitempty"`
	Staff     *Staff     `json:"staff,omitempty"`
}

type ReturnBorrowingRequest struct {
	BorrowingID string     `param:"id" validate:"required,uuid"`
	StaffID     string     `json:"staff_id" validate:"omitempty,uuid"`
	ReturnedAt  *time.Time `json:"returned_at" validate:"omitempty"`
	Fine        *int       `json:"fine" validate:"omitempty"`
}

func (s *Server) ReturnBorrowing(ctx echo.Context) error {
	var req ReturnBorrowingRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	borrowingID, _ := uuid.Parse(req.BorrowingID)
	staffID, _ := uuid.Parse(req.StaffID)
	// NOTE: usecase will calculate the fine
	// based on due date if fine is negative
	var fine = -1
	if req.Fine != nil {
		// FIXME: to be implemented in validator
		if *req.Fine < 0 {
			return ctx.JSON(400, map[string]string{"error": "fine must be positive"})
		}
		fine = *req.Fine
	}

	// default to now if not provided
	var returnedAt = time.Now()
	if req.ReturnedAt != nil {
		returnedAt = *req.ReturnedAt
	}

	borrow, err := s.server.ReturnBorrowing(ctx.Request().Context(), borrowingID, usecase.Returning{
		StaffID:    staffID,
		ReturnedAt: returnedAt,
		Fine:       fine,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	var r *Returning
	if borrow.Returning != nil {
		r = &Returning{
			ID:          borrow.Returning.ID.String(),
			BorrowingID: borrow.Returning.BorrowingID.String(),
			StaffID:     borrow.Returning.StaffID.String(),
			ReturnedAt:  borrow.Returning.ReturnedAt,
			Fine:        borrow.Returning.Fine,
			CreatedAt:   borrow.Returning.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   borrow.Returning.UpdatedAt.Format(time.RFC3339),
			// DeletedAt:   borrow.Returning.DeletedAt,
		}
	}
	return ctx.JSON(200, Res{Data: Borrowing{
		ID:             borrow.ID.String(),
		BookID:         borrow.BookID.String(),
		SubscriptionID: borrow.SubscriptionID.String(),
		StaffID:        borrow.StaffID.String(),
		BorrowedAt:     borrow.BorrowedAt.Format(time.RFC3339),
		DueAt:          borrow.DueAt.Format(time.RFC3339),
		Returning:      r,
		CreatedAt:      borrow.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      borrow.UpdatedAt.Format(time.RFC3339),
	}})
}

type DeleteReturnRequest struct {
	BorrowingID string `param:"id" validate:"required,uuid"`
}

func (s *Server) DeleteReturn(ctx echo.Context) error {
	var req DeleteReturnRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}
	borrowingID, _ := uuid.Parse(req.BorrowingID)

	if err := s.server.DeleteReturn(ctx.Request().Context(), borrowingID); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusNoContent, nil)

}
