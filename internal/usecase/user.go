package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeleteAt  *time.Time
}

func (u Usecase) ListUsers(ctx context.Context) ([]User, int, error) {
	users, total, err := u.repo.ListUsers(ctx)
	if err != nil {
		return nil, 0, err
	}

	var userList []User
	for _, user := range users {
		userList = append(userList, User{
			ID:   user.ID,
			Name: user.Name,
		})
	}

	return userList, total, nil
}
