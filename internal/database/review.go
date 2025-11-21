package database

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/usecase"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Review struct {
	ID          uuid.UUID  `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	BorrowingID uuid.UUID  `gorm:"column:borrowing_id;type:uuid;uniqueIndex:,where:deleted_at IS NULL"`
	Borrowing   *Borrowing `gorm:"foreignKey:BorrowingID;references:ID"`
	Rating      int        `gorm:"column:rating;not null"`
	Comment     *string    `gorm:"column:comment;type:text"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
	DeletedAt   *gorm.DeletedAt
}

func (Review) TableName() string {
	return "reviews"
}

func (s service) ListReviews(ctx context.Context, opt usecase.ListReviewsOption) ([]usecase.Review, int, error) {
	var (
		reviews  []Review
		ureviews []usecase.Review
		count    int64
	)

	db := s.db.Model([]Review{}).
		WithContext(ctx).
		Preload("Borrowing").
		Preload("Borrowing.Book").
		Preload("Borrowing.Subscription").
		Preload("Borrowing.Subscription.User")

	if opt.BorrowingID != uuid.Nil {
		db = db.Where("id = ?", opt.BorrowingID)
	}
	if opt.BookID != uuid.Nil {
		db = db.Where("book_id = ?", opt.BookID)
	}
	if opt.UserID != uuid.Nil {
		db = db.Where("user_id = ?", opt.UserID)
	}
	if opt.Rating != nil {
		db = db.Where("rating = ?", *opt.Rating)
	}
	if opt.Comment != nil {
		db = db.Where("comment ILIKE ?", "%"+*opt.Comment+"%")
	}

	var (
		orderIn = "DESC"
		orderBy = "created_at"
	)
	if opt.SortBy != "" {
		orderBy = opt.SortBy
	}
	if opt.SortIn != "" {
		orderIn = opt.SortIn
	}

	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if opt.Limit > 0 {
		db = db.Limit(opt.Limit)
	}
	if opt.Skip > 0 {
		db = db.Offset(opt.Skip)
	}

	if err := db.
		Order(orderBy + " " + orderIn).
		Find(&reviews).
		Error; err != nil {

		return nil, 0, err
	}

	for _, r := range reviews {
		ureviews = append(ureviews, r.ConvertToUsecase())
	}
	return ureviews, int(count), nil
}

// Convert core model to Usecase
func (m Review) ConvertToUsecase() usecase.Review {
	var d *time.Time
	if m.DeletedAt != nil {
		d = &m.DeletedAt.Time
	}
	var (
		borw = new(usecase.Borrowing)
		book = new(usecase.Book)
		user = new(usecase.User)
	)
	if m.Borrowing != nil {
		*borw = m.Borrowing.ConvertToUsecase()

		if b := m.Borrowing.Book; b != nil {
			*book = b.ConvertToUsecase()
		}

		if m.Borrowing.Subscription != nil &&
			m.Borrowing.Subscription.User != nil {

			*user = m.Borrowing.Subscription.User.ConvertToUsecase()
		}
	}
	return usecase.Review{
		ID:          m.ID,
		BorrowingID: m.BorrowingID,
		Rating:      m.Rating,
		Comment:     m.Comment,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   d,
		Borrowing:   borw,
		Book:        book,
		User:        user,
	}
}

