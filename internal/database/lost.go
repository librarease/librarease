package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/usecase"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Lost struct {
	ID          uuid.UUID  `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	BorrowingID uuid.UUID  `gorm:"column:borrowing_id;type:uuid;not null;index:"`
	Borrowing   *Borrowing `gorm:"foreignKey:BorrowingID;references:ID"`
	StaffID     uuid.UUID  `gorm:"column:staff_id;type:uuid;"`
	Staff       *Staff     `gorm:"foreignKey:StaffID;references:ID"`
	ReportedAt  time.Time  `gorm:"column:reported_at;default:now()"`
	Fine        int        `gorm:"column:fine;type:int"`
	Note        string     `gorm:"column:note;type:text"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
	DeletedAt   *gorm.DeletedAt
}

func (Lost) TableName() string {
	return "losts"
}

func (l Lost) ConvertToUsecase() usecase.Lost {
	return usecase.Lost{
		ID:          l.ID,
		BorrowingID: l.BorrowingID,
		StaffID:     l.StaffID,
		ReportedAt:  l.ReportedAt,
		Fine:        l.Fine,
		Note:        l.Note,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}
}

func (s *service) CreateLost(ctx context.Context, l usecase.Lost) (usecase.Lost, error) {
	lost := &Lost{
		BorrowingID: l.BorrowingID,
		StaffID:     l.StaffID,
		ReportedAt:  l.ReportedAt,
		Fine:        l.Fine,
		Note:        l.Note,
	}
	if err := s.db.WithContext(ctx).Create(lost).Error; err != nil {
		return usecase.Lost{}, err
	}

	// Invalidate borrowing cache
	pattern := fmt.Sprintf("borrowing:%s:*", l.BorrowingID.String())
	s.cache.Del(ctx, pattern)

	return lost.ConvertToUsecase(), nil
}

func (s *service) DeleteLost(ctx context.Context, id uuid.UUID) error {
	// Get borrowing ID before deleting
	var lost Lost
	if err := s.db.WithContext(ctx).First(&lost, "id = ?", id).Error; err != nil {
		return err
	}

	if err := s.db.Clauses(clause.Returning{}).Delete(&Lost{ID: id}).Error; err != nil {
		return err
	}

	// Invalidate borrowing cache
	pattern := fmt.Sprintf("borrowing:%s:*", lost.BorrowingID.String())
	s.cache.Del(ctx, pattern)

	return nil
}

func (s *service) UpdateLost(ctx context.Context, id uuid.UUID, l usecase.Lost) error {

	err := s.db.
		WithContext(ctx).
		Model(&Lost{}).
		Where("id = ?", id).
		Updates(Lost{
			ReportedAt: l.ReportedAt,
			Fine:       l.Fine,
			Note:       l.Note,
		}).Error

	if err != nil {
		return err
	}

	// Invalidate borrowing cache
	pattern := fmt.Sprintf("borrowing:%s:*", l.BorrowingID.String())
	s.cache.Del(ctx, pattern)

	return nil
}
