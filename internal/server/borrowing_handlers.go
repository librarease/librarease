package server

import (
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Borrowing struct {
	ID             string  `json:"id"`
	BookID         string  `json:"book_id"`
	SubscriptionID string  `json:"subscription_id"`
	StaffID        string  `json:"staff_id"`
	BorrowedAt     string  `json:"borrowed_at"`
	DueAt          string  `json:"due_at"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
	DeletedAt      *string `json:"deleted_at,omitempty"`

	Book         *Book         `json:"book"`
	Subscription *Subscription `json:"subscription"`
	Staff        *Staff        `json:"staff"`
	Returning    *Returning    `json:"returning,omitempty"`
}

type ListBorrowingsOption struct {
	Skip   int    `query:"skip"`
	Limit  int    `query:"limit" validate:"required,gte=1,lte=100"`
	SortBy string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at"`
	SortIn string `query:"sort_in" validate:"omitempty,oneof=asc desc"`

	BookID         string `query:"book_id" validate:"omitempty,uuid"`
	SubscriptionID string `query:"subscription_id" validate:"omitempty,uuid"`
	MembershipID   string `query:"membership_id" validate:"omitempty,uuid"`
	LibraryID      string `query:"library_id" validate:"omitempty,uuid"`
	UserID         string `query:"user_id" validate:"omitempty,uuid"`
	ReturningID    string `query:"returning_id" validate:"omitempty,uuid"`
	BorrowedAt     string `query:"borrowed_at" validate:"omitempty"`
	DueAt          string `query:"due_at" validate:"omitempty"`
	IsActive       bool   `query:"is_active"`
	IsExpired      bool   `query:"is_expired"`

	BorrowStaffID string  `query:"borrow_staff_id" validate:"omitempty,uuid"`
	ReturnStaffID string  `query:"return_staff_id" validate:"omitempty,uuid"`
	ReturnedAt    *string `query:"returned_at" validate:"omitempty"`
}

func (s *Server) ListBorrowings(ctx echo.Context) error {
	var req ListBorrowingsOption
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	var bookIDs uuid.UUIDs
	if req.BookID != "" {
		id, _ := uuid.Parse(req.BookID)
		bookIDs = append(bookIDs, id)
	}

	var subscriptionIDs uuid.UUIDs
	if req.SubscriptionID != "" {
		id, _ := uuid.Parse(req.SubscriptionID)
		subscriptionIDs = append(subscriptionIDs, id)
	}

	var borrowStaffIDs uuid.UUIDs
	if req.BorrowStaffID != "" {
		id, _ := uuid.Parse(req.BorrowStaffID)
		borrowStaffIDs = append(borrowStaffIDs, id)
	}

	var returnStaffIDs uuid.UUIDs
	if req.ReturnStaffID != "" {
		id, _ := uuid.Parse(req.ReturnStaffID)
		returnStaffIDs = append(returnStaffIDs, id)
	}

	var membershipIDs uuid.UUIDs
	if req.MembershipID != "" {
		id, _ := uuid.Parse(req.MembershipID)
		membershipIDs = append(membershipIDs, id)
	}

	var userIDs uuid.UUIDs
	if req.UserID != "" {
		id, _ := uuid.Parse(req.UserID)
		userIDs = append(userIDs, id)
	}

	// NOTE: libIDs.Strings() initializes a slice of strings
	// meaning non-nil but empty
	var libIDs uuid.UUIDs
	if req.LibraryID != "" {
		id, _ := uuid.Parse(req.LibraryID)
		libIDs = append(libIDs, id)
	}

	var returningIDs uuid.UUIDs
	if req.ReturningID != "" {
		id, _ := uuid.Parse(req.ReturningID)
		returningIDs = append(returningIDs, id)
	}

	var borrowedAt time.Time
	if req.BorrowedAt != "" {
		t, err := time.Parse(time.RFC3339, req.BorrowedAt)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": err.Error()})
		}
		borrowedAt = t
	}

	var dueAt time.Time
	if req.DueAt != "" {
		t, err := time.Parse(time.RFC3339, req.DueAt)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": err.Error()})
		}
		dueAt = t
	}

	var returnedAt *time.Time
	if req.ReturnedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ReturnedAt)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": err.Error()})
		}
		returnedAt = &t
	}

	borrows, total, err := s.server.ListBorrowings(ctx.Request().Context(), usecase.ListBorrowingsOption{
		Skip:            req.Skip,
		Limit:           req.Limit,
		SortBy:          req.SortBy,
		SortIn:          req.SortIn,
		BookIDs:         bookIDs,
		SubscriptionIDs: subscriptionIDs,
		BorrowStaffIDs:  borrowStaffIDs,
		ReturnStaffIDs:  returnStaffIDs,
		MembershipIDs:   membershipIDs,
		LibraryIDs:      libIDs,
		UserIDs:         userIDs,
		ReturningIDs:    returningIDs,
		BorrowedAt:      borrowedAt,
		DueAt:           dueAt,
		ReturnedAt:      returnedAt,
		IsActive:        req.IsActive,
		IsExpired:       req.IsExpired,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	list := make([]Borrowing, 0, len(borrows))

	for _, borrow := range borrows {
		var d *string
		if borrow.DeletedAt != nil {
			tmp := borrow.DeletedAt.Format(time.RFC3339)
			d = &tmp
		}
		var returning *Returning
		if borrow.Returning != nil {
			r := Returning{
				ID:         borrow.Returning.ID.String(),
				ReturnedAt: borrow.Returning.ReturnedAt,
			}
			returning = &r
		}
		m := Borrowing{
			ID:             borrow.ID.String(),
			BookID:         borrow.BookID.String(),
			SubscriptionID: borrow.SubscriptionID.String(),
			StaffID:        borrow.StaffID.String(),
			BorrowedAt:     borrow.BorrowedAt.Format(time.RFC3339),
			DueAt:          borrow.DueAt.Format(time.RFC3339),
			Returning:      returning,
			CreatedAt:      borrow.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      borrow.UpdatedAt.Format(time.RFC3339),
			DeletedAt:      d,
		}

		if borrow.Returning != nil {
			returning := Returning{
				ID:          borrow.Returning.ID.String(),
				BorrowingID: borrow.Returning.BorrowingID.String(),
				StaffID:     borrow.Returning.StaffID.String(),
				ReturnedAt:  borrow.Returning.ReturnedAt,
				Fine:        borrow.Returning.Fine,
			}
			if borrow.Returning.Staff != nil {
				staff := Staff{
					ID:   borrow.Returning.Staff.ID.String(),
					Name: borrow.Returning.Staff.Name,
				}
				returning.Staff = &staff
			}
			m.Returning = &returning
		}

		if borrow.Book != nil {
			book := Book{
				ID:    borrow.Book.ID.String(),
				Code:  borrow.Book.Code,
				Title: borrow.Book.Title,
				Cover: borrow.Book.Cover,
				// Author:    borrow.Book.Author,
				// Year:      borrow.Book.Year,
				// LibraryID: borrow.Book.LibraryID,
				// CreatedAt: borrow.Book.CreatedAt,
				// UpdatedAt: borrow.Book.UpdatedAt,
				// DeletedAt: borrow.Book.DeletedAt,
			}
			m.Book = &book
		}

		if borrow.Staff != nil {
			staff := Staff{
				ID:   borrow.Staff.ID.String(),
				Name: borrow.Staff.Name,
			}
			m.Staff = &staff
		}

		if borrow.Subscription != nil {
			sub := Subscription{
				ID:           borrow.SubscriptionID.String(),
				UserID:       borrow.Subscription.UserID.String(),
				MembershipID: borrow.Subscription.MembershipID.String(),
			}
			if borrow.Subscription.User != nil {
				sub.User = &User{
					ID:   borrow.Subscription.User.ID.String(),
					Name: borrow.Subscription.User.Name,
				}
			}
			if borrow.Subscription.Membership != nil {
				m := Membership{
					ID:        borrow.Subscription.Membership.ID.String(),
					Name:      borrow.Subscription.Membership.Name,
					LibraryID: borrow.Subscription.Membership.LibraryID.String(),
				}

				if borrow.Subscription.Membership.Library != nil {
					l := Library{
						ID:   borrow.Subscription.Membership.Library.ID.String(),
						Name: borrow.Subscription.Membership.Library.Name,
					}
					m.Library = &l
				}
				sub.Membership = &m
			}
			m.Subscription = &sub
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

type GetBorrowingByIDRequest struct {
	ID string `param:"id" validate:"required,uuid"`
}

func (s *Server) GetBorrowingByID(ctx echo.Context) error {
	var req GetBorrowingByIDRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.ID)
	borrow, err := s.server.GetBorrowingByID(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	var d *string
	if borrow.DeletedAt != nil {
		tmp := borrow.DeletedAt.Format(time.RFC3339)
		d = &tmp
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
			// DeletedAt: d,
		}
		if borrow.Returning.Staff != nil {
			staff := Staff{
				ID:   borrow.Staff.ID.String(),
				Name: borrow.Staff.Name,
			}
			r.Staff = &staff
		}
	}
	m := Borrowing{
		ID:             borrow.ID.String(),
		BookID:         borrow.BookID.String(),
		SubscriptionID: borrow.SubscriptionID.String(),
		StaffID:        borrow.StaffID.String(),
		BorrowedAt:     borrow.BorrowedAt.Format(time.RFC3339),
		DueAt:          borrow.DueAt.Format(time.RFC3339),
		Returning:      r,
		CreatedAt:      borrow.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      borrow.UpdatedAt.Format(time.RFC3339),
		DeletedAt:      d,
	}

	if borrow.Book != nil {
		book := Book{
			ID:        borrow.Book.ID.String(),
			Code:      borrow.Book.Code,
			Title:     borrow.Book.Title,
			Author:    borrow.Book.Author,
			Year:      borrow.Book.Year,
			Cover:     borrow.Book.Cover,
			LibraryID: borrow.Book.LibraryID.String(),
			CreatedAt: borrow.Book.CreatedAt.Format(time.RFC3339),
			UpdatedAt: borrow.Book.UpdatedAt.Format(time.RFC3339),
			// DeletedAt: borrow.Book.DeletedAt,
		}
		m.Book = &book
	}

	if borrow.Staff != nil {
		staff := Staff{
			ID:        borrow.Staff.ID.String(),
			Name:      borrow.Staff.Name,
			Role:      string(borrow.Staff.Role),
			UserID:    borrow.Staff.UserID.String(),
			LibraryID: borrow.Staff.LibraryID.String(),
			CreatedAt: borrow.Staff.CreatedAt.Format(time.RFC3339),
			UpdatedAt: borrow.Staff.UpdatedAt.Format(time.RFC3339),
		}
		if borrow.Staff.User != nil {
			staff.User = &User{
				ID:        borrow.Staff.User.ID.String(),
				Name:      borrow.Staff.User.Name,
				CreatedAt: borrow.Staff.User.CreatedAt.Format(time.RFC3339),
				UpdatedAt: borrow.Staff.User.UpdatedAt.Format(time.RFC3339),
			}
		}
		if borrow.Staff.Library != nil {
			staff.Library = &Library{
				ID:        borrow.Staff.Library.ID.String(),
				Name:      borrow.Staff.Library.Name,
				CreatedAt: borrow.Staff.Library.CreatedAt.Format(time.RFC3339),
				UpdatedAt: borrow.Staff.Library.UpdatedAt.Format(time.RFC3339),
			}
		}
		m.Staff = &staff
	}

	if borrow.Subscription != nil {
		sub := Subscription{
			ID:              borrow.SubscriptionID.String(),
			UserID:          borrow.Subscription.UserID.String(),
			MembershipID:    borrow.Subscription.MembershipID.String(),
			CreatedAt:       borrow.Subscription.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       borrow.Subscription.UpdatedAt.Format(time.RFC3339),
			ExpiresAt:       borrow.Subscription.ExpiresAt.Format(time.RFC3339),
			FinePerDay:      borrow.Subscription.FinePerDay,
			LoanPeriod:      borrow.Subscription.LoanPeriod,
			ActiveLoanLimit: borrow.Subscription.ActiveLoanLimit,
			UsageLimit:      borrow.Subscription.UsageLimit,
		}
		if borrow.Subscription.User != nil {
			sub.User = &User{
				ID:        borrow.Subscription.User.ID.String(),
				Name:      borrow.Subscription.User.Name,
				CreatedAt: borrow.Subscription.User.CreatedAt.Format(time.RFC3339),
				UpdatedAt: borrow.Subscription.User.UpdatedAt.Format(time.RFC3339),
			}
		}
		if borrow.Subscription.Membership != nil {
			m := Membership{
				ID:              borrow.Subscription.Membership.ID.String(),
				Name:            borrow.Subscription.Membership.Name,
				LibraryID:       borrow.Subscription.Membership.LibraryID.String(),
				Duration:        borrow.Subscription.Membership.Duration,
				ActiveLoanLimit: borrow.Subscription.Membership.ActiveLoanLimit,
				LoanPeriod:      borrow.Subscription.Membership.LoanPeriod,
				FinePerDay:      borrow.Subscription.Membership.FinePerDay,
				CreatedAt:       borrow.Subscription.Membership.CreatedAt.Format(time.RFC3339),
				UpdatedAt:       borrow.Subscription.Membership.UpdatedAt.Format(time.RFC3339),
			}

			if borrow.Subscription.Membership.Library != nil {
				m.Library = &Library{
					ID:        borrow.Subscription.Membership.Library.ID.String(),
					Name:      borrow.Subscription.Membership.Library.Name,
					CreatedAt: borrow.Subscription.Membership.Library.CreatedAt.Format(time.RFC3339),
					UpdatedAt: borrow.Subscription.Membership.Library.UpdatedAt.Format(time.RFC3339),
				}
			}
			sub.Membership = &m
		}
		m.Subscription = &sub
	}

	return ctx.JSON(200, Res{Data: m})
}

type CreateBorrowingRequest struct {
	BookID         string `json:"book_id" validate:"required,uuid"`
	SubscriptionID string `json:"subscription_id" validate:"required,uuid"`
	StaffID        string `json:"staff_id" validate:"required,uuid"`
	BorrowedAt     string `json:"borrowed_at" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	DueAt          string `json:"due_at" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

func (s *Server) CreateBorrowing(ctx echo.Context) error {
	var req CreateBorrowingRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	bookID, _ := uuid.Parse(req.BookID)
	subscriptionID, _ := uuid.Parse(req.SubscriptionID)
	staffID, _ := uuid.Parse(req.StaffID)

	var borrowedAt time.Time
	if req.BorrowedAt != "" {
		t, err := time.Parse(time.RFC3339, req.BorrowedAt)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": err.Error()})
		}
		borrowedAt = t
	}

	var dueAt time.Time
	if req.DueAt != "" {
		t, err := time.Parse(time.RFC3339, req.DueAt)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": err.Error()})
		}
		dueAt = t
	}

	borrow, err := s.server.CreateBorrowing(ctx.Request().Context(), usecase.Borrowing{
		BookID:         bookID,
		SubscriptionID: subscriptionID,
		StaffID:        staffID,
		BorrowedAt:     borrowedAt,
		DueAt:          dueAt,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(201, Res{Data: Borrowing{
		ID:             borrow.ID.String(),
		BookID:         borrow.BookID.String(),
		SubscriptionID: borrow.SubscriptionID.String(),
		StaffID:        borrow.StaffID.String(),
		BorrowedAt:     borrow.BorrowedAt.Format(time.RFC3339),
		DueAt:          borrow.DueAt.Format(time.RFC3339),
		CreatedAt:      borrow.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      borrow.UpdatedAt.Format(time.RFC3339),
	}})
}

type UpdateBorrowingRequest struct {
	ID             string     `param:"id" validate:"required,uuid"`
	BookID         string     `json:"book_id" validate:"omitempty,uuid"`
	SubscriptionID string     `json:"subscription_id" validate:"omitempty,uuid"`
	StaffID        string     `json:"staff_id" validate:"omitempty,uuid"`
	BorrowedAt     string     `json:"borrowed_at" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	DueAt          string     `json:"due_at" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	ReturningID    *string    `json:"returning_id" validate:"omitempty,uuid"`
	Returning      *Returning `json:"returning,omitempty"`
}

func (s *Server) UpdateBorrowing(ctx echo.Context) error {
	var req UpdateBorrowingRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.ID)
	bookID, _ := uuid.Parse(req.BookID)
	subscriptionID, _ := uuid.Parse(req.SubscriptionID)
	staffID, _ := uuid.Parse(req.StaffID)

	var borrowedAt time.Time
	if req.BorrowedAt != "" {
		t, err := time.Parse(time.RFC3339, req.BorrowedAt)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": err.Error()})
		}
		borrowedAt = t
	}

	var dueAt time.Time
	if req.DueAt != "" {
		t, err := time.Parse(time.RFC3339, req.DueAt)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": err.Error()})
		}
		dueAt = t
	}

	var r *usecase.Returning
	if req.Returning != nil {
		r = &usecase.Returning{
			ReturnedAt: req.Returning.ReturnedAt,
			Fine:       req.Returning.Fine,
		}
	}

	borrow, err := s.server.UpdateBorrowing(ctx.Request().Context(), usecase.Borrowing{
		ID:             id,
		BookID:         bookID,
		SubscriptionID: subscriptionID,
		StaffID:        staffID,
		BorrowedAt:     borrowedAt,
		DueAt:          dueAt,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	if r != nil {
		if err := s.server.UpdateReturn(ctx.Request().Context(), id, usecase.Returning{
			ReturnedAt: r.ReturnedAt,
			Fine:       r.Fine,
		}); err != nil {
			return ctx.JSON(500, map[string]string{"error": err.Error()})
		}
	}

	return ctx.JSON(200, Res{Data: Borrowing{
		ID:             borrow.ID.String(),
		BookID:         borrow.BookID.String(),
		SubscriptionID: borrow.SubscriptionID.String(),
		StaffID:        borrow.StaffID.String(),
		BorrowedAt:     borrow.BorrowedAt.Format(time.RFC3339),
		DueAt:          borrow.DueAt.Format(time.RFC3339),
		CreatedAt:      borrow.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      borrow.UpdatedAt.Format(time.RFC3339),
	}})
}
