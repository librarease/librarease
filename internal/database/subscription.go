package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Subscription represents user's purchase of a membership
type Subscription struct {
	ID           uuid.UUID  `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID       uuid.UUID  `gorm:"column:user_id;type:uuid;"`
	User         User       `gorm:"foreignKey:UserID;references:ID"`
	MembershipID uuid.UUID  `gorm:"column:membership_id;type:uuid;"`
	Membership   Membership `gorm:"foreignKey:MembershipID;references:ID"`
	// since memberships can be updated, we need to store expiry date
	ExpiresAt  time.Time       `gorm:"column:expires_at"`
	CreatedAt  time.Time       `gorm:"column:created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at"`
	DeletedAt  *gorm.DeletedAt `gorm:"column:deleted_at"`
	Borrowings []Borrowing
}

func (Subscription) TableName() string {
	return "subscriptions"
}
