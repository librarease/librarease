package database

import (
	"context"
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	CreatedAt      time.Time     `gorm:"column:created_at"`
	UpdatedAt      time.Time     `gorm:"column:updated_at"`
	DeletedAt      *gorm.DeletedAt

	Returning *Returning
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

	db := s.db.Model([]Borrowing{}).WithContext(ctx).
		// Joins("LEFT JOIN returnings r ON borrowings.returning_id = r.id")
		Preload("Returning")

	if len(opt.BookIDs) > 0 {
		db = db.Where("book_id IN ?", opt.BookIDs)
	}
	if len(opt.SubscriptionIDs) > 0 {
		db = db.Where("subscription_id IN ?", opt.SubscriptionIDs)
	}
	if len(opt.BorrowStaffIDs) > 0 {
		db = db.Where("staff_id IN ?", opt.BorrowStaffIDs)
	}
	if !opt.BorrowedAt.IsZero() {
		db = db.Where("borrowed_at = ?", opt.BorrowedAt)
	}
	if !opt.DueAt.IsZero() {
		db = db.Where("due_at = ?", opt.DueAt)
	}
	if opt.ReturnedAt != nil {
		db = db.Where("returnings.returned_at = ?", opt.ReturnedAt)
	}
	if opt.IsActive {
		// Filter borrowings that do not have a corresponding entry in the returnings table
		db = db.Where("NOT EXISTS (SELECT NULL FROM returnings r WHERE r.borrowing_id = borrowings.id)")
	}
	if opt.IsExpired {
		db = db.Where("due_at < now() AND returning_id IS NULL")
	}
	if len(opt.MembershipIDs) > 0 {
		db = db.Joins("Subscription").Where("membership_id IN ?", opt.MembershipIDs)
	}
	if len(opt.UserIDs) > 0 {
		db = db.Joins("Subscription").Where("user_id IN ?", opt.UserIDs)
	}
	if len(opt.LibraryIDs) > 0 {
		db = db.Joins("Book").Where("library_id IN ?", opt.LibraryIDs)
		// db = db.Joins("JOIN subscriptions s ON borrowings.subscription_id = s.id").
		// 	Joins("JOIN memberships m ON s.membership_id = m.id").
		// 	Where("m.library_id = ?", opt.LibraryID)
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
		Preload("Book").
		Preload("Staff").
		Preload("Subscription").
		Preload("Subscription.User").
		Preload("Subscription.Membership").
		Preload("Subscription.Membership.Library").
		Count(&count).
		Limit(opt.Limit).
		Offset(opt.Skip).
		Order(orderBy + " " + orderIn).
		Find(&borrows).
		Error

	if err != nil {
		return nil, 0, err
	}

	for _, b := range borrows {
		ub := b.ConvertToUsecase()

		if b.Returning != nil {
			returning := b.Returning.ConvertToUsecase()
			ub.Returning = &returning
		}

		if b.Book != nil {
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
		if b.Staff != nil {
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
		Preload("Returning").
		Preload("Returning.Staff").
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

	if b.Returning != nil {
		returning := b.Returning.ConvertToUsecase()
		ub.Returning = &returning

		if b.Returning.Staff != nil {
			staff := b.Returning.Staff.ConvertToUsecase()
			returning.Staff = &staff
		}
	}

	if b.Book != nil {
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

	if b.Staff != nil {
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
		// FIXME: accept returning id / create returning
		// ReturnedAt:     b.ReturnedAt,
	}

	if err := s.db.
		WithContext(ctx).
		Clauses(clause.Returning{}).
		Create(&borrow).Error; err != nil {

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
		// FIXME: accept returning id / create returning
		// ReturnedAt:     b.ReturnedAt,
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
		CreatedAt:      b.CreatedAt,
		UpdatedAt:      b.UpdatedAt,
		DeletedAt:      d,
	}
}
