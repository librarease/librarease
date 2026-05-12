package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"

	"github.com/librarease/librarease/internal/queue"
	"github.com/librarease/librarease/internal/usecase"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	// ListUsers returns a list of users.
	// FIXME: user model, input params
	ListUsers(context.Context, usecase.ListUsersOption) ([]usecase.User, int, error)
	GetUserByID(context.Context, uuid.UUID, usecase.GetUserByIDOption) (usecase.User, error)
	CreateUser(context.Context, usecase.User) (usecase.User, error)
	UpdateUser(context.Context, uuid.UUID, usecase.User) (usecase.User, error)
	DeleteUser(context.Context, uuid.UUID) error
	GetAuthUserByUID(context.Context, string) (usecase.AuthUser, error)
	GetAuthUserByUserID(context.Context, string) (usecase.AuthUser, error)
	GetMe(context.Context) (usecase.MeUser, error)

	ListLibraries(context.Context, usecase.ListLibrariesOption) ([]usecase.Library, int, error)
	GetLibraryByID(context.Context, uuid.UUID) (usecase.Library, error)
	CreateLibrary(context.Context, usecase.Library) (usecase.Library, error)
	UpdateLibrary(context.Context, uuid.UUID, usecase.Library) (usecase.Library, error)
	DeleteLibrary(context.Context, uuid.UUID) error

	ListStaffs(context.Context, usecase.ListStaffsOption) ([]usecase.Staff, int, error)
	CreateStaff(context.Context, usecase.Staff) (usecase.Staff, error)
	GetStaffByID(context.Context, string) (usecase.Staff, error)
	UpdateStaff(context.Context, usecase.Staff) (usecase.Staff, error)
	DeleteStaff(context.Context, uuid.UUID) error

	ListBooks(context.Context, usecase.ListBooksOption) ([]usecase.Book, int, error)
	GetBookByID(context.Context, uuid.UUID, usecase.GetBookByIDOption) (usecase.Book, error)
	CreateBook(context.Context, usecase.Book) (usecase.Book, error)
	UpdateBook(context.Context, uuid.UUID, usecase.Book) (usecase.Book, error)
	DeleteBook(context.Context, uuid.UUID) error
	PreviewImportBooks(context.Context, uuid.UUID, string) (usecase.PreviewImportBooksResult, error)
	ConfirmImportBooks(context.Context, uuid.UUID, string) (string, error)

	ListMemberships(context.Context, usecase.ListMembershipsOption) ([]usecase.Membership, int, error)
	GetMembershipByID(context.Context, string) (usecase.Membership, error)
	CreateMembership(context.Context, usecase.Membership) (usecase.Membership, error)
	UpdateMembership(context.Context, usecase.Membership) (usecase.Membership, error)
	DeleteMembership(context.Context, uuid.UUID) error

	ListSubscriptions(context.Context, usecase.ListSubscriptionsOption) ([]usecase.Subscription, int, error)
	GetSubscriptionByID(context.Context, uuid.UUID) (usecase.Subscription, error)
	CreateSubscription(context.Context, usecase.Subscription) (usecase.Subscription, error)
	UpdateSubscription(context.Context, usecase.Subscription) (usecase.Subscription, error)
	DeleteSubscription(context.Context, uuid.UUID) error

	ListBorrowings(context.Context, usecase.ListBorrowingsOption) ([]usecase.Borrowing, int, error)
	GetBorrowingByID(context.Context, uuid.UUID, usecase.BorrowingsOption) (usecase.Borrowing, error)
	CreateBorrowing(context.Context, usecase.Borrowing) (usecase.Borrowing, error)
	UpdateBorrowing(context.Context, usecase.Borrowing) (usecase.Borrowing, error)
	DeleteBorrowing(context.Context, uuid.UUID) error
	ExportBorrowings(context.Context, usecase.ExportBorrowingsOption) (string, error)

	ReturnBorrowing(context.Context, uuid.UUID, usecase.Returning) (usecase.Borrowing, error)
	DeleteReturn(context.Context, uuid.UUID) error
	UpdateReturn(context.Context, uuid.UUID, usecase.Returning) error

	LostBorrowing(context.Context, uuid.UUID, usecase.Lost) (usecase.Lost, error)
	UpdateLost(context.Context, uuid.UUID, usecase.Lost) (usecase.Lost, error)
	DeleteLost(context.Context, uuid.UUID) error

	RegisterUser(context.Context, usecase.RegisterUser) (usecase.User, error)
	VerifyIDToken(context.Context, string) (string, error)

	GetAnalysis(context.Context, usecase.GetAnalysisOption) (usecase.Analysis, error)
	OverdueAnalysis(context.Context, *time.Time, *time.Time, string) ([]usecase.OverdueAnalysis, error)
	BorrowingHeatmap(context.Context, uuid.UUID, *time.Time, *time.Time) ([]usecase.HeatmapCell, error)
	ReturningHeatmap(context.Context, uuid.UUID, *time.Time, *time.Time) ([]usecase.HeatmapCell, error)
	GetPowerUsers(context.Context, usecase.GetPowerUsersOption) ([]usecase.PowerUser, int, error)
	GetLongestUnreturned(context.Context, usecase.GetOverdueBorrowsOption) ([]usecase.OverdueBorrow, int, error)

	GetDocs(context.Context, usecase.GetDocsOption) (string, error)

	GetTempUploadURL(context.Context, string) (string, string, error)

	ListNotifications(context.Context, usecase.ListNotificationsOption) ([]usecase.Notification, int, int, error)
	ReadNotification(context.Context, uuid.UUID) error
	ReadAllNotifications(context.Context) error
	StreamNotifications(context.Context, uuid.UUID) (<-chan usecase.Notification, error)
	CreateNotification(context.Context, usecase.Notification) error

	SavePushToken(context.Context, string, usecase.PushProvider) error

	// watchlist
	CreateWatchlist(context.Context, usecase.Watchlist) (usecase.Watchlist, error)
	DeleteWatchlist(context.Context, usecase.Watchlist) error

	// collection
	ListCollections(context.Context, usecase.ListCollectionsOption) ([]usecase.Collection, int, error)
	GetCollectionByID(context.Context, uuid.UUID, usecase.GetCollectionOption) (usecase.Collection, error)
	CreateCollection(context.Context, usecase.Collection) (usecase.Collection, error)
	UpdateCollection(context.Context, uuid.UUID, usecase.UpdateCollectionRequest) (usecase.Collection, error)
	DeleteCollection(context.Context, uuid.UUID) error

	// collection books
	ListCollectionBooks(context.Context, uuid.UUID, usecase.ListCollectionBooksOption) ([]usecase.CollectionBook, int, error)
	UpdateCollectionBooks(context.Context, uuid.UUID, []uuid.UUID) ([]usecase.CollectionBook, error)

	// collection followers
	ListCollectionFollowers(context.Context, usecase.ListCollectionFollowersOption) ([]usecase.CollectionFollower, int, error)
	CreateCollectionFollower(context.Context, uuid.UUID) (usecase.CollectionFollower, error)
	DeleteCollectionFollower(context.Context, uuid.UUID) error

	// job
	ListJobs(context.Context, usecase.ListJobsOption) ([]usecase.Job, int, error)
	GetJobByID(context.Context, uuid.UUID) (usecase.Job, error)
	// do not expose CreateJob - jobs are created internally by the system
	CreateJob(context.Context, usecase.Job) (usecase.Job, error)
	UpdateJob(context.Context, usecase.Job) (usecase.Job, error)
	DeleteJob(context.Context, uuid.UUID) error
	DownloadJobAsset(context.Context, uuid.UUID) (string, error)

	// review
	ListReviews(context.Context, usecase.ListReviewsOption) ([]usecase.Review, int, error)
	GetReview(context.Context, uuid.UUID, usecase.ReviewsOption) (usecase.Review, error)
	CreateReview(context.Context, usecase.Review) (usecase.Review, error)
	UpdateReview(context.Context, uuid.UUID, usecase.Review) (usecase.Review, error)
	DeleteReview(context.Context, uuid.UUID) error
}

