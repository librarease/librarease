package usecase

import (
	"context"

	"github.com/google/uuid"
)

func New(repo Repository, ip IdentityProvider) Usecase {
	return Usecase{
		repo:             repo,
		identityProvider: ip,
	}
}

type Repository interface {
	Health() map[string]string
	Close() error

	// user
	ListUsers(context.Context, ListUsersOption) ([]User, int, error)
	GetUserByID(context.Context, string, GetUserByIDOption) (User, error)
	CreateUser(context.Context, User) (User, error)
	UpdateUser(context.Context, User) (User, error)
	DeleteUser(context.Context, string) error

	// library
	ListLibraries(context.Context, ListLibrariesOption) ([]Library, int, error)
	GetLibraryByID(context.Context, string) (Library, error)
	CreateLibrary(context.Context, Library) (Library, error)
	UpdateLibrary(context.Context, Library) (Library, error)
	DeleteLibrary(context.Context, string) error

	// book
	ListBooks(context.Context, ListBooksOption) ([]Book, int, error)
	GetBookByID(context.Context, uuid.UUID) (Book, error)
	CreateBook(context.Context, Book) (Book, error)
	UpdateBook(context.Context, Book) (Book, error)

	// staff
	ListStaffs(context.Context, ListStaffsOption) ([]Staff, int, error)
	CreateStaff(context.Context, Staff) (Staff, error)
	GetStaffByID(context.Context, uuid.UUID) (Staff, error)
	UpdateStaff(context.Context, Staff) (Staff, error)

	// membership
	ListMemberships(context.Context, ListMembershipsOption) ([]Membership, int, error)
	GetMembershipByID(context.Context, uuid.UUID) (Membership, error)
	CreateMembership(context.Context, Membership) (Membership, error)
	UpdateMembership(context.Context, Membership) (Membership, error)
	// DeleteMembership(context.Context, string) error

	// subscription
	ListSubscriptions(context.Context, ListSubscriptionsOption) ([]Subscription, int, error)
	GetSubscriptionByID(context.Context, uuid.UUID) (Subscription, error)
	CreateSubscription(context.Context, Subscription) (Subscription, error)
	UpdateSubscription(context.Context, Subscription) (Subscription, error)

	// borrowing
	ListBorrowings(context.Context, ListBorrowingsOption) ([]Borrowing, int, error)
	GetBorrowingByID(context.Context, uuid.UUID) (Borrowing, error)
	CreateBorrowing(context.Context, Borrowing) (Borrowing, error)
	UpdateBorrowing(context.Context, Borrowing) (Borrowing, error)

	// auth user
	CreateAuthUser(context.Context, AuthUser) (AuthUser, error)
	GetAuthUser(context.Context, GetAuthUserOption) (AuthUser, error)
}

type IdentityProvider interface {
	CreateUser(context.Context, RegisterUser) (string, error)
}

type Usecase struct {
	repo             Repository
	identityProvider IdentityProvider
}

func (u Usecase) Health() map[string]string {
	return u.repo.Health()
}

func (u Usecase) Close() error {
	return u.repo.Close()
}
