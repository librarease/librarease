package database

import (
	"context"
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Collection struct {
	ID        uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	LibraryID uuid.UUID       `gorm:"column:library_id;type:uuid;not null"`
	Title     string          `gorm:"column:title;type:varchar(255);not null"`
	CreatedAt time.Time       `gorm:"column:created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"column:deleted_at"`

	Library   *Library              `gorm:"foreignKey:LibraryID;references:ID"`
	Books     []CollectionBooks     `gorm:"foreignKey:CollectionID"`
	Followers []CollectionFollowers `gorm:"foreignKey:CollectionID"`
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

// Collection CRUD methods

// ListCollections retrieves collections with optional filtering
func (s *service) ListCollections(ctx context.Context, opt usecase.ListCollectionsOption) ([]usecase.Collection, int, error) {
	var (
		collections  []Collection
		ucollections []usecase.Collection
		count        int64
	)

	db := s.db.Model([]Collection{}).WithContext(ctx)

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

	if err := db.Find(&collections).Error; err != nil {
		return nil, 0, err
	}

	for _, c := range collections {
		uc := usecase.Collection{
			ID:        c.ID,
			LibraryID: c.LibraryID,
			Title:     c.Title,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
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

		ucollections = append(ucollections, uc)
	}

	return ucollections, int(count), nil
}

// GetCollectionByID retrieves a collection by ID
func (s *service) GetCollectionByID(ctx context.Context, id uuid.UUID) (usecase.Collection, error) {
	var collection Collection

	db := s.db.WithContext(ctx).Preload("Library").Preload("Books.Book").Preload("Followers.User")

	if err := db.First(&collection, id).Error; err != nil {
		return usecase.Collection{}, err
	}

	uc := usecase.Collection{
		ID:        collection.ID,
		LibraryID: collection.LibraryID,
		Title:     collection.Title,
		CreatedAt: collection.CreatedAt,
		UpdatedAt: collection.UpdatedAt,
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

	return uc, nil
}

// CreateCollection creates a new collection
func (s *service) CreateCollection(ctx context.Context, c usecase.Collection) (usecase.Collection, error) {
	collection := Collection{
		LibraryID: c.LibraryID,
		Title:     c.Title,
	}

	if err := s.db.WithContext(ctx).Create(&collection).Error; err != nil {
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
func (s *service) UpdateCollection(ctx context.Context, id uuid.UUID, c usecase.Collection) (usecase.Collection, error) {
	var collection Collection

	if err := s.db.WithContext(ctx).First(&collection, id).Error; err != nil {
		return usecase.Collection{}, err
	}

	collection.Title = c.Title
	collection.LibraryID = c.LibraryID

	if err := s.db.WithContext(ctx).Save(&collection).Error; err != nil {
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

// DeleteCollection deletes a collection
func (s *service) DeleteCollection(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Delete(&Collection{}, id).Error
}

// CollectionBooks CRUD methods

// ListCollectionBooks retrieves collection books with optional filtering
func (s *service) ListCollectionBooks(ctx context.Context, opt usecase.ListCollectionBooksOption) ([]usecase.CollectionBook, int, error) {
	var (
		collectionBooks  []CollectionBooks
		ucollectionBooks []usecase.CollectionBook
		count            int64
	)

	db := s.db.Model([]CollectionBooks{}).WithContext(ctx)

	if opt.CollectionID != uuid.Nil {
		db = db.Where("collection_id = ?", opt.CollectionID)
	}

	if opt.BookID != uuid.Nil {
		db = db.Where("book_id = ?", opt.BookID)
	}

	if opt.IncludeCollection {
		db = db.Preload("Collection")
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

	if err := db.Find(&collectionBooks).Error; err != nil {
		return nil, 0, err
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

// DeleteCollectionBook removes a book from a collection
func (s *service) DeleteCollectionBook(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Delete(&CollectionBooks{}, id).Error
}

// CollectionFollowers CRUD methods

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
