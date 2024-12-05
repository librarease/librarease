package database

import (
	"context"
	"librarease/internal/usecase"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name      string          `gorm:"column:name;type:varchar(255)"`
	CreatedAt time.Time       `gorm:"column:created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"column:deleted_at"`

	Staffs        []Staff
	Subscriptions []Subscription
}

func (User) TableName() string {
	return "users"
}

func (s *service) ListUsers(ctx context.Context) ([]usecase.User, int, error) {
	var (
		users []usecase.User
		count int64
	)

	db := s.db.Table("users")
	db.Count(&count)

	err := db.Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	for _, u := range users {
		uu := usecase.User{
			ID:        u.ID,
			Name:      u.Name,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
			DeleteAt:  u.DeleteAt,
		}
		users = append(users, uu)
	}

	return users, int(count), nil
}
