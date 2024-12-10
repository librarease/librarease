package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Borrowing struct {
	ID             uuid.UUID
	BookID         uuid.UUID
	SubscriptionID uuid.UUID
	StaffID        uuid.UUID
	BorrowedAt     time.Time
	DueAt          time.Time
	ReturnedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time

	Book         *Book
	Subscription *Subscription
	Staff        *Staff
}

type ListBorrowingsOption struct {
	Skip           int
	Limit          int
	BookID         string
	SubscriptionID string
	StaffID        string

	MembershipID string
	LibraryID    string
	UserID       string
	BorrowedAt   time.Time
	DueAt        time.Time
	ReturnedAt   *time.Time
	IsActive     bool
	IsExpired    bool
}

func (u Usecase) ListBorrowings(ctx context.Context, opt ListBorrowingsOption) ([]Borrowing, int, error) {
	return u.repo.ListBorrowings(ctx, opt)
}

func (u Usecase) GetBorrowingByID(ctx context.Context, id uuid.UUID) (Borrowing, error) {
	return u.repo.GetBorrowingByID(ctx, id)
}

func (u Usecase) CreateBorrowing(ctx context.Context, borrow Borrowing) (Borrowing, error) {

	// 1. Check if the membership subscription is still active
	s, err := u.repo.GetSubscriptionByID(ctx, borrow.SubscriptionID)
	if err != nil {
		return Borrowing{}, err
	}
	// TODO: ErrMembershipExpired
	if s.ExpiresAt.Before(time.Now()) {
		fmt.Println(s.ExpiresAt, time.Now())
		return Borrowing{}, fmt.Errorf("membership subscription %s expired", s.ID)
	}

	// 2. Check if the user has reached the maximum borrowing limit
	_, activeCount, err := u.repo.ListBorrowings(ctx, ListBorrowingsOption{
		SubscriptionID: s.ID.String(),
		IsActive:       true,
	})
	if err != nil {
		return Borrowing{}, err
	}
	// TODO: ErrActiveLoanLimitReached
	if s.ActiveLoanLimit <= activeCount {
		return Borrowing{}, fmt.Errorf("user %s has reached the active loan limit", s.UserID)
	}

	// 3. Check if the book is available
	_, count, err := u.repo.ListBorrowings(ctx, ListBorrowingsOption{
		BookID:   borrow.BookID.String(),
		IsActive: true,
	})
	if err != nil {
		return Borrowing{}, err
	}
	// TODO: ErrBookNotAvailable
	if count > 0 {
		return Borrowing{}, fmt.Errorf("book %s is not available", borrow.BookID)
	}

	// 4. Check if the book is in the same library
	book, err := u.repo.GetBookByID(ctx, borrow.BookID)
	if err != nil {
		return Borrowing{}, err
	}
	m, err := u.repo.GetMembershipByID(ctx, s.MembershipID)
	if err != nil {
		return Borrowing{}, err
	}
	// TODO: ErrBookNotAvailable
	if book.LibraryID != m.LibraryID {
		return Borrowing{}, fmt.Errorf("book %s is not in library %s", book.ID, m.LibraryID)
	}

	// 5. Check if staff exists
	staff, err := u.repo.GetStaffByID(ctx, borrow.StaffID)
	if err != nil {
		return Borrowing{}, err
	}
	if staff.LibraryID != m.LibraryID {
		return Borrowing{}, fmt.Errorf("staff %s is not from library %s", staff.ID, m.LibraryID)
	}

	// 6. All checks passed, create borrowing
	// Set the borrowed at time if not set
	if borrow.BorrowedAt.IsZero() {
		borrow.BorrowedAt = time.Now()
	}
	// Set the due at time if not set
	if borrow.DueAt.IsZero() {
		borrow.DueAt = time.Now().AddDate(0, 0, s.LoanPeriod)
	}

	bw, err := u.repo.CreateBorrowing(ctx, borrow)
	if err != nil {
		return Borrowing{}, err
	}
	return bw, nil
}

func (u Usecase) UpdateBorrowing(ctx context.Context, borrow Borrowing) (Borrowing, error) {
	return u.repo.UpdateBorrowing(ctx, borrow)
}
