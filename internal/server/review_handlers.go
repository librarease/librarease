package server

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/librarease/librarease/internal/usecase"
)

type Review struct {
	ID          string  `json:"id"`
	BorrowingID string  `json:"borrowing_id"`
	Rating      int     `json:"rating"`
	Comment     *string `json:"comment,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	DeletedAt   *string `json:"deleted_at,omitempty"`

	Borrowing *Borrowing `json:"borrowing,omitempty"`
	User      *User      `json:"user,omitempty"`
	Book      *Book      `json:"book,omitempty"`

	PrevID *string `json:"prev_id,omitempty"`
	NextID *string `json:"next_id,omitempty"`
}

type ReviewsOption struct {
	SortBy      string  `query:"sort_by" validate:"omitempty,oneof=created_at updated_at"`
	SortIn      string  `query:"sort_in" validate:"omitempty,oneof=asc desc"`
	BorrowingID string  `query:"borrowing_id" validate:"omitempty,uuid"`
	UserID      string  `query:"user_id" validate:"omitempty,uuid"`
	BookID      string  `query:"book_id" validate:"omitempty,uuid"`
	Rating      *int    `query:"rating" validate:"omitempty"`
	Comment     *string `query:"comment" validate:"omitempty"`
}

type ListReviewsOption struct {
	Skip  int `query:"skip"`
	Limit int `query:"limit"`
	ReviewsOption
}

func (s *Server) ListReviews(ctx echo.Context) error {
	var req ListReviewsOption
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	var borrowingID uuid.UUID
	if req.BorrowingID != "" {
		borrowingID, _ = uuid.Parse(req.BorrowingID)
	}

	var userID uuid.UUID
	if req.UserID != "" {
		userID, _ = uuid.Parse(req.UserID)
	}

	var bookID uuid.UUID
	if req.BookID != "" {
		bookID, _ = uuid.Parse(req.BookID)
	}

	reviews, count, err := s.server.ListReviews(ctx.Request().Context(), usecase.ListReviewsOption{
		Skip:  req.Skip,
		Limit: req.Limit,
		ReviewsOption: usecase.ReviewsOption{
			SortBy:      req.SortBy,
			SortIn:      req.SortIn,
			BorrowingID: borrowingID,
			BookID:      bookID,
			UserID:      userID,
			Rating:      req.Rating,
			Comment:     req.Comment,
		},
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	list := make([]Review, 0, len(reviews))
	for _, r := range reviews {
		var prevID *string
		if r.PrevID != nil {
			id := r.PrevID.String()
			prevID = &id
		}
		var nextID *string
		if r.NextID != nil {
			id := r.NextID.String()
			nextID = &id
		}
		var deletedAt *string
		if r.DeletedAt != nil {
			t := r.DeletedAt.UTC().Format(time.RFC3339)
			deletedAt = &t
		}
		var borw *Borrowing
		if r.Borrowing != nil {
			borw = &Borrowing{
				ID:             r.Borrowing.ID.String(),
				BookID:         r.Borrowing.BookID.String(),
				SubscriptionID: r.Borrowing.SubscriptionID.String(),
				BorrowedAt:     r.Borrowing.BorrowedAt.UTC().Format(time.RFC3339),
				DueAt:          r.Borrowing.DueAt.UTC().Format(time.RFC3339),
			}
		}
		var user *User
		if r.User != nil {
			user = &User{
				ID:   r.User.ID.String(),
				Name: r.User.Name,
			}
		}
		var book *Book
		if r.Book != nil {
			book = &Book{
				ID:        r.Book.ID.String(),
				Title:     r.Book.Title,
				Colors:    r.Book.Colors,
				Author:    r.Book.Author,
				Code:      r.Book.Code,
				Cover:     r.Book.Cover,
				LibraryID: r.Book.LibraryID.String(),
			}
		}
		list = append(list, Review{
			ID:          r.ID.String(),
			BorrowingID: r.BorrowingID.String(),
			Rating:      r.Rating,
			Comment:     r.Comment,
			CreatedAt:   r.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:   r.UpdatedAt.UTC().Format(time.RFC3339),
			DeletedAt:   deletedAt,
			PrevID:      prevID,
			NextID:      nextID,
			Borrowing:   borw,
			Book:        book,
			User:        user,
		})
	}

	return ctx.JSON(200, Res{
		Data: list,
		Meta: &Meta{
			Total: count,
			Skip:  req.Skip,
			Limit: req.Limit,
		},
	})
}

type CreateReviewRequest struct {
	BorrowingID string  `json:"borrowing_id" validate:"required,uuid"`
	Rating      int     `json:"rating" validate:"required,min=0,max=5"`
	Comment     *string `json:"comment"`
}

func (s *Server) CreateReview(ctx echo.Context) error {
	var req CreateReviewRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	borrowingID, _ := uuid.Parse(req.BorrowingID)

	review, err := s.server.CreateReview(ctx.Request().Context(), usecase.Review{
		BorrowingID: borrowingID,
		Rating:      req.Rating,
		Comment:     req.Comment,
	})

	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(201, Res{
		Data: Review{
			ID:        review.ID.String(),
			CreatedAt: review.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: review.UpdatedAt.UTC().Format(time.RFC3339),
		},
	})
}

type GetReviewRequest struct {
	ID string `param:"id" validate:"required,uuid"`
	ReviewsOption
}

func (s *Server) GetReview(ctx echo.Context) error {
	var req GetReviewRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.ID)

	var borrowingID uuid.UUID
	if req.BorrowingID != "" {
		borrowingID, _ = uuid.Parse(req.BorrowingID)
	}

	var userID uuid.UUID
	if req.UserID != "" {
		userID, _ = uuid.Parse(req.UserID)
	}

	var bookID uuid.UUID
	if req.BookID != "" {
		bookID, _ = uuid.Parse(req.BookID)
	}

	r, err := s.server.GetReview(ctx.Request().Context(), id, usecase.ReviewsOption{
		SortBy:      req.SortBy,
		SortIn:      req.SortIn,
		BorrowingID: borrowingID,
		BookID:      bookID,
		UserID:      userID,
		Rating:      req.Rating,
		Comment:     req.Comment,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	var prevID *string
	if r.PrevID != nil {
		id := r.PrevID.String()
		prevID = &id
	}
	var nextID *string
	if r.NextID != nil {
		id := r.NextID.String()
		nextID = &id
	}
	var deletedAt *string
	if r.DeletedAt != nil {
		t := r.DeletedAt.UTC().Format(time.RFC3339)
		deletedAt = &t
	}
	var borw *Borrowing
	if r.Borrowing != nil {
		borw = &Borrowing{
			ID:             r.Borrowing.ID.String(),
			BookID:         r.Borrowing.BookID.String(),
			SubscriptionID: r.Borrowing.SubscriptionID.String(),
			BorrowedAt:     r.Borrowing.BorrowedAt.UTC().Format(time.RFC3339),
			DueAt:          r.Borrowing.DueAt.UTC().Format(time.RFC3339),
		}
	}
	var user *User
	if r.User != nil {
		user = &User{
			ID:   r.User.ID.String(),
			Name: r.User.Name,
		}
	}
	var book *Book
	if r.Book != nil {
		book = &Book{
			ID:        r.Book.ID.String(),
			Title:     r.Book.Title,
			Colors:    r.Book.Colors,
			Author:    r.Book.Author,
			Code:      r.Book.Code,
			Cover:     r.Book.Cover,
			LibraryID: r.Book.LibraryID.String(),
		}
	}

	data := Review{
		ID:          r.ID.String(),
		BorrowingID: r.BorrowingID.String(),
		Rating:      r.Rating,
		Comment:     r.Comment,
		CreatedAt:   r.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   r.UpdatedAt.UTC().Format(time.RFC3339),
		DeletedAt:   deletedAt,
		PrevID:      prevID,
		NextID:      nextID,
		Borrowing:   borw,
		Book:        book,
		User:        user,
	}

	return ctx.JSON(200, Res{
		Data: data,
	})
}

type UpdateReviewRequest struct {
	ID      string  `param:"id" validate:"required,uuid"`
	Rating  int     `json:"rating" validate:"required,min=0,max=5"`
	Comment *string `json:"comment"`
}

func (s *Server) UpdateReview(ctx echo.Context) error {
	var req UpdateReviewRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.ID)
	review, err := s.server.UpdateReview(ctx.Request().Context(), id, usecase.Review{
		Rating:  req.Rating,
		Comment: req.Comment,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(200, Res{
		Data: Review{
			ID:        review.ID.String(),
			Rating:    review.Rating,
			Comment:   review.Comment,
			CreatedAt: review.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: review.UpdatedAt.UTC().Format(time.RFC3339),
		},
	})
}

type DeleteReviewRequest struct {
	ID string `param:"id" validate:"required,uuid"`
}

func (s *Server) DeleteReview(ctx echo.Context) error {
	var req DeleteReviewRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}
	id, _ := uuid.Parse(req.ID)
	if err := s.server.DeleteReview(ctx.Request().Context(), id); err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(200, Res{
		Data: map[string]string{"id": id.String()},
	})
}
