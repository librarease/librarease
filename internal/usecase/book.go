package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/config"
)

type BookStats struct {
	BorrowCount     int
	ActiveBorrowing *Borrowing // nil if available, populated if currently borrowed or lost
	Rating          float64
	ReviewCount     int
}

type Book struct {
	ID          uuid.UUID
	Title       string
	Author      string
	Year        int
	Code        string
	Cover       string
	LibraryID   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	Library     *Library
	Stats       *BookStats
	Colors      json.RawMessage
	Description *string

	// UpdateCover is used to update cover
	UpdateCover *string

	// For watchlist
	Watchlists []Watchlist

	// For collection
	Collections []CollectionBook
}

type ListBooksOption struct {
	Skip         int
	Limit        int
	ID           string
	LibraryIDs   uuid.UUIDs
	IDs          uuid.UUIDs
	Title        string
	SortBy       string
	SortIn       string
	IncludeStats bool

	// For watchlist
	IncludeWatchlists bool
	WatchlistUserID   uuid.UUID
}

func (u Usecase) ListBooks(ctx context.Context, opt ListBooksOption) ([]Book, int, error) {
	books, total, err := u.repo.ListBooks(ctx, opt)
	if err != nil {
		return nil, 0, err
	}

	var list []Book
	for _, b := range books {
		if b.Cover != "" {
			b.Cover = u.fileStorageProvider.GetPublicURL(b.Cover)
		}
		list = append(list, b)
	}

	return list, total, err
}

func (u Usecase) CreateBook(ctx context.Context, book Book) (Book, error) {

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
		fmt.Println("[DEBUG] global superadmin")
		// ALLOW ALL
	case "ADMIN":
		fmt.Println("[DEBUG] global admin")
		// ALLOW ALL
	case "USER":
		fmt.Println("[DEBUG] global user")
		staffs, _, err := u.repo.ListStaffs(ctx, ListStaffsOption{
			UserID:     userID.String(),
			LibraryIDs: uuid.UUIDs{book.LibraryID},
		})
		if err != nil {
			return Book{}, err
		}
		if len(staffs) == 0 {
			return Book{}, fmt.Errorf("user is not staff of the library")
		}
	}

	book.ID = uuid.New()

	if book.Cover != "" {
		var err error
		coverPath := fmt.Sprintf("public/books/%s/cover", book.ID.String())
		book.Cover, err = u.fileStorageProvider.CopyFilePreserveFilename(ctx, book.Cover, coverPath)
		if err != nil {
			log.Printf("CreateBook: copy cover for book %s failed: %v", book.ID, err)
			// don't save cover if copy failed
			book.Cover = ""
		}
	}

	b, err := u.repo.CreateBook(ctx, book)
	if err != nil {
		return Book{}, err
	}

	if b.Cover != "" {
		b.Cover = u.fileStorageProvider.GetPublicURL(b.Cover)
	}

	return b, err
}

type GetBookByIDOption struct {
	IncludeWatchlists bool
	WatchlistUserID   uuid.UUID
}

func (u Usecase) GetBookByID(ctx context.Context, id uuid.UUID, opt GetBookByIDOption) (Book, error) {

	book, err := u.repo.GetBookByID(ctx, id)
	if err != nil {
		return Book{}, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Go(func() {

		mu.Lock()
		if book.Cover != "" {
			book.Cover = u.fileStorageProvider.GetPublicURL(book.Cover)
		}
		if book.Library != nil && book.Library.Logo != "" {
			book.Library.Logo = u.fileStorageProvider.GetPublicURL(book.Library.Logo)
		}
		mu.Unlock()
	})

	wg.Go(func() {
		if opt.IncludeWatchlists && opt.WatchlistUserID != uuid.Nil {
			list, _, err := u.repo.ListWatchlists(ctx, ListWatchlistsOption{
				BookID: book.ID,
				UserID: opt.WatchlistUserID,
			})
			if err != nil {
				log.Printf("err_GetBookByID_ListWatchlists: %v\n", err)
				return
			}
			mu.Lock()
			book.Watchlists = list
			mu.Unlock()
		}
	})

	wg.Wait()

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
		var err error
		coverPath := fmt.Sprintf("public/books/%s/cover", book.ID.String())
		book.Cover, err = u.fileStorageProvider.CopyFilePreserveFilename(ctx, *book.UpdateCover, coverPath)
		if err != nil {
			log.Printf("UpdateBook: copy cover for book %s failed: %v", id, err)
			// don't update cover if copy failed
			book.Cover = ""
		}
	}

	b, err := u.repo.UpdateBook(ctx, id, book)
	if err != nil {
		return Book{}, err
	}

	if b.Cover != "" {
		b.Cover = u.fileStorageProvider.GetPublicURL(b.Cover)
	}

	return b, nil
}
func (u Usecase) DeleteBook(ctx context.Context, id uuid.UUID) error {
	role, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
	if !ok {
		return fmt.Errorf("user role not found in context")
	}
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return fmt.Errorf("user id not found in context")
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
			return err
		}
		if len(staffs) == 0 {
			// TODO: implement error
			return fmt.Errorf("you are not right staff for the book")
		}
	}

	// check borrowings
	b, _, err := u.repo.ListBorrowings(ctx, ListBorrowingsOption{
		BorrowingsOption: BorrowingsOption{
			BookIDs: uuid.UUIDs{id},
		},
	})
	if err != nil {
		return err
	}
	if len(b) > 0 {
		return fmt.Errorf("book has %d borrowings", len(b))
	}

	// u.repo.

	return nil
}
