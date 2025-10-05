package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/usecase"
)

func (s *service) GetAnalysis(
	ctx context.Context,
	opt usecase.GetAnalysisOption) (
	usecase.Analysis, error) {

	var borrowing []usecase.BorrowingAnalysis
	if err := s.db.WithContext(ctx).Raw(`
		WITH date_series AS (
			SELECT generate_series(
				date_trunc('day', ?::timestamp),
				date_trunc('day', ?::timestamp),
				'1 day'::interval
			) AS date_val
		),
		borrowing_counts AS (
			SELECT 
				DATE_TRUNC('day', b.borrowed_at) AS date_val,
				COUNT(b.id) AS total_borrow
			FROM borrowings b
			JOIN books bk ON b.book_id = bk.id
			WHERE b.borrowed_at BETWEEN ? AND ?
				AND bk.library_id = ?
				AND b.deleted_at IS NULL
			GROUP BY DATE_TRUNC('day', b.borrowed_at)
		),
		returning_counts AS (
			SELECT 
				DATE_TRUNC('day', r.returned_at) AS date_val,
				COUNT(r.id) AS total_return
			FROM returnings r
			JOIN borrowings b ON r.borrowing_id = b.id
			JOIN books bk ON b.book_id = bk.id
			WHERE r.returned_at BETWEEN ? AND ?
				AND bk.library_id = ?
				AND r.deleted_at IS NULL
				AND b.deleted_at IS NULL
			GROUP BY DATE_TRUNC('day', r.returned_at)
		)
		SELECT 
			ds.date_val AS timestamp,
			COALESCE(bc.total_borrow, 0) AS total_borrow,
			COALESCE(rc.total_return, 0) AS total_return
		FROM date_series ds
		LEFT JOIN borrowing_counts bc ON ds.date_val = bc.date_val
		LEFT JOIN returning_counts rc ON ds.date_val = rc.date_val
		WHERE COALESCE(bc.total_borrow, 0) > 0 OR COALESCE(rc.total_return, 0) > 0
		ORDER BY ds.date_val ASC
	`, opt.From, opt.To, opt.From, opt.To, opt.LibraryID, opt.From, opt.To, opt.LibraryID).
		Scan(&borrowing).Error; err != nil {

		return usecase.Analysis{}, err
	}

	var revenue []usecase.RevenueAnalysis
	if err := s.db.WithContext(ctx).Raw(`
		WITH date_series AS (
			SELECT generate_series(
				date_trunc('day', ?::timestamp),
				date_trunc('day', ?::timestamp),
				'1 day'::interval
			) AS date_val
		),
		fine_data AS (
			SELECT 
				DATE_TRUNC('day', r.returned_at) AS date_val,
				SUM(r.fine) AS fine
			FROM returnings r
			JOIN borrowings b ON r.borrowing_id = b.id
			JOIN subscriptions s ON b.subscription_id = s.id
			JOIN memberships m ON s.membership_id = m.id
			WHERE r.deleted_at IS NULL
				AND r.fine > 0
				AND r.returned_at BETWEEN ? AND ?
				AND m.library_id = ?
			GROUP BY DATE_TRUNC('day', r.returned_at)
		),
		subscription_data AS (
			SELECT 
				DATE_TRUNC('day', s.created_at) AS date_val,
				SUM(s.amount) AS subscription
			FROM subscriptions s
			JOIN memberships m ON s.membership_id = m.id
			WHERE s.created_at BETWEEN ? AND ?
				AND m.library_id = ?
			GROUP BY DATE_TRUNC('day', s.created_at)
		)
		SELECT 
			ds.date_val AS timestamp,
			COALESCE(fd.fine, 0) AS fine,
			COALESCE(sd.subscription, 0) AS subscription
		FROM date_series ds
		LEFT JOIN fine_data fd ON ds.date_val = fd.date_val
		LEFT JOIN subscription_data sd ON ds.date_val = sd.date_val
		WHERE COALESCE(fd.fine, 0) > 0 OR COALESCE(sd.subscription, 0) > 0
		ORDER BY ds.date_val ASC
	`, opt.From, opt.To, opt.From, opt.To, opt.LibraryID, opt.From, opt.To, opt.LibraryID).
		Scan(&revenue).Error; err != nil {

		return usecase.Analysis{}, err
	}

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
		Joins("LEFT JOIN returnings r ON r.borrowing_id = b.id AND r.deleted_at IS NULL").
		Joins("LEFT JOIN losts l ON l.borrowing_id = b.id AND l.deleted_at IS NULL").
		Select(`
		m.id,
		m.name AS membership,
		COUNT(*) AS total,
		SUM(CASE 
			WHEN l.id IS NULL AND r.id IS NOT NULL AND r.returned_at > b.due_at THEN 1 
			ELSE 0 
		END) AS overdue,
		(SUM(CASE 
			WHEN l.id IS NULL AND r.id IS NOT NULL AND r.returned_at > b.due_at THEN 1 
			ELSE 0 
		END)::float / NULLIF(COUNT(*), 0)) AS overdue_rate
	`).
		Where("m.library_id = ?", libraryID).
		Where("b.deleted_at IS NULL")

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

func (r *service) BorrowingHeatmap(
	ctx context.Context,
	libraryID uuid.UUID,
	start, end *time.Time) (
	[]usecase.HeatmapCell, error) {

	q := `
        SELECT
            EXTRACT(DOW FROM b.borrowed_at)::int AS day_of_week,
            EXTRACT(HOUR FROM b.borrowed_at)::int AS hour_of_day,
            CASE WHEN EXTRACT(MINUTE FROM b.borrowed_at)::int >= 30 THEN 30 ELSE 0 END AS minute_of_hour,
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
        GROUP BY day_of_week, hour_of_day, minute_of_hour
        ORDER BY day_of_week, hour_of_day, minute_of_hour
    `

	rows, err := r.db.WithContext(ctx).Raw(q, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cells []usecase.HeatmapCell
	for rows.Next() {
		var c usecase.HeatmapCell
		if err := rows.Scan(&c.DayOfWeek, &c.HourOfDay, &c.MinuteOfHour, &c.Count); err != nil {
			return nil, err
		}
		cells = append(cells, c)
	}
	return cells, rows.Err()
}

func (r *service) ReturningHeatmap(
	ctx context.Context,
	libraryID uuid.UUID,
	start, end *time.Time) (
	[]usecase.HeatmapCell, error) {

	q := `
        SELECT
            EXTRACT(DOW FROM r.returned_at)::int AS day_of_week,
            EXTRACT(HOUR FROM r.returned_at)::int AS hour_of_day,
            CASE WHEN EXTRACT(MINUTE FROM r.returned_at)::int >= 30 THEN 30 ELSE 0 END AS minute_of_hour,
            COUNT(r.id) AS count
        FROM returnings r
		JOIN borrowings b ON r.borrowing_id = b.id
		JOIN subscriptions s ON b.subscription_id = s.id
		JOIN memberships m ON s.membership_id = m.id
        WHERE r.deleted_at IS NULL
          AND b.deleted_at IS NULL
          AND m.library_id = $1
    `
	args := []any{libraryID}
	i := 2
	if start != nil {
		q += fmt.Sprintf(" AND r.returned_at >= $%d", i)
		args = append(args, *start)
		i++
	}
	if end != nil {
		q += fmt.Sprintf(" AND r.returned_at <= $%d", i)
		args = append(args, *end)
		i++
	}
	q += `
        GROUP BY day_of_week, hour_of_day, minute_of_hour
        ORDER BY day_of_week, hour_of_day, minute_of_hour
    `

	rows, err := r.db.WithContext(ctx).Raw(q, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cells []usecase.HeatmapCell
	for rows.Next() {
		var c usecase.HeatmapCell
		if err := rows.Scan(&c.DayOfWeek, &c.HourOfDay, &c.MinuteOfHour, &c.Count); err != nil {
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
