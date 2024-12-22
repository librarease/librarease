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

// get auth user by firebase uid
func (u Usecase) GetAuthUserByUID(ctx context.Context, uid string) (AuthUser, error) {
	authUser, err := u.repo.GetAuthUserByUID(ctx, uid)
	if err != nil {
		return AuthUser{}, err
	}
	return authUser, nil
}

// get auth user by user id
func (u Usecase) GetAuthUserByUserID(ctx context.Context, id string) (AuthUser, error) {
	authUser, err := u.repo.GetAuthUserByUserID(ctx, id)
	if err != nil {
		return AuthUser{}, err
	}
	return authUser, nil
}

// used by middleware
func (u Usecase) VerifyIDToken(ctx context.Context, token string) (string, error) {
	return u.identityProvider.VerifyIDToken(ctx, token)
}
