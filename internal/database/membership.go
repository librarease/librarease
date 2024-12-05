package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Membership struct {
	ID              uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name            string          `gorm:"column:name;type:varchar(255)"`
	LibraryID       uuid.UUID       `gorm:"column:library_id;type:uuid;"`
	Library         Library         `gorm:"foreignKey:LibraryID;references:ID"`
	Duration        int             `gorm:"column:duration;type:int"`
	ActiveLoanLimit int             `gorm:"column:active_loan_limit;type:int"`
	FinePerDay      int             `gorm:"column:fine_per_day;type:int"`
	CreatedAt       time.Time       `gorm:"column:created_at"`
	UpdatedAt       time.Time       `gorm:"column:updated_at"`
	DeletedAt       *gorm.DeletedAt `gorm:"column:deleted_at"`

	Subscriptions []Subscription
}

func (Membership) TableName() string {
	return "memberships"
}
