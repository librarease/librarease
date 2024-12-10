package server

import (
	"librarease/internal/usecase"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Membership struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	LibraryID       string   `json:"library_id"`
	Duration        int      `json:"duration,omitempty"`
	ActiveLoanLimit int      `json:"active_loan_limit,omitempty"`
	LoanPeriod      int      `json:"loan_period,omitempty"`
	FinePerDay      int      `json:"fine_per_day,omitempty"`
	CreatedAt       string   `json:"created_at,omitempty"`
	UpdatedAt       string   `json:"updated_at,omitempty"`
	DeletedAt       string   `json:"deleted_at,omitempty"`
	Library         *Library `json:"library,omitempty"`
}

type ListMembershipsRequest struct {
	LibraryID string `query:"library_id" validate:"omitempty,uuid"`
	Skip      int    `query:"skip"`
	Limit     int    `query:"limit" validate:"required,gte=1,lte=100"`
}

func (s *Server) ListMemberships(ctx echo.Context) error {
	var req ListMembershipsRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	memberships, _, err := s.server.ListMemberships(ctx.Request().Context(), usecase.ListMembershipsOption{
		Skip:      req.Skip,
		Limit:     req.Limit,
		LibraryID: req.LibraryID,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}
	list := make([]Membership, 0, len(memberships))

	for _, mem := range memberships {
		var d string
		if mem.DeletedAt != nil {
			d = mem.DeletedAt.String()
		}
		m := Membership{
			ID:              mem.ID.String(),
			Name:            mem.Name,
			LibraryID:       mem.LibraryID.String(),
			Duration:        mem.Duration,
			ActiveLoanLimit: mem.ActiveLoanLimit,
			LoanPeriod:      mem.LoanPeriod,
			FinePerDay:      mem.FinePerDay,
			CreatedAt:       mem.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       mem.UpdatedAt.Format(time.RFC3339),
			DeletedAt:       d,
		}
		if mem.Library != nil {
			m.Library = &Library{
				ID:   mem.Library.ID.String(),
				Name: mem.Library.Name,
				// CreatedAt: mem.Library.CreatedAt.Format(time.RFC3339),
				// UpdatedAt: mem.Library.UpdateAt.String(),
			}
		}
		list = append(list, m)
	}
	return ctx.JSON(200, list)
}

type GetMembershipByIDRequest struct {
	ID string `param:"id" validate:"required,uuid"`
}

func (s *Server) GetMembershipByID(ctx echo.Context) error {
	var req GetMembershipByIDRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	mem, err := s.server.GetMembershipByID(ctx.Request().Context(), req.ID)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	var d string
	if mem.DeletedAt != nil {
		d = mem.DeletedAt.String()
	}
	m := Membership{
		ID:              mem.ID.String(),
		Name:            mem.Name,
		LibraryID:       mem.LibraryID.String(),
		Duration:        mem.Duration,
		ActiveLoanLimit: mem.ActiveLoanLimit,
		LoanPeriod:      mem.LoanPeriod,
		FinePerDay:      mem.FinePerDay,
		CreatedAt:       mem.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       mem.UpdatedAt.Format(time.RFC3339),
		DeletedAt:       d,
	}
	if mem.Library != nil {
		m.Library = &Library{
			ID:        mem.Library.ID.String(),
			Name:      mem.Library.Name,
			CreatedAt: mem.Library.CreatedAt.Format(time.RFC3339),
			UpdatedAt: mem.Library.UpdatedAt.Format(time.RFC3339),
		}
	}
	return ctx.JSON(200, m)
}

type CreateMembershipRequest struct {
	Name            string `json:"name" validate:"required"`
	LibraryID       string `json:"library_id" validate:"required,uuid"`
	Duration        int    `json:"duration" validate:"required,number"`
	ActiveLoanLimit int    `json:"active_loan_limit" validate:"required,number"`
	LoanPeriod      int    `json:"loan_period" validate:"required,number"`
	FinePerDay      int    `json:"fine_per_day" validate:"number"`
}

func (s *Server) CreateMembership(ctx echo.Context) error {
	var req CreateMembershipRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	uid, _ := uuid.Parse(req.LibraryID)
	mem, err := s.server.CreateMembership(ctx.Request().Context(), usecase.Membership{
		Name:            req.Name,
		LibraryID:       uid,
		Duration:        req.Duration,
		ActiveLoanLimit: req.ActiveLoanLimit,
		LoanPeriod:      req.LoanPeriod,
		FinePerDay:      req.FinePerDay,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(201, Membership{
		ID:              mem.ID.String(),
		Name:            mem.Name,
		LibraryID:       mem.LibraryID.String(),
		Duration:        mem.Duration,
		ActiveLoanLimit: mem.ActiveLoanLimit,
		LoanPeriod:      mem.LoanPeriod,
		FinePerDay:      mem.FinePerDay,
		CreatedAt:       mem.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       mem.UpdatedAt.Format(time.RFC3339),
	})
}

type UpdateMembershipRequest struct {
	ID              string `param:"id" validate:"required,uuid"`
	Name            string `json:"name"`
	LibraryID       string `json:"library_id" validate:"omitempty,uuid"`
	Duration        int    `json:"duration" validate:"number"`
	ActiveLoanLimit int    `json:"active_loan_limit" validate:"number"`
	LoanPeriod      int    `json:"loan_period" validate:"number"`
	FinePerDay      int    `json:"fine_per_day" validate:"number"`
}

func (s *Server) UpdateMembership(ctx echo.Context) error {
	var req UpdateMembershipRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.ID)
	mem, err := s.server.UpdateMembership(ctx.Request().Context(), usecase.Membership{
		ID:   id,
		Name: req.Name,
		// LibraryID:       uid,
		Duration:        req.Duration,
		ActiveLoanLimit: req.ActiveLoanLimit,
		LoanPeriod:      req.LoanPeriod,
		FinePerDay:      req.FinePerDay,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(200, Membership{
		ID:              mem.ID.String(),
		Name:            mem.Name,
		LibraryID:       mem.LibraryID.String(),
		Duration:        mem.Duration,
		ActiveLoanLimit: mem.ActiveLoanLimit,
		LoanPeriod:      mem.LoanPeriod,
		FinePerDay:      mem.FinePerDay,
		CreatedAt:       mem.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       mem.UpdatedAt.Format(time.RFC3339),
	})
}
