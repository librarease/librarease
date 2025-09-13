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
	Code       string          `gorm:"column:code;type:varchar(255);uniqueIndex:idx_lib_code,where:deleted_at IS NULL"`
	Count      int             `gorm:"column:count;type:int;default:1"`
	Cover      string          `gorm:"column:cover;type:varchar(255)"`
	CreatedAt  time.Time       `gorm:"column:created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at"`
	DeletedAt  *gorm.DeletedAt `gorm:"column:deleted_at"`
	LibraryID  uuid.UUID       `gorm:"uniqueIndex:idx_lib_code,where:deleted_at IS NULL"`
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

	if opt.ID != "" {
		db = db.Where("books.id::text ILIKE ?", "%"+opt.ID+"%")
	}

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

	// Convert to usecase models
	for _, b := range books {
		ub := b.ConvertToUsecase()
		if b.Library != nil {
			lib := b.Library.ConvertToUsecase()
			ub.Library = &lib
		}
		ubooks = append(ubooks, ub)
	}

	// Fetch stats separately if requested
	if opt.IncludeStats && len(ubooks) > 0 {
		statsMap, err := s.getBookStats(ctx, ubooks)
		if err != nil {
			return nil, 0, err
		}

		// Apply stats to books
		for i := range ubooks {
			if stats, exists := statsMap[ubooks[i].ID]; exists {
				ubooks[i].Stats = &stats
			}
		}
	}

	return ubooks, int(count), nil
}

func (s *service) getBookStats(ctx context.Context, books []usecase.Book) (map[uuid.UUID]usecase.BookStats, error) {
	if len(books) == 0 {
		return map[uuid.UUID]usecase.BookStats{}, nil
	}

	// Extract book IDs
	bookIDs := make([]uuid.UUID, 0, len(books))
	for _, book := range books {
		bookIDs = append(bookIDs, book.ID)
	}

	type bookStat struct {
		BookID      uuid.UUID `gorm:"column:book_id"`
		LibraryID   uuid.UUID `gorm:"column:library_id"`
		BorrowCount int       `gorm:"column:borrow_count"`
	}

	var stats []bookStat
	err := s.db.WithContext(ctx).Raw(`
	       SELECT 
		       b.id as book_id,
		       b.library_id,
		       COUNT(br.id) as borrow_count
	       FROM books b
	       LEFT JOIN borrowings br ON br.book_id = b.id
	       WHERE b.id IN ?
	       GROUP BY b.id, b.library_id
       `, bookIDs).Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	// Query availability for all book IDs
	var availResults []struct {
		ID          uuid.UUID
		IsAvailable bool
	}
	err2 := s.db.Raw(`
			   SELECT b.id, (b.count > COALESCE(SUM(
				   CASE 
					   WHEN br.id IS NOT NULL AND br.deleted_at IS NULL AND r.id IS NULL THEN 1 
					   ELSE 0 
				   END
			   ), 0)) AS is_available
			   FROM books b
			   LEFT JOIN borrowings br ON br.book_id = b.id AND br.deleted_at IS NULL
			   LEFT JOIN returnings r ON r.borrowing_id = br.id AND r.deleted_at IS NULL
			   WHERE b.id IN ? AND b.deleted_at IS NULL
			   GROUP BY b.id, b.count
		   `, bookIDs).Scan(&availResults).Error
	availMap := make(map[uuid.UUID]bool)
	if err2 != nil {
		return nil, err2
	}
	for _, res := range availResults {
		availMap[res.ID] = res.IsAvailable
	}

	// Convert to map
	statsMap := make(map[uuid.UUID]usecase.BookStats)
	for _, stat := range stats {
		statsMap[stat.BookID] = usecase.BookStats{
			BorrowCount: stat.BorrowCount,
			IsAvailable: availMap[stat.BookID],
		}
	}
	// Also handle books with no borrowings (not in stats)
	for _, id := range bookIDs {
		if _, ok := statsMap[id]; !ok {
			statsMap[id] = usecase.BookStats{
				BorrowCount: 0,
				IsAvailable: availMap[id],
			}
		}
	}
	return statsMap, nil
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

	// Fetch stats for this book
	statsMap, err := s.getBookStats(ctx, []usecase.Book{book})
	if err != nil {
		return usecase.Book{}, err
	}
	if stats, exists := statsMap[book.ID]; exists {
		book.Stats = &stats
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
