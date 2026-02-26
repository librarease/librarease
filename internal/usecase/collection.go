package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/config"
)

// Collection structures
type Collection struct {
	ID          uuid.UUID
	LibraryID   uuid.UUID
	Title       string
	Cover       string
	Colors      json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Description string

	Library *Library
	Stats   *CollectionStats

	BookCount     int
	FollowerCount int
	BookIDs       []uuid.UUID
}

type CollectionStats struct {
	FollowedAt *time.Time
}

type ListCollectionsOption struct {
	LibraryID        uuid.UUID
	Title            string
	Limit            int
	Offset           int
	IncludeLibrary   bool
	IncludeBooks     bool
	IncludeFollowers bool
	IncludeStats     bool
	FollowedUserID   uuid.UUID
	SortBy           string
	SortIn           string
}

type GetCollectionOption struct {
	IncludeBookIDs bool
	IncludeStats   bool
	FollowedUserID uuid.UUID
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
	IncludeBook bool
	Limit       int
	Skip        int
	SortBy      string
	SortIn      string
	// For book
	BookTitle  string
	BookSortBy string
	BookSortIn string
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

// Collection usecase methods
func (u Usecase) ListCollections(ctx context.Context, opt ListCollectionsOption) ([]Collection, int, error) {
	if opt.IncludeStats {
		role, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
		if !ok {
			return nil, 0, fmt.Errorf("user role not found in context")
		}
		userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
		if !ok {
			return nil, 0, fmt.Errorf("user id not found in context")
		}
		switch role {
		case "SUPERADMIN", "ADMIN", "USER":
			opt.FollowedUserID = userID
		default:
			return nil, 0, fmt.Errorf("invalid user role: %s", role)
		}
	}

	collections, count, err := u.repo.ListCollections(ctx, opt)
	if err != nil {
		return nil, 0, err
	}

	var list []Collection
	for _, c := range collections {
		if c.Cover != "" {
			c.Cover = u.fileStorageProvider.GetPublicURL(c.Cover)
		}
		list = append(list, c)
	}

	return list, count, nil
}

func (u Usecase) GetCollectionByID(ctx context.Context, id uuid.UUID, opt GetCollectionOption) (Collection, error) {
	if opt.IncludeStats {
		role, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
		if !ok {
			return Collection{}, fmt.Errorf("user role not found in context")
		}
		userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
		if !ok {
			return Collection{}, fmt.Errorf("user id not found in context")
		}
		switch role {
		case "SUPERADMIN", "ADMIN", "USER":
			opt.FollowedUserID = userID
		default:
			return Collection{}, fmt.Errorf("invalid user role: %s", role)
		}
	}

	collection, err := u.repo.GetCollectionByID(ctx, id, opt)
	if err != nil {
		return Collection{}, err
	}

	if collection.Cover != "" {
		collection.Cover = u.fileStorageProvider.GetPublicURL(collection.Cover)
	}

	return collection, nil
}

func (u Usecase) CreateCollection(ctx context.Context, c Collection) (Collection, error) {
	if c.Cover != "" {
		if c.ID == uuid.Nil {
			c.ID = uuid.New()
		}

		coverPath := fmt.Sprintf("public/collections/%s/cover", c.ID.String())
		storedCoverPath, err := u.fileStorageProvider.CopyFilePreserveFilename(ctx, c.Cover, coverPath)
		if err != nil {
			log.Printf("CreateCollection: copy cover for collection %s failed: %v", c.ID, err)
			c.Cover = ""
		} else {
			c.Cover = storedCoverPath
		}
	}

	col, err := u.repo.CreateCollection(ctx, c)
	if err != nil {
		return Collection{}, err
	}

	if col.Cover != "" {
		col.Cover = u.fileStorageProvider.GetPublicURL(col.Cover)
	}
	return col, nil
}

type UpdateCollectionRequest struct {
	Title       string
	Description string
	UpdateCover *string
	Cover       *string
	Colors      json.RawMessage
}

func (u Usecase) UpdateCollection(ctx context.Context, id uuid.UUID, req UpdateCollectionRequest) (Collection, error) {
	if req.UpdateCover != nil {
		coverPath := fmt.Sprintf("public/collections/%s/cover", id.String())
		storedCoverPath, err := u.fileStorageProvider.CopyFilePreserveFilename(ctx, *req.UpdateCover, coverPath)
		if err != nil {
			log.Printf("err_UpdateCollection_fileStorageProvider.CopyFilePreserveFilename: %v", err)
			req.Cover = nil
		} else {
			req.Cover = &storedCoverPath
		}
	}

	c, err := u.repo.UpdateCollection(ctx, id, req)
	if err != nil {
		return Collection{}, err
	}
	if c.Cover != "" {
		c.Cover = u.fileStorageProvider.GetPublicURL(c.Cover)
	}
	return c, nil
}

func (u Usecase) DeleteCollection(ctx context.Context, id uuid.UUID) error {
	return u.repo.DeleteCollection(ctx, id)
}

func (u Usecase) ListCollectionBooks(ctx context.Context, id uuid.UUID, opt ListCollectionBooksOption) ([]CollectionBook, int, error) {
	list, count, err := u.repo.ListCollectionBooks(ctx, id, opt)
	if err != nil {
		return nil, 0, err
	}

	for i, cb := range list {
		if cb.Book != nil && cb.Book.Cover != "" {
			cb.Book.Cover = u.fileStorageProvider.GetPublicURL(cb.Book.Cover)
			list[i] = cb
		}
	}

	return list, count, nil
}

func (u Usecase) UpdateCollectionBooks(ctx context.Context, id uuid.UUID, bookIDs []uuid.UUID) ([]CollectionBook, error) {
	newIDs := make(map[uuid.UUID]struct{})
	for _, id := range bookIDs {
		newIDs[id] = struct{}{}
	}

	existing, _, err := u.repo.ListCollectionBooks(ctx, id, ListCollectionBooksOption{})
	if err != nil {
		return nil, err
	}

	existingIDs := make(map[uuid.UUID]struct{})
	for _, e := range existing {
		existingIDs[e.BookID] = struct{}{}
	}

	addedIDs := make([]uuid.UUID, 0)
	for nid := range newIDs {
		if _, found := existingIDs[nid]; !found {
			addedIDs = append(addedIDs, nid)
		}
	}

	removedIDs := make([]uuid.UUID, 0)
	for eid := range existingIDs {
		if _, found := newIDs[eid]; !found {
			removedIDs = append(removedIDs, eid)
		}
	}

	if len(removedIDs) > 0 {
		if err := u.repo.DeleteCollectionBooks(ctx, id, removedIDs); err != nil {
			return nil, err
		}
	}

	var created []CollectionBook
	if len(bookIDs) > 0 {
		created, err = u.repo.UpdateCollectionBooks(ctx, id, bookIDs)
		if err != nil {
			return nil, err
		}
	}

	if len(addedIDs) > 0 {
		go func(collectionID uuid.UUID, added []uuid.UUID) {
			bg := context.Background()

			collection, err := u.repo.GetCollectionByID(bg, collectionID, GetCollectionOption{})
			if err != nil {
				log.Printf("err_UpdateCollectionBooks_GetCollectionByID: %v", err)
				return
			}

			books, _, err := u.repo.ListBooks(bg, ListBooksOption{
				IDs: added,
			})
			if err != nil {
				log.Printf("err_UpdateCollectionBooks_ListBooks: %v", err)
				return
			}

			followers, _, err := u.repo.ListCollectionFollowers(bg, ListCollectionFollowersOption{
				CollectionID: collectionID,
			})
			if err != nil {
				log.Printf("err_UpdateCollectionBooks_ListCollectionFollowers: %v", err)
				return
			}

			titles := make([]string, 0, len(books))
			for _, b := range books {
				if b.Title != "" {
					titles = append(titles, b.Title)
				}
			}

			for _, follower := range followers {
				message := ""
				if len(titles) > 0 && len(titles) <= 3 {
					message = "New books added to " + collection.Title + ": " + strings.Join(titles, ", ")
				} else if len(titles) > 3 {
					message = fmt.Sprintf("New books added to %s: %d new books", collection.Title, len(titles))
				} else {
					message = fmt.Sprintf("New books added to %s", collection.Title)
				}

				if err := u.CreateNotification(bg, Notification{
					Title:         "Collection Updated",
					Message:       message,
					UserID:        follower.UserID,
					ReferenceID:   &collectionID,
					ReferenceType: "COLLECTION",
				}); err != nil {
					log.Printf("err_UpdateCollectionBooks_CreateNotification: %v", err)
				}
			}
		}(id, addedIDs)
	}

	if len(created) > 0 {
		return created, nil
	}
	return nil, nil
}

// CollectionFollower usecase methods
func (u Usecase) ListCollectionFollowers(ctx context.Context, opt ListCollectionFollowersOption) ([]CollectionFollower, int, error) {
	return u.repo.ListCollectionFollowers(ctx, opt)
}

func (u Usecase) CreateCollectionFollower(ctx context.Context, collectionID uuid.UUID) (CollectionFollower, error) {
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return CollectionFollower{}, fmt.Errorf("user id not found in context")
	}
	return u.repo.CreateCollectionFollower(ctx, CollectionFollower{
		CollectionID: collectionID,
		UserID:       userID,
	})
}

func (u Usecase) DeleteCollectionFollower(ctx context.Context, collectionID uuid.UUID) error {
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return fmt.Errorf("user id not found in context")
	}
	return u.repo.DeleteCollectionFollower(ctx, CollectionFollower{
		CollectionID: collectionID,
		UserID:       userID,
	})
}