type Server struct {
	port int

	server    Service
	validator *validator.Validate
	logger    *slog.Logger
}

type App struct {
	httpServer  *http.Server
	sqlDB       *sql.DB
	notifyConn  *pgx.Conn
	redisClient *redis.Client
	queueClient *queue.Client
	otelCleanup func(context.Context) error
	logger      *slog.Logger
}

type AppDeps struct {
	Service     Service
	SQL         *sql.DB
	NotifyConn  *pgx.Conn
	RedisClient *redis.Client
	QueueClient *queue.Client
	OTELCleanup func(context.Context) error
	Logger      *slog.Logger
	Port        int
}

func (a *App) ListenAndServe() error {
	if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server error: %w", err)
	}
	return nil
}

func (a *App) Addr() string {
	return a.httpServer.Addr
}

func (a *App) Shutdown(ctx context.Context) error {
	var errs []error

	if err := a.notifyConn.Close(ctx); err != nil {
		errs = append(errs, fmt.Errorf("notify connection close: %w", err))
	}

	if err := a.sqlDB.Close(); err != nil {
		errs = append(errs, fmt.Errorf("sql db close: %w", err))
	}

	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("redis close: %w", err))
		}
	}

	if a.queueClient != nil {
		if err := a.queueClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("queue client close: %w", err))
		}
	}

	if err := a.httpServer.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("http server shutdown: %w", err))
	}

	if err := a.otelCleanup(ctx); err != nil {
		errs = append(errs, fmt.Errorf("telemetry cleanup: %w", err))
	}

	return errors.Join(errs...)
}

func NewApp(deps AppDeps) (*App, error) {
	v := validator.New()

	s := &Server{
		port:      deps.Port,
		server:    deps.Service,
		validator: v,
		logger:    deps.Logger.With(slog.String("component", "server")),
	}

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0,
	}

	return &App{
		httpServer:  httpServer,
		sqlDB:       deps.SQL,
		notifyConn:  deps.NotifyConn,
		redisClient: deps.RedisClient,
		queueClient: deps.QueueClient,
		otelCleanup: deps.OTELCleanup,
		logger:      deps.Logger,
	}, nil
}
