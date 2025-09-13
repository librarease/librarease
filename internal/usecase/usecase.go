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
	dp Dispatcher,
) Usecase {
	return Usecase{
		repo:                repo,
		identityProvider:    ip,
		fileStorageProvider: fsp,
		mailer:              mp,
		dispatcher:          dp,
	}
}

type Repository interface {
	Health() map[string]string
	Close() error

	// user
	ListUsers(context.Context, ListUsersOption) ([]User, int, error)
	GetUserByID(context.Context, uuid.UUID, GetUserByIDOption) (User, error)
	CreateUser(context.Context, User) (User, error)
	UpdateUser(context.Context, uuid.UUID, User) (User, error)
	DeleteUser(context.Context, uuid.UUID) error

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
	CreateNotification(context.Context, Notification) (Notification, error)

	// push token
	SavePushToken(context.Context, uuid.UUID, string, PushProvider) error
	ListPushTokens(context.Context, ListPushTokensOption) ([]PushToken, int, error)
	DeletePushToken(context.Context, uuid.UUID) error

	// watchlist
	ListWatchlists(context.Context, ListWatchlistsOption) ([]Watchlist, int, error)
	// GetWatchlistByID(context.Context, uuid.UUID) (Watchlist, error)
	CreateWatchlist(context.Context, Watchlist) (Watchlist, error)
	DeleteWatchlist(context.Context, Watchlist) error

	// collection
	// ListCollections(context.Context, ListCollectionsOption) ([]Collection, int, error)
	// GetCollectionByID(context.Context, uuid.UUID) (Collection, error)
	// CreateCollection(context.Context, Collection) (Collection, error)
	// UpdateCollection(context.Context, uuid.UUID, Collection) (Collection, error)
	// DeleteCollection(context.Context, uuid.UUID) error

	// collection books
	// ListCollectionBooks(context.Context, ListCollectionBooksOption) ([]CollectionBook, int, error)
	// CreateCollectionBook(context.Context, CollectionBook) (CollectionBook, error)
	// DeleteCollectionBook(context.Context, uuid.UUID) error

	// collection followers
	// ListCollectionFollowers(context.Context, ListCollectionFollowersOption) ([]CollectionFollower, int, error)
	// CreateCollectionFollower(context.Context, CollectionFollower) (CollectionFollower, error)
	// DeleteCollectionFollower(context.Context, uuid.UUID) error
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

type Dispatcher interface {
	Send(context.Context, []PushToken, Notification) error
}

type Usecase struct {
	repo                Repository
	identityProvider    IdentityProvider
	fileStorageProvider FileStorageProvider
	mailer              Mailer
	dispatcher          Dispatcher
}

func (u Usecase) Health() map[string]string {
	return u.repo.Health()
}

func (u Usecase) Close() error {
	return u.repo.Close()
}