func (s service) GetReview(ctx context.Context, id uuid.UUID, opt usecase.ReviewsOption) (usecase.Review, error) {

	var (
		r              Review
		prevID, nextID *uuid.UUID
	)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.db.
			WithContext(ctx).
			Preload("Borrowing").
			Preload("Borrowing.Book").
			Preload("Borrowing.Subscription").
			Preload("Borrowing.Subscription.User").
			First(&r, "id = ?", id).
			Error
	})

	if !reflect.DeepEqual(opt, usecase.ReviewsOption{}) {
		g.Go(func() error {
			p, n, err := s.GetReviewSiblings(ctx, id, opt)
			if err != nil {
				return err
			}
			prevID, nextID = p, n
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return usecase.Review{}, err
	}

	ur := r.ConvertToUsecase()

	ur.PrevID = prevID
	ur.NextID = nextID

	return ur, nil
}

func (s service) GetReviewSiblings(ctx context.Context, id uuid.UUID, opt usecase.ReviewsOption) (*uuid.UUID, *uuid.UUID, error) {
	if opt.SortBy == "" {
		opt.SortBy = "r.created_at"
	}
	if !strings.Contains(opt.SortBy, ".") {
		opt.SortBy = "r." + opt.SortBy
	}
	if opt.SortIn == "" {
		opt.SortIn = "DESC"
	}

	var joins []string
	var where []string
	var args []any

	where = append(where, "r.deleted_at IS NULL")

	// Determine if we need the borrowings join
	needBorrowings := opt.BookID != uuid.Nil || opt.UserID != uuid.Nil
	needSubscriptions := opt.UserID != uuid.Nil

	if needBorrowings {
		joins = append(joins, "JOIN borrowings b ON b.id = r.borrowing_id AND b.deleted_at IS NULL")
	}

	if needSubscriptions {
		joins = append(joins, "JOIN subscriptions s ON s.id = b.subscription_id")
	}

	if opt.BorrowingID != uuid.Nil {
		where = append(where, "r.borrowing_id = ?")
		args = append(args, opt.BorrowingID)
	}

	if opt.BookID != uuid.Nil {
		where = append(where, "b.book_id = ?")
		args = append(args, opt.BookID)
	}

	if opt.UserID != uuid.Nil {
		where = append(where, "s.user_id = ?")
		args = append(args, opt.UserID)
	}

	if opt.Rating != nil {
		where = append(where, "r.rating = ?")
		args = append(args, *opt.Rating)
	}

	if opt.Comment != nil {
		where = append(where, "r.comment ILIKE ?")
		args = append(args, "%"+*opt.Comment+"%")
	}

	joinsSQL := ""
	if len(joins) > 0 {
		joinsSQL = strings.Join(joins, "\n")
	}
	whereSQL := "WHERE " + strings.Join(where, "\nAND ")

	sql := fmt.Sprintf(`
WITH filtered AS (
    SELECT r.id, %s AS sort_col
    FROM reviews r
    %s
    %s
),
ordered AS (
    SELECT id,
           LAG(id) OVER (ORDER BY sort_col %s, id) AS prev_id,
           LEAD(id) OVER (ORDER BY sort_col %s, id) AS next_id
    FROM filtered
)
SELECT prev_id, next_id FROM ordered WHERE id = ?
`, opt.SortBy, joinsSQL, whereSQL, opt.SortIn, opt.SortIn)

	args = append(args, id)

	var out struct {
		PrevID *uuid.UUID
		NextID *uuid.UUID
	}
	if err := s.db.Raw(sql, args...).Scan(&out).Error; err != nil {
		return nil, nil, err
	}
	return out.PrevID, out.NextID, nil
}

func (s service) CreateReview(ctx context.Context, r usecase.Review) (usecase.Review, error) {
	review := Review{
		BorrowingID: r.BorrowingID,
		Rating:      r.Rating,
		Comment:     r.Comment,
	}

	if err := s.db.
		WithContext(ctx).
		Clauses(clause.Returning{}).
		Create(&review).
		Error; err != nil {

		return usecase.Review{}, err
	}

	return review.ConvertToUsecase(), nil
}

func (s service) UpdateReview(ctx context.Context, id uuid.UUID, r usecase.Review) (usecase.Review, error) {
	review := Review{
		ID:      r.ID,
		Rating:  r.Rating,
		Comment: r.Comment,
	}

	if err := s.db.
		WithContext(ctx).
		Model(&review).
		Where("id = ?", id).
		Select("rating", "comment").
		Updates(&review).
		Clauses(clause.Returning{}).
		Error; err != nil {

		return usecase.Review{}, err
	}

	return review.ConvertToUsecase(), nil
}

func (s service) DeleteReview(ctx context.Context, id uuid.UUID) error {
	if err := s.db.
		WithContext(ctx).
		Delete(&Review{}, "id = ?", id).
		Error; err != nil {

		return err
	}
	return nil
}
