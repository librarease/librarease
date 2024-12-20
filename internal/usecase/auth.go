package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AuthUser struct {
	UID        string
	UserID     uuid.UUID
	GlobalRole string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
	User       *User
}

type RegisterUser struct {
	Name     string
	Email    string
	Password string
}

func (u Usecase) RegisterUser(ctx context.Context, ru RegisterUser) (User, error) {
	uid, err := u.identityProvider.CreateUser(ctx, ru)
	if err != nil {
		return User{}, err
	}

	user, err := u.CreateUser(ctx, User{
		Name:  ru.Name,
		Email: ru.Email,
	})
	if err != nil {
		return User{}, err
	}

	_, err = u.repo.CreateAuthUser(ctx, AuthUser{
		UID:    uid,
		UserID: user.ID,
	})
	if err != nil {
		return User{}, err
	}
	return user, nil
}

type GetAuthUserOption struct {
	UID        string
	ID         uuid.UUID
	UserID     uuid.UUID
	GlobalRole string
}

// get auth user by firebase uid
func (u Usecase) GetAuthUser(ctx context.Context, opt GetAuthUserOption) (AuthUser, error) {
	authUser, err := u.repo.GetAuthUser(ctx, opt)
	if err != nil {
		return AuthUser{}, err
	}
	return authUser, nil
}
