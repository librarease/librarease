package usecase

import (
	"context"
	"fmt"
	"librarease/internal/config"
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

	role, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
	if !ok {
		return Library{}, fmt.Errorf("user role not found in context")
	}
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return Library{}, fmt.Errorf("user id not found in context")
	}

	switch role {
	case "SUPERADMIN":
		// ALLOW
	case "ADMIN":
		// ALLlOW
	case "USER":
		staffs, _, err := u.ListStaffs(ctx, ListStaffsOption{
			UserID:    userID.String(),
			LibraryID: library.ID.String(),
		})
		if err != nil {
			return Library{}, err
		}
		if len(staffs) == 0 {
			// TODO: implement error
			return Library{}, fmt.Errorf("you are not staff of this library")
		}

		if staffs[0].Role != StaffRoleAdmin {
			// TODO: implement error
			return Library{}, fmt.Errorf("you are not allowed to update library")
		}

		// if len(opt.IDs) > 0 {
		// 	for _, id := range opt.IDs {
		// 		if slices.Contains(staffLibIDs, id) {
		// 			opt.IDs = append(opt.IDs, id)
		// 		}
		// 	}
		// } else {
		// 	opt.IDs = staffLibIDs
		// }

	}
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

	role, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
	if !ok {
		return fmt.Errorf("user role not found in context")
	}

	switch role {
	case "SUPERADMIN":
		// ALLOW
	case "ADMIN":
		// ALLlOW
	default:
		// TODO: implement error
		return fmt.Errorf("you are not allowed to delete library")
	}

	err = u.repo.DeleteLibrary(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
