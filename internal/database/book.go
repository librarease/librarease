package database

import (
	"context"
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Book struct {
	ID         uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	Title      string          `gorm:"column:title;type:varchar(255)"`
	Author     string          `gorm:"column:author;type:varchar(255)"`
	Year       int             `gorm:"column:year;type:int"`
	Code       string          `gorm:"column:code;type:varchar(255);uniqueIndex:idx_lib_code"`
	Count      int             `gorm:"column:count;type:int;default:1"`
	Cover      string          `gorm:"column:cover;type:varchar(255)"`
	CreatedAt  time.Time       `gorm:"column:created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at"`
	DeletedAt  *gorm.DeletedAt `gorm:"column:deleted_at"`
	LibraryID  uuid.UUID       `gorm:"uniqueIndex:idx_lib_code"`
	Library    *Library        `gorm:"foreignKey:LibraryID;"`
	Borrowings []Borrowing
}

func (Book) TableName() string {
	return "books"
}

func (s *service) ListBooks(ctx context.Context, opt usecase.ListBooksOption) ([]usecase.Book, int, error) {
	var (
		books  []Book
		ubooks []usecase.Book
		count  int64
	)

	db := s.db.Model([]Book{}).WithContext(ctx)

	if opt.LibraryIDs != nil {
		db = db.Where("library_id IN ?", opt.LibraryIDs)
	}

	if opt.Title != "" {
		db = db.Where("title ILIKE ?", "%"+opt.Title+"%")
	}

	if opt.IDs != nil {
		db = db.Where("id IN ?", opt.IDs)
	}

	var (
		orderIn = "DESC"
		orderBy = "created_at"
	)
	if opt.SortBy != "" {
		orderBy = opt.SortBy
	}
	if opt.SortIn != "" {
		orderIn = opt.SortIn
	}

	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := db.
		Joins("Library").
		Limit(opt.Limit).
		Offset(opt.Skip).
		Order(orderBy + " " + orderIn).
		Find(&books).
		Error; err != nil {

		return nil, 0, err
	}

	for _, b := range books {
		ub := b.ConvertToUsecase()
		if b.Library != nil {
			lib := b.Library.ConvertToUsecase()
			ub.Library = &lib
		}
		ubooks = append(ubooks, ub)
	}

	return ubooks, int(count), nil
}

func (s *service) GetBookByID(ctx context.Context, id uuid.UUID) (usecase.Book, error) {
	var b Book

	err := s.db.WithContext(ctx).Preload("Library").Where("id = ?", id).First(&b).Error
	if err != nil {
		return usecase.Book{}, err
	}

	book := b.ConvertToUsecase()
	if b.Library != nil {
		lib := b.Library.ConvertToUsecase()
		book.Library = &lib
	}

	return book, nil
}

func (s *service) CreateBook(ctx context.Context, book usecase.Book) (usecase.Book, error) {
	b := Book{
		Title:     book.Title,
		Author:    book.Author,
		Year:      book.Year,
		Code:      book.Code,
		Count:     book.Count,
		Cover:     book.Cover,
		LibraryID: book.LibraryID,
	}

	err := s.db.WithContext(ctx).Create(&b).Error
	if err != nil {
		return usecase.Book{}, err
	}
	return b.ConvertToUsecase(), nil
}

func (s *service) UpdateBook(ctx context.Context, id uuid.UUID, book usecase.Book) (usecase.Book, error) {
	b := Book{
		Title:     book.Title,
		Author:    book.Author,
		Year:      book.Year,
		Code:      book.Code,
		Count:     book.Count,
		Cover:     book.Cover,
		LibraryID: book.LibraryID,
	}

	err := s.db.
		WithContext(ctx).
		Model(&b).
		Clauses(clause.Returning{}).
		Where("id = ?", id).
		Updates(&b).
		Error
	if err != nil {
		return usecase.Book{}, err
	}
	return b.ConvertToUsecase(), nil
}

// Convert core model to Usecase
func (b Book) ConvertToUsecase() usecase.Book {
	var d *time.Time
	if b.DeletedAt != nil {
		d = &b.DeletedAt.Time
	}
	return usecase.Book{
		ID:        b.ID,
		Title:     b.Title,
		Author:    b.Author,
		Year:      b.Year,
		Code:      b.Code,
		Count:     b.Count,
		Cover:     b.Cover,
		LibraryID: b.LibraryID,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
		DeletedAt: d,
	}
}
