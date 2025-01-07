package database

import (
	"context"
	"librarease/internal/usecase"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Returning struct {
	ID          uuid.UUID  `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	BorrowingID uuid.UUID  `gorm:"column:borrowing_id;type:uuid;"`
	Borrowing   *Borrowing `gorm:"foreignKey:BorrowingID;references:ID"`
	StaffID     uuid.UUID  `gorm:"column:staff_id;type:uuid;"`
	Staff       *Staff     `gorm:"foreignKey:StaffID;references:ID"`
	ReturnedAt  time.Time  `gorm:"column:returned_at;default:now()"`
	Fine        int        `gorm:"column:fine;type:int"`
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
	db := s.db.WithContext(ctx)

	var borrowing = Borrowing{}

	var returning = Returning{
		BorrowingID: borrowingID,
		StaffID:     r.StaffID,
		ReturnedAt:  r.ReturnedAt,
		Fine:        r.Fine,
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where("id = ?", borrowingID).
			Where("returning_id IS NULL").
			First(&borrowing).
			Error; err != nil {

			return err
		}

		returning.BorrowingID = borrowingID

		if err := tx.Create(&returning).Error; err != nil {
			return err
		}

		borrowing.ReturningID = &returning.ID

		if err := tx.Save(&borrowing).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return usecase.Borrowing{}, err
	}

	uborrow := borrowing.ConvertToUsecase()
	ureturn := returning.ConvertToUsecase()
	uborrow.Returning = &ureturn

	return uborrow, nil
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
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		DeletedAt:   d,
	}
}
