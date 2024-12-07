package usecase

import (
	"context"

	"github.com/google/uuid"
)

func New(repo Repository) Usecase {
	return Usecase{repo: repo}
}

type Repository interface {
	Health() map[string]string
	Close() error

	// user
	ListUsers(context.Context) ([]User, int, error)
	GetUserByID(context.Context, string, GetUserByIDOption) (User, error)
	CreateUser(context.Context, User) (User, error)
	UpdateUser(context.Context, User) (User, error)
	DeleteUser(context.Context, string) error

	// library
	ListLibraries(context.Context) ([]Library, int, error)
	GetLibraryByID(context.Context, string) (Library, error)
	CreateLibrary(context.Context, Library) (Library, error)
	UpdateLibrary(context.Context, Library) (Library, error)
	DeleteLibrary(context.Context, string) error

	// staff
	ListStaffs(context.Context, ListStaffsOption) ([]Staff, int, error)
	CreateStaff(context.Context, Staff) (Staff, error)
	GetStaffByID(context.Context, uuid.UUID) (Staff, error)
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
