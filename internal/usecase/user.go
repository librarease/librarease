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

	Staffs []Staff
}

func (u Usecase) ListUsers(ctx context.Context) ([]User, int, error) {
	users, total, err := u.repo.ListUsers(ctx)
	if err != nil {
		return nil, 0, err
	}

	var userList []User
	for _, user := range users {
		userList = append(userList, User{
			ID:        user.ID,
			Name:      user.Name,
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
	// uid, err := uuid.Parse(id)
	// if err != nil {
	// 	return User{}, err
	// }
	// user, err := u.repo.GetUserByID(ctx, uid, opt)
	// if err != nil {
	// 	return User{}, err
	// }

	// return User{
	// 	ID:        user.ID,
	// 	Name:      user.Name,
	// 	CreatedAt: user.CreatedAt,
	// 	UpdatedAt: user.UpdatedAt,
	// 	DeleteAt:  user.DeleteAt,
	// }, nil
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
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
	}, nil
}

func (u Usecase) UpdateUser(ctx context.Context, user User) (User, error) {
	updatedUser, err := u.repo.UpdateUser(ctx, user)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:        updatedUser.ID,
		Name:      updatedUser.Name,
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
