package usecase

import "context"

type User struct {
	ID   string
	Name string
}

func (u Usecase) ListUsers(ctx context.Context) ([]User, error) {
	users, err := u.repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	var userList []User
	for _, user := range users {
		userList = append(userList, User{
			ID:   user.ID,
			Name: user.Name,
		})
	}

	return userList, nil
}
