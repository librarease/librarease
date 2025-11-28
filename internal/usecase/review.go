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
	ReviewedAt  time.Time
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

	reviews, total, err := u.repo.ListReviews(ctx, opt)
	if err != nil {
		return nil, 0, err
	}
	for _, review := range reviews {
		if review.Book != nil && review.Book.Cover != "" {
			review.Book.Cover = u.fileStorageProvider.GetPublicURL(review.Book.Cover)
		}
	}
	return reviews, total, nil
}

func (u Usecase) GetReview(ctx context.Context, id uuid.UUID, opt ReviewsOption) (Review, error) {
	review, err := u.repo.GetReview(ctx, id, opt)
	if err != nil {
		return Review{}, err
	}
	if review.Book != nil && review.Book.Cover != "" {
		review.Book.Cover = u.fileStorageProvider.GetPublicURL(review.Book.Cover)
	}
	return review, nil

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
