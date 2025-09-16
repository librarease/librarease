package database

import (
	"context"
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Collection struct {
	ID          uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	LibraryID   uuid.UUID       `gorm:"column:library_id;type:uuid;not null"`
	Title       string          `gorm:"column:title;type:varchar(255);not null"`
	Description string          `gorm:"column:description;type:text"`
	CreatedAt   time.Time       `gorm:"column:created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at"`
	DeletedAt   *gorm.DeletedAt `gorm:"column:deleted_at"`

	Library   *Library              `gorm:"foreignKey:LibraryID;references:ID"`
	Books     []CollectionBooks     `gorm:"foreignKey:CollectionID"`
	Followers []CollectionFollowers `gorm:"foreignKey:CollectionID"`

	Cover *Asset `gorm:"polymorphicType:OwnerType;polymorphicId:OwnerID;polymorphicValue:collections"`

	BookCount     int `gorm:"->;-:migration"`
	FollowerCount int `gorm:"->;-:migration"`
}

func (Collection) TableName() string {
	return "collections"
}

type CollectionBooks struct {
	ID           uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	CollectionID uuid.UUID       `gorm:"column:collection_id;type:uuid;not null;uniqueIndex:idx_collection_book,where:deleted_at IS NULL"`
	BookID       uuid.UUID       `gorm:"column:book_id;type:uuid;not null;uniqueIndex:idx_collection_book,where:deleted_at IS NULL"`
	CreatedAt    time.Time       `gorm:"column:created_at"`
	UpdatedAt    time.Time       `gorm:"column:updated_at"`
	DeletedAt    *gorm.DeletedAt `gorm:"column:deleted_at"`

	Collection *Collection `gorm:"foreignKey:CollectionID;references:ID"`
	Book       *Book       `gorm:"foreignKey:BookID;references:ID"`
}

func (CollectionBooks) TableName() string {
	return "collection_books"
}

type CollectionFollowers struct {
	ID           uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	CollectionID uuid.UUID       `gorm:"column:collection_id;type:uuid;not null;uniqueIndex:idx_collection_follower,where:deleted_at IS NULL"`
	UserID       uuid.UUID       `gorm:"column:user_id;type:uuid;not null;uniqueIndex:idx_collection_follower,where:deleted_at IS NULL"`
	CreatedAt    time.Time       `gorm:"column:created_at"`
	UpdatedAt    time.Time       `gorm:"column:updated_at"`
	DeletedAt    *gorm.DeletedAt `gorm:"column:deleted_at"`

	Collection *Collection `gorm:"foreignKey:CollectionID;references:ID"`
	User       *User       `gorm:"foreignKey:UserID;references:ID"`
}

func (CollectionFollowers) TableName() string {
	return "collection_followers"
}

// ListCollections retrieves collections with optional filtering
func (s *service) ListCollections(ctx context.Context, opt usecase.ListCollectionsOption) ([]usecase.Collection, int, error) {
	var (
		collections  []Collection
		ucollections []usecase.Collection
		count        int64
	)

	db := s.db.
		Model(&Collection{}).
		WithContext(ctx).
		Preload("Cover", func(db *gorm.DB) *gorm.DB {
			return db.Order("updated_at DESC")
		}).
		Select(`*,
			(SELECT COUNT(*) FROM collection_books WHERE collection_books.collection_id = collections.id AND collection_books.deleted_at IS NULL) AS book_count`)

	if opt.LibraryID != uuid.Nil {
		db = db.Where("library_id = ?", opt.LibraryID)
	}

	if opt.Title != "" {
		db = db.Where("title ILIKE ?", "%"+opt.Title+"%")
	}

	if opt.IncludeLibrary {
		db = db.Preload("Library")
	}

	if opt.IncludeBooks {
		db = db.Preload("Books.Book")
	}

	if opt.IncludeFollowers {
		db = db.Preload("Followers.User")
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

	sortBy := "created_at"
	sortIn := "desc"
	if opt.SortBy != "" {
		sortBy = opt.SortBy
	}
	if opt.SortIn != "" {
		sortIn = opt.SortIn
	}

	if err := db.
		Order(sortBy + " " + sortIn).
		Find(&collections).
		Error; err != nil {
		return nil, 0, err
	}

	for _, c := range collections {
		uc := usecase.Collection{
			ID:            c.ID,
			LibraryID:     c.LibraryID,
			Title:         c.Title,
			Description:   c.Description,
			CreatedAt:     c.CreatedAt,
			UpdatedAt:     c.UpdatedAt,
			BookCount:     c.BookCount,
			FollowerCount: c.FollowerCount,
		}

		if c.Library != nil {
			uc.Library = &usecase.Library{
				ID:        c.Library.ID,
				Name:      c.Library.Name,
				Address:   c.Library.Address,
				Phone:     c.Library.Phone,
				CreatedAt: c.Library.CreatedAt,
				UpdatedAt: c.Library.UpdatedAt,
			}
		}

		if c.Cover != nil {
			uc.Cover = &usecase.Asset{
				ID:        c.Cover.ID,
				Path:      c.Cover.Path,
				Colors:    c.Cover.Colors,
				OwnerID:   c.Cover.OwnerID,
				OwnerType: c.Cover.OwnerType,
				Kind:      c.Cover.Kind,
				IsPrimary: c.Cover.IsPrimary,
				Position:  c.Cover.Position,
				CreatedAt: c.Cover.CreatedAt,
				UpdatedAt: c.Cover.UpdatedAt,
			}
		}

		ucollections = append(ucollections, uc)
	}

	return ucollections, int(count), nil
}

// GetCollectionByID retrieves a collection by ID
func (s *service) GetCollectionByID(ctx context.Context, id uuid.UUID) (usecase.Collection, error) {
	var collection Collection

	db := s.db.
		WithContext(ctx).
		Preload("Library").
		Preload("Cover", func(db *gorm.DB) *gorm.DB {
			return db.Order("updated_at DESC")
		}).
		Select(`*,
			(SELECT COUNT(*) FROM collection_books WHERE collection_books.collection_id = collections.id AND collection_books.deleted_at IS NULL) AS book_count`)

	if err := db.First(&collection, id).Error; err != nil {
		return usecase.Collection{}, err
	}

	uc := usecase.Collection{
		ID:            collection.ID,
		LibraryID:     collection.LibraryID,
		Title:         collection.Title,
		Description:   collection.Description,
		CreatedAt:     collection.CreatedAt,
		UpdatedAt:     collection.UpdatedAt,
		BookCount:     collection.BookCount,
		FollowerCount: collection.FollowerCount,
	}

	if collection.Library != nil {
		uc.Library = &usecase.Library{
			ID:        collection.Library.ID,
			Name:      collection.Library.Name,
			Address:   collection.Library.Address,
			Phone:     collection.Library.Phone,
			CreatedAt: collection.Library.CreatedAt,
			UpdatedAt: collection.Library.UpdatedAt,
		}
	}

	if collection.Cover != nil {
		uc.Cover = &usecase.Asset{
			ID:        collection.Cover.ID,
			Path:      collection.Cover.Path,
			Colors:    collection.Cover.Colors,
			OwnerID:   collection.Cover.OwnerID,
			OwnerType: collection.Cover.OwnerType,
			Kind:      collection.Cover.Kind,
			IsPrimary: collection.Cover.IsPrimary,
			Position:  collection.Cover.Position,
			CreatedAt: collection.Cover.CreatedAt,
			UpdatedAt: collection.Cover.UpdatedAt,
		}
	}

	return uc, nil
}

// CreateCollection creates a new collection
func (s *service) CreateCollection(ctx context.Context, c usecase.Collection) (usecase.Collection, error) {
	var cover *Asset
	if c.Cover != nil {
		cover = &Asset{
			Path:   c.Cover.Path,
			Colors: c.Cover.Colors,
		}
	}
	collection := Collection{
		LibraryID:   c.LibraryID,
		Title:       c.Title,
		Cover:       cover,
		Description: c.Description,
	}

	if err := s.db.
		WithContext(ctx).
		Create(&collection).
		Error; err != nil {

		return usecase.Collection{}, err
	}

	return usecase.Collection{
		ID:        collection.ID,
		LibraryID: collection.LibraryID,
		Title:     collection.Title,
		CreatedAt: collection.CreatedAt,
		UpdatedAt: collection.UpdatedAt,
	}, nil
}

// UpdateCollection updates an existing collection
func (s *service) UpdateCollection(ctx context.Context, id uuid.UUID, req usecase.UpdateCollectionRequest) (usecase.Collection, error) {

	var cover *Asset
	if req.Cover != nil {
		cover = &Asset{
			Path:   req.Cover.Path,
			Colors: req.Cover.Colors,
		}
	}

	update := Collection{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Cover:       cover,
	}

	if err := s.db.
		WithContext(ctx).
		Clauses(clause.Returning{}).
		Updates(&update).
		Error; err != nil {
		return usecase.Collection{}, err
	}

	return usecase.Collection{
		ID:          update.ID,
		LibraryID:   update.LibraryID,
		Title:       update.Title,
		Description: update.Description,
		CreatedAt:   update.CreatedAt,
		UpdatedAt:   update.UpdatedAt,
	}, nil
}

func (s *service) DeleteCollection(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Delete(&Collection{}, id).Error
}

func (s *service) ListCollectionBooks(ctx context.Context, id uuid.UUID, opt usecase.ListCollectionBooksOption) ([]usecase.CollectionBook, int, error) {
	var (
		collectionBooks  []CollectionBooks
		ucollectionBooks []usecase.CollectionBook
		count            int64
	)

	db := s.db.
		Model([]CollectionBooks{}).
		WithContext(ctx).
		Where("collection_id = ?", id)

	if opt.IncludeBook {
		db = db.Preload("Book")
	}

	// For book
	if opt.BookTitle != "" || opt.BookSortBy != "" || opt.BookSortIn != "" {
		db = db.Joins("JOIN books ON books.id = collection_books.book_id")
	}
	if opt.BookTitle != "" {
		db = db.Where("books.title ILIKE ?", "%"+opt.BookTitle+"%")
	}
	if opt.BookSortBy != "" {
		sortBy := opt.BookSortBy
		sortIn := "desc"
		if opt.BookSortIn != "" {
			sortIn = opt.BookSortIn
		}
		db = db.Order("books." + sortBy + " " + sortIn)
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

	sortBy := "updated_at"
	sortIn := "desc"
	if opt.SortBy != "" {
		sortBy = opt.SortBy
	}
	if opt.SortIn != "" {
		sortIn = opt.SortIn
	}

	if err := db.
		Order(sortBy + " " + sortIn).
		Find(&collectionBooks).
		Error; err != nil {
		return nil, 0, err
	}

	var (
		statsMap map[uuid.UUID]usecase.BookStats
		err      error
	)

	bookIDs := make([]uuid.UUID, 0, len(collectionBooks))
	for _, cb := range collectionBooks {
		bookIDs = append(bookIDs, cb.BookID)
	}
	if opt.IncludeBook && len(bookIDs) > 0 {
		statsMap, err = s.getBookStats(ctx, bookIDs)
		if err != nil {
			return nil, 0, err
		}
	}

	for _, cb := range collectionBooks {
		ucb := usecase.CollectionBook{
			ID:           cb.ID,
			CollectionID: cb.CollectionID,
			BookID:       cb.BookID,
			CreatedAt:    cb.CreatedAt,
			UpdatedAt:    cb.UpdatedAt,
		}

		if cb.Collection != nil {
			ucb.Collection = &usecase.Collection{
				ID:        cb.Collection.ID,
				LibraryID: cb.Collection.LibraryID,
				Title:     cb.Collection.Title,
				CreatedAt: cb.Collection.CreatedAt,
				UpdatedAt: cb.Collection.UpdatedAt,
			}
		}

		if cb.Book != nil {
			ucb.Book = &usecase.Book{
				ID:        cb.Book.ID,
				Title:     cb.Book.Title,
				Author:    cb.Book.Author,
				Year:      cb.Book.Year,
				Code:      cb.Book.Code,
				Count:     cb.Book.Count,
				Cover:     cb.Book.Cover,
				LibraryID: cb.Book.LibraryID,
				CreatedAt: cb.Book.CreatedAt,
				UpdatedAt: cb.Book.UpdatedAt,
			}
			if stats, exists := statsMap[cb.Book.ID]; exists {
				ucb.Book.Stats = &stats
			}
		}
		ucollectionBooks = append(ucollectionBooks, ucb)
	}

	return ucollectionBooks, int(count), nil
}

// CreateCollectionBook adds a book to a collection
func (s *service) CreateCollectionBook(ctx context.Context, cb usecase.CollectionBook) (usecase.CollectionBook, error) {
	collectionBook := CollectionBooks{
		CollectionID: cb.CollectionID,
		BookID:       cb.BookID,
	}

	if err := s.db.WithContext(ctx).Create(&collectionBook).Error; err != nil {
		return usecase.CollectionBook{}, err
	}

	return usecase.CollectionBook{
		ID:           collectionBook.ID,
		CollectionID: collectionBook.CollectionID,
		BookID:       collectionBook.BookID,
		CreatedAt:    collectionBook.CreatedAt,
		UpdatedAt:    collectionBook.UpdatedAt,
	}, nil
}

func (s *service) UpdateCollectionBooks(
	ctx context.Context,
	collectionID uuid.UUID,
	bookIDs []uuid.UUID,
) ([]usecase.CollectionBook, error) {

	var collectionBooks []CollectionBooks
	for _, bookID := range bookIDs {
		collectionBooks = append(collectionBooks, CollectionBooks{
			CollectionID: collectionID,
			BookID:       bookID,
		})
	}

	if err := s.db.
		WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:     []clause.Column{{Name: "collection_id"}, {Name: "book_id"}},
			DoNothing:   true,
			TargetWhere: clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "deleted_at", Value: nil}}},
		}).
		Create(&collectionBooks).
		Error; err != nil {
		return nil, err
	}

	var createdCollectionBooks []usecase.CollectionBook
	for _, cb := range collectionBooks {
		createdCollectionBooks = append(createdCollectionBooks, usecase.CollectionBook{
			ID:           cb.ID,
			CollectionID: cb.CollectionID,
			BookID:       cb.BookID,
			CreatedAt:    cb.CreatedAt,
			UpdatedAt:    cb.UpdatedAt,
		})
	}

	return createdCollectionBooks, nil
}

