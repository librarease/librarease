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

// Collection structures
type Collection struct {
	ID        uuid.UUID
	LibraryID uuid.UUID
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time

	Library   *Library
	Books     []CollectionBook
	Followers []CollectionFollower
}

type ListCollectionsOption struct {
	LibraryID        uuid.UUID
	Title            string
	Limit            int
	Offset           int
	IncludeLibrary   bool
	IncludeBooks     bool
	IncludeFollowers bool
}

// CollectionBook structures
type CollectionBook struct {
	ID           uuid.UUID
	CollectionID uuid.UUID
	BookID       uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Collection *Collection
	Book       *Book
}

type ListCollectionBooksOption struct {
	CollectionID      uuid.UUID
	BookID            uuid.UUID
	Limit             int
	Offset            int
	IncludeCollection bool
	IncludeBook       bool
}

// CollectionFollower structures
type CollectionFollower struct {
	ID           uuid.UUID
	CollectionID uuid.UUID
	UserID       uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Collection *Collection
	User       *User
}

type ListCollectionFollowersOption struct {
	CollectionID      uuid.UUID
	UserID            uuid.UUID
	Limit             int
	Offset            int
	IncludeCollection bool
	IncludeUser       bool
}

// Watchlist usecase methods
// func (u Usecase) ListWatchlists(ctx context.Context, opt ListWatchlistsOption) ([]Watchlist, int, error) {
// 	return u.repo.ListWatchlists(ctx, opt)
// }

// func (u Usecase) GetWatchlistByID(ctx context.Context, id uuid.UUID) (Watchlist, error) {
// 	return u.repo.GetWatchlistByID(ctx, id)
// }

func (u Usecase) CreateWatchlist(ctx context.Context, w Watchlist) (Watchlist, error) {
	return u.repo.CreateWatchlist(ctx, w)
}

func (u Usecase) DeleteWatchlist(ctx context.Context, w Watchlist) error {
	return u.repo.DeleteWatchlist(ctx, w)
}

// Collection usecase methods
// func (u Usecase) ListCollections(ctx context.Context, opt ListCollectionsOption) ([]Collection, int, error) {
// 	return u.repo.ListCollections(ctx, opt)
// }

// func (u Usecase) GetCollectionByID(ctx context.Context, id uuid.UUID) (Collection, error) {
// 	return u.repo.GetCollectionByID(ctx, id)
// }

// func (u Usecase) CreateCollection(ctx context.Context, c Collection) (Collection, error) {
// 	return u.repo.CreateCollection(ctx, c)
// }

// func (u Usecase) UpdateCollection(ctx context.Context, id uuid.UUID, c Collection) (Collection, error) {
// 	return u.repo.UpdateCollection(ctx, id, c)
// }

// func (u Usecase) DeleteCollection(ctx context.Context, id uuid.UUID) error {
// 	return u.repo.DeleteCollection(ctx, id)
// }

// // CollectionBook usecase methods
// func (u Usecase) ListCollectionBooks(ctx context.Context, opt ListCollectionBooksOption) ([]CollectionBook, int, error) {
// 	return u.repo.ListCollectionBooks(ctx, opt)
// }

// func (u Usecase) CreateCollectionBook(ctx context.Context, cb CollectionBook) (CollectionBook, error) {
// 	return u.repo.CreateCollectionBook(ctx, cb)
// }

// func (u Usecase) DeleteCollectionBook(ctx context.Context, id uuid.UUID) error {
// 	return u.repo.DeleteCollectionBook(ctx, id)
// }

// // CollectionFollower usecase methods
// func (u Usecase) ListCollectionFollowers(ctx context.Context, opt ListCollectionFollowersOption) ([]CollectionFollower, int, error) {
// 	return u.repo.ListCollectionFollowers(ctx, opt)
// }

// func (u Usecase) CreateCollectionFollower(ctx context.Context, cf CollectionFollower) (CollectionFollower, error) {
// 	return u.repo.CreateCollectionFollower(ctx, cf)
// }

// func (u Usecase) DeleteCollectionFollower(ctx context.Context, id uuid.UUID) error {
// 	return u.repo.DeleteCollectionFollower(ctx, id)
// }
