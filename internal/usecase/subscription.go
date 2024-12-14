package usecase

import (
	"context"
	"fmt"
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
	LibraryID      string
	MembershipName string
	IsActive       bool
}

func (u Usecase) ListSubscriptions(ctx context.Context, opt ListSubscriptionsOption) ([]Subscription, int, error) {
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
