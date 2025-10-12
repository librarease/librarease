package server

import (
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Membership struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	LibraryID       string   `json:"library_id"`
	Duration        int      `json:"duration,omitempty"`
	ActiveLoanLimit int      `json:"active_loan_limit,omitempty"`
	UsageLimit      int      `json:"usage_limit,omitempty"`
	LoanPeriod      int      `json:"loan_period,omitempty"`
	FinePerDay      int      `json:"fine_per_day,omitempty"`
	Price           int      `json:"price,omitempty"`
	CreatedAt       string   `json:"created_at,omitempty"`
	UpdatedAt       string   `json:"updated_at,omitempty"`
	DeletedAt       string   `json:"deleted_at,omitempty"`
	Library         *Library `json:"library,omitempty"`
}

type ListMembershipsRequest struct {
	Skip   int    `query:"skip"`
	Limit  int    `query:"limit" validate:"required,gte=1,lte=100"`
	SortBy string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at"`
	SortIn string `query:"sort_in" validate:"omitempty,oneof=asc desc"`

	Name      string `query:"name" validate:"omitempty"`
	LibraryID string `query:"library_id" validate:"omitempty,uuid"`
}

func (s *Server) ListMemberships(ctx echo.Context) error {
	var req = ListMembershipsRequest{Limit: 20}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	var libIDs uuid.UUIDs
	if req.LibraryID != "" {
		id, _ := uuid.Parse(req.LibraryID)
		libIDs = append(libIDs, id)
	}

	memberships, total, err := s.server.ListMemberships(ctx.Request().Context(), usecase.ListMembershipsOption{
		Skip:   req.Skip,
		Limit:  req.Limit,
		SortBy: req.SortBy,
		SortIn: req.SortIn,

		Name:       req.Name,
		LibraryIDs: libIDs,
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
			UsageLimit:      mem.UsageLimit,
			LoanPeriod:      mem.LoanPeriod,
			FinePerDay:      mem.FinePerDay,
			Price:           mem.Price,
			CreatedAt:       mem.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:       mem.UpdatedAt.UTC().Format(time.RFC3339),
			DeletedAt:       d,
		}
		if mem.Library != nil {
			m.Library = &Library{
				ID:   mem.Library.ID.String(),
				Name: mem.Library.Name,
				// CreatedAt: mem.Library.CreatedAt.UTC().Format(time.RFC3339),
				// UpdatedAt: mem.Library.UpdateAt.String(),
			}
		}
		list = append(list, m)
	}

	return ctx.JSON(200, Res{
		Data: list,
		Meta: &Meta{
			Total: total,
			Skip:  req.Skip,
			Limit: req.Limit,
		},
	})
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
		UsageLimit:      mem.UsageLimit,
		LoanPeriod:      mem.LoanPeriod,
		FinePerDay:      mem.FinePerDay,
		Price:           mem.Price,
		CreatedAt:       mem.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       mem.UpdatedAt.UTC().Format(time.RFC3339),
		DeletedAt:       d,
	}
	if mem.Library != nil {
		m.Library = &Library{
			ID:        mem.Library.ID.String(),
			Name:      mem.Library.Name,
			CreatedAt: mem.Library.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: mem.Library.UpdatedAt.UTC().Format(time.RFC3339),
		}
	}
	return ctx.JSON(200, Res{Data: m})
}

type CreateMembershipRequest struct {
	Name            string `json:"name" validate:"required"`
	LibraryID       string `json:"library_id" validate:"required,uuid"`
	Duration        int    `json:"duration" validate:"required,number"`
	ActiveLoanLimit int    `json:"active_loan_limit" validate:"required,number"`
	UsageLimit      int    `json:"usage_limit" validate:"required,number"`
	LoanPeriod      int    `json:"loan_period" validate:"required,number"`
	FinePerDay      int    `json:"fine_per_day" validate:"number"`
	Price           int    `json:"price" validate:"number"`
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
		UsageLimit:      req.UsageLimit,
		LoanPeriod:      req.LoanPeriod,
		FinePerDay:      req.FinePerDay,
		Price:           req.Price,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(201, Res{Data: Membership{
		ID:              mem.ID.String(),
		Name:            mem.Name,
		LibraryID:       mem.LibraryID.String(),
		Duration:        mem.Duration,
		ActiveLoanLimit: mem.ActiveLoanLimit,
		UsageLimit:      mem.UsageLimit,
		LoanPeriod:      mem.LoanPeriod,
		FinePerDay:      mem.FinePerDay,
		Price:           mem.Price,
		CreatedAt:       mem.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       mem.UpdatedAt.UTC().Format(time.RFC3339),
	}})
}

type UpdateMembershipRequest struct {
	ID              string `param:"id" validate:"required,uuid"`
	Name            string `json:"name"`
	LibraryID       string `json:"library_id" validate:"omitempty,uuid"`
	Duration        int    `json:"duration" validate:"number"`
	ActiveLoanLimit int    `json:"active_loan_limit" validate:"number"`
	UsageLimit      int    `json:"usage_limit" validate:"number"`
	LoanPeriod      int    `json:"loan_period" validate:"number"`
	FinePerDay      int    `json:"fine_per_day" validate:"number"`
	Price           int    `json:"price" validate:"number"`
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
		UsageLimit:      req.UsageLimit,
		LoanPeriod:      req.LoanPeriod,
		FinePerDay:      req.FinePerDay,
		Price:           req.Price,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(200, Res{Data: Membership{
		ID:              mem.ID.String(),
		Name:            mem.Name,
		LibraryID:       mem.LibraryID.String(),
		Duration:        mem.Duration,
		ActiveLoanLimit: mem.ActiveLoanLimit,
		UsageLimit:      mem.UsageLimit,
		LoanPeriod:      mem.LoanPeriod,
		FinePerDay:      mem.FinePerDay,
		Price:           mem.Price,
		CreatedAt:       mem.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       mem.UpdatedAt.UTC().Format(time.RFC3339),
	}})
}
