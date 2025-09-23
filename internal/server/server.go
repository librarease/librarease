package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/librarease/librarease/internal/config"
	"github.com/librarease/librarease/internal/database"
	"github.com/librarease/librarease/internal/email"
	"github.com/librarease/librarease/internal/filestorage"
	"github.com/librarease/librarease/internal/firebase"
	"github.com/librarease/librarease/internal/push"
	"github.com/librarease/librarease/internal/telemetry"
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

	ListBooks(context.Context, usecase.ListBooksOption) ([]usecase.Book, int, error)
	GetBookByID(context.Context, uuid.UUID, usecase.GetBookByIDOption) (usecase.Book, error)
	CreateBook(context.Context, usecase.Book) (usecase.Book, error)
	UpdateBook(context.Context, uuid.UUID, usecase.Book) (usecase.Book, error)

	ListMemberships(context.Context, usecase.ListMembershipsOption) ([]usecase.Membership, int, error)
	GetMembershipByID(context.Context, string) (usecase.Membership, error)
	CreateMembership(context.Context, usecase.Membership) (usecase.Membership, error)
	UpdateMembership(context.Context, usecase.Membership) (usecase.Membership, error)
	// DeleteMembership(context.Context, string) error

	ListSubscriptions(context.Context, usecase.ListSubscriptionsOption) ([]usecase.Subscription, int, error)
	GetSubscriptionByID(context.Context, uuid.UUID) (usecase.Subscription, error)
	CreateSubscription(context.Context, usecase.Subscription) (usecase.Subscription, error)
	UpdateSubscription(context.Context, usecase.Subscription) (usecase.Subscription, error)

	ListBorrowings(context.Context, usecase.ListBorrowingsOption) ([]usecase.Borrowing, int, error)
	GetBorrowingByID(context.Context, uuid.UUID) (usecase.Borrowing, error)
	CreateBorrowing(context.Context, usecase.Borrowing) (usecase.Borrowing, error)
	UpdateBorrowing(context.Context, usecase.Borrowing) (usecase.Borrowing, error)

	ReturnBorrowing(context.Context, uuid.UUID, usecase.Returning) (usecase.Borrowing, error)
	DeleteReturn(context.Context, uuid.UUID) error
	UpdateReturn(context.Context, uuid.UUID, usecase.Returning) error

	RegisterUser(context.Context, usecase.RegisterUser) (usecase.User, error)
	VerifyIDToken(context.Context, string) (string, error)

	GetAnalysis(context.Context, usecase.GetAnalysisOption) (usecase.Analysis, error)
	OverdueAnalysis(context.Context, *time.Time, *time.Time, string) ([]usecase.OverdueAnalysis, error)
	BookUtilization(context.Context, usecase.GetBookUtilizationOption) ([]usecase.BookUtilization, int, error)
	BorrowingHeatmap(context.Context, uuid.UUID, *time.Time, *time.Time) ([]usecase.BorrowHeatmapCell, error)
	GetPowerUsers(context.Context, usecase.GetPowerUsersOption) ([]usecase.PowerUser, int, error)
	GetLongestUnreturned(context.Context, usecase.GetOverdueBorrowsOption) ([]usecase.OverdueBorrow, int, error)

	GetDocs(context.Context, usecase.GetDocsOption) (string, error)

	GetTempUploadURL(context.Context, string) (string, error)

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
	// ListCollectionFollowers(context.Context, usecase.ListCollectionFollowersOption) ([]usecase.CollectionFollower, int, error)
	// CreateCollectionFollower(context.Context, usecase.CollectionFollower) (usecase.CollectionFollower, error)
	// DeleteCollectionFollower(context.Context, uuid.UUID) error
}

type Server struct {
	port int

	server    Service
	validator *validator.Validate
}

