package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Watchlist structures
type Watchlist struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	BookID    uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time

	User *User
	Book *Book
}

type ListWatchlistsOption struct {
	UserID      uuid.UUID
	BookID      uuid.UUID
	Limit       int
	Offset      int
	IncludeUser bool
	IncludeBook bool
	LibraryID   uuid.UUID
	Title       string
	SortIn      string
}

func (u Usecase) CreateWatchlist(ctx context.Context, w Watchlist) (Watchlist, error) {
	return u.repo.CreateWatchlist(ctx, w)
}

func (u Usecase) DeleteWatchlist(ctx context.Context, w Watchlist) error {
	return u.repo.DeleteWatchlist(ctx, w)
}
