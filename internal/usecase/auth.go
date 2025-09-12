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

type CustomClaims struct {
	ID        uuid.UUID
	Role      string
	AdminLibs uuid.UUIDs
	StaffLibs uuid.UUIDs
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

	// set custom claims
	err = u.refreshCustomClaims(ctx, user.ID)
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

// Deprecated: only global role is used which is included in GetUserByID
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

// helper function to set custom claims
func (u Usecase) refreshCustomClaims(ctx context.Context, id uuid.UUID) error {
	user, err := u.repo.GetUserByID(ctx, id, GetUserByIDOption{
		IncludeStaffs: true,
	})
	if err != nil {
		return err
	}
	var (
		adminLibs = make(uuid.UUIDs, 0)
		staffLibs = make(uuid.UUIDs, 0)
	)
	for _, staff := range user.Staffs {
		if staff.Role == "ADMIN" {
			adminLibs = append(adminLibs, staff.LibraryID)
		}
		if staff.Role == "STAFF" {
			staffLibs = append(staffLibs, staff.LibraryID)
		}
	}

	claims := CustomClaims{
		ID:        id,
		Role:      user.AuthUser.GlobalRole,
		AdminLibs: adminLibs,
		StaffLibs: staffLibs,
	}

	return u.identityProvider.SetCustomClaims(ctx, user.AuthUser.UID, claims)
}
