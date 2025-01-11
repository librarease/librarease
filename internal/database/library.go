package database

import (
	"context"
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Library struct {
	ID          uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name        string          `gorm:"column:name;type:varchar(255)"`
	Logo        string          `gorm:"column:logo;type:varchar(255)"`
	Address     string          `gorm:"column:address;type:varchar(255)"`
	Phone       string          `gorm:"column:phone;type:varchar(255)"`
	Email       string          `gorm:"column:email;type:varchar(255)"`
	Description string          `gorm:"column:description;type:text"`
	CreatedAt   time.Time       `gorm:"column:created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at"`
	DeletedAt   *gorm.DeletedAt `gorm:"column:deleted_at"`

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

	err := db.
		Count(&count).
		Offset(opt.Skip).
		Limit(opt.Limit).
		Order(orderBy + " " + orderIn).
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

func (s *service) GetLibraryByID(ctx context.Context, id uuid.UUID) (usecase.Library, error) {
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
		ID:          library.ID,
		Name:        library.Name,
		Logo:        library.Logo,
		Address:     library.Address,
		Phone:       library.Phone,
		Email:       library.Email,
		Description: library.Description,
	}

	err := s.db.WithContext(ctx).Model(&l).Clauses(clause.Returning{}).Create(&l).Error
	if err != nil {
		return usecase.Library{}, err
	}

	return usecase.Library{
		ID:          l.ID,
		Name:        l.Name,
		Logo:        l.Logo,
		Address:     l.Address,
		Phone:       l.Phone,
		Email:       l.Email,
		Description: l.Description,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}, nil
}

func (s *service) UpdateLibrary(ctx context.Context, id uuid.UUID, library usecase.Library) (usecase.Library, error) {
	l := Library{
		Name:        library.Name,
		Logo:        library.Logo,
		Address:     library.Address,
		Phone:       library.Phone,
		Email:       library.Email,
		Description: library.Description,
	}

	err := s.db.WithContext(ctx).Model(&l).Clauses(clause.Returning{}).Where("id = ?", id).Updates(&l).Error
	if err != nil {
		return usecase.Library{}, err
	}

	return usecase.Library{
		ID:          l.ID,
		Name:        l.Name,
		Logo:        l.Logo,
		Address:     l.Address,
		Phone:       l.Phone,
		Email:       l.Email,
		Description: l.Description,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}, nil
}

func (s *service) DeleteLibrary(ctx context.Context, id uuid.UUID) error {
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
		ID:          lib.ID,
		Name:        lib.Name,
		Logo:        lib.Logo,
		Address:     lib.Address,
		Phone:       lib.Phone,
		Email:       lib.Email,
		Description: lib.Description,
		CreatedAt:   lib.CreatedAt,
		UpdatedAt:   lib.UpdatedAt,
		DeleteAt:    d,
	}
}
