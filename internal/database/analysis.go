package database

import (
	"context"
	"slices"
	"time"

	"github.com/librarease/librarease/internal/usecase"
)

func (s *service) GetAnalysis(ctx context.Context, opt usecase.GetAnalysisOption) (usecase.Analysis, error) {
	var borrowing []usecase.BorrowingAnalysis
	err := s.db.WithContext(ctx).Table("borrowings b").
		Joins("JOIN books bk ON b.book_id = bk.id").
		Select("DATE_TRUNC('day', b.borrowed_at) AS timestamp, COUNT(*) AS count").
		Group("timestamp").
		Order("timestamp DESC").
		Offset(opt.Skip).
		Limit(opt.Limit).
		Where("b.borrowed_at BETWEEN ? AND ?", opt.From, opt.To).
		Where("bk.library_id = ?", opt.LibraryID).
		Scan(&borrowing).Error
	if err != nil {
		return usecase.Analysis{}, err
	}
	slices.Reverse(borrowing)

	var fineData []usecase.RevenueAnalysis
	err = s.db.WithContext(ctx).Table("borrowings b").
		Joins("JOIN subscriptions s ON b.subscription_id = s.id").
		Joins("JOIN memberships m ON s.membership_id = m.id").
		Joins("JOIN returnings r ON r.borrowing_id = b.id").
		Select(`
			DATE_TRUNC('day', r.returned_at) AS timestamp,
			-- SUM((EXTRACT(DAY FROM r.returned_at - b.due_at)) * s.fine_per_day) AS predicted_fine,
			SUM(r.fine) AS fine
		`).
		Where("r.returned_at > b.due_at").
		Where("r.returned_at BETWEEN ? AND ?", opt.From, opt.To).
		Where("m.library_id = ?", opt.LibraryID).
		Group("timestamp").
		Order("timestamp DESC").
		Offset(opt.Skip).
		Limit(opt.Limit).
		Scan(&fineData).Error
	if err != nil {
		return usecase.Analysis{}, err
	}
	var subscriptionData []usecase.RevenueAnalysis
	err = s.db.WithContext(ctx).Table("subscriptions s").
		Joins("JOIN memberships m ON s.membership_id = m.id").
		Select("DATE_TRUNC('day', s.created_at) AS timestamp, SUM(s.amount) AS subscription").
		Group("timestamp").
		Order("timestamp DESC").
		Offset(opt.Skip).
		Limit(opt.Limit).
		Where("s.created_at BETWEEN ? AND ?", opt.From, opt.To).
		Where("m.library_id = ?", opt.LibraryID).
		Scan(&subscriptionData).Error
	if err != nil {
		return usecase.Analysis{}, err
	}
	revenueMap := make(map[time.Time]usecase.RevenueAnalysis)
	for _, r := range fineData {
		revenueMap[r.Timestamp] = r
	}

	for _, r := range subscriptionData {
		if v, exists := revenueMap[r.Timestamp]; exists {
			v.Subscription = r.Subscription
			revenueMap[r.Timestamp] = v
		} else {
			revenueMap[r.Timestamp] = usecase.RevenueAnalysis{
				Timestamp:    r.Timestamp,
				Subscription: r.Subscription,
			}
		}
	}

	revenue := make([]usecase.RevenueAnalysis, 0, len(revenueMap))
	for _, r := range revenueMap {
		revenue = append(revenue, r)
	}

	slices.SortFunc(revenue, func(a, b usecase.RevenueAnalysis) int {
		if a.Timestamp.Before(b.Timestamp) {
			return -1
		}
		return 1
	})

	var book []usecase.BookAnalysis
	err = s.db.WithContext(ctx).Table("borrowings b").
		Joins("JOIN books bk ON b.book_id = bk.id").
		Select("bk.title, COUNT(*) AS count").
		Group("bk.title").
		Order("count DESC").
		Offset(opt.Skip).
		Limit(opt.Limit).
		Where("b.borrowed_at BETWEEN ? AND ?", opt.From, opt.To).
		Where("bk.library_id = ?", opt.LibraryID).
		Scan(&book).Error

	if err != nil {
		return usecase.Analysis{}, err
	}

	var membership []usecase.MembershipAnalysis
	err = s.db.WithContext(ctx).Table("subscriptions s").
		Joins("JOIN memberships m ON s.membership_id = m.id").
		Select("m.name, COUNT(*) AS count").
		Group("m.name").
		Order("count DESC").
		Offset(opt.Skip).
		Limit(opt.Limit).
		Where("s.created_at BETWEEN ? AND ?", opt.From, opt.To).
		Where("m.library_id = ?", opt.LibraryID).
		Scan(&membership).Error
	if err != nil {
		return usecase.Analysis{}, err
	}

	return usecase.Analysis{
		Borrowing:  borrowing,
		Revenue:    revenue,
		Book:       book,
		Membership: membership,
	}, nil
}
