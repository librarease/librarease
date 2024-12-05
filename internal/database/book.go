package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Book struct {
	ID         uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	Title      string          `gorm:"column:title;type:varchar(255)"`
	Author     string          `gorm:"column:author;type:varchar(255)"`
	Year       int             `gorm:"column:year;type:int"`
	CreatedAt  time.Time       `gorm:"column:created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at"`
	DeletedAt  *gorm.DeletedAt `gorm:"column:deleted_at"`
	LibraryID  uuid.UUID
	Library    Library `gorm:"foreignKey:LibraryID"`
	Borrowings []Borrowing
}

func (Book) TableName() string {
	return "books"
}