type App struct {
	httpServer  *http.Server
	gormDB      *gorm.DB
	sqlDB       *sql.DB
	notifyConn  *pgx.Conn
	otelCleanup func(context.Context) error
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

	if err := a.httpServer.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("http server shutdown: %w", err))
	}

	if err := a.otelCleanup(ctx); err != nil {
		errs = append(errs, fmt.Errorf("telemetry cleanup: %w", err))
	}

	return errors.Join(errs...)
}

func NewApp() (*App, error) {

	var (
		dbname = os.Getenv(config.ENV_KEY_DB_DATABASE)
		dbpass = os.Getenv(config.ENV_KEY_DB_PASSWORD)
		dbuser = os.Getenv(config.ENV_KEY_DB_USER)
		dbport = os.Getenv(config.ENV_KEY_DB_PORT)
		dbhost = os.Getenv(config.ENV_KEY_DB_HOST)
	)

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbuser, dbpass, dbhost, dbport, dbname)
	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	var maxOpenConnections int
	if m, err := strconv.Atoi(
		os.Getenv("DB_MAX_OPEN_CONNECTIONS")); err == nil {
		maxOpenConnections = m
	}
	sqlDB.SetMaxOpenConns(maxOpenConnections)

	// Create GORM DB with the configured sql.DB
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm database connection: %w", err)
	}

	// Create separate pgx.Conn for notifications
	notifyConn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		sqlDB.Close() // cleanup previous connection
		return nil, fmt.Errorf("failed to connect to database for notification: %w", err)
	}

	// rdb := redis.NewClient(&redis.Options{
	// 	Addr:     "localhost:6379",
	// 	Password: "", // no password set
	// 	DB:       0,  // use default DB
	// })
	repo, err := database.New(gormDB, notifyConn, nil)
	if err != nil {
		sqlDB.Close()
		notifyConn.Close(context.Background())
		return nil, fmt.Errorf("failed to create database repository: %w", err)
	}
	fb := firebase.New()
	mp := email.NewEmailProvider(
		os.Getenv(config.ENV_KEY_SMTP_HOST),
		os.Getenv(config.ENV_KEY_SMTP_USERNAME),
		os.Getenv(config.ENV_KEY_SMTP_PASSWORD),
		os.Getenv(config.ENV_KEY_SMTP_PORT),
	)

	// AWS S3
	// var (
	// 	bucket   = os.Getenv(config.ENV_KEY_S3_BUCKET)
	// 	tempPath = os.Getenv(config.ENV_KEY_S3_TEMP_PATH)
	// )
	// fsp := filestorage.NewS3Storage(bucket, tempPath)

	// MinIO (S3 compatible)
	var (
		bucket    = os.Getenv(config.ENV_KEY_MINIO_BUCKET)
		temp      = os.Getenv(config.ENV_KEY_MINIO_TEMP_PATH)
		public    = os.Getenv(config.ENV_KEY_MINIO_PUBLIC_PATH)
		endpoint  = os.Getenv(config.ENV_KEY_MINIO_ENDPOINT)
		accessKey = os.Getenv(config.ENV_KEY_MINIO_ACCESS_KEY)
		secretKey = os.Getenv(config.ENV_KEY_MINIO_SECRET_KEY)
	)
	fsp := filestorage.NewMinIOStorage(bucket, temp, public, endpoint, accessKey, secretKey)

	dp := push.NewPushDispatcher(fb)

	sv := usecase.New(repo, fb, fsp, mp, dp)
	v := validator.New()

	port, _ := strconv.Atoi(os.Getenv(config.ENV_KEY_PORT))
	s := &Server{
		port:      port,
		server:    sv,
		validator: v,
	}

	// Set up OpenTelemetry.
	otelShutdown, err := telemetry.SetupOTelSDK(context.Background())
	if err != nil {
		sqlDB.Close()
		notifyConn.Close(context.Background())
		return nil, fmt.Errorf("failed to set up OpenTelemetry: %w", err)
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
		gormDB:      gormDB,
		sqlDB:       sqlDB,
		notifyConn:  notifyConn,
		otelCleanup: otelShutdown,
	}, nil
}
