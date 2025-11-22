package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID          uuid.UUID
	BorrowingID uuid.UUID
	Rating      int
	Comment     *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time

	Borrowing *Borrowing
	User      *User
	Book      *Book

	PrevID *uuid.UUID
	NextID *uuid.UUID
}

type ReviewsOption struct {
	SortBy      string
	SortIn      string
	BorrowingID uuid.UUID
	UserID      uuid.UUID
	BookID      uuid.UUID
	LibraryID   uuid.UUID
	Rating      *int
	Comment     *string
}

type ListReviewsOption struct {
	Skip  int
	Limit int
	ReviewsOption
}

func (u Usecase) ListReviews(ctx context.Context, opt ListReviewsOption) ([]Review, int, error) {
	return u.repo.ListReviews(ctx, opt)
}

func (u Usecase) GetReview(ctx context.Context, id uuid.UUID, opt ReviewsOption) (Review, error) {
	return u.repo.GetReview(ctx, id, opt)
}

func (u Usecase) CreateReview(ctx context.Context, review Review) (Review, error) {
	return u.repo.CreateReview(ctx, review)
}

func (u Usecase) UpdateReview(ctx context.Context, id uuid.UUID, review Review) (Review, error) {
	return u.repo.UpdateReview(ctx, id, review)
}

func (u Usecase) DeleteReview(ctx context.Context, id uuid.UUID) error {
	return u.repo.DeleteReview(ctx, id)
}
