package database

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/usecase"
)

func (s *service) GetAnalysis(
	ctx context.Context,
	opt usecase.GetAnalysisOption) (
	usecase.Analysis, error) {

	var borrowing []usecase.BorrowingAnalysis
	if err := s.db.WithContext(ctx).Table("borrowings b").
		Joins("JOIN books bk ON b.book_id = bk.id").
		Joins("LEFT JOIN returnings r ON r.borrowing_id = b.id AND r.deleted_at IS NULL").
		Select(`
			DATE_TRUNC('day', b.borrowed_at) AS timestamp, 
			COUNT(b.id) AS total_borrow,
			COUNT(r.id) AS total_return
		`).
		Group("DATE_TRUNC('day', b.borrowed_at)").
		Order("DATE_TRUNC('day', b.borrowed_at) DESC").
		Where("b.borrowed_at BETWEEN ? AND ?", opt.From, opt.To).
		Where("bk.library_id = ?", opt.LibraryID).
		Scan(&borrowing).
		Error; err != nil {

		return usecase.Analysis{}, err
	}
	slices.Reverse(borrowing)

	var fineData []usecase.RevenueAnalysis
	if err := s.db.WithContext(ctx).Table("borrowings b").
		Joins("JOIN subscriptions s ON b.subscription_id = s.id").
		Joins("JOIN memberships m ON s.membership_id = m.id").
		Joins("JOIN returnings r ON r.borrowing_id = b.id").
		Select(`
			DATE_TRUNC('day', r.returned_at) AS timestamp,
			-- SUM((EXTRACT(DAY FROM r.returned_at - b.due_at)) * s.fine_per_day) AS predicted_fine,
			SUM(r.fine) AS fine
		`).
		// Where("r.returned_at > b.due_at").
		Where("r.deleted_at IS NULL").
		Where("r.fine > 0").
		Where("r.returned_at BETWEEN ? AND ?", opt.From, opt.To).
		Where("m.library_id = ?", opt.LibraryID).
		Group("DATE_TRUNC('day', r.returned_at)").
		Order("DATE_TRUNC('day', r.returned_at) DESC").
		Scan(&fineData).
		Error; err != nil {

		return usecase.Analysis{}, err
	}

	var subscriptionData []usecase.RevenueAnalysis
	if err := s.db.WithContext(ctx).Table("subscriptions s").
		Joins("JOIN memberships m ON s.membership_id = m.id").
		Select("DATE_TRUNC('day', s.created_at) AS timestamp, SUM(s.amount) AS subscription").
		Group("DATE_TRUNC('day', s.created_at)").
		Order("DATE_TRUNC('day', s.created_at) DESC").
		Where("s.created_at BETWEEN ? AND ?", opt.From, opt.To).
		Where("m.library_id = ?", opt.LibraryID).
		Scan(&subscriptionData).
		Error; err != nil {

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
	if err := s.db.WithContext(ctx).Table("borrowings b").
		Joins("JOIN books bk ON b.book_id = bk.id").
		Select("bk.id, bk.title, COUNT(*) AS count").
		Group("bk.id, bk.title").
		Order("count DESC").
		Offset(opt.Skip).
		Limit(opt.Limit).
		Where("b.borrowed_at BETWEEN ? AND ?", opt.From, opt.To).
		Where("bk.library_id = ?", opt.LibraryID).
		Scan(&book).
		Error; err != nil {

		return usecase.Analysis{}, err
	}

	var membership []usecase.MembershipAnalysis
	if err := s.db.WithContext(ctx).Table("subscriptions s").
		Joins("JOIN memberships m ON s.membership_id = m.id").
		Select("m.id, m.name, COUNT(*) AS count").
		Group("m.id, m.name").
		Order("count DESC").
		Offset(opt.Skip).
		Limit(opt.Limit).
		Where("s.created_at BETWEEN ? AND ?", opt.From, opt.To).
		Where("m.library_id = ?", opt.LibraryID).
		Scan(&membership).
		Error; err != nil {

		return usecase.Analysis{}, err
	}

	return usecase.Analysis{
		Borrowing:  borrowing,
		Revenue:    revenue,
		Book:       book,
		Membership: membership,
	}, nil
}

func (s *service) OverdueAnalysis(
	ctx context.Context,
	from, to *time.Time,
	libraryID string) ([]usecase.OverdueAnalysis, error) {

	var result []struct {
		ID      uuid.UUID `gorm:"column:id"`
		Name    string    `gorm:"column:membership"`
		Total   int       `gorm:"column:total"`
		Overdue int       `gorm:"column:overdue"`
		Rate    float64   `gorm:"column:overdue_rate"`
	}

	db := s.db.WithContext(ctx).
		Table("borrowings b").
		Joins("JOIN subscriptions s ON b.subscription_id = s.id").
		Joins("JOIN memberships m ON s.membership_id = m.id").
		Joins("JOIN returnings r ON r.borrowing_id = b.id").
		Select(`
		m.id,
		m.name AS membership,
		COUNT(*) AS total,
		SUM(CASE WHEN r.returned_at > b.due_at THEN 1 ELSE 0 END) AS overdue,
		(SUM(CASE WHEN r.returned_at > b.due_at THEN 1 ELSE 0 END)::float / COUNT(*)) AS overdue_rate
	`).
		Where("m.library_id = ?", libraryID).
		Where("r.deleted_at IS NULL")

	if from != nil && to != nil {
		db = db.Where("b.borrowed_at BETWEEN ? AND ?", *from, *to)
	}

	if err := db.
		Group("m.id, m.name").
		Order("overdue_rate DESC").
		Scan(&result).Error; err != nil {
		return nil, err
	}

	analysis := make([]usecase.OverdueAnalysis, len(result))
	for i, r := range result {
		analysis[i] = usecase.OverdueAnalysis{
			MembershipID:   r.ID,
			MembershipName: r.Name,
			Total:          r.Total,
			Overdue:        r.Overdue,
			Rate:           r.Rate,
		}
	}

	return analysis, nil
}

func (s *service) BookUtilization(
	ctx context.Context,
	opt usecase.GetBookUtilizationOption) (
	[]usecase.BookUtilization, int, error) {

	// First, get the total count without limit/skip
	var totalCount int64
	countDB := s.db.WithContext(ctx).
		Table("books bk").
		Joins("LEFT JOIN borrowings b ON bk.id = b.book_id AND b.deleted_at IS NULL").
		Where("bk.library_id = ?", opt.LibraryID).
		Where("bk.deleted_at IS NULL")

	if opt.From != nil && opt.To != nil {
		countDB = countDB.Where("b.borrowed_at BETWEEN ? AND ?", *opt.From, *opt.To)
	}

	if err := countDB.
		Group("bk.id").
		Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Now get the actual data with limit/skip
	var result []struct {
		ID              uuid.UUID `gorm:"column:id"`
		Title           string    `gorm:"column:title"`
		Count           int       `gorm:"column:count"`
		TotalBorrowings int       `gorm:"column:total_borrowings"`
		UtilizationRate float64   `gorm:"column:utilization_rate"`
	}

	db := s.db.WithContext(ctx).
		Table("books bk").
		Select(`
		bk.id, bk.title, bk.count AS count,
		COUNT(b.id) AS total_borrowings,
		(COUNT(b.id)::float / NULLIF(bk.count, 0)) AS utilization_rate
	`).
		Joins("LEFT JOIN borrowings b ON bk.id = b.book_id AND b.deleted_at IS NULL").
		Where("bk.library_id = ?", opt.LibraryID).
		Where("bk.deleted_at IS NULL")

	if opt.From != nil && opt.To != nil {
		db = db.Where("b.borrowed_at BETWEEN ? AND ?", *opt.From, *opt.To)
	}
	if opt.Skip > 0 {
		db = db.Offset(opt.Skip)
	}
	if opt.Limit > 0 {
		db = db.Limit(opt.Limit)
	}

	if err := db.
		Group("bk.id, bk.title, bk.count").
		Order("utilization_rate DESC").
		Scan(&result).Error; err != nil {

		return nil, 0, err
	}

	analysis := make([]usecase.BookUtilization, len(result))
	for i, r := range result {
		analysis[i] = usecase.BookUtilization{
			BookID:          r.ID,
			BookTitle:       r.Title,
			Copies:          r.Count,
			TotalBorrowings: r.TotalBorrowings,
			UtilizationRate: r.UtilizationRate,
		}
	}

	return analysis, int(totalCount), nil
}

func (r *service) BorrowingHeatmap(
	ctx context.Context,
	libraryID uuid.UUID,
	start, end *time.Time) (
	[]usecase.BorrowHeatmapCell, error) {

	q := `
        SELECT
            EXTRACT(DOW FROM b.borrowed_at)::int AS day_of_week,
            EXTRACT(HOUR FROM b.borrowed_at)::int AS hour_of_day,
            COUNT(b.id) AS count
        FROM borrowings b
		JOIN subscriptions s ON b.subscription_id = s.id
		JOIN memberships m ON s.membership_id = m.id
        WHERE b.deleted_at IS NULL
          AND m.library_id = $1
    `
	args := []any{libraryID}
	i := 2
	if start != nil {
		q += fmt.Sprintf(" AND b.borrowed_at >= $%d", i)
		args = append(args, *start)
		i++
	}
	if end != nil {
		q += fmt.Sprintf(" AND b.borrowed_at <= $%d", i)
		args = append(args, *end)
		i++
	}
	q += `
        GROUP BY day_of_week, hour_of_day
        ORDER BY day_of_week, hour_of_day
    `

	rows, err := r.db.WithContext(ctx).Raw(q, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cells []usecase.BorrowHeatmapCell
	for rows.Next() {
		var c usecase.BorrowHeatmapCell
		if err := rows.Scan(&c.DayOfWeek, &c.HourOfDay, &c.Count); err != nil {
			return nil, err
		}
		cells = append(cells, c)
	}
	return cells, rows.Err()
}

func (s *service) GetPowerUsers(
	ctx context.Context,
	opt usecase.GetPowerUsersOption) (
	[]usecase.PowerUser, int, error) {

	// First, get the total count without limit/skip
	var totalCount int64
	countDB := s.db.WithContext(ctx).
		Table("borrowings b").
		Joins("JOIN subscriptions s ON b.subscription_id = s.id").
		Joins("JOIN memberships m ON s.membership_id = m.id").
		Joins("JOIN users u ON s.user_id = u.id").
		Where("m.library_id = ?", opt.LibraryID).
		Where("b.deleted_at IS NULL")

	if opt.From != nil && opt.To != nil {
		countDB = countDB.Where("b.borrowed_at BETWEEN ? AND ?", *opt.From, *opt.To)
	}

	if err := countDB.
		Group("u.id, u.name").
		Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Now get the actual data with limit/skip
	db := s.db.WithContext(ctx).
		Table("borrowings b").
		Joins("JOIN subscriptions s ON b.subscription_id = s.id").
		Joins("JOIN memberships m ON s.membership_id = m.id").
		Joins("JOIN users u ON s.user_id = u.id").
		Select("u.id as user_id, u.name as user_name, u.email as user_email, COUNT(b.id) as total_books").
		Where("m.library_id = ?", opt.LibraryID).
		Where("b.deleted_at IS NULL")

	if opt.From != nil && opt.To != nil {
		db = db.Where("b.borrowed_at BETWEEN ? AND ?", *opt.From, *opt.To)
	}
	if opt.Skip > 0 {
		db = db.Offset(opt.Skip)
	}
	if opt.Limit > 0 {
		db = db.Limit(opt.Limit)
	}

	var result []usecase.PowerUser
	err := db.
		Group("u.id, u.name, u.email").
		Order("total_books DESC").
		Scan(&result).Error

	return result, int(totalCount), err
}

func (s *service) GetLongestUnreturned(
	ctx context.Context,
	opt usecase.GetOverdueBorrowsOption) (
	[]usecase.OverdueBorrow, int, error) {

	// First, get the total count without limit/skip
	var totalCount int64
	countDB := s.db.WithContext(ctx).
		Table("borrowings b").
		Joins("JOIN subscriptions s ON b.subscription_id = s.id").
		Joins("JOIN memberships m ON s.membership_id = m.id").
		Joins("JOIN users u ON s.user_id = u.id").
		Joins("JOIN books bk ON b.book_id = bk.id").
		Where("m.library_id = ?", opt.LibraryID).
		Where("b.deleted_at IS NULL").
		Where("NOT EXISTS (SELECT 1 FROM returnings r WHERE r.borrowing_id = b.id AND r.deleted_at IS NULL)")

	if opt.From != nil && opt.To != nil {
		countDB = countDB.Where("b.borrowed_at BETWEEN ? AND ?", *opt.From, *opt.To)
	}

	if err := countDB.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Now get the actual data with limit/skip
	var result []usecase.OverdueBorrow

	db := s.db.WithContext(ctx).
		Table("borrowings b").
		Joins("JOIN subscriptions s ON b.subscription_id = s.id").
		Joins("JOIN memberships m ON s.membership_id = m.id").
		Joins("JOIN users u ON s.user_id = u.id").
		Joins("JOIN books bk ON b.book_id = bk.id").
		Where("m.library_id = ?", opt.LibraryID).
		Where("b.deleted_at IS NULL").
		Where("NOT EXISTS (SELECT 1 FROM returnings r WHERE r.borrowing_id = b.id AND r.deleted_at IS NULL)").
		Select(`
			b.id as borrowing_id,
			b.borrowed_at,
			u.id as user_id,
			u.name as user_name,
			bk.id as book_id,
			bk.title as book_title,
			EXTRACT(DAY FROM (NOW() - b.borrowed_at))::int as days_out
		`)

	if opt.From != nil && opt.To != nil {
		db = db.Where("b.borrowed_at BETWEEN ? AND ?", *opt.From, *opt.To)
	}
	if opt.Skip > 0 {
		db = db.Offset(opt.Skip)
	}
	if opt.Limit > 0 {
		db = db.Limit(opt.Limit)
	}

	err := db.
		Order("days_out DESC").
		Scan(&result).Error

	return result, int(totalCount), err
}
