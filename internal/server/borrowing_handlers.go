package server

import (
	"errors"
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
	Lost         *Lost         `json:"lost,omitempty"`

	PrevID *string `json:"prev_id,omitempty"`
	NextID *string `json:"next_id,omitempty"`
}

type BorrowingsOption struct {
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
	IsOverdue      bool   `query:"is_overdue"`
	IsReturned     bool   `query:"is_returned"`
	IsLost         bool   `query:"is_lost"`

	BorrowStaffID string  `query:"borrow_staff_id" validate:"omitempty,uuid"`
	ReturnStaffID string  `query:"return_staff_id" validate:"omitempty,uuid"`
	ReturnedAt    *string `query:"returned_at" validate:"omitempty"`
	LostAt        *string `query:"lost_at" validate:"omitempty"`
}

type ListBorrowingsOption struct {
	Skip  int `query:"skip"`
	Limit int `query:"limit" validate:"required,gte=1,lte=100"`
	BorrowingsOption
}

func (s *Server) ListBorrowings(ctx echo.Context) error {
	var req = ListBorrowingsOption{Limit: 20}
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

	var lostAt *time.Time
	if req.LostAt != nil {
		t, err := time.Parse(time.RFC3339, *req.LostAt)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": err.Error()})
		}
		lostAt = &t
	}

	borrows, total, err := s.server.ListBorrowings(ctx.Request().Context(), usecase.ListBorrowingsOption{
		Skip:  req.Skip,
		Limit: req.Limit,
		BorrowingsOption: usecase.BorrowingsOption{
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
			LostAt:          lostAt,
			IsActive:        req.IsActive,
			IsOverdue:       req.IsOverdue,
			IsReturned:      req.IsReturned,
			IsLost:          req.IsLost,
		},
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	list := make([]Borrowing, 0, len(borrows))

	for _, borrow := range borrows {
		var d *string
		if borrow.DeletedAt != nil {
			tmp := borrow.DeletedAt.UTC().Format(time.RFC3339)
			d = &tmp
		}
		m := Borrowing{
			ID:             borrow.ID.String(),
			BookID:         borrow.BookID.String(),
			SubscriptionID: borrow.SubscriptionID.String(),
			StaffID:        borrow.StaffID.String(),
			BorrowedAt:     borrow.BorrowedAt.UTC().Format(time.RFC3339),
			DueAt:          borrow.DueAt.UTC().Format(time.RFC3339),
			CreatedAt:      borrow.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:      borrow.UpdatedAt.UTC().Format(time.RFC3339),
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

		if borrow.Lost != nil {
			l := Lost{
				ID:          borrow.Lost.ID.String(),
				ReportedAt:  borrow.Lost.ReportedAt,
				BorrowingID: borrow.Lost.BorrowingID.String(),
				StaffID:     borrow.Lost.StaffID.String(),
				Fine:        borrow.Lost.Fine,
				Note:        borrow.Lost.Note,
			}
			if borrow.Lost.Staff != nil {
				staff := Staff{
					ID:   borrow.Lost.Staff.ID.String(),
					Name: borrow.Lost.Staff.Name,
				}
				l.Staff = &staff
			}
			m.Lost = &l
		}

		if borrow.Book != nil {
			book := Book{
				ID:     borrow.Book.ID.String(),
				Code:   borrow.Book.Code,
				Title:  borrow.Book.Title,
				Cover:  borrow.Book.Cover,
				Colors: borrow.Book.Colors,
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
	BorrowingsOption
}

func (s *Server) GetBorrowingByID(ctx echo.Context) error {
	var req GetBorrowingByIDRequest
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

	var lostAt *time.Time
	if req.LostAt != nil {
		t, err := time.Parse(time.RFC3339, *req.LostAt)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": err.Error()})
		}
		lostAt = &t
	}

	id, _ := uuid.Parse(req.ID)
	borrow, err := s.server.GetBorrowingByID(ctx.Request().Context(), id, usecase.BorrowingsOption{
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
		LostAt:          lostAt,
		IsActive:        req.IsActive,
		IsOverdue:       req.IsOverdue,
		IsReturned:      req.IsReturned,
		IsLost:          req.IsLost,
	})
	if err != nil {
		var notFoundErr usecase.ErrNotFound
		if errors.As(err, &notFoundErr) {
			return ctx.JSON(404, map[string]any{
				"error":   notFoundErr.Error(),
				"code":    notFoundErr.Code,
				"message": notFoundErr.Message,
			})
		}

		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	var d *string
	if borrow.DeletedAt != nil {
		tmp := borrow.DeletedAt.UTC().Format(time.RFC3339)
		d = &tmp
	}
	var prevID *string
	if borrow.PrevID != nil {
		s := borrow.PrevID.String()
		prevID = &s
	}
	var nextID *string
	if borrow.NextID != nil {
		s := borrow.NextID.String()
		nextID = &s
	}
	var r *Returning
	if borrow.Returning != nil {
		r = &Returning{
			ID:          borrow.Returning.ID.String(),
			BorrowingID: borrow.Returning.BorrowingID.String(),
			StaffID:     borrow.Returning.StaffID.String(),
			ReturnedAt:  borrow.Returning.ReturnedAt,
			Fine:        borrow.Returning.Fine,
			CreatedAt:   borrow.Returning.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:   borrow.Returning.UpdatedAt.UTC().Format(time.RFC3339),
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
	var l *Lost
	if borrow.Lost != nil {
		l = &Lost{
			ID:          borrow.Lost.ID.String(),
			BorrowingID: borrow.Lost.BorrowingID.String(),
			StaffID:     borrow.Lost.StaffID.String(),
			ReportedAt:  borrow.Lost.ReportedAt,
			Fine:        borrow.Lost.Fine,
			Note:        borrow.Lost.Note,
			CreatedAt:   borrow.Lost.CreatedAt.UTC().String(),
			UpdatedAt:   borrow.Lost.UpdatedAt.UTC().String(),
			// DeletedAt: d,
		}
		if borrow.Lost.Staff != nil {
			staff := Staff{
				ID:   borrow.Staff.ID.String(),
				Name: borrow.Staff.Name,
			}
			l.Staff = &staff
		}
	}
	b := Borrowing{
		ID:             borrow.ID.String(),
		BookID:         borrow.BookID.String(),
		SubscriptionID: borrow.SubscriptionID.String(),
		StaffID:        borrow.StaffID.String(),
		BorrowedAt:     borrow.BorrowedAt.UTC().Format(time.RFC3339),
		DueAt:          borrow.DueAt.UTC().Format(time.RFC3339),
		Returning:      r,
		Lost:           l,
		CreatedAt:      borrow.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      borrow.UpdatedAt.UTC().Format(time.RFC3339),
		DeletedAt:      d,
		PrevID:         prevID,
		NextID:         nextID,
	}

	if borrow.Book != nil {
		book := Book{
			ID:        borrow.Book.ID.String(),
			Code:      borrow.Book.Code,
			Title:     borrow.Book.Title,
			Author:    borrow.Book.Author,
			Year:      borrow.Book.Year,
			Cover:     borrow.Book.Cover,
			Colors:    borrow.Book.Colors,
			LibraryID: borrow.Book.LibraryID.String(),
			CreatedAt: borrow.Book.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: borrow.Book.UpdatedAt.UTC().Format(time.RFC3339),
			// DeletedAt: borrow.Book.DeletedAt,
		}
		b.Book = &book
	}

	if borrow.Staff != nil {
		staff := Staff{
			ID:        borrow.Staff.ID.String(),
			Name:      borrow.Staff.Name,
			Role:      string(borrow.Staff.Role),
			UserID:    borrow.Staff.UserID.String(),
			LibraryID: borrow.Staff.LibraryID.String(),
			CreatedAt: borrow.Staff.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: borrow.Staff.UpdatedAt.UTC().Format(time.RFC3339),
		}
		if borrow.Staff.User != nil {
			staff.User = &User{
				ID:        borrow.Staff.User.ID.String(),
				Name:      borrow.Staff.User.Name,
				CreatedAt: borrow.Staff.User.CreatedAt.UTC().Format(time.RFC3339),
				UpdatedAt: borrow.Staff.User.UpdatedAt.UTC().Format(time.RFC3339),
			}
		}
		if borrow.Staff.Library != nil {
			staff.Library = &Library{
				ID:        borrow.Staff.Library.ID.String(),
				Name:      borrow.Staff.Library.Name,
				CreatedAt: borrow.Staff.Library.CreatedAt.UTC().Format(time.RFC3339),
				UpdatedAt: borrow.Staff.Library.UpdatedAt.UTC().Format(time.RFC3339),
			}
		}
		b.Staff = &staff
	}

	if borrow.Subscription != nil {
		sub := Subscription{
			ID:              borrow.SubscriptionID.String(),
			UserID:          borrow.Subscription.UserID.String(),
			MembershipID:    borrow.Subscription.MembershipID.String(),
			CreatedAt:       borrow.Subscription.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:       borrow.Subscription.UpdatedAt.UTC().Format(time.RFC3339),
			ExpiresAt:       borrow.Subscription.ExpiresAt.UTC().Format(time.RFC3339),
			FinePerDay:      borrow.Subscription.FinePerDay,
			LoanPeriod:      borrow.Subscription.LoanPeriod,
			ActiveLoanLimit: borrow.Subscription.ActiveLoanLimit,
			UsageLimit:      borrow.Subscription.UsageLimit,
		}
		if borrow.Subscription.User != nil {
			sub.User = &User{
				ID:        borrow.Subscription.User.ID.String(),
				Name:      borrow.Subscription.User.Name,
				CreatedAt: borrow.Subscription.User.CreatedAt.UTC().Format(time.RFC3339),
				UpdatedAt: borrow.Subscription.User.UpdatedAt.UTC().Format(time.RFC3339),
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
				CreatedAt:       borrow.Subscription.Membership.CreatedAt.UTC().Format(time.RFC3339),
				UpdatedAt:       borrow.Subscription.Membership.UpdatedAt.UTC().Format(time.RFC3339),
			}

			if borrow.Subscription.Membership.Library != nil {
				m.Library = &Library{
					ID:        borrow.Subscription.Membership.Library.ID.String(),
					Name:      borrow.Subscription.Membership.Library.Name,
					Logo:      borrow.Subscription.Membership.Library.Logo,
					CreatedAt: borrow.Subscription.Membership.Library.CreatedAt.UTC().Format(time.RFC3339),
					UpdatedAt: borrow.Subscription.Membership.Library.UpdatedAt.UTC().Format(time.RFC3339),
				}
			}
			sub.Membership = &m
		}
		b.Subscription = &sub
	}

	return ctx.JSON(200, Res{Data: b})
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
		BorrowedAt:     borrow.BorrowedAt.UTC().Format(time.RFC3339),
		DueAt:          borrow.DueAt.UTC().Format(time.RFC3339),
		CreatedAt:      borrow.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      borrow.UpdatedAt.UTC().Format(time.RFC3339),
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
	Lost           *Lost      `json:"lost,omitempty"`
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

	var r *usecase.Returning
	if req.Returning != nil {
		r = &usecase.Returning{
			ReturnedAt: req.Returning.ReturnedAt,
			Fine:       req.Returning.Fine,
		}
	}

	if r != nil {
		if err := s.server.UpdateReturn(ctx.Request().Context(), id, usecase.Returning{
			ReturnedAt: r.ReturnedAt,
			Fine:       r.Fine,
		}); err != nil {
			return ctx.JSON(500, map[string]string{"error": err.Error()})
		}
	}

	var l *usecase.Lost
	if req.Lost != nil {
		l = &usecase.Lost{
			ReportedAt: req.Lost.ReportedAt,
			Fine:       req.Lost.Fine,
			Note:       req.Lost.Note,
		}
	}

	if l != nil {
		if _, err := s.server.UpdateLost(ctx.Request().Context(), id, usecase.Lost{
			ReportedAt: l.ReportedAt,
			Fine:       l.Fine,
			Note:       l.Note,
		}); err != nil {
			return ctx.JSON(500, map[string]string{"error": err.Error()})
		}
	}

	return ctx.JSON(200, Res{Data: Borrowing{
		ID:             borrow.ID.String(),
		BookID:         borrow.BookID.String(),
		SubscriptionID: borrow.SubscriptionID.String(),
		StaffID:        borrow.StaffID.String(),
		BorrowedAt:     borrow.BorrowedAt.UTC().Format(time.RFC3339),
		DueAt:          borrow.DueAt.UTC().Format(time.RFC3339),
		CreatedAt:      borrow.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      borrow.UpdatedAt.UTC().Format(time.RFC3339),
	}})
}

func (s *Server) DeleteBorrowing(ctx echo.Context) error {
	var req GetBorrowingByIDRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.ID)
	if err := s.server.DeleteBorrowing(ctx.Request().Context(), id); err != nil {
		var notFoundErr usecase.ErrNotFound
		if errors.As(err, &notFoundErr) {
			return ctx.JSON(404, map[string]any{
				"error":   notFoundErr.Error(),
				"code":    notFoundErr.Code,
				"message": notFoundErr.Message,
			})
		}

		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(200, Res{Message: "Borrowing deleted successfully"})
}

type ExportBorrowingsRequest struct {
	LibraryID string `json:"library_id" validate:"required,uuid"`

	IsActive       bool    `json:"is_active"`
	IsOverdue      bool    `json:"is_overdue"`
	IsReturned     bool    `json:"is_returned"`
	IsLost         bool    `json:"is_lost"`
	BorrowedAtFrom *string `json:"borrowed_at_from" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	BorrowedAtTo   *string `json:"borrowed_at_to" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

func (s *Server) ExportBorrowings(ctx echo.Context) error {
	var req ExportBorrowingsRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	libID, _ := uuid.Parse(req.LibraryID)
	opt := usecase.ExportBorrowingsOption{
		LibraryID:  libID,
		IsActive:   req.IsActive,
		IsOverdue:  req.IsOverdue,
		IsReturned: req.IsReturned,
		IsLost:     req.IsLost,
	}
	var from *time.Time
	if v := req.BorrowedAtFrom; v != nil {
		if t, err := time.Parse(time.RFC3339, *v); err == nil {
			from = &t
		}
	}
	var to *time.Time
	if v := req.BorrowedAtTo; v != nil {
		if t, err := time.Parse(time.RFC3339, *v); err == nil {
			to = &t
		}
	}
	opt.BorrowedAtFrom = from
	opt.BorrowedAtTo = to
	id, err := s.server.ExportBorrowings(ctx.Request().Context(), opt)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(202, Res{
		Message: "Export job has been queued. You will be notified when it's ready.",
		Data:    map[string]string{"id": id},
	})
}
