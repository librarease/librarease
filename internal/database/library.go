package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Library struct {
	ID        uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name      string          `gorm:"column:name;type:varchar(255)"`
	CreatedAt time.Time       `gorm:"column:created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"column:deleted_at"`

	Staffs      []Staff
	Books       []Book
	Memberships []Membership
}

func (Library) TableName() string {
	return "libraries"
}
