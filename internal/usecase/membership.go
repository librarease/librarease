package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/config"
)

type Membership struct {
	ID              uuid.UUID
	Name            string
	LibraryID       uuid.UUID
	Duration        int
	ActiveLoanLimit int
	UsageLimit      int
	LoanPeriod      int
	FinePerDay      int
	Price           int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time

	Library *Library
}

type ListMembershipsOption struct {
	Skip   int
	Limit  int
	SortBy string
	SortIn string

	Name       string
	LibraryIDs uuid.UUIDs
}

func (u Usecase) ListMemberships(ctx context.Context, opt ListMembershipsOption) ([]Membership, int, error) {
	return u.repo.ListMemberships(ctx, opt)
}

func (u Usecase) CreateMembership(ctx context.Context, membership Membership) (Membership, error) {
	return u.repo.CreateMembership(ctx, membership)
}

func (u Usecase) GetMembershipByID(ctx context.Context, id string) (Membership, error) {
	mid, err := uuid.Parse(id)
	if err != nil {
		return Membership{}, err
	}
	return u.repo.GetMembershipByID(ctx, mid)
}

func (u Usecase) UpdateMembership(ctx context.Context, membership Membership) (Membership, error) {
	return u.repo.UpdateMembership(ctx, membership)
}

func (u Usecase) DeleteMembership(ctx context.Context, id uuid.UUID) error {
	role, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
	if !ok {
		return fmt.Errorf("user role not found in context")
	}
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return fmt.Errorf("user id not found in context")
	}

	var isRequireToCheckUser bool

	switch role {
	case "SUPERADMIN":
		// ALLOW
	case "ADMIN":
		// ALLlOW
	case "USER":
		staffs, _, err := u.repo.ListStaffs(ctx, ListStaffsOption{
			UserID: userID.String(),
		})
		if err != nil {
			return err
		}
		if len(staffs) == 0 {
			return fmt.Errorf("user %s is not staff", userID)
		}
		isRequireToCheckUser = true
	}

	// check subscriptions
	sub, _, err := u.repo.ListSubscriptions(ctx, ListSubscriptionsOption{
		MembershipID: id.String(),
	})
	if err != nil {
		return err
	}
	if len(sub) > 0 {
		return fmt.Errorf("cannot delete membership with subscriptions")
	}

	if !isRequireToCheckUser {
		return u.repo.DeleteMembership(ctx, id)
	}

	mem, err := u.repo.GetMembershipByID(ctx, id)
	if err != nil {
		return err
	}

	staffs, _, err := u.repo.ListStaffs(ctx, ListStaffsOption{
		UserID:     userID.String(),
		LibraryIDs: uuid.UUIDs{mem.LibraryID},
	})
	if err != nil {
		return err
	}
	if len(staffs) == 0 {
		return fmt.Errorf("user %s is not staff in library %s", userID, mem.LibraryID)
	}

	return u.repo.DeleteMembership(ctx, id)

}
