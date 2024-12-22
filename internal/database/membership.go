package database

import (
	"context"
	"librarease/internal/usecase"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Membership struct {
	ID              uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name            string          `gorm:"column:name;type:varchar(255)"`
	LibraryID       uuid.UUID       `gorm:"column:library_id;type:uuid;"`
	Library         *Library        `gorm:"foreignKey:LibraryID;references:ID"`
	Duration        int             `gorm:"column:duration;type:int"`
	ActiveLoanLimit int             `gorm:"column:active_loan_limit;type:int"`
	LoanPeriod      int             `gorm:"column:loan_period;type:int"`
	FinePerDay      int             `gorm:"column:fine_per_day;type:int"`
	CreatedAt       time.Time       `gorm:"column:created_at"`
	UpdatedAt       time.Time       `gorm:"column:updated_at"`
	DeletedAt       *gorm.DeletedAt `gorm:"column:deleted_at"`

	Subscriptions []Subscription
}

func (Membership) TableName() string {
	return "memberships"
}

func (s *service) ListMemberships(ctx context.Context, opt usecase.ListMembershipsOption) ([]usecase.Membership, int, error) {
	var (
		mems  []Membership
		ums   []usecase.Membership
		count int64
	)

	db := s.db.Model([]Membership{}).WithContext(ctx)

	if opt.Name != "" {
		db = db.Where("memberships.name ILIKE ?", "%"+opt.Name+"%")
	}

	if len(opt.LibraryIDs) > 0 {
		db = db.Where("l.id IN ?", opt.LibraryIDs)
	}

	err := db.
		Joins("JOIN libraries l on l.id = memberships.library_id AND l.deleted_at IS NULL").
		Count(&count).
		Limit(opt.Limit).
		Offset(opt.Skip).
		Find(&mems).
		Error

	if err != nil {
		return nil, 0, err
	}

	for _, mem := range mems {
		umem := mem.ConvertToUsecase()
		if mem.Library != nil {
			lib := mem.Library.ConvertToUsecase()
			umem.Library = &lib
		}
		ums = append(ums, umem)
	}

	return ums, int(count), nil
}

func (s *service) GetMembershipByID(ctx context.Context, id uuid.UUID) (usecase.Membership, error) {
	var m Membership
	err := s.db.
		WithContext(ctx).
		Preload("Library").
		Where("id = ?", id).
		First(&m).
		Error
	if err != nil {
		return usecase.Membership{}, err
	}

	mem := m.ConvertToUsecase()
	if m.Library != nil {
		lib := m.Library.ConvertToUsecase()
		mem.Library = &lib
	}
	return mem, nil
}

func (s *service) CreateMembership(ctx context.Context, m usecase.Membership) (usecase.Membership, error) {
	mem := Membership{
		Name:            m.Name,
		LibraryID:       m.LibraryID,
		Duration:        m.Duration,
		ActiveLoanLimit: m.ActiveLoanLimit,
		LoanPeriod:      m.LoanPeriod,
		FinePerDay:      m.FinePerDay,
	}

	if err := s.db.WithContext(ctx).Create(&mem).Error; err != nil {
		return usecase.Membership{}, err
	}

	return mem.ConvertToUsecase(), nil
}

func (s *service) UpdateMembership(ctx context.Context, m usecase.Membership) (usecase.Membership, error) {
	mem := Membership{
		ID:              m.ID,
		Name:            m.Name,
		LibraryID:       m.LibraryID,
		Duration:        m.Duration,
		ActiveLoanLimit: m.ActiveLoanLimit,
		LoanPeriod:      m.LoanPeriod,
		FinePerDay:      m.FinePerDay,
	}

	err := s.db.WithContext(ctx).Updates(&mem).Error
	if err != nil {
		return usecase.Membership{}, err
	}

	return mem.ConvertToUsecase(), nil
}

func (s *service) DeleteMembership(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Delete(&Membership{}, id).Error
}

// Convert core model to Usecase
func (m Membership) ConvertToUsecase() usecase.Membership {
	var d *time.Time
	if m.DeletedAt != nil {
		d = &m.DeletedAt.Time
	}
	return usecase.Membership{
		ID:              m.ID,
		Name:            m.Name,
		LibraryID:       m.LibraryID,
		Duration:        m.Duration,
		ActiveLoanLimit: m.ActiveLoanLimit,
		LoanPeriod:      m.LoanPeriod,
		FinePerDay:      m.FinePerDay,
		CreatedAt:       m.CreatedAt,
		DeletedAt:       d,
	}
}
