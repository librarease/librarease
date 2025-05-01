package usecase

import (
	"context"

	"github.com/google/uuid"
)

func New(
	repo Repository,
	ip IdentityProvider,
	fsp FileStorageProvider,
	mp Mailer,
) Usecase {
	return Usecase{
		repo:                repo,
		identityProvider:    ip,
		fileStorageProvider: fsp,
		mailer:              mp,
	}
}

type Repository interface {
	Health() map[string]string
	Close() error

	// user
	ListUsers(context.Context, ListUsersOption) ([]User, int, error)
	GetUserByID(context.Context, string, GetUserByIDOption) (User, error)
	CreateUser(context.Context, User) (User, error)
	UpdateUser(context.Context, uuid.UUID, User) (User, error)
	DeleteUser(context.Context, string) error

	// library
	ListLibraries(context.Context, ListLibrariesOption) ([]Library, int, error)
	GetLibraryByID(context.Context, uuid.UUID) (Library, error)
	CreateLibrary(context.Context, Library) (Library, error)
	UpdateLibrary(context.Context, uuid.UUID, Library) (Library, error)
	DeleteLibrary(context.Context, uuid.UUID) error

	// book
	ListBooks(context.Context, ListBooksOption) ([]Book, int, error)
	GetBookByID(context.Context, uuid.UUID) (Book, error)
	CreateBook(context.Context, Book) (Book, error)
	UpdateBook(context.Context, uuid.UUID, Book) (Book, error)

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

	// returning
	ReturnBorrowing(context.Context, uuid.UUID, Returning) (Borrowing, error)
	DeleteReturn(ctx context.Context, id uuid.UUID) error
	UpdateReturn(ctx context.Context, id uuid.UUID, r Returning) error

	// auth user
	CreateAuthUser(context.Context, AuthUser) (AuthUser, error)
	GetAuthUserByUID(context.Context, string) (AuthUser, error)
	GetAuthUserByUserID(context.Context, string) (AuthUser, error)

	// analysis
	GetAnalysis(context.Context, GetAnalysisOption) (Analysis, error)

	// notification
	SubscribeNotifications(context.Context, chan<- Notification) error
	UnsubscribeNotifications(context.Context, chan<- Notification) error
	ListNotifications(context.Context, ListNotificationsOption) ([]Notification, int, int, error)
	ReadNotification(context.Context, uuid.UUID) error
	ReadAllNotifications(context.Context, uuid.UUID) error
	CountUnreadNotifications(context.Context, uuid.UUID) (int, error)
}

type IdentityProvider interface {
	CreateUser(context.Context, RegisterUser) (string, error)
	VerifyIDToken(context.Context, string) (string, error)
	SetCustomClaims(context.Context, string, CustomClaims) error
}

type FileStorageProvider interface {
	GetTempUploadURL(context.Context, string) (string, error)
	// MoveTempFilePublic moves source in temp+path to public+dest
	MoveTempFilePublic(ctx context.Context, source string, dest string) error
	GetPublicURL(context.Context) (string, error)
}

type Mailer interface {
	SendEmail(context.Context, Email) error
}

type Usecase struct {
	repo                Repository
	identityProvider    IdentityProvider
	fileStorageProvider FileStorageProvider
	mailer              Mailer
}

func (u Usecase) Health() map[string]string {
	return u.repo.Health()
}

func (u Usecase) Close() error {
	return u.repo.Close()
}
