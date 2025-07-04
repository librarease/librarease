package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/config"
)

type Book struct {
	ID        uuid.UUID
	Title     string
	Author    string
	Year      int
	Code      string
	Count     int
	Cover     string
	LibraryID uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Library   *Library

	// UpdateCover is used to update cover
	UpdateCover *string
}

type ListBooksOption struct {
	Skip       int
	Limit      int
	ID         string
	LibraryIDs uuid.UUIDs
	IDs        uuid.UUIDs
	Title      string
	SortBy     string
	SortIn     string
}

func (u Usecase) ListBooks(ctx context.Context, opt ListBooksOption) ([]Book, int, error) {
	books, total, err := u.repo.ListBooks(ctx, opt)
	if err != nil {
		return nil, 0, err
	}

	publicURL, _ := u.fileStorageProvider.GetPublicURL(ctx)

	var list []Book
	for _, b := range books {
		if b.Cover != "" {
			b.Cover = fmt.Sprintf("%s/books/%s/cover/%s", publicURL, b.ID, b.Cover)
		}
		list = append(list, b)
	}

	return list, total, err
}

func (u Usecase) CreateBook(ctx context.Context, book Book) (Book, error) {
	book, err := u.repo.CreateBook(ctx, book)
	if err != nil {
		return Book{}, err
	}

	var cover = book.Cover
	if cover != "" {
		var coverPath = fmt.Sprintf("books/%s/cover", book.ID)
		err = u.fileStorageProvider.MoveTempFilePublic(ctx, cover, coverPath)
		if err != nil {
			fmt.Printf("failed to move file for book %s: %v\n", book.ID, err)
			// don't save cover if failed to move file
			cover = ""
		}
	}

	publicURL, _ := u.fileStorageProvider.GetPublicURL(ctx)
	if cover != "" {
		book.Cover = fmt.Sprintf("%s/books/%s/cover/%s", publicURL, book.ID, book.Cover)
	}

	return book, err
}

func (u Usecase) GetBookByID(ctx context.Context, id uuid.UUID) (Book, error) {
	book, err := u.repo.GetBookByID(ctx, id)
	if err != nil {
		return Book{}, err
	}

	publicURL, _ := u.fileStorageProvider.GetPublicURL(ctx)
	if book.Cover != "" {
		book.Cover = fmt.Sprintf("%s/books/%s/cover/%s", publicURL, book.ID, book.Cover)
	}

	if book.Library != nil && book.Library.Logo != "" {
		book.Library.Logo = fmt.Sprintf("%s/libraries/%s/logo/%s", publicURL, book.Library.ID, book.Library.Logo)
	}

	return book, nil
}

func (u Usecase) UpdateBook(ctx context.Context, id uuid.UUID, book Book) (Book, error) {
	role, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
	if !ok {
		return Book{}, fmt.Errorf("user role not found in context")
	}
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return Book{}, fmt.Errorf("user id not found in context")
	}

	switch role {
	case "SUPERADMIN":
		// ALLOW
	case "ADMIN":
		// ALLlOW
	case "USER":
		staffs, _, err := u.repo.ListStaffs(ctx, ListStaffsOption{
			UserID: userID.String(),
			// FIXME: no way to check book's library
			// LibraryIDs: uuid.UUIDs{library.ID},
		})
		if err != nil {
			return Book{}, err
		}
		if len(staffs) == 0 {
			// TODO: implement error
			return Book{}, fmt.Errorf("you are not right staff for the book")
		}
	}

	if book.UpdateCover != nil {
		coverPath := fmt.Sprintf("books/%s/cover", id)
		err := u.fileStorageProvider.MoveTempFilePublic(ctx, *book.UpdateCover, coverPath)
		if err != nil {
			fmt.Printf("failed to move file for book %s: %v\n", id, err)
			return Book{}, err
		}
		book.Cover = *book.UpdateCover
	}

	b, err := u.repo.UpdateBook(ctx, id, book)
	if err != nil {
		return Book{}, err
	}

	publicURL, _ := u.fileStorageProvider.GetPublicURL(ctx)
	if b.Cover != "" {
		b.Cover = fmt.Sprintf("%s/books/%s/cover/%s", publicURL, b.ID, b.Cover)
	}

	return b, nil
}
