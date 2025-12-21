package database

import (
	"context"
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Subscription represents user's purchase of a membership
type Subscription struct {
	ID           uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID       uuid.UUID       `gorm:"column:user_id;type:uuid;"`
	User         *User           `gorm:"foreignKey:UserID;references:ID"`
	MembershipID uuid.UUID       `gorm:"column:membership_id;type:uuid;"`
	Membership   *Membership     `gorm:"foreignKey:MembershipID;references:ID"`
	SubscribedAt time.Time       `gorm:"column:subscribed_at;default:now()"`
	Note         *string         `gorm:"column:note;type:text"`
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
	UsageLimit      int       `gorm:"column:usage_limit;type:int"`
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

	var (
		now                time.Time
		usageCountSubQuery *gorm.DB
	)
	if opt.IsActive || opt.IsExpired {
		now = time.Now()
		usageCountSubQuery = s.db.
			WithContext(ctx).
			Model(&Borrowing{}).
			Select("COUNT(*)").
			Where("borrowings.subscription_id = subscriptions.id").
			Where("borrowings.deleted_at IS NULL")
	}

	if opt.ID != "" {
		db = db.Where("subscriptions.id::text ILIKE ?", "%"+opt.ID+"%")
	}

	if opt.UserID != "" {
		db = db.Where("user_id = ?", opt.UserID)
	}
	if opt.MembershipID != "" {
		db = db.Where("membership_id = ?", opt.MembershipID)
	}
	if opt.IsActive {
		db = db.Where("(expires_at > ?) AND (subscriptions.usage_limit <= 0 OR (subscriptions.usage_limit > 0 AND (?) < subscriptions.usage_limit))", now, usageCountSubQuery)
	}
	if opt.IsExpired {
		db = db.Where("(expires_at <= ?) OR (subscriptions.usage_limit > 0 AND (?) >= subscriptions.usage_limit)", now, usageCountSubQuery)
	}
	if len(opt.LibraryIDs) > 0 {
		db = db.Joins("Membership").
			Where("library_id IN ?", opt.LibraryIDs)
	}
	if opt.MembershipName != "" {
		db = db.
			Joins("Membership").
			Where("name ILIKE ?", "%"+opt.MembershipName+"%")
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

	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if opt.Limit > 0 {
		db = db.Limit(opt.Limit)
	}

	if opt.Skip > 0 {
		db = db.Offset(opt.Skip)
	}

	if err := db.
		Preload("User").
		Preload("Membership").
		Preload("Membership.Library").
		Order(orderBy + " " + orderIn).
		Find(&subs).
		Error; err != nil {

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
		UsageLimit:      sub.UsageLimit,
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

	var borrowingsCount *int
	if err := s.db.
		WithContext(ctx).
		Table("subscriptions s").
		Select("COUNT(b.id)").
		Joins("JOIN borrowings b ON s.id = b.subscription_id AND b.deleted_at IS NULL").
		Where("s.id = ?", id).
		Scan(&borrowingsCount).
		Error; err != nil {
		return usecase.Subscription{}, err
	}

	var activeLoanCount *int
	if err := s.db.
		WithContext(ctx).
		Table("subscriptions s").
		Select("COUNT(b.id)").
		Joins("JOIN borrowings b ON s.id = b.subscription_id").
		Joins("LEFT JOIN returnings r ON b.id = r.borrowing_id AND r.deleted_at IS NULL").
		Joins("LEFT JOIN losts l ON b.id = l.borrowing_id AND l.deleted_at IS NULL").
		Where("s.id = ?", id).
		Where("r.id IS NULL").
		Where("l.id IS NULL").
		Scan(&activeLoanCount).
		Error; err != nil {
		return usecase.Subscription{}, err
	}

	usub := sub.ConvertToUsecase()
	usub.UsageCount = borrowingsCount
	usub.ActiveLoanCount = activeLoanCount
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
		UsageLimit:      sub.UsageLimit,
	}
	err := s.db.
		WithContext(ctx).
		Clauses(clause.Returning{}).
		Select("expires_at", "amount", "fine_per_day", "loan_period", "active_loan_limit", "usage_limit").
		Updates(&d).
		Error
	if err != nil {
		return usecase.Subscription{}, err
	}
	return d.ConvertToUsecase(), nil
}

func (s *service) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	return s.db.
		WithContext(ctx).
		Where("id = ?", id).
		Delete(&Subscription{}).
		Error
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
		UsageLimit:      s.UsageLimit,
		SubscribedAt:    s.SubscribedAt,
		Note:            s.Note,
	}
}
