package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/config"
)

type Lost struct {
	ID          uuid.UUID
	BorrowingID uuid.UUID
	StaffID     uuid.UUID
	ReportedAt  time.Time
	Fine        int
	Note        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time

	Borrowing *Borrowing
	Staff     *Staff
}

func (u Usecase) LostBorrowing(ctx context.Context, borrowingID uuid.UUID, l Lost) (Lost, error) {
	role, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
	if !ok {
		return Lost{}, fmt.Errorf("user role not found in context")
	}
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return Lost{}, fmt.Errorf("user id not found in context")
	}

	// Get the borrowing record
	borrow, err := u.repo.GetBorrowingByID(ctx, borrowingID)
	if err != nil {
		return Lost{}, err
	}

	// Check if borrowing is already returned
	if borrow.Returning != nil {
		return Lost{}, fmt.Errorf("borrowing already returned")
	}

	// Check if borrowing is already lost
	if borrow.Lost != nil {
		return Lost{}, fmt.Errorf("borrowing already lost")
	}

	// Check if borrowing is latest borrowing for the book
	latestBorrow, _, err := u.repo.ListBorrowings(ctx, ListBorrowingsOption{
		BookIDs: []uuid.UUID{borrow.BookID},
		Limit:   1,
	})
	if err != nil {
		return Lost{}, err
	}
	if len(latestBorrow) == 0 || latestBorrow[0].ID != borrow.ID {
		return Lost{}, fmt.Errorf("borrowing is not the latest borrowing for the book")
	}

	// Validate reported date
	if l.ReportedAt.Before(borrow.BorrowedAt) {
		return Lost{}, fmt.Errorf("reported at date is before borrowed at date")
	}

	// Permission check
	switch role {
	case "SUPERADMIN":
		fmt.Println("[DEBUG] global superadmin")
		// ALLOW ALL
	case "ADMIN":
		fmt.Println("[DEBUG] global admin")
		// ALLOW ALL
	case "USER":
		fmt.Println("[DEBUG] global user")
		staffs, _, err := u.repo.ListStaffs(ctx, ListStaffsOption{
			UserID:     userID.String(),
			LibraryIDs: uuid.UUIDs{borrow.Subscription.Membership.LibraryID},
			Limit:      500,
		})
		if err != nil {
			return Lost{}, err
		}
		// user is not staff
		if len(staffs) == 0 {
			return Lost{}, fmt.Errorf("user %s is not staff", userID)
		}
		// user is library staff
		if st := staffs[0]; st.Role == StaffRoleStaff {
			fmt.Println("[DEBUG] user is library staff")
			l.StaffID = st.ID
			break
		}

		// user is library admin
		fmt.Println("[DEBUG] user is library admin")
		// ALLOW ALL
	}

	// Set borrowing ID
	l.BorrowingID = borrowingID

	// Create lost report
	lost, err := u.repo.CreateLost(ctx, l)
	if err != nil {
		return Lost{}, err
	}

	// Send notification to user
	go func() {
		if err := u.CreateNotification(context.Background(), Notification{
			Title:         "Book Reported Lost",
			Message:       fmt.Sprintf("Book %s has been reported lost.", borrow.Book.Title),
			UserID:        borrow.Subscription.UserID,
			ReferenceID:   &borrowingID,
			ReferenceType: "BORROWING",
		}); err != nil {
			fmt.Printf("lost: failed to create notification: %v\n", err)
		}
	}()

	return lost, nil
}

func (u Usecase) UpdateLost(ctx context.Context, borrowingID uuid.UUID, l Lost) (Lost, error) {
	borrow, err := u.repo.GetBorrowingByID(ctx, borrowingID)
	if err != nil {
		return Lost{}, err
	}

	if borrow.Lost == nil {
		return Lost{}, fmt.Errorf("no lost record found for borrowing %s", l.BorrowingID)
	}

	if !l.ReportedAt.IsZero() && l.ReportedAt.Before(borrow.BorrowedAt) {
		return Lost{}, fmt.Errorf("reported at date is before borrowed at date")
	}

	return l, u.repo.UpdateLost(ctx, borrow.Lost.ID, l)

}

func (u Usecase) DeleteLost(ctx context.Context, borrowingId uuid.UUID) error {
	borrow, err := u.repo.GetBorrowingByID(ctx, borrowingId)
	if err != nil {
		return err
	}
	if borrow.Lost == nil {
		return fmt.Errorf("borrow has not been reported lost yet: %s", borrowingId)
	}
	return u.repo.DeleteLost(ctx, borrow.Lost.ID)
}
