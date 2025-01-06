package database

import (
	"context"
	"librarease/internal/usecase"
	"time"
)

func (s *service) GetAnalysis(ctx context.Context, opt usecase.GetAnalysisOption) (usecase.Analysis, error) {
	var borrowing []usecase.BorrowingAnalysis
	err := s.db.WithContext(ctx).Table("borrowings b").
		Joins("JOIN books bk ON b.book_id = bk.id").
		Select("DATE_TRUNC('month', b.borrowed_at) AS timestamp, COUNT(*) AS count").
		Group("timestamp").
		Offset(opt.Skip).
		Limit(opt.Limit).
		Order("timestamp").
		Where("b.borrowed_at BETWEEN ? AND ?", opt.From, opt.To).
		Where("bk.library_id = ?", opt.LibraryID).
		Scan(&borrowing).Error
	if err != nil {
		return usecase.Analysis{}, err
	}

	var revenue []usecase.RevenueAnalysis
	err = s.db.WithContext(ctx).Table("borrowings b").
		Joins("JOIN subscriptions s ON b.subscription_id = s.id").
		Joins("JOIN memberships m ON s.membership_id = m.id").
		Select(`
			DATE_TRUNC('month', b.returned_at) AS timestamp,
			SUM((EXTRACT(DAY FROM b.returned_at - b.due_at)) * s.fine_per_day) AS fine
		`).
		Where("b.returned_at IS NOT NULL").
		Where("b.returned_at > b.due_at").
		Where("b.returned_at BETWEEN ? AND ?", opt.From, opt.To).
		Where("m.library_id = ?", opt.LibraryID).
		Group("timestamp").
		Order("timestamp").
		Offset(opt.Skip).
		Limit(opt.Limit).
		Scan(&revenue).Error
	if err != nil {
		return usecase.Analysis{}, err
	}
	var subscriptionData []usecase.RevenueAnalysis
	err = s.db.WithContext(ctx).Table("subscriptions s").
		Joins("JOIN memberships m ON s.membership_id = m.id").
		Select("DATE_TRUNC('month', s.created_at) AS timestamp, COUNT(*) AS subscription").
		Group("timestamp").
		Order("timestamp").
		Offset(opt.Skip).
		Limit(opt.Limit).
		Where("s.created_at BETWEEN ? AND ?", opt.From, opt.To).
		Where("m.library_id = ?", opt.LibraryID).
		Scan(&subscriptionData).Error
	if err != nil {
		return usecase.Analysis{}, err
	}
	revenueMap := make(map[time.Time]*usecase.RevenueAnalysis)
	for i := range revenue {
		revenueMap[revenue[i].Timestamp] = &revenue[i]
	}

	for _, sub := range subscriptionData {
		if entry, exists := revenueMap[sub.Timestamp]; exists {
			entry.Subscription = sub.Subscription
		} else {
			revenue = append(revenue, sub)
		}
	}

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
