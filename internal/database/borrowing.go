package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Borrowing struct {
	ID             uuid.UUID    `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	BookID         uuid.UUID    `gorm:"column:book_id;type:uuid;"`
	Book           Book         `gorm:"foreignKey:BookID;references:ID"`
	SubscriptionID uuid.UUID    `gorm:"column:subscription_id;type:uuid;"`
	Subscription   Subscription `gorm:"foreignKey:SubscriptionID;references:ID"`
	StaffID        uuid.UUID    `gorm:"column:staff_id;type:uuid;"`
	Staff          Staff        `gorm:"foreignKey:StaffID;references:ID"`
	BorrowedAt     time.Time    `gorm:"column:borrowed_at;default:now()"`
	DueAt          time.Time    `gorm:"column:due_at"`
	ReturnedAt     *time.Time   `gorm:"column:returned_at"`
	CreatedAt      time.Time    `gorm:"column:created_at"`
	UpdatedAt      time.Time    `gorm:"column:updated_at"`
	DeletedAt      *gorm.DeletedAt
}

func (Borrowing) TableName() string {
	return "borrowings"
}
