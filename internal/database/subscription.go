package database

import (
	"context"
	"librarease/internal/usecase"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Subscription represents user's purchase of a membership
type Subscription struct {
	ID           uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID       uuid.UUID       `gorm:"column:user_id;type:uuid;"`
	User         *User           `gorm:"foreignKey:UserID;references:ID"`
	MembershipID uuid.UUID       `gorm:"column:membership_id;type:uuid;"`
	Membership   *Membership     `gorm:"foreignKey:MembershipID;references:ID"`
	CreatedAt    time.Time       `gorm:"column:created_at"`
	UpdatedAt    time.Time       `gorm:"column:updated_at"`
	DeletedAt    *gorm.DeletedAt `gorm:"column:deleted_at"`
	Borrowings   []Borrowing

	// Granfathering the membership
	ExpiresAt       time.Time `gorm:"column:expires_at"`
	Amount          int       `gorm:"column:amount;type:int"`
	FinePerDay      int       `gorm:"column:fine_per_day;type:int"`
	LoanPeriod      int       `gorm:"column:loan_period;type:int"`
	ActiveLoanLimit int       `gorm:"column:active_loan_limit;type:int"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}

func (s *service) ListSubscriptions(ctx context.Context, opt usecase.ListSubscriptionsOption) ([]usecase.Subscription, int, error) {
	var (
		subs  []Subscription
		usubs []usecase.Subscription
		count int64
	)

	db := s.db.Model([]Subscription{}).WithContext(ctx)

	if opt.UserID != "" {
		db = db.Where("user_id = ?", opt.UserID)
	}
	if opt.MembershipID != "" {
		db = db.Where("membership_id = ?", opt.MembershipID)
	}
	if len(opt.LibraryIDs) > 0 {
		db = db.Joins("Membership").Where("library_id IN ?", opt.LibraryIDs)
	}
	if opt.IsActive {
		db = db.Where("expires_at > ?", time.Now())
	}
	if opt.MembershipName != "" {
		db = db.
			Joins("JOIN memberships m ON subscriptions.membership_id = m.id").
			Where("m.name ILIKE ?", "%"+opt.MembershipName+"%")
	}

	err := db.
		Preload("User").
		Preload("Membership").
		Preload("Membership.Library").
		Count(&count).
		Limit(opt.Limit).
		Offset(opt.Skip).
		Find(&subs).
		Error

	if err != nil {
		return nil, 0, err
	}

	for _, sub := range subs {
		usub := sub.ConvertToUsecase()
		if sub.User != nil {
			user := sub.User.ConvertToUsecase()
			usub.User = &user
		}
		if sub.Membership != nil {
			mem := sub.Membership.ConvertToUsecase()
			usub.Membership = &mem
			if sub.Membership.Library != nil {
				lib := sub.Membership.Library.ConvertToUsecase()
				mem.Library = &lib
			}
		}
		usubs = append(usubs, usub)
	}

	return usubs, int(count), nil
}

func (s *service) CreateSubscription(ctx context.Context, sub usecase.Subscription) (usecase.Subscription, error) {
	d := Subscription{
		ID:              sub.ID,
		UserID:          sub.UserID,
		MembershipID:    sub.MembershipID,
		CreatedAt:       sub.CreatedAt,
		UpdatedAt:       sub.UpdatedAt,
		ExpiresAt:       sub.ExpiresAt,
		Amount:          sub.Amount,
		FinePerDay:      sub.FinePerDay,
		LoanPeriod:      sub.LoanPeriod,
		ActiveLoanLimit: sub.ActiveLoanLimit,
	}
	err := s.db.
		WithContext(ctx).
		Create(&d).
		Error
	if err != nil {
		return usecase.Subscription{}, err
	}
	return d.ConvertToUsecase(), nil
}

func (s *service) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (usecase.Subscription, error) {
	var sub Subscription
	err := s.db.
		WithContext(ctx).
		Preload("User").
		Preload("Membership").
		Preload("Membership.Library").
		Where("id = ?", id).
		First(&sub).
		Error
	if err != nil {
		return usecase.Subscription{}, err
	}
	usub := sub.ConvertToUsecase()
	if sub.User != nil {
		user := sub.User.ConvertToUsecase()
		usub.User = &user
	}
	if sub.Membership != nil {
		mem := sub.Membership.ConvertToUsecase()
		usub.Membership = &mem
		if sub.Membership.Library != nil {
			lib := sub.Membership.Library.ConvertToUsecase()
			mem.Library = &lib
		}
	}
	return usub, nil
}

func (s *service) UpdateSubscription(ctx context.Context, sub usecase.Subscription) (usecase.Subscription, error) {
	d := Subscription{
		ID:              sub.ID,
		UserID:          sub.UserID,
		MembershipID:    sub.MembershipID,
		CreatedAt:       sub.CreatedAt,
		UpdatedAt:       sub.UpdatedAt,
		ExpiresAt:       sub.ExpiresAt,
		Amount:          sub.Amount,
		FinePerDay:      sub.FinePerDay,
		LoanPeriod:      sub.LoanPeriod,
		ActiveLoanLimit: sub.ActiveLoanLimit,
	}
	err := s.db.
		WithContext(ctx).
		Updates(&d).
		Error
	if err != nil {
		return usecase.Subscription{}, err
	}
	return d.ConvertToUsecase(), nil
}

// Convert core model to Usecase
func (s Subscription) ConvertToUsecase() usecase.Subscription {
	var d *time.Time
	if s.DeletedAt != nil {
		d = &s.DeletedAt.Time
	}
	return usecase.Subscription{
		ID:              s.ID,
		UserID:          s.UserID,
		MembershipID:    s.MembershipID,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
		DeletedAt:       d,
		ExpiresAt:       s.ExpiresAt,
		Amount:          s.Amount,
		FinePerDay:      s.FinePerDay,
		LoanPeriod:      s.LoanPeriod,
		ActiveLoanLimit: s.ActiveLoanLimit,
	}
}
