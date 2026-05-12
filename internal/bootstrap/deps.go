package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"

	"github.com/librarease/librarease/internal/database"
	"github.com/librarease/librarease/internal/email"
	"github.com/librarease/librarease/internal/filestorage"
	"github.com/librarease/librarease/internal/firebase"
	"github.com/librarease/librarease/internal/push"
	"github.com/librarease/librarease/internal/queue"
	"github.com/librarease/librarease/internal/telemetry"
	"github.com/librarease/librarease/internal/usecase"
)

type SQLStore struct {
	SQL  *sql.DB
	GORM *gorm.DB
}

type APIService struct {
	Service     usecase.Usecase
	SQL         *sql.DB
	NotifyConn  *pgx.Conn
	RedisClient *redis.Client
	QueueClient *queue.Client
	OTELCleanup func(context.Context) error
	Port        int
}

type WorkerService struct {
	Service       usecase.Usecase
	SQL           *sql.DB
	QueueClient   *queue.Client
	OTELCleanup   func(context.Context) error
	Concurrency   int
	RedisAddr     string
	RedisPassword string
}

type SchedulerService struct {
	OTELCleanup   func(context.Context) error
	RedisAddr     string
	RedisPassword string
}

type SQLStoreOptions struct {
	ConfigurePool bool
	TraceGORM     bool
}

func NewSQLStore(cfg Config, logger *slog.Logger, opts SQLStoreOptions) (*SQLStore, error) {
	sqlDB, err := sql.Open("pgx", cfg.DB.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if opts.ConfigurePool {
		sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConnections)
		sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConnections)
		sqlDB.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)
		sqlDB.SetConnMaxIdleTime(cfg.DB.ConnMaxIdleTime)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: database.NewSlogGormLogger(logger),
	})
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to open gorm database connection: %w", err)
	}

	if opts.TraceGORM {
		if err := gormDB.Use(tracing.NewPlugin()); err != nil {
			sqlDB.Close()
			return nil, fmt.Errorf("failed to enable gorm tracing: %w", err)
		}
	}

	return &SQLStore{SQL: sqlDB, GORM: gormDB}, nil
}

func NewNotificationConn(ctx context.Context, cfg Config) (*pgx.Conn, error) {
	notifyConn, err := pgx.Connect(ctx, cfg.DB.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database for notification: %w", err)
	}
	return notifyConn, nil
}

func NewRedisCache(cfg Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       0,
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
	})

	if err := redisotel.InstrumentTracing(client); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to instrument redis tracing: %w", err)
	}

	if err := redisotel.InstrumentMetrics(client); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to instrument redis metrics: %w", err)
	}

	return client, nil
}

func NewUsecase(logger *slog.Logger, repo usecase.Repository, cfg Config) (usecase.Usecase, *queue.Client) {
	fb := firebase.New()
	mailer := email.NewEmailProvider(
		cfg.SMTP.Host,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.Port,
	)
	fileStorage := filestorage.NewMinIOStorage(
		cfg.MinIO.Bucket,
		cfg.MinIO.TempPath,
		cfg.MinIO.PublicPath,
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKey,
		cfg.MinIO.SecretKey,
	)
	queueClient := queue.NewClient(cfg.Redis.Addr(), cfg.Redis.Password)
	dispatcher := push.NewPushDispatcher(fb)

	return usecase.New(
		repo,
		fb,
		fileStorage,
		mailer,
		dispatcher,
		queueClient,
		logger.With(slog.String("component", "usecase")),
	), queueClient
}

func SetupTelemetry(ctx context.Context) (func(context.Context) error, error) {
	otelShutdown, err := telemetry.SetupOTelSDK(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to set up OpenTelemetry: %w", err)
	}
	return otelShutdown, nil
}

func NewAPIService(ctx context.Context, cfg Config, logger *slog.Logger) (*APIService, error) {
	store, err := NewSQLStore(cfg, logger, SQLStoreOptions{
		ConfigurePool: true,
		TraceGORM:     true,
	})
	if err != nil {
		return nil, err
	}

	notifyConn, err := NewNotificationConn(ctx, cfg)
	if err != nil {
		store.SQL.Close()
		return nil, err
	}

	redisClient, err := NewRedisCache(cfg)
	if err != nil {
		store.SQL.Close()
		notifyConn.Close(context.Background())
		return nil, err
	}

	repo, err := database.New(store.GORM, notifyConn, redisClient)
	if err != nil {
		store.SQL.Close()
		notifyConn.Close(context.Background())
		redisClient.Close()
		return nil, fmt.Errorf("failed to create database repository: %w", err)
	}

	service, queueClient := NewUsecase(logger, repo, cfg)

	otelShutdown, err := SetupTelemetry(ctx)
	if err != nil {
		store.SQL.Close()
		notifyConn.Close(context.Background())
		redisClient.Close()
		queueClient.Close()
		return nil, err
	}

	return &APIService{
		Service:     service,
		SQL:         store.SQL,
		NotifyConn:  notifyConn,
		RedisClient: redisClient,
		QueueClient: queueClient,
		OTELCleanup: otelShutdown,
		Port:        cfg.Port,
	}, nil
}

func NewWorkerService(ctx context.Context, cfg Config, logger *slog.Logger) (*WorkerService, error) {
	store, err := NewSQLStore(cfg, logger, SQLStoreOptions{})
	if err != nil {
		return nil, err
	}

	repo, err := database.New(store.GORM, nil, nil)
	if err != nil {
		store.SQL.Close()
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	service, queueClient := NewUsecase(logger, repo, cfg)

	otelShutdown, err := SetupTelemetry(ctx)
	if err != nil {
		store.SQL.Close()
		queueClient.Close()
		return nil, err
	}

	return &WorkerService{
		Service:       service,
		SQL:           store.SQL,
		QueueClient:   queueClient,
		OTELCleanup:   otelShutdown,
		Concurrency:   cfg.Worker.Concurrency,
		RedisAddr:     cfg.Redis.Addr(),
		RedisPassword: cfg.Redis.Password,
	}, nil
}

func NewSchedulerService(ctx context.Context, cfg Config) (*SchedulerService, error) {
	otelShutdown, err := SetupTelemetry(ctx)
	if err != nil {
		return nil, err
	}

	return &SchedulerService{
		OTELCleanup:   otelShutdown,
		RedisAddr:     cfg.Redis.Addr(),
		RedisPassword: cfg.Redis.Password,
	}, nil
}
