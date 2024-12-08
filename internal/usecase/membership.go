package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Membership struct {
	ID              uuid.UUID
	Name            string
	LibraryID       uuid.UUID
	Duration        int
	ActiveLoanLimit int
	LoanPeriod      int
	FinePerDay      int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time

	Library *Library
}

type ListMembershipsOption struct {
	Skip      int
	Limit     int
	LibraryID string
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

// func (u Usecase) DeleteMembership(ctx context.Context, id string) error {
// 	mid, err := uuid.Parse(id)
// 	if err != nil {
// 		return err
// 	}
// 	return u.repo.DeleteMembership(ctx, mid)
// }
