package database

import (
	"context"
	"librarease/internal/usecase"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Borrowing struct {
	ID             uuid.UUID     `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	BookID         uuid.UUID     `gorm:"column:book_id;type:uuid;"`
	Book           *Book         `gorm:"foreignKey:BookID;references:ID"`
	SubscriptionID uuid.UUID     `gorm:"column:subscription_id;type:uuid;"`
	Subscription   *Subscription `gorm:"foreignKey:SubscriptionID;references:ID"`
	StaffID        uuid.UUID     `gorm:"column:staff_id;type:uuid;"`
	Staff          *Staff        `gorm:"foreignKey:StaffID;references:ID"`
	BorrowedAt     time.Time     `gorm:"column:borrowed_at;default:now()"`
	DueAt          time.Time     `gorm:"column:due_at"`
	ReturnedAt     *time.Time    `gorm:"column:returned_at"`
	CreatedAt      time.Time     `gorm:"column:created_at"`
	UpdatedAt      time.Time     `gorm:"column:updated_at"`
	DeletedAt      *gorm.DeletedAt
}

func (Borrowing) TableName() string {
	return "borrowings"
}

// NOTE: would need to optimize this query
func (s *service) ListBorrowings(ctx context.Context, opt usecase.ListBorrowingsOption) ([]usecase.Borrowing, int, error) {
	var (
		borrows  []Borrowing
		uborrows []usecase.Borrowing
		count    int64
	)

	db := s.db.Model([]Borrowing{}).WithContext(ctx)

	if opt.BookID != "" {
		db = db.Where("book_id = ?", opt.BookID)
	}
	if opt.SubscriptionID != "" {
		db = db.Where("subscription_id = ?", opt.SubscriptionID)
	}
	if opt.StaffID != "" {
		db = db.Where("staff_id = ?", opt.StaffID)
	}
	if !opt.BorrowedAt.IsZero() {
		db = db.Where("borrowed_at = ?", opt.BorrowedAt)
	}
	if !opt.DueAt.IsZero() {
		db = db.Where("due_at = ?", opt.DueAt)
	}
	if opt.ReturnedAt != nil {
		db = db.Where("returned_at = ?", opt.ReturnedAt)
	}
	if opt.IsActive {
		db = db.Where("returned_at IS NULL")
	}
	if opt.IsExpired {
		db = db.Where("due_at < now() AND returned_at IS NULL")
	}
	if opt.MembershipID != "" {
		db = db.Joins("Subscription").Where("membership_id = ?", opt.MembershipID)
	}
	if opt.UserID != "" {
		db = db.Joins("Subscription").Where("user_id = ?", opt.UserID)
	}
	if opt.LibraryID != "" {
		db = db.Joins("Book").Where("library_id = ?", opt.LibraryID)
		// db = db.Joins("JOIN subscriptions s ON borrowings.subscription_id = s.id").
		// 	Joins("JOIN memberships m ON s.membership_id = m.id").
		// 	Where("m.library_id = ?", opt.LibraryID)
	}

	err := db.
		Preload("Book").
		Preload("Staff").
		Preload("Subscription").
		Preload("Subscription.User").
		Preload("Subscription.Membership").
		Preload("Subscription.Membership.Library").
		Count(&count).
		Limit(opt.Limit).
		Offset(opt.Skip).
		Find(&borrows).
		Error

	if err != nil {
		return nil, 0, err
	}

	for _, b := range borrows {
		ub := b.ConvertToUsecase()

		if b.Book.ID != uuid.Nil {
			book := b.Book.ConvertToUsecase()
			ub.Book = &book
		}

		if b.Subscription != nil {
			sub := b.Subscription.ConvertToUsecase()
			ub.Subscription = &sub

			if b.Subscription.User != nil {
				user := b.Subscription.User.ConvertToUsecase()
				sub.User = &user
			}

			if b.Subscription.Membership != nil {
				mem := b.Subscription.Membership.ConvertToUsecase()
				sub.Membership = &mem

				if b.Subscription.Membership.Library != nil {
					lib := b.Subscription.Membership.Library.ConvertToUsecase()
					mem.Library = &lib
				}
			}

		}
		if b.Staff.ID != uuid.Nil {
			staff := b.Staff.ConvertToUsecase()
			ub.Staff = &staff
		}
		uborrows = append(uborrows, ub)
	}

	return uborrows, int(count), nil
}

func (s *service) GetBorrowingByID(ctx context.Context, id uuid.UUID) (usecase.Borrowing, error) {
	var b Borrowing

	err := s.db.
		Model(Borrowing{}).
		WithContext(ctx).
		Preload("Book").
		Preload("Staff").
		Preload("Subscription").
		Preload("Subscription.User").
		Preload("Subscription.Membership").
		Preload("Subscription.Membership.Library").
		Where("id = ?", id).
		First(&b).
		Error
	if err != nil {
		return usecase.Borrowing{}, err
	}

	ub := b.ConvertToUsecase()

	if b.Book.ID != uuid.Nil {
		book := b.Book.ConvertToUsecase()
		ub.Book = &book
	}

	if b.Subscription != nil {
		sub := b.Subscription.ConvertToUsecase()
		ub.Subscription = &sub

		if b.Subscription.User != nil {
			user := b.Subscription.User.ConvertToUsecase()
			sub.User = &user
		}

		if b.Subscription.Membership != nil {
			mem := b.Subscription.Membership.ConvertToUsecase()
			sub.Membership = &mem

			if b.Subscription.Membership.Library != nil {
				lib := b.Subscription.Membership.Library.ConvertToUsecase()
				mem.Library = &lib
			}
		}
	}

	if b.Staff.ID != uuid.Nil {
		staff := b.Staff.ConvertToUsecase()
		ub.Staff = &staff
	}

	return ub, nil
}

func (s *service) CreateBorrowing(ctx context.Context, b usecase.Borrowing) (usecase.Borrowing, error) {
	borrow := Borrowing{
		BookID:         b.BookID,
		SubscriptionID: b.SubscriptionID,
		StaffID:        b.StaffID,
		BorrowedAt:     b.BorrowedAt,
		DueAt:          b.DueAt,
		ReturnedAt:     b.ReturnedAt,
	}

	if err := s.db.WithContext(ctx).Create(&borrow).Error; err != nil {
		return usecase.Borrowing{}, err
	}

	return borrow.ConvertToUsecase(), nil
}

func (s *service) UpdateBorrowing(ctx context.Context, b usecase.Borrowing) (usecase.Borrowing, error) {
	borrow := Borrowing{
		ID:             b.ID,
		BookID:         b.BookID,
		SubscriptionID: b.SubscriptionID,
		StaffID:        b.StaffID,
		BorrowedAt:     b.BorrowedAt,
		DueAt:          b.DueAt,
		ReturnedAt:     b.ReturnedAt,
	}

	err := s.db.WithContext(ctx).Updates(&borrow).Error
	if err != nil {
		return usecase.Borrowing{}, err
	}

	return borrow.ConvertToUsecase(), nil
}

// Convert core model to Usecase
func (b Borrowing) ConvertToUsecase() usecase.Borrowing {
	var d *time.Time
	if b.DeletedAt != nil {
		d = &b.DeletedAt.Time
	}
	return usecase.Borrowing{
		ID:             b.ID,
		BookID:         b.BookID,
		SubscriptionID: b.SubscriptionID,
		StaffID:        b.StaffID,
		BorrowedAt:     b.BorrowedAt,
		DueAt:          b.DueAt,
		ReturnedAt:     b.ReturnedAt,
		CreatedAt:      b.CreatedAt,
		UpdatedAt:      b.UpdatedAt,
		DeletedAt:      d,
	}
}
