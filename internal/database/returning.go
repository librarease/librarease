package database

import (
	"context"
	"fmt"
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Returning struct {
	ID          uuid.UUID  `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	BorrowingID uuid.UUID  `gorm:"column:borrowing_id;type:uuid;not null;index:"`
	Borrowing   *Borrowing `gorm:"foreignKey:BorrowingID;references:ID"`
	StaffID     uuid.UUID  `gorm:"column:staff_id;type:uuid;"`
	Staff       *Staff     `gorm:"foreignKey:StaffID;references:ID"`
	ReturnedAt  time.Time  `gorm:"column:returned_at;default:now()"`
	Fine        int        `gorm:"column:fine;type:int"`
	Note        *string    `gorm:"column:note;type:text"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
	DeletedAt   *gorm.DeletedAt
}

func (Returning) TableName() string {
	return "returnings"
}

// func (s *service) ListReturning(ctx context.Context, opt usecase.ListReturningOption) ([]usecase.Returning, int, error) {
// 	var (
// 		returns  []Returning
// 		ureturns []usecase.Returning
// 		count    int64
// 	)

// 	db := s.db.Model([]Returning{}).WithContext(ctx)

// 	if opt.BorrowingID != "" {
// 		db = db.Where("borrowing_id = ?", opt.BorrowingID)
// 	}
// 	if opt.StaffID != "" {
// 		db = db.Where("staff_id = ?", opt.StaffID)
// 	}
// 	if !opt.ReturnedAt.IsZero() {
// 		db = db.Where("returned_at = ?", opt.ReturnedAt)
// 	}

// 	if opt.Fine != 0 {
// 		db = db.Where("fine = ?", opt.Fine)
// 	}

// 	db.Count(&count)
// 	db = db.Offset(opt.Skip).Limit(opt.Limit)

// 	if err := db.Find(&returns).Error; err != nil {
// 		return nil, 0, err
// 	}

// 	for i := range returns {
// 		ureturns = append(ureturns, returns[i].ToUsecase())
// 	}

// 	return ureturns, int(count), nil
// }

func (s *service) ReturnBorrowing(ctx context.Context, borrowingID uuid.UUID, r usecase.Returning) (usecase.Borrowing, error) {

	var returning = Returning{
		BorrowingID: borrowingID,
		StaffID:     r.StaffID,
		ReturnedAt:  r.ReturnedAt,
		Fine:        r.Fine,
		Note:        r.Note,
	}

	err := s.db.WithContext(ctx).
		Clauses(clause.Returning{}).
		Create(&returning).
		Error

	if err != nil {
		return usecase.Borrowing{}, err
	}

	var borrowing Borrowing
	err = s.db.WithContext(ctx).
		Model(&Borrowing{}).
		First(&borrowing, "id = ?", borrowingID).
		Preload("Returning").
		Error

	if err != nil {
		return usecase.Borrowing{}, err
	}

	// Invalidate borrowing cache
	pattern := fmt.Sprintf("borrowing:%s:*", returning.BorrowingID.String())
	if keys, err := s.cache.Keys(ctx, pattern).Result(); err == nil && len(keys) > 0 {
		s.cache.Del(ctx, keys...)
	}

	return borrowing.ConvertToUsecase(), nil
}

// Convert core model to Usecase
func (r Returning) ConvertToUsecase() usecase.Returning {
	var d *time.Time
	if r.DeletedAt != nil {
		d = &r.DeletedAt.Time
	}
	return usecase.Returning{
		ID:          r.ID,
		BorrowingID: r.BorrowingID,
		StaffID:     r.StaffID,
		ReturnedAt:  r.ReturnedAt,
		Fine:        r.Fine,
		Note:        r.Note,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		DeletedAt:   d,
	}
}

func (s service) DeleteReturn(ctx context.Context, id uuid.UUID) error {
	// Get borrowing ID before deleting
	var returning Returning
	if err := s.db.WithContext(ctx).First(&returning, "id = ?", id).Error; err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).
		Delete(&Returning{
			ID: id,
		}).
		Error; err != nil {
		return err
	}

	// Invalidate borrowing cache
	pattern := fmt.Sprintf("borrowing:%s:*", returning.BorrowingID.String())
	if keys, err := s.cache.Keys(ctx, pattern).Result(); err == nil && len(keys) > 0 {
		s.cache.Del(ctx, keys...)
	}

	return nil
}

func (s service) UpdateReturn(ctx context.Context, id uuid.UUID, r usecase.Returning) error {
	var returning = Returning{
		ID:          id,
		BorrowingID: r.BorrowingID,
		StaffID:     r.StaffID,
		ReturnedAt:  r.ReturnedAt,
		Fine:        r.Fine,
		Note:        r.Note,
	}

	if err := s.db.WithContext(ctx).
		Model(&Returning{}).
		Where("id = ?", id).
		Updates(&returning).
		Error; err != nil {
		return err
	}

	pattern := fmt.Sprintf("borrowing:%s:*", r.BorrowingID.String())
	if keys, err := s.cache.Keys(ctx, pattern).Result(); err == nil && len(keys) > 0 {
		s.cache.Del(ctx, keys...)
	}

	return nil
}