func (s *service) DeleteCollectionBook(ctx context.Context, id uuid.UUID) error {
	return s.db.
		WithContext(ctx).
		Delete(&CollectionBooks{}, id).
		Error
}

func (s *service) DeleteCollectionBooks(ctx context.Context, id uuid.UUID, ids []uuid.UUID) error {
	return s.db.
		WithContext(ctx).
		Where("collection_id = ? AND book_id IN ?", id, ids).
		Delete(&CollectionBooks{}).
		Error
}

// ListCollectionFollowers retrieves collection followers with optional filtering
func (s *service) ListCollectionFollowers(ctx context.Context, opt usecase.ListCollectionFollowersOption) ([]usecase.CollectionFollower, int, error) {
	var (
		collectionFollowers  []CollectionFollowers
		ucollectionFollowers []usecase.CollectionFollower
		count                int64
	)

	db := s.db.Model([]CollectionFollowers{}).WithContext(ctx)

	if opt.CollectionID != uuid.Nil {
		db = db.Where("collection_id = ?", opt.CollectionID)
	}

	if opt.UserID != uuid.Nil {
		db = db.Where("user_id = ?", opt.UserID)
	}

	if opt.IncludeCollection {
		db = db.Preload("Collection")
	}

	if opt.IncludeUser {
		db = db.Preload("User")
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

	if err := db.Find(&collectionFollowers).Error; err != nil {
		return nil, 0, err
	}

	for _, cf := range collectionFollowers {
		ucf := usecase.CollectionFollower{
			ID:           cf.ID,
			CollectionID: cf.CollectionID,
			UserID:       cf.UserID,
			CreatedAt:    cf.CreatedAt,
			UpdatedAt:    cf.UpdatedAt,
		}

		if cf.Collection != nil {
			ucf.Collection = &usecase.Collection{
				ID:        cf.Collection.ID,
				LibraryID: cf.Collection.LibraryID,
				Title:     cf.Collection.Title,
				CreatedAt: cf.Collection.CreatedAt,
				UpdatedAt: cf.Collection.UpdatedAt,
			}
		}

		if cf.User != nil {
			ucf.User = &usecase.User{
				ID:        cf.User.ID,
				Name:      cf.User.Name,
				Email:     cf.User.Email,
				Phone:     cf.User.Phone,
				CreatedAt: cf.User.CreatedAt,
				UpdatedAt: cf.User.UpdatedAt,
			}
		}

		ucollectionFollowers = append(ucollectionFollowers, ucf)
	}

	return ucollectionFollowers, int(count), nil
}

// CreateCollectionFollower adds a follower to a collection
func (s *service) CreateCollectionFollower(ctx context.Context, cf usecase.CollectionFollower) (usecase.CollectionFollower, error) {
	collectionFollower := CollectionFollowers{
		CollectionID: cf.CollectionID,
		UserID:       cf.UserID,
	}

	if err := s.db.WithContext(ctx).Create(&collectionFollower).Error; err != nil {
		return usecase.CollectionFollower{}, err
	}

	return usecase.CollectionFollower{
		ID:           collectionFollower.ID,
		CollectionID: collectionFollower.CollectionID,
		UserID:       collectionFollower.UserID,
		CreatedAt:    collectionFollower.CreatedAt,
		UpdatedAt:    collectionFollower.UpdatedAt,
	}, nil
}

// DeleteCollectionFollower removes a follower from a collection
func (s *service) DeleteCollectionFollower(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Delete(&CollectionFollowers{}, id).Error
}
