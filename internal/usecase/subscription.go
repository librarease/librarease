package usecase

import (
	"context"
	"fmt"
	"librarease/internal/config"
	"slices"
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	MembershipID uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time

	// Granfathering the membership
	ExpiresAt       time.Time
	FinePerDay      int
	LoanPeriod      int
	ActiveLoanLimit int

	User       *User
	Membership *Membership
}

type ListSubscriptionsOption struct {
	Skip           int
	Limit          int
	UserID         string
	MembershipID   string
	LibraryIDs     uuid.UUIDs
	MembershipName string
	IsActive       bool
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
		fmt.Println("[DEBUG] global superadmin")
		// ALLOW ALL
	case "ADMIN":
		fmt.Println("[DEBUG] global admin")
		// ALLOW ALL
	case "USER":
		fmt.Println("[DEBUG] global user")
		staffs, _, err := u.ListStaffs(ctx, ListStaffsOption{
			UserID: userID.String(),
			// Using a limit of 500 for now, adjust as needed based on expected data size
			Limit: 500,
		})
		if err != nil {
			return nil, 0, err
		}
		// user is not staff
		if len(staffs) == 0 {
			fmt.Println("[DEBUG] user is not staff, filtering by user id")
			opt.UserID = userID.String()
			break
		}

		// user is staff
		fmt.Println("[DEBUG] user is staff")
		var staffLibIDs uuid.UUIDs
		for _, staff := range staffs {
			staffLibIDs = append(staffLibIDs, staff.LibraryID)
		}
		// user is staff, filtering by library ids
		if len(opt.LibraryIDs) == 0 {
			// user is staff, filters default to assigned libraries
			fmt.Println("[DEBUG] filtering by default assigned libraries")
			opt.LibraryIDs = staffLibIDs
			break
		}

		fmt.Println("[DEBUG] filtering by library ids query")
		var intersectLibIDs uuid.UUIDs
		for _, id := range opt.LibraryIDs {
			// filter out library ids that are not assigned to the staff
			if slices.Contains(staffLibIDs, id) {
				intersectLibIDs = append(intersectLibIDs, id)
			}
		}

		if len(intersectLibIDs) == 0 {
			// user is filtering by library ids but none of the ids are assigned to the staff
			fmt.Println("[DEBUG] staff filters by lib ids but none assigned")
			opt.LibraryIDs = staffLibIDs
			break
		}

		// user is filtering by library ids and some of the ids are assigned to the staff
		fmt.Println("[DEBUG] staff filters by lib ids and some assigned")
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
	sub.LoanPeriod = m.LoanPeriod
	sub.FinePerDay = m.FinePerDay
	sub.ActiveLoanLimit = m.ActiveLoanLimit

	return u.repo.CreateSubscription(ctx, sub)
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
