package database

import (
	"context"
	"librarease/internal/usecase"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Library struct {
	ID        uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name      string          `gorm:"column:name;type:varchar(255)"`
	CreatedAt time.Time       `gorm:"column:created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"column:deleted_at"`

	Staffs      []Staff
	Books       []Book
	Memberships []Membership
}

func (Library) TableName() string {
	return "libraries"
}

func (s *service) ListLibraries(ctx context.Context, opt usecase.ListLibrariesOption) ([]usecase.Library, int, error) {
	var (
		libs  []Library
		ulibs []usecase.Library
		count int64
	)

	db := s.db.Model([]Library{}).WithContext(ctx)

	if opt.Name != "" {
		db = db.Where("name ILIKE ?", "%"+opt.Name+"%")
	}

	err := db.
		Count(&count).
		Offset(opt.Skip).
		Limit(opt.Limit).
		Find(&libs).
		Error

	if err != nil {
		return nil, 0, err
	}

	for _, l := range libs {
		ul := l.ConvertToUsecase()
		ulibs = append(ulibs, ul)
	}

	return ulibs, int(count), nil
}

func (s *service) GetLibraryByID(ctx context.Context, id string) (usecase.Library, error) {
	var l Library

	err := s.db.WithContext(ctx).Where("id = ?", id).First(&l).Error
	if err != nil {
		return usecase.Library{}, err
	}

	lib := l.ConvertToUsecase()

	return lib, nil
}

func (s *service) CreateLibrary(ctx context.Context, library usecase.Library) (usecase.Library, error) {
	l := Library{
		ID:        library.ID,
		Name:      library.Name,
		CreatedAt: library.CreatedAt,
		UpdatedAt: library.UpdatedAt,
	}

	err := s.db.WithContext(ctx).Create(&l).Error
	if err != nil {
		return usecase.Library{}, err
	}

	return usecase.Library{
		ID:        l.ID,
		Name:      l.Name,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}, nil
}

func (s *service) UpdateLibrary(ctx context.Context, library usecase.Library) (usecase.Library, error) {
	l := Library{
		ID:        library.ID,
		Name:      library.Name,
		CreatedAt: library.CreatedAt,
		UpdatedAt: library.UpdatedAt,
	}

	err := s.db.WithContext(ctx).Save(&l).Error
	if err != nil {
		return usecase.Library{}, err
	}

	return usecase.Library{
		ID:        l.ID,
		Name:      l.Name,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}, nil
}

func (s *service) DeleteLibrary(ctx context.Context, id string) error {
	err := s.db.WithContext(ctx).Where("id = ?", id).Delete(&Library{}).Error
	if err != nil {
		return err
	}

	return nil
}

// Convert core model to Usecase
func (lib Library) ConvertToUsecase() usecase.Library {
	var d *time.Time
	if lib.DeletedAt != nil {
		d = &lib.DeletedAt.Time
	}
	return usecase.Library{
		ID:        lib.ID,
		Name:      lib.Name,
		CreatedAt: lib.CreatedAt,
		UpdatedAt: lib.UpdatedAt,
		DeleteAt:  d,
	}
}
