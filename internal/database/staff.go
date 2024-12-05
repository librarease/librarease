package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Staff struct {
	ID         uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name       string          `gorm:"column:name;type:varchar(255)"`
	LibraryID  uuid.UUID       `gorm:"column:library_id;type:uuid;uniqueIndex:idx_user_library"`
	Library    Library         `gorm:"foreignKey:LibraryID;references:ID"`
	UserID     uuid.UUID       `gorm:"column:user_id;type:uuid;uniqueIndex:idx_user_library"`
	User       User            `gorm:"foreignKey:UserID;references:ID"`
	CreatedAt  time.Time       `gorm:"column:created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at"`
	DeletedAt  *gorm.DeletedAt `gorm:"column:deleted_at;"`
	Borrowings []Borrowing
}

func (Staff) TableName() string {
	return "staffs"
}
