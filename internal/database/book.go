package database

import (
	"context"
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Book struct {
	ID         uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	Title      string          `gorm:"column:title;type:varchar(255)"`
	Author     string          `gorm:"column:author;type:varchar(255)"`
	Year       int             `gorm:"column:year;type:int"`
	Code       string          `gorm:"column:code;type:varchar(255);uniqueIndex:idx_lib_code,where:deleted_at IS NULL"`
	Cover      string          `gorm:"column:cover;type:varchar(255)"`
	Colors     datatypes.JSON  `gorm:"column:colors"`
	CreatedAt  time.Time       `gorm:"column:created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at"`
	DeletedAt  *gorm.DeletedAt `gorm:"column:deleted_at"`
	LibraryID  uuid.UUID       `gorm:"uniqueIndex:idx_lib_code,where:deleted_at IS NULL"`
	Library    *Library        `gorm:"foreignKey:LibraryID;"`
	Borrowings []Borrowing
	Watchlists []Watchlist
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
		db = db.Where("books.id IN ?", opt.IDs)
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

	if opt.IncludeWatchlists {

		if opt.WatchlistUserID != uuid.Nil {
			db = db.
				Preload("Watchlists", "user_id = ?", opt.WatchlistUserID).
				Joins("JOIN watchlists w ON w.book_id = books.id AND w.user_id = ? AND w.deleted_at IS NULL", opt.WatchlistUserID)
		} else {
			db = db.Preload("Watchlists")
		}
	}

	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if opt.Limit > 0 {
		db = db.Limit(opt.Limit)
	}

	if opt.Skip > 0 {
		db = db.Offset(opt.Skip)
	}

	if err := db.
		Joins("Library").
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
		if b.Watchlists != nil {
			for _, w := range b.Watchlists {
				uw := usecase.Watchlist{
					ID:        w.ID,
					UserID:    w.UserID,
					BookID:    w.BookID,
					CreatedAt: w.CreatedAt,
					UpdatedAt: w.UpdatedAt,
				}
				ub.Watchlists = append(ub.Watchlists, uw)
			}
		}
		ubooks = append(ubooks, ub)
	}

	bookIDs := make([]uuid.UUID, 0, len(ubooks))
	for _, b := range ubooks {
		bookIDs = append(bookIDs, b.ID)
	}

	// Fetch stats separately if requested
	if opt.IncludeStats && len(ubooks) > 0 {
		statsMap, err := s.getBookStats(ctx, bookIDs)
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

func (s *service) getBookStats(ctx context.Context, bookIDs []uuid.UUID) (map[uuid.UUID]usecase.BookStats, error) {
	if len(bookIDs) == 0 {
		return map[uuid.UUID]usecase.BookStats{}, nil
	}

	stats := make(map[uuid.UUID]usecase.BookStats)

	// 1. Count borrowings per book
	type CountResult struct {
		BookID uuid.UUID
		Count  int
	}
	var counts []CountResult
	if err := s.db.WithContext(ctx).
		Model(&Borrowing{}).
		Select("book_id, COUNT(*) AS count").
		Where("book_id IN ?", bookIDs).
		Group("book_id").
		Scan(&counts).Error; err != nil {
		return nil, err
	}
	for _, c := range counts {
		stats[c.BookID] = usecase.BookStats{BorrowCount: c.Count}
	}

	// 2. Get latest borrowings (no preload)
	var latestBorrowings []Borrowing
	if err := s.db.WithContext(ctx).
		Raw(`
			SELECT DISTINCT ON (book_id) *
			FROM borrowings
			WHERE book_id IN ? AND deleted_at IS NULL
			ORDER BY book_id, created_at DESC
		`, bookIDs).
		Scan(&latestBorrowings).Error; err != nil {
		return nil, err
	}

	// 3. Get Returning and Lost for those borrowings
	borrowingIDs := make([]uuid.UUID, 0, len(latestBorrowings))
	for _, b := range latestBorrowings {
		borrowingIDs = append(borrowingIDs, b.ID)
	}

	var returnings []Returning
	var losts []Lost
	if len(borrowingIDs) > 0 {
		if err := s.db.WithContext(ctx).
			Where("borrowing_id IN ?", borrowingIDs).
			Find(&returnings).Error; err != nil {
			return nil, err
		}
		if err := s.db.WithContext(ctx).
			Where("borrowing_id IN ?", borrowingIDs).
			Find(&losts).Error; err != nil {
			return nil, err
		}
	}

	// Map them for easy access
	retMap := make(map[uuid.UUID]*Returning)
	for i := range returnings {
		retMap[returnings[i].BorrowingID] = &returnings[i]
	}
	lostMap := make(map[uuid.UUID]*Lost)
	for i := range losts {
		lostMap[losts[i].BorrowingID] = &losts[i]
	}

	// 4. Merge all into final stats
	for _, b := range latestBorrowings {
		s := stats[b.BookID]
		var returning *usecase.Returning
		if r, exists := retMap[b.ID]; exists {
			returning = &usecase.Returning{
				ReturnedAt: r.ReturnedAt,
			}
		}
		var lost *usecase.Lost
		if lo, exists := lostMap[b.ID]; exists {
			lost = &usecase.Lost{
				ReportedAt: lo.ReportedAt,
			}
		}
		s.ActiveBorrowing = &usecase.Borrowing{
			ID:         b.ID,
			DueAt:      b.DueAt,
			BorrowedAt: b.BorrowedAt,
			Returning:  returning,
			Lost:       lost,
		}
		stats[b.BookID] = s
	}

	return stats, nil
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
	statsMap, err := s.getBookStats(ctx, []uuid.UUID{id})
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
		ID:        book.ID,
		Title:     book.Title,
		Author:    book.Author,
		Year:      book.Year,
		Code:      book.Code,
		Cover:     book.Cover,
		LibraryID: book.LibraryID,
		Colors:    []byte(book.Colors),
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
		Cover:     book.Cover,
		LibraryID: book.LibraryID,
		Colors:    []byte(book.Colors),
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
		Cover:     b.Cover,
		LibraryID: b.LibraryID,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
		Colors:    []byte(b.Colors),
		DeletedAt: d,
	}
}
