package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Book struct {
	ID        uuid.UUID
	Title     string
	Author    string
	Year      int
	Code      string
	LibraryID uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Library   *Library
}

type ListBooksOption struct {
	Skip      int
	Limit     int
	LibraryID uuid.UUIDs
	IDs       uuid.UUIDs
	Title     string
	SortBy    string
	SortIn    string
}

func (u Usecase) ListBooks(ctx context.Context, opt ListBooksOption) ([]Book, int, error) {
	return u.repo.ListBooks(ctx, opt)
}

func (u Usecase) CreateBook(ctx context.Context, book Book) (Book, error) {
	return u.repo.CreateBook(ctx, book)
}

func (u Usecase) GetBookByID(ctx context.Context, id uuid.UUID) (Book, error) {
	return u.repo.GetBookByID(ctx, id)
}

func (u Usecase) UpdateBook(ctx context.Context, book Book) (Book, error) {
	return u.repo.UpdateBook(ctx, book)
}
