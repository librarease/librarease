package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// borrowing count
type BorrowingAnalysis struct {
	Timestamp   time.Time
	TotalBorrow int
	TotalReturn int
}

// revenue amount
type RevenueAnalysis struct {
	Timestamp    time.Time
	Subscription int
	Fine         int
}

// book borrowing count
type BookAnalysis struct {
	Count int
	Title string
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

func (u Usecase) OverdueAnalysis(ctx context.Context, from, to *time.Time, libraryID string) ([]OverdueAnalysis, error) {
	return u.repo.OverdueAnalysis(ctx, from, to, libraryID)
}

func (u Usecase) BookUtilization(ctx context.Context, opt GetBookUtilizationOption) ([]BookUtilization, int, error) {
	return u.repo.BookUtilization(ctx, opt)
}

func (u Usecase) BorrowingHeatmap(ctx context.Context, libraryID uuid.UUID, start, end *time.Time) ([]BorrowHeatmapCell, error) {
	return u.repo.BorrowingHeatmap(ctx, libraryID, start, end)
}

func (u Usecase) GetPowerUsers(ctx context.Context, opt GetPowerUsersOption) ([]PowerUser, int, error) {
	return u.repo.GetPowerUsers(ctx, opt)
}

func (u Usecase) GetLongestUnreturned(ctx context.Context, opt GetOverdueBorrowsOption) ([]OverdueBorrow, int, error) {
	return u.repo.GetLongestUnreturned(ctx, opt)
}

type OverdueAnalysis struct {
	MembershipID   uuid.UUID
	MembershipName string
	Total          int
	Overdue        int
	Rate           float64
}

type BookUtilization struct {
	BookID          uuid.UUID
	BookTitle       string
	Copies          int
	TotalBorrowings int
	UtilizationRate float64
}

type BorrowHeatmapCell struct {
	DayOfWeek int // 0=Sunday ... 6=Saturday
	HourOfDay int // 0-23
	Count     int
}

type PowerUser struct {
	UserID     uuid.UUID
	UserName   string
	UserEmail  string
	TotalBooks int
}

type GetPowerUsersOption struct {
	LibraryID uuid.UUID
	From      *time.Time
	To        *time.Time
	Limit     int
	Skip      int
}

type GetBookUtilizationOption struct {
	LibraryID string
	From      *time.Time
	To        *time.Time
	Limit     int
	Skip      int
}

type OverdueBorrow struct {
	BorrowingID uuid.UUID `json:"borrowing_id"`
	BorrowedAt  time.Time `json:"borrowed_at"`
	UserID      uuid.UUID `json:"user_id"`
	UserName    string    `json:"user_name"`
	BookID      uuid.UUID `json:"book_id"`
	BookTitle   string    `json:"book_title"`
	DaysOut     int       `json:"days_out"`
}

type GetOverdueBorrowsOption struct {
	LibraryID uuid.UUID
	From      *time.Time
	To        *time.Time
	Limit     int
	Skip      int
}
