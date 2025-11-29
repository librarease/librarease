package usecase

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/librarease/librarease/internal/config"

	"github.com/google/uuid"
)

type Subscription struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	MembershipID uuid.UUID
	Note         *string
	SubscribedAt time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time

	// Granfathering the membership
	ExpiresAt       time.Time
	Amount          int
	FinePerDay      int
	LoanPeriod      int
	ActiveLoanLimit int
	UsageLimit      int

	User       *User
	Membership *Membership

	UsageCount      *int
	ActiveLoanCount *int
}

type ListSubscriptionsOption struct {
	Skip   int
	Limit  int
	SortBy string
	SortIn string

	ID             string
	UserID         string
	MembershipID   string
	LibraryIDs     uuid.UUIDs
	MembershipName string
	IsActive       bool
	IsExpired      bool
}

func (u Usecase) ListSubscriptions(ctx context.Context, opt ListSubscriptionsOption) ([]Subscription, int, error) {

	role, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
	if !ok {
		return nil, 0, fmt.Errorf("user role not found in context")
	}
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return nil, 0, fmt.Errorf("user id not found in context")
	}

	switch role {
	case "SUPERADMIN":
		u.logger.InfoContext(ctx, "[DEBUG] global superadmin")
		// ALLOW ALL
	case "ADMIN":
		u.logger.InfoContext(ctx, "[DEBUG] global admin")
		// ALLOW ALL
	case "USER":
		u.logger.InfoContext(ctx, "[DEBUG] global user")
		staffs, _, err := u.repo.ListStaffs(ctx, ListStaffsOption{
			UserID: userID.String(),
			// Using a limit of 500 for now, adjust as needed based on expected data size
			Limit: 500,
		})
		if err != nil {
			return nil, 0, err
		}
		// user is not staff
		if len(staffs) == 0 {
			u.logger.InfoContext(ctx, "[DEBUG] user is not staff, filtering by user id")
			opt.UserID = userID.String()
			break
		}

		// user is staff
		u.logger.InfoContext(ctx, "[DEBUG] user is staff")
		var staffLibIDs uuid.UUIDs
		for _, staff := range staffs {
			staffLibIDs = append(staffLibIDs, staff.LibraryID)
		}
		// user is staff, filtering by library ids
		if len(opt.LibraryIDs) == 0 {
			// user is staff, filters default to assigned libraries
			u.logger.InfoContext(ctx, "[DEBUG] filtering by default assigned libraries")
			opt.LibraryIDs = staffLibIDs
			break
		}

		u.logger.InfoContext(ctx, "[DEBUG] filtering by library ids query")
		var intersectLibIDs uuid.UUIDs
		for _, id := range opt.LibraryIDs {
			// filter out library ids that are not assigned to the staff
			if slices.Contains(staffLibIDs, id) {
				intersectLibIDs = append(intersectLibIDs, id)
			}
		}

		if len(intersectLibIDs) == 0 {
			// user is filtering by library ids but none of the ids are assigned to the staff
			u.logger.InfoContext(ctx, "[DEBUG] staff filters by lib ids but none assigned")
			opt.LibraryIDs = staffLibIDs
			break
		}

		// user is filtering by library ids and some of the ids are assigned to the staff
		u.logger.InfoContext(ctx, "[DEBUG] staff filters by lib ids and some assigned")
		opt.LibraryIDs = intersectLibIDs
	}
	return u.repo.ListSubscriptions(ctx, opt)
}

func (u Usecase) CreateSubscription(ctx context.Context, sub Subscription) (Subscription, error) {
	m, err := u.GetMembershipByID(ctx, sub.MembershipID.String())
	if err != nil {
		return Subscription{}, err
	}
	if m.DeletedAt != nil {
		return Subscription{}, fmt.Errorf("membership %s is deleted", m.ID)
	}
	// Granfathering the membership
	sub.ExpiresAt = time.Now().AddDate(0, 0, m.Duration)
	sub.Amount = m.Price
	sub.LoanPeriod = m.LoanPeriod
	sub.FinePerDay = m.FinePerDay
	sub.ActiveLoanLimit = m.ActiveLoanLimit
	sub.UsageLimit = m.UsageLimit

	s, err := u.repo.CreateSubscription(ctx, sub)
	if err != nil {
		return Subscription{}, err
	}

	go func() {
		// if err := u.SendBorrowingEmail(context.Background(), bw.ID); err != nil {
		// 	fmt.Printf("borrowing: failed to send email: %v\n", err)
		// }

		if err := u.CreateNotification(context.Background(), Notification{
			Title:         "Membership Activated",
			Message:       fmt.Sprintf("Your membership \"%s\" is now active.", m.Name),
			UserID:        s.UserID,
			ReferenceType: "SUBSCRIPTION",
			ReferenceID:   &s.ID,
		}); err != nil {
			fmt.Printf("borrowing: failed to create notification: %v\n", err)
		}

	}()

	return s, nil
}

func (u Usecase) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (Subscription, error) {
	return u.repo.GetSubscriptionByID(ctx, id)
}

func (u Usecase) UpdateSubscription(ctx context.Context, sub Subscription) (Subscription, error) {
	if sub.ExpiresAt.IsZero() {
		s, err := u.GetSubscriptionByID(ctx, sub.ID)
		if err != nil {
			return Subscription{}, err
		}
		sub.ExpiresAt = s.ExpiresAt
	}
	return u.repo.UpdateSubscription(ctx, sub)
}

func (u Usecase) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	_, count, err := u.ListBorrowings(ctx, ListBorrowingsOption{
		BorrowingsOption: BorrowingsOption{
			SubscriptionIDs: uuid.UUIDs{id},
		},
	})
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("subscription %s has %d borrowings", id, count)
	}

	return u.repo.DeleteSubscription(ctx, id)
}
