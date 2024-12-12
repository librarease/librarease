package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Library struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeleteAt  *time.Time
}
type ListLibrariesOption struct {
	Skip   int
	Limit  int
	Name   string
	IDs    uuid.UUIDs
	SortBy string
	SortIn string
}

func (u Usecase) ListLibraries(ctx context.Context, opt ListLibrariesOption) ([]Library, int, error) {
	libs, total, err := u.repo.ListLibraries(ctx, opt)
	if err != nil {
		return nil, 0, err
	}

	var userList []Library
	for _, lib := range libs {
		userList = append(userList, Library{
			ID:        lib.ID,
			Name:      lib.Name,
			CreatedAt: lib.CreatedAt,
			UpdatedAt: lib.UpdatedAt,
			DeleteAt:  lib.DeleteAt,
		})
	}

	return userList, total, nil
}

func (u Usecase) GetLibraryByID(ctx context.Context, id string) (Library, error) {
	err := uuid.Validate(id)
	if err != nil {
		return Library{}, err
	}
	lib, err := u.repo.GetLibraryByID(ctx, id)
	if err != nil {
		return Library{}, err
	}

	return Library{
		ID:        lib.ID,
		Name:      lib.Name,
		CreatedAt: lib.CreatedAt,
		UpdatedAt: lib.UpdatedAt,
		DeleteAt:  lib.DeleteAt,
	}, nil
}

func (u Usecase) CreateLibrary(ctx context.Context, library Library) (Library, error) {
	lib, err := u.repo.CreateLibrary(ctx, library)
	if err != nil {
		return Library{}, err
	}

	return Library{
		ID:        lib.ID,
		Name:      lib.Name,
		CreatedAt: lib.CreatedAt,
		UpdatedAt: lib.UpdatedAt,
	}, nil
}

func (u Usecase) UpdateLibrary(ctx context.Context, library Library) (Library, error) {
	lib, err := u.repo.UpdateLibrary(ctx, library)
	if err != nil {
		return Library{}, err
	}

	return Library{
		ID:        lib.ID,
		Name:      lib.Name,
		CreatedAt: lib.CreatedAt,
		UpdatedAt: lib.UpdatedAt,
	}, nil
}

func (u Usecase) DeleteLibrary(ctx context.Context, id string) error {
	err := uuid.Validate(id)
	if err != nil {
		return err
	}
	err = u.repo.DeleteLibrary(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
