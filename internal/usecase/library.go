package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/librarease/librarease/internal/config"

	"github.com/google/uuid"
)

type Library struct {
	ID          uuid.UUID
	Name        string
	Logo        string
	Address     string
	Phone       string
	Email       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeleteAt    *time.Time

	// UpdateLogo is used to update logo
	UpdateLogo *string
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

	var libraries []Library
	publicURL, _ := u.fileStorageProvider.GetPublicURL(ctx)

	for _, lib := range libs {

		var logo string
		if lib.Logo != "" {
			logo = fmt.Sprintf("%s/libraries/%s/logo/%s", publicURL, lib.ID, lib.Logo)
		}

		libraries = append(libraries, Library{
			ID:          lib.ID,
			Name:        lib.Name,
			Logo:        logo,
			Address:     lib.Address,
			Phone:       lib.Phone,
			Email:       lib.Email,
			Description: lib.Description,
			CreatedAt:   lib.CreatedAt,
			UpdatedAt:   lib.UpdatedAt,
			DeleteAt:    lib.DeleteAt,
		})
	}

	return libraries, total, nil
}

func (u Usecase) GetLibraryByID(ctx context.Context, id uuid.UUID) (Library, error) {

	lib, err := u.repo.GetLibraryByID(ctx, id)
	if err != nil {
		return Library{}, err
	}

	var logo string
	publicURL, _ := u.fileStorageProvider.GetPublicURL(ctx)
	if lib.Logo != "" {
		logo = fmt.Sprintf("%s/libraries/%s/logo/%s", publicURL, lib.ID, lib.Logo)
	}

	return Library{
		ID:          lib.ID,
		Name:        lib.Name,
		Logo:        logo,
		Address:     lib.Address,
		Phone:       lib.Phone,
		Email:       lib.Email,
		Description: lib.Description,
		CreatedAt:   lib.CreatedAt,
		UpdatedAt:   lib.UpdatedAt,
		DeleteAt:    lib.DeleteAt,
	}, nil
}

func (u Usecase) CreateLibrary(ctx context.Context, library Library) (Library, error) {
	lib, err := u.repo.CreateLibrary(ctx, library)
	if err != nil {
		return Library{}, err
	}

	var logo = library.Logo
	if logo != "" {
		var logoPath = fmt.Sprintf("libraries/%s/logo", lib.ID.String())
		err = u.fileStorageProvider.MoveTempFilePublic(ctx, logo, logoPath)
		if err != nil {
			fmt.Printf("failed to move file for lib %s: %v\n", lib.ID, err)
			// don't save logo if failed to move file
			logo = ""
		}
	}
	publicURL, _ := u.fileStorageProvider.GetPublicURL(ctx)
	if logo != "" {
		logo = fmt.Sprintf("%s/libraries/%s/logo/%s", publicURL, lib.ID, lib.Logo)
	}

	return Library{
		ID:          lib.ID,
		Name:        lib.Name,
		Logo:        logo,
		Address:     lib.Address,
		Phone:       lib.Phone,
		Email:       lib.Email,
		Description: lib.Description,
	}, nil
}

func (u Usecase) UpdateLibrary(ctx context.Context, id uuid.UUID, library Library) (Library, error) {

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
		staffs, _, err := u.repo.ListStaffs(ctx, ListStaffsOption{
			UserID:     userID.String(),
			LibraryIDs: uuid.UUIDs{library.ID},
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

	if library.UpdateLogo != nil {
		logoPath := fmt.Sprintf("libraries/%s/logo", id)
		err := u.fileStorageProvider.MoveTempFilePublic(ctx, *library.UpdateLogo, logoPath)
		if err != nil {
			fmt.Printf("failed to move file for lib %s: %v\n", id, err)
			return Library{}, err
		}
		library.Logo = *library.UpdateLogo
	}

	lib, err := u.repo.UpdateLibrary(ctx, id, library)
	if err != nil {
		return Library{}, err
	}

	var logo string
	publicURL, _ := u.fileStorageProvider.GetPublicURL(ctx)
	if lib.Logo != "" {
		logo = fmt.Sprintf("%s/libraries/%s/logo/%s", publicURL, lib.ID, lib.Logo)
	}

	return Library{
		ID:          lib.ID,
		Name:        lib.Name,
		Logo:        logo,
		Address:     lib.Address,
		Phone:       lib.Phone,
		Email:       lib.Email,
		Description: lib.Description,
		CreatedAt:   lib.CreatedAt,
		UpdatedAt:   lib.UpdatedAt,
	}, nil
}

func (u Usecase) DeleteLibrary(ctx context.Context, id uuid.UUID) error {

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

	err := u.repo.DeleteLibrary(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
