package usecase

import "context"

func New(repo Repository) Usecase {
	return Usecase{repo: repo}
}

type Repository interface {
	Health() map[string]string
	Close() error
	ListUsers(context.Context) ([]User, error)
}

type Usecase struct {
	repo Repository
}

func (u Usecase) Health() map[string]string {
	return u.repo.Health()
}

func (u Usecase) Close() error {
	return u.repo.Close()
}
