package usecase

import (
	"context"
	"time"
)

// borrowing count
type BorrowingAnalysis struct {
	Timestamp time.Time
	Count     int
}

// revenue amount
type RevenueAnalysis struct {
	Timestamp    time.Time
	Subscription int
	Fine         int
}

// book borrowing count
type BookAnalysis struct {
	Timestamp time.Time
	Count     int
	Title     string
}

// membership purchasing count
type MembershipAnalysis struct {
	Name  string
	Count int
}

type Analysis struct {
	Borrowing  []BorrowingAnalysis
	Revenue    []RevenueAnalysis
	Book       []BookAnalysis
	Membership []MembershipAnalysis
}

type GetAnalysisOption struct {
	From      time.Time
	To        time.Time
	Limit     int
	Skip      int
	LibraryID string
}

func (u Usecase) GetAnalysis(ctx context.Context, opt GetAnalysisOption) (Analysis, error) {
	return u.repo.GetAnalysis(ctx, opt)
}
