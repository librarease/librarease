package database

import (
	"context"
	"errors"
	"fmt"
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
	Lost      *Lost
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
		db = db.Where("borrowed_at >= ? AND borrowed_at < ?", opt.BorrowedAt, opt.BorrowedAt.Add(24*time.Hour))
	}
	if !opt.DueAt.IsZero() {
		db = db.Where("due_at >= ? AND due_at < ?", opt.DueAt, opt.DueAt.Add(24*time.Hour))
	}
	if opt.ReturnedAt != nil {
		db = db.Where("EXISTS (SELECT NULL FROM returnings r WHERE r.borrowing_id = borrowings.id AND r.deleted_at IS NULL AND r.returned_at >= ? AND r.returned_at < ?)", opt.ReturnedAt, opt.ReturnedAt.Add(24*time.Hour))
	}
	if opt.LostAt != nil {
		db = db.Where("EXISTS (SELECT NULL FROM losts l WHERE l.borrowing_id = borrowings.id AND l.deleted_at IS NULL AND l.reported_at >= ? AND l.reported_at < ?)", opt.LostAt, opt.LostAt.Add(24*time.Hour))
	}
	if opt.IsActive || opt.IsOverdue {
		db = db.
			// - do not have a corresponding entry in the losts table
			Where("NOT EXISTS (SELECT NULL FROM losts l WHERE l.borrowing_id = borrowings.id AND l.deleted_at IS NULL)").
			// - do not have a corresponding entry in the returnings table
			Where("NOT EXISTS (SELECT NULL FROM returnings r WHERE r.borrowing_id = borrowings.id AND r.deleted_at IS NULL)")
		if opt.IsOverdue {
			db = db.Where("due_at < now()")
		}
	}
	if opt.IsReturned {
		db = db.Where("EXISTS (SELECT NULL FROM returnings r WHERE r.borrowing_id = borrowings.id AND r.deleted_at IS NULL)")
	}
	if opt.IsLost {
		db = db.Where("EXISTS (SELECT NULL FROM losts l WHERE l.borrowing_id = borrowings.id AND l.deleted_at IS NULL)")
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
	if opt.BorrowedAtFrom != nil {
		db = db.Where("borrowed_at >= ?", opt.BorrowedAtFrom)
	}
	if opt.BorrowedAtTo != nil {
		db = db.Where("borrowed_at <= ?", opt.BorrowedAtTo)
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
		Preload("Book").
		Preload("Staff").
		Preload("Lost").
		Preload("Subscription").
		Preload("Subscription.User").
		Preload("Subscription.Membership").
		Preload("Subscription.Membership.Library").
		Order(orderBy + " " + orderIn).
		Find(&borrows).
		Error; err != nil {

		return nil, 0, err
	}

	for _, b := range borrows {
		ub := b.ConvertToUsecase()

		if b.Returning != nil {
			returning := b.Returning.ConvertToUsecase()
			ub.Returning = &returning
		}

		if b.Lost != nil {
			lost := b.Lost.ConvertToUsecase()
			ub.Lost = &lost
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

// ListBorrowingSummariesForNotifications returns lightweight borrowing summaries with precise notification filters
func (s *service) ListBorrowingSummariesForNotifications(ctx context.Context, opt usecase.NotificationFiltersOption) ([]usecase.BorrowingSummary, error) {
	type result struct {
		ID          uuid.UUID `gorm:"column:id"`
		UserID      uuid.UUID `gorm:"column:user_id"`
		DueAt       time.Time `gorm:"column:due_at"`
		BookTitle   string    `gorm:"column:book_title"`
		LibraryName string    `gorm:"column:library_name"`
	}

	var results []result

	// Build optimized query for notification processing
	db := s.db.WithContext(ctx).
		Table("borrowings").
		Joins("JOIN subscriptions s ON borrowings.subscription_id = s.id").
		Joins("JOIN books b ON borrowings.book_id = b.id").
		Joins("JOIN libraries l ON b.library_id = l.id").
		Select(`
			borrowings.id,
			s.user_id,
			borrowings.due_at,
			b.title as book_title,
			l.name as library_name
		`).
		Where("borrowings.deleted_at IS NULL").
		// Must be active (not returned or lost)
		Where("NOT EXISTS (SELECT 1 FROM losts WHERE losts.borrowing_id = borrowings.id AND losts.deleted_at IS NULL)").
		Where("NOT EXISTS (SELECT 1 FROM returnings WHERE returnings.borrowing_id = borrowings.id AND returnings.deleted_at IS NULL)")

	// Apply precise time filters for near-due notifications
	if opt.DueAtFrom != nil && opt.DueAtTo != nil {
		db = db.Where("borrowings.due_at >= ? AND borrowings.due_at <= ?", *opt.DueAtFrom, *opt.DueAtTo)
	}

	// Optional library filtering
	if len(opt.LibraryIDs) > 0 {
		db = db.Where("l.id IN ?", opt.LibraryIDs)
	}

	// Apply limit
	if opt.Limit > 0 {
		db = db.Limit(opt.Limit)
	}

	if err := db.Find(&results).Error; err != nil {
		return nil, err
	}

	// Convert to usecase.BorrowingSummary
	summaries := make([]usecase.BorrowingSummary, len(results))
	for i, r := range results {
		summaries[i] = usecase.BorrowingSummary{
			ID:          r.ID,
			UserID:      r.UserID,
			DueAt:       r.DueAt,
			BookTitle:   r.BookTitle,
			LibraryName: r.LibraryName,
		}
	}

	return summaries, nil
}

func (s *service) GetBorrowingByID(ctx context.Context, id uuid.UUID) (usecase.Borrowing, error) {
	var b Borrowing

	err := s.db.
		Model(Borrowing{}).
		WithContext(ctx).
		Preload("Returning").
		Preload("Returning.Staff").
		Preload("Lost").
		Preload("Lost.Staff").
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return usecase.Borrowing{}, usecase.ErrNotFound{
				ID:      id,
				Code:    "borrowing_not_found",
				Message: fmt.Sprintf("borrowing with id %s not found", id),
			}
		}
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

	if b.Lost != nil {
		lost := b.Lost.ConvertToUsecase()
		ub.Lost = &lost

		if b.Lost.Staff != nil {
			staff := b.Lost.Staff.ConvertToUsecase()
			lost.Staff = &staff
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

func (s *service) DeleteBorrowing(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Where("id = ?", id).Delete(&Borrowing{}).Error
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
