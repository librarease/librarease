package database

import (
	"context"
	"librarease/internal/usecase"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthUser struct {
	UID        string          `gorm:"column:uid;primaryKey;type:varchar(255);"`
	UserID     uuid.UUID       `gorm:"column:user_id;type:uuid;uniqueIndex"`
	User       *User           `gorm:"foreignKey:UserID;references:ID"`
	GlobalRole string          `gorm:"column:global_role;check:global_role IN ('SUPERADMIN', 'ADMIN', 'USER');default:'USER'"`
	CreateAt   time.Time       `gorm:"column:created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at"`
	DeletedAt  *gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (AuthUser) TableName() string {
	return "auth_users"
}

func (s *service) CreateAuthUser(ctx context.Context, au usecase.AuthUser) (usecase.AuthUser, error) {
	u := AuthUser{
		UID:        au.UID,
		UserID:     au.UserID,
		GlobalRole: au.GlobalRole,
	}
	err := s.db.WithContext(ctx).Create(&u).Error
	if err != nil {
		return usecase.AuthUser{}, err
	}

	return u.ConvertToUsecase(), nil
}

func (a AuthUser) ConvertToUsecase() usecase.AuthUser {
	return usecase.AuthUser{
		UID:        a.UID,
		UserID:     a.UserID,
		GlobalRole: a.GlobalRole,
		CreatedAt:  a.CreateAt,
		UpdatedAt:  a.UpdatedAt,
	}
}
