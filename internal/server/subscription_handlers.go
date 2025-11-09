package server

import (
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Subscription struct {
	ID           string      `json:"id"`
	UserID       string      `json:"user_id"`
	MembershipID string      `json:"membership_id"`
	CreatedAt    string      `json:"created_at,omitempty"`
	UpdatedAt    string      `json:"updated_at,omitempty"`
	DeletedAt    *string     `json:"deleted_at,omitempty"`
	User         *User       `json:"user,omitempty"`
	Membership   *Membership `json:"membership,omitempty"`

	// Granfathering the membership
	ExpiresAt       string `json:"expires_at,omitempty"`
	Amount          int    `json:"amount,omitempty"`
	FinePerDay      int    `json:"fine_per_day,omitempty"`
	LoanPeriod      int    `json:"loan_period,omitempty"`
	ActiveLoanLimit int    `json:"active_loan_limit,omitempty"`
	UsageLimit      int    `json:"usage_limit,omitempty"`

	UsageCount      *int `json:"usage_count,omitempty"`
	ActiveLoanCount *int `json:"active_loan_count,omitempty"`
}

type ListSubscriptionsRequest struct {
	Skip   int    `query:"skip"`
	Limit  int    `query:"limit"`
	SortBy string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at name"`
	SortIn string `query:"sort_in" validate:"omitempty,oneof=asc desc"`

	ID             string `query:"id" validate:"omitempty"`
	UserID         string `query:"user_id" validate:"omitempty,uuid"`
	MembershipID   string `query:"membership_id" validate:"omitempty,uuid"`
	LibraryID      string `query:"library_id" validate:"omitempty,uuid"`
	MembershipName string `query:"membership_name" validate:"omitempty"`
	IsActive       bool   `query:"is_active"`
	IsExpired      bool   `query:"is_expired"`
}

func (s *Server) ListSubscriptions(ctx echo.Context) error {
	var req ListSubscriptionsRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	// NOTE: since Echo's default binding doesn't support binding slice of UUIDs
	// we only accept single string in the query parameter
	var libIDs uuid.UUIDs
	if req.LibraryID != "" {
		id, _ := uuid.Parse(req.LibraryID)
		libIDs = append(libIDs, id)
	}

	subs, total, err := s.server.ListSubscriptions(ctx.Request().Context(), usecase.ListSubscriptionsOption{
		ID:             req.ID,
		Skip:           req.Skip,
		Limit:          req.Limit,
		SortBy:         req.SortBy,
		SortIn:         req.SortIn,
		UserID:         req.UserID,
		MembershipID:   req.MembershipID,
		LibraryIDs:     libIDs,
		MembershipName: req.MembershipName,
		IsActive:       req.IsActive,
		IsExpired:      req.IsExpired,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}
	list := make([]Subscription, 0, len(subs))

	for _, sub := range subs {
		var d *string
		if sub.DeletedAt != nil {
			tmp := sub.DeletedAt.String()
			d = &tmp
		}
		m := Subscription{
			ID:              sub.ID.String(),
			UserID:          sub.UserID.String(),
			MembershipID:    sub.MembershipID.String(),
			CreatedAt:       sub.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:       sub.UpdatedAt.UTC().Format(time.RFC3339),
			DeletedAt:       d,
			ExpiresAt:       sub.ExpiresAt.UTC().Format(time.RFC3339),
			Amount:          sub.Amount,
			FinePerDay:      sub.FinePerDay,
			LoanPeriod:      sub.LoanPeriod,
			ActiveLoanLimit: sub.ActiveLoanLimit,
			UsageLimit:      sub.UsageLimit,
		}
		if sub.User != nil {
			m.User = &User{
				ID:   sub.User.ID.String(),
				Name: sub.User.Name,
			}
		}
		if sub.Membership != nil {
			m.Membership = &Membership{
				ID:        sub.Membership.ID.String(),
				Name:      sub.Membership.Name,
				LibraryID: sub.Membership.LibraryID.String(),
			}

			if lib := sub.Membership.Library; lib != nil {
				m.Membership.Library = &Library{
					ID:   lib.ID.String(),
					Name: lib.Name,
				}
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

type GetSubscriptionByIDRequest struct {
	ID string `param:"id" validate:"required,uuid"`
}

func (s *Server) GetSubscriptionByID(ctx echo.Context) error {
	var req GetSubscriptionByIDRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.ID)

	sub, err := s.server.GetSubscriptionByID(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	var d *string
	if sub.DeletedAt != nil {
		tmp := sub.DeletedAt.String()
		d = &tmp
	}
	m := Subscription{
		ID:              sub.ID.String(),
		UserID:          sub.UserID.String(),
		MembershipID:    sub.MembershipID.String(),
		CreatedAt:       sub.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       sub.UpdatedAt.UTC().Format(time.RFC3339),
		DeletedAt:       d,
		ExpiresAt:       sub.ExpiresAt.UTC().Format(time.RFC3339),
		Amount:          sub.Amount,
		FinePerDay:      sub.FinePerDay,
		LoanPeriod:      sub.LoanPeriod,
		ActiveLoanLimit: sub.ActiveLoanLimit,
		UsageLimit:      sub.UsageLimit,
		UsageCount:      sub.UsageCount,
		ActiveLoanCount: sub.ActiveLoanCount,
	}
	if sub.User != nil {
		m.User = &User{
			ID:   sub.User.ID.String(),
			Name: sub.User.Name,
		}
	}
	if sub.Membership != nil {
		m.Membership = &Membership{
			ID:              sub.Membership.ID.String(),
			Name:            sub.Membership.Name,
			LibraryID:       sub.Membership.LibraryID.String(),
			Duration:        sub.Membership.Duration,
			ActiveLoanLimit: sub.Membership.ActiveLoanLimit,
			UsageLimit:      sub.Membership.UsageLimit,
			LoanPeriod:      sub.Membership.LoanPeriod,
			FinePerDay:      sub.Membership.FinePerDay,
			Price:           sub.Membership.Price,
			CreatedAt:       sub.Membership.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:       sub.Membership.UpdatedAt.UTC().Format(time.RFC3339),
		}

		if lib := sub.Membership.Library; lib != nil {
			m.Membership.Library = &Library{
				ID:   lib.ID.String(),
				Name: lib.Name,
			}
		}
	}

	return ctx.JSON(200, Res{Data: m})
}

type CreateSubscriptionRequest struct {
	UserID       string `json:"user_id" validate:"required,uuid"`
	MembershipID string `json:"membership_id" validate:"required,uuid"`
}

func (s *Server) CreateSubscription(ctx echo.Context) error {
	var req CreateSubscriptionRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	userID, _ := uuid.Parse(req.UserID)
	membershipID, _ := uuid.Parse(req.MembershipID)

	id, err := s.server.CreateSubscription(ctx.Request().Context(), usecase.Subscription{
		UserID:       userID,
		MembershipID: membershipID,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	// FIXME: return the created subscription
	return ctx.JSON(200, Res{Data: id})
}

type UpdateSubscriptionRequest struct {
	ID              string `param:"id" validate:"required,uuid"`
	UserID          string `json:"user_id" validate:"omitempty,uuid"`
	MembershipID    string `json:"membership_id" validate:"omitempty,uuid"`
	ExpiresAt       string `json:"expires_at" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Amount          int    `json:"amount" validate:"omitempty,number"`
	FinePerDay      int    `json:"fine_per_day" validate:"omitempty,number"`
	LoanPeriod      int    `json:"loan_period" validate:"omitempty,number"`
	ActiveLoanLimit int    `json:"active_loan_limit" validate:"omitempty,number"`
	UsageLimit      int    `json:"usage_limit" validate:"omitempty,number"`
}

func (s *Server) UpdateSubscription(ctx echo.Context) error {
	var req UpdateSubscriptionRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.ID)
	userID, _ := uuid.Parse(req.UserID)
	membershipID, _ := uuid.Parse(req.MembershipID)

	var (
		exp time.Time
		err error
	)
	if req.ExpiresAt != "" {
		exp, err = time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			return ctx.JSON(422, map[string]string{"error": "invalid expires_at"})
		}
	}

	sub, err := s.server.UpdateSubscription(ctx.Request().Context(), usecase.Subscription{
		ID:              id,
		UserID:          userID,
		MembershipID:    membershipID,
		ExpiresAt:       exp,
		Amount:          req.Amount,
		FinePerDay:      req.FinePerDay,
		LoanPeriod:      req.LoanPeriod,
		ActiveLoanLimit: req.ActiveLoanLimit,
		UsageLimit:      req.UsageLimit,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(200, Res{Data: Subscription{
		ID:              sub.ID.String(),
		UserID:          sub.UserID.String(),
		MembershipID:    sub.MembershipID.String(),
		CreatedAt:       sub.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       sub.UpdatedAt.UTC().Format(time.RFC3339),
		ExpiresAt:       sub.ExpiresAt.UTC().Format(time.RFC3339),
		Amount:          sub.Amount,
		FinePerDay:      sub.FinePerDay,
		LoanPeriod:      sub.LoanPeriod,
		ActiveLoanLimit: sub.ActiveLoanLimit,
		UsageLimit:      sub.UsageLimit,
	}})
}

type DeleteSubscriptionRequest struct {
	ID string `param:"id" validate:"required,uuid"`
}

func (s *Server) DeleteSubscription(ctx echo.Context) error {
	var req DeleteSubscriptionRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.ID)

	err := s.server.DeleteSubscription(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.NoContent(204)
}
