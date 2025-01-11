package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/librarease/librarease/internal/config"

	"github.com/google/uuid"
)

type GlobalRole string

const (
	GlobalRoleSuperAdmin GlobalRole = "SUPERADMIN"
	GlobalRoleAdmin      GlobalRole = "ADMIN"
	GlobalRoleUser       GlobalRole = "USER"
)

type User struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeleteAt  *time.Time

	Staffs   []Staff
	AuthUser *AuthUser
}

type ListUsersOption struct {
	Skip       int
	Limit      int
	SortBy     string
	SortIn     string
	Name       string
	Email      string
	Phone      string
	IDs        uuid.UUIDs
	GlobalRole GlobalRole
}

func (u Usecase) ListUsers(ctx context.Context, opt ListUsersOption) ([]User, int, error) {
	users, total, err := u.repo.ListUsers(ctx, opt)
	if err != nil {
		return nil, 0, err
	}

	var userList []User
	for _, user := range users {
		userList = append(userList, User{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Phone:     user.Phone,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			DeleteAt:  user.DeleteAt,
		})
	}

	return userList, total, nil
}

type GetUserByIDOption struct {
	IncludeStaffs bool
}

func (u Usecase) GetUserByID(ctx context.Context, id string, opt GetUserByIDOption) (User, error) {
	return u.repo.GetUserByID(ctx, id, opt)
}

func (u Usecase) CreateUser(ctx context.Context, user User) (User, error) {
	createdUser, err := u.repo.CreateUser(ctx, user)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:        createdUser.ID,
		Name:      createdUser.Name,
		Email:     createdUser.Email,
		Phone:     createdUser.Phone,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
	}, nil
}

func (u Usecase) UpdateUser(ctx context.Context, id uuid.UUID, user User) (User, error) {

	role, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
	if !ok {
		return User{}, fmt.Errorf("user role not found in context")
	}
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return User{}, fmt.Errorf("user id not found in context")
	}

	switch role {
	case "SUPERADMIN":
		// ALLOW
	case "ADMIN":
		if user.AuthUser != nil {
			fmt.Println("[DEBUG] admin can't update auth user")
			return User{}, fmt.Errorf("admin can't update auth user")
		}
	case "USER":
		if id != userID {
			fmt.Printf("[DEBUG] user can only update their own data, id: %s, userID: %s\n", id, userID)
			return User{}, fmt.Errorf("user can only update their own data")
		}
		if user.AuthUser != nil {
			fmt.Println("[DEBUG] user can't update auth user")
			return User{}, fmt.Errorf("user can't update auth user")
		}
	}

	updatedUser, err := u.repo.UpdateUser(ctx, id, user)
	if err != nil {
		return User{}, err
	}

	if user.AuthUser != nil {
		fmt.Printf("[DEBUG] updating custom claims for user id: %s\n", updatedUser.ID)
		err = u.refreshCustomClaims(ctx, updatedUser.ID)
		if err != nil {
			fmt.Printf("[ERROR] failed to refresh custom claims for user id: %s, error: %v\n", updatedUser.ID, err)
			return User{}, err
		}
	}

	return User{
		ID:        updatedUser.ID,
		Name:      updatedUser.Name,
		Email:     updatedUser.Email,
		Phone:     updatedUser.Phone,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
	}, nil
}

func (u Usecase) DeleteUser(ctx context.Context, id string) error {
	err := uuid.Validate(id)
	if err != nil {
		return err
	}
	err = u.repo.DeleteUser(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
