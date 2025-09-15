package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"log"
	"net/http"
	"time"

	"github.com/cenkalti/dominantcolor"
	"github.com/google/uuid"

	_ "image/jpeg"
	_ "image/png"
)

// Collection structures
type Collection struct {
	ID          uuid.UUID
	LibraryID   uuid.UUID
	Title       string
	Cover       *Asset
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Description string

	Library *Library

	BookCount     int
	FollowerCount int
}

type ListCollectionsOption struct {
	LibraryID        uuid.UUID
	Title            string
	Limit            int
	Offset           int
	IncludeLibrary   bool
	IncludeBooks     bool
	IncludeFollowers bool
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
	return u.repo.ListCollections(ctx, opt)
}

func (u Usecase) GetCollectionByID(ctx context.Context, id uuid.UUID) (Collection, error) {
	return u.repo.GetCollectionByID(ctx, id)
}

func (u Usecase) CreateCollection(ctx context.Context, c Collection) (Collection, error) {

	if c.Cover != nil {
		temp := u.fileStorageProvider.TempPath()
		url, err := u.fileStorageProvider.GetPresignedURL(ctx, temp+"/"+c.Cover.Path)
		if err != nil {
			return Collection{}, err
		}
		colors, err := ExtractColors(ctx, url)
		if err != nil {
			return Collection{}, err
		}
		c.Cover.Colors = colors
	}

	col, err := u.repo.CreateCollection(ctx, c)
	if err != nil {
		return Collection{}, err
	}

	if c.Cover != nil {
		if err := u.fileStorageProvider.MoveTempFilePublic(ctx, c.Cover.Path, "collections/covers"); err != nil {
			log.Printf("err_CreateCollection_fileStorageProvider.MoveTempFilePublic: %v", err)
			return col, nil
		}

	}
	return col, nil
}

type UpdateCollectionRequest struct {
	Title       string
	Description string
	UpdateCover *string
	Cover       *Asset
}

func (u Usecase) UpdateCollection(ctx context.Context, id uuid.UUID, req UpdateCollectionRequest) (Collection, error) {
	if req.UpdateCover != nil {
		err := u.fileStorageProvider.MoveTempFilePublic(ctx, *req.UpdateCover, "collections/covers")
		if err != nil {
			log.Printf("err_UpdateCollection_fileStorageProvider.MoveTempFilePublic: %v", err)
			return Collection{}, err
		}
		publicPrefix, _ := u.fileStorageProvider.GetPublicURL(ctx)

		colors, err := ExtractColors(ctx, publicPrefix+"/collections/covers/"+*req.UpdateCover)
		if err != nil {
			log.Printf("err_UpdateCollection_ExtractColors: %v", err)
			return Collection{}, err
		}
		req.Cover = &Asset{
			Path:   *req.UpdateCover,
			Colors: colors,
		}
	}
	return u.repo.UpdateCollection(ctx, id, req)
}

func (u Usecase) DeleteCollection(ctx context.Context, id uuid.UUID) error {
	return u.repo.DeleteCollection(ctx, id)
}

func (u Usecase) ListCollectionBooks(ctx context.Context, id uuid.UUID, opt ListCollectionBooksOption) ([]CollectionBook, int, error) {
	list, count, err := u.repo.ListCollectionBooks(ctx, id, opt)
	if err != nil {
		return nil, 0, err
	}

	publicURL, _ := u.fileStorageProvider.GetPublicURL(ctx)

	for i, cb := range list {
		if cb.Book != nil && cb.Book.Cover != "" {
			cb.Book.Cover = fmt.Sprintf("%s/books/%s/cover/%s", publicURL, cb.Book.ID, cb.Book.Cover)
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

	if len(bookIDs) > 0 {
		return u.repo.UpdateCollectionBooks(ctx, id, bookIDs)
	}

	return nil, nil
}

// // CollectionFollower usecase methods
// func (u Usecase) ListCollectionFollowers(ctx context.Context, opt ListCollectionFollowersOption) ([]CollectionFollower, int, error) {
// 	return u.repo.ListCollectionFollowers(ctx, opt)
// }

// func (u Usecase) CreateCollectionFollower(ctx context.Context, cf CollectionFollower) (CollectionFollower, error) {
// 	return u.repo.CreateCollectionFollower(ctx, cf)
// }

// func (u Usecase) DeleteCollectionFollower(ctx context.Context, id uuid.UUID) error {
// 	return u.repo.DeleteCollectionFollower(ctx, id)
// }

func ExtractColors(ctx context.Context, url string) ([]byte, error) {
	f, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer f.Body.Close()

	img, _, err := image.Decode(f.Body)
	if err != nil {
		return nil, err
	}

	colors := make(map[int][4]uint8)
	dColors := dominantcolor.FindN(img, 4)
	for i, color := range dColors {
		colors[i] = [4]uint8{color.R, color.G, color.B, color.A}
	}

	return json.Marshal(colors)
}
