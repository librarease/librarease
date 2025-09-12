package usecase

import (
	"context"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/librarease/librarease/internal/config"

	"github.com/google/uuid"
)

type Returning struct {
	ID          uuid.UUID
	BorrowingID uuid.UUID
	StaffID     uuid.UUID
	ReturnedAt  time.Time
	Fine        int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time

	Borrowing *Borrowing
	Staff     *Staff
}

type ListReturningOption struct {
	Skip   int
	Limit  int
	SortBy string
	SortIn string

	BorrowingIDs uuid.UUIDs
	StaffIDs     uuid.UUIDs
	ReturnedAt   time.Time
	Fine         *int
}

func (u Usecase) ReturnBorrowing(ctx context.Context, borrowingID uuid.UUID, r Returning) (Borrowing, error) {

	role, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
	if !ok {
		return Borrowing{}, fmt.Errorf("user role not found in context")
	}
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return Borrowing{}, fmt.Errorf("user id not found in context")
	}

	borrow, err := u.repo.GetBorrowingByID(ctx, borrowingID)
	if err != nil {
		return Borrowing{}, err
	}
	if borrow.Returning != nil {
		return Borrowing{}, fmt.Errorf("borrowing already returned")
	}

	if r.ReturnedAt.Before(borrow.BorrowedAt) {
		return Borrowing{}, fmt.Errorf("returned at date is before borrowed at date")
	}

	// calculate fine only if fine is negative (not provided)
	if r.Fine < 0 {
		r.Fine = 0
		if r.ReturnedAt.After(borrow.DueAt) {

			overdueHours := r.ReturnedAt.Sub(borrow.DueAt).Hours()
			days := int(math.Floor(overdueHours / 24))
			fine := days * borrow.Subscription.FinePerDay
			r.Fine = fine
		}
	}

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
			// Using a limit of 500 for now, adjust as needed based on expected data size
			Limit: 500,
		})
		if err != nil {
			return Borrowing{}, err
		}
		// user is not staff
		if len(staffs) == 0 {
			return Borrowing{}, fmt.Errorf("user is not staff")
		}
		// user is library staff
		if st := staffs[0]; st.Role == StaffRoleStaff {
			fmt.Println("[DEBUG] user is library staff")
			r.StaffID = st.ID
			break
		}

		// user is library admin
		fmt.Println("[DEBUG] user is library admin")
		// ALLOW ALL
	}

	staffs, _, err := u.repo.ListStaffs(ctx, ListStaffsOption{
		// get all staffs in the library
		LibraryIDs: uuid.UUIDs{borrow.Subscription.Membership.LibraryID},
		// Using a limit of 500 for now, adjust as needed based on expected data size
		Limit: 500,
	})
	if err != nil {
		return Borrowing{}, err
	}
	var staffIDs uuid.UUIDs
	for _, staff := range staffs {
		staffIDs = append(staffIDs, staff.ID)
	}
	if !slices.Contains(staffIDs, r.StaffID) {
		return Borrowing{}, fmt.Errorf("staff %s is not from the library", r.StaffID)
	}

	rb, err := u.repo.ReturnBorrowing(ctx, borrowingID, r)
	if err != nil {
		return Borrowing{}, err
	}

	go func() {
		if err := u.CreateNotification(context.Background(), Notification{
			Title:         "Book Returned",
			Message:       fmt.Sprintf("Book %s has been returned", borrow.Book.Title),
			UserID:        borrow.Subscription.UserID,
			ReferenceID:   &borrowingID,
			ReferenceType: "BORROWING",
		}); err != nil {
			fmt.Printf("returning: failed to create notification: %v\n", err)
		}
	}()

	return rb, nil
}

func (u Usecase) DeleteReturn(ctx context.Context, borrowingId uuid.UUID) error {
	borrow, err := u.repo.GetBorrowingByID(ctx, borrowingId)
	if err != nil {
		return err
	}
	if borrow.Returning == nil {
		return fmt.Errorf("borrow has not returned yet: %s", borrowingId)
	}
	if err := u.repo.DeleteReturn(ctx, borrow.Returning.ID); err != nil {
		return err
	}

	go func() {
		if err := u.CreateNotification(context.Background(), Notification{
			Title:         "Undo Book Return",
			Message:       fmt.Sprintf("Return of book %s has been undone", borrow.Book.Title),
			UserID:        borrow.Subscription.UserID,
			ReferenceID:   &borrowingId,
			ReferenceType: "BORROWING",
		}); err != nil {
			fmt.Printf("returning: failed to create notification: %v\n", err)
		}
	}()

	return nil
}

func (u Usecase) UpdateReturn(ctx context.Context, borrowingId uuid.UUID, r Returning) error {
	borrow, err := u.repo.GetBorrowingByID(ctx, borrowingId)
	if err != nil {
		return err
	}
	if borrow.Returning == nil {
		return fmt.Errorf("borrow has not returned yet: %s", borrowingId)
	}

	if !r.ReturnedAt.IsZero() && r.ReturnedAt.Before(borrow.BorrowedAt) {
		return fmt.Errorf("returned at date is before borrowed at date")
	}

	// calculate fine only if fine is negative (not provided)
	// if r.Fine < 0 {
	// 	r.Fine = 0
	// 	if r.ReturnedAt.After(borrow.DueAt) {

	// 		overdueHours := r.ReturnedAt.Sub(borrow.DueAt).Hours()
	// 		days := int(math.Floor(overdueHours / 24))
	// 		fine := days * borrow.Subscription.FinePerDay
	// 		r.Fine = fine
	// 	}
	// }

	return u.repo.UpdateReturn(ctx, borrow.Returning.ID, r)
}
