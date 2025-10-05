package database

import (
	"context"
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Watchlist struct {
	ID        uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID    uuid.UUID       `gorm:"column:user_id;type:uuid;not null;uniqueIndex:idx_watchlist_book,where:deleted_at IS NULL"`
	BookID    uuid.UUID       `gorm:"column:book_id;type:uuid;not null;uniqueIndex:idx_watchlist_book,where:deleted_at IS NULL"`
	CreatedAt time.Time       `gorm:"column:created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"column:deleted_at"`

	User *User `gorm:"foreignKey:UserID;references:ID"`
	Book *Book `gorm:"foreignKey:BookID;references:ID"`
}

func (Watchlist) TableName() string {
	return "watchlists"
}

// ListWatchlists retrieves watchlists with optional filtering
func (s *service) ListWatchlists(ctx context.Context, opt usecase.ListWatchlistsOption) ([]usecase.Watchlist, int, error) {
	var (
		watchlists  []Watchlist
		uwatchlists []usecase.Watchlist
		count       int64
	)

	db := s.db.Model([]Watchlist{}).WithContext(ctx)

	if opt.UserID != uuid.Nil {
		db = db.Where("user_id = ?", opt.UserID)
	}

	if opt.BookID != uuid.Nil {
		db = db.Where("book_id = ?", opt.BookID)
	}

	if opt.IncludeUser {
		db = db.Preload("User")
	}

	if opt.IncludeBook {
		db = db.Preload("Book")
	}

	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if opt.Limit > 0 {
		db = db.Limit(opt.Limit)
	}

	if opt.Offset > 0 {
		db = db.Offset(opt.Offset)
	}

	if err := db.Find(&watchlists).Error; err != nil {
		return nil, 0, err
	}

	for _, w := range watchlists {
		uw := usecase.Watchlist{
			ID:        w.ID,
			UserID:    w.UserID,
			BookID:    w.BookID,
			CreatedAt: w.CreatedAt,
			UpdatedAt: w.UpdatedAt,
		}

		if w.User != nil {
			uw.User = &usecase.User{
				ID:        w.User.ID,
				Name:      w.User.Name,
				Email:     w.User.Email,
				Phone:     w.User.Phone,
				CreatedAt: w.User.CreatedAt,
				UpdatedAt: w.User.UpdatedAt,
			}
		}

		if w.Book != nil {
			uw.Book = &usecase.Book{
				ID:        w.Book.ID,
				Title:     w.Book.Title,
				Author:    w.Book.Author,
				Year:      w.Book.Year,
				Code:      w.Book.Code,
				Cover:     w.Book.Cover,
				LibraryID: w.Book.LibraryID,
				CreatedAt: w.Book.CreatedAt,
				UpdatedAt: w.Book.UpdatedAt,
			}
		}

		uwatchlists = append(uwatchlists, uw)
	}

	return uwatchlists, int(count), nil
}

// GetWatchlistByID retrieves a watchlist by ID
func (s *service) GetWatchlistByID(ctx context.Context, id uuid.UUID) (usecase.Watchlist, error) {
	var watchlist Watchlist

	db := s.db.WithContext(ctx).Preload("User").Preload("Book")

	if err := db.First(&watchlist, id).Error; err != nil {
		return usecase.Watchlist{}, err
	}

	uw := usecase.Watchlist{
		ID:        watchlist.ID,
		UserID:    watchlist.UserID,
		BookID:    watchlist.BookID,
		CreatedAt: watchlist.CreatedAt,
		UpdatedAt: watchlist.UpdatedAt,
	}

	if watchlist.User != nil {
		uw.User = &usecase.User{
			ID:        watchlist.User.ID,
			Name:      watchlist.User.Name,
			Email:     watchlist.User.Email,
			Phone:     watchlist.User.Phone,
			CreatedAt: watchlist.User.CreatedAt,
			UpdatedAt: watchlist.User.UpdatedAt,
		}
	}

	if watchlist.Book != nil {
		uw.Book = &usecase.Book{
			ID:        watchlist.Book.ID,
			Title:     watchlist.Book.Title,
			Author:    watchlist.Book.Author,
			Year:      watchlist.Book.Year,
			Code:      watchlist.Book.Code,
			Cover:     watchlist.Book.Cover,
			LibraryID: watchlist.Book.LibraryID,
			CreatedAt: watchlist.Book.CreatedAt,
			UpdatedAt: watchlist.Book.UpdatedAt,
		}
	}

	return uw, nil
}

// CreateWatchlist creates a new watchlist entry
func (s *service) CreateWatchlist(ctx context.Context, w usecase.Watchlist) (usecase.Watchlist, error) {
	watchlist := Watchlist{
		UserID: w.UserID,
		BookID: w.BookID,
	}

	if err := s.db.WithContext(ctx).Create(&watchlist).Error; err != nil {
		return usecase.Watchlist{}, err
	}

	return usecase.Watchlist{
		ID:        watchlist.ID,
		UserID:    watchlist.UserID,
		BookID:    watchlist.BookID,
		CreatedAt: watchlist.CreatedAt,
		UpdatedAt: watchlist.UpdatedAt,
	}, nil
}

// DeleteWatchlist deletes a watchlist entry
func (s *service) DeleteWatchlist(ctx context.Context, w usecase.Watchlist) error {
	return s.db.
		WithContext(ctx).
		Where("user_id = ? AND book_id = ?", w.UserID, w.BookID).
		Delete(&Watchlist{}).
		Error
}
