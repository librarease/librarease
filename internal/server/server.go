package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"

	"github.com/librarease/librarease/internal/config"
	"github.com/librarease/librarease/internal/database"
	"github.com/librarease/librarease/internal/email"
	"github.com/librarease/librarease/internal/filestorage"
	"github.com/librarease/librarease/internal/firebase"
	"github.com/librarease/librarease/internal/push"
	"github.com/librarease/librarease/internal/queue"
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
	// ListCollectionFollowers(context.Context, usecase.ListCollectionFollowersOption) ([]usecase.CollectionFollower, int, error)
	// CreateCollectionFollower(context.Context, usecase.CollectionFollower) (usecase.CollectionFollower, error)
	// DeleteCollectionFollower(context.Context, uuid.UUID) error

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
	gormDB      *gorm.DB
	sqlDB       *sql.DB
	notifyConn  *pgx.Conn
	otelCleanup func(context.Context) error
	logger      *slog.Logger
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

func NewApp(logger *slog.Logger) (*App, error) {

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

	// Configure connection pool
	maxOpenConnections := 25
	if m, err := strconv.Atoi(os.Getenv(config.ENV_KEY_DB_MAX_OPEN_CONNECTIONS)); err == nil && m > 0 {
		maxOpenConnections = m
	}
	sqlDB.SetMaxOpenConns(maxOpenConnections)

	maxIdleConnections := 10
	if m, err := strconv.Atoi(os.Getenv(config.ENV_KEY_DB_MAX_IDLE_CONNECTIONS)); err == nil && m > 0 {
		maxIdleConnections = m
	}
	sqlDB.SetMaxIdleConns(maxIdleConnections)

	connMaxLifetime := 5 * time.Minute
	if m, err := strconv.Atoi(os.Getenv(config.ENV_KEY_DB_CONN_MAX_LIFETIME_MINUTES)); err == nil && m > 0 {
		connMaxLifetime = time.Duration(m) * time.Minute
	}
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	connMaxIdleTime := 2 * time.Minute
	if m, err := strconv.Atoi(os.Getenv(config.ENV_KEY_DB_CONN_MAX_IDLE_TIME_MINUTES)); err == nil && m > 0 {
		connMaxIdleTime = time.Duration(m) * time.Minute
	}
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	// Create GORM DB with the configured sql.DB
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: database.NewSlogGormLogger(logger),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm database connection: %w", err)
	}
	if err := gormDB.Use(tracing.NewPlugin()); err != nil {
		return nil, err
	}

	// Create separate pgx.Conn for notifications
	notifyConn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		sqlDB.Close() // cleanup previous connection
		return nil, fmt.Errorf("failed to connect to database for notification: %w", err)
	}
	var (
		redisAddr     = os.Getenv(config.ENV_KEY_REDIS_HOST) + ":" + os.Getenv(config.ENV_KEY_REDIS_PORT)
		redisPassword = os.Getenv(config.ENV_KEY_REDIS_PASSWORD)
	)

	redis := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0, // use default DB
	})

	// Enable OpenTelemetry tracing for Redis
	if err := redisotel.InstrumentTracing(redis); err != nil {
		sqlDB.Close()
		notifyConn.Close(context.Background())
		return nil, fmt.Errorf("failed to instrument redis tracing: %w", err)
	}

	// Optional: Enable metrics
	if err := redisotel.InstrumentMetrics(redis); err != nil {
		sqlDB.Close()
		notifyConn.Close(context.Background())
		return nil, fmt.Errorf("failed to instrument redis metrics: %w", err)
	}

	repo, err := database.New(gormDB, notifyConn, redis)
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

	qc := queue.NewClient(redisAddr, redisPassword)

	sv := usecase.New(repo, fb, fsp, mp, dp, qc, logger.With(slog.String("component", "usecase")))
	v := validator.New()

	port, _ := strconv.Atoi(os.Getenv(config.ENV_KEY_PORT))
	s := &Server{
		port:      port,
		server:    sv,
		validator: v,
		logger:    logger.With(slog.String("component", "server")),
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
		logger:      logger,
	}, nil
}
