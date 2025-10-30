package queue

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/hibiken/asynq"
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
	"github.com/librarease/librarease/internal/queue/handlers"
	"github.com/librarease/librarease/internal/telemetry"
	"github.com/librarease/librarease/internal/usecase"
)

// Server wraps asynq.Server for processing tasks
type Server struct {
	asynqServer *asynq.Server
	mux         *asynq.ServeMux
	gormDB      *gorm.DB
	sqlDB       *sql.DB
}

// Worker represents a worker application with all its dependencies
type Worker struct {
	server      *Server
	otelCleanup func(context.Context) error
}

// Scheduler represents a scheduler application with all its dependencies
type Scheduler struct {
	scheduler   *asynq.Scheduler
	otelCleanup func(context.Context) error
}

// NewWorker creates a fully configured worker with all dependencies
func NewWorker() (*Worker, error) {
	log.Println("Initializing worker dependencies...")

	// Setup database connection
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

	// Create GORM DB with the configured sql.DB
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to open gorm database connection: %w", err)
	}

	// Setup repository (workers don't need pgx.Conn for notifications)
	repo, err := database.New(gormDB, nil, nil)
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// Setup providers
	fb := firebase.New()
	mp := email.NewEmailProvider(
		os.Getenv(config.ENV_KEY_SMTP_HOST),
		os.Getenv(config.ENV_KEY_SMTP_USERNAME),
		os.Getenv(config.ENV_KEY_SMTP_PASSWORD),
		os.Getenv(config.ENV_KEY_SMTP_PORT),
	)

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

	// Create usecase without queue client (workers don't need to enqueue)
	uc := usecase.New(repo, fb, fsp, mp, dp, nil)

	// Setup Asynq server
	redisAddr := fmt.Sprintf("%s:%s",
		os.Getenv(config.ENV_KEY_REDIS_HOST),
		os.Getenv(config.ENV_KEY_REDIS_PORT),
	)
	redisPassword := os.Getenv(config.ENV_KEY_REDIS_PASSWORD)

	workerConcurrency := 10
	if wc := os.Getenv(config.ENV_KEY_WORKER_CONCURRENCY); wc != "" {
		var n int
		if _, err := fmt.Sscanf(wc, "%d", &n); err == nil && n > 0 {
			workerConcurrency = n
		}
	}

	asynqServer := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: redisPassword,
		},
		asynq.Config{
			Concurrency: workerConcurrency,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	mux := asynq.NewServeMux()
	h := handlers.NewHandlers(uc)

	mux.HandleFunc("export:borrowings", h.HandleExportBorrowings)
	mux.HandleFunc("notification:check-overdue", h.HandleCheckOverdue)
	mux.HandleFunc("import:books", h.HandleImportBooks)

	log.Println("Worker registered handlers:")
	log.Println(" - export:borrowings")
	log.Println(" - notification:check-overdue")
	log.Println(" - import:books")

	// Set up OpenTelemetry
	otelShutdown, err := telemetry.SetupOTelSDK(context.Background())
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to set up OpenTelemetry: %w", err)
	}

	server := &Server{
		asynqServer: asynqServer,
		mux:         mux,
		gormDB:      gormDB,
		sqlDB:       sqlDB,
	}

	return &Worker{
		server:      server,
		otelCleanup: otelShutdown,
	}, nil
}

// Start starts the worker server
func (w *Worker) Start() error {
	log.Println("Worker started successfully")
	return w.server.asynqServer.Start(w.server.mux)
}

// Stop stops the worker server gracefully
func (w *Worker) Stop() {
	log.Println("Stopping worker...")
	w.server.asynqServer.Shutdown()

	// Close database connections
	if w.server.sqlDB != nil {
		if err := w.server.sqlDB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}

	// Cleanup OpenTelemetry
	if w.otelCleanup != nil {
		ctx := context.Background()
		if err := w.otelCleanup(ctx); err != nil {
			log.Printf("Error cleaning up OpenTelemetry: %v", err)
		}
	}
}

// NewScheduler creates a fully configured scheduler with all dependencies
func NewScheduler() (*Scheduler, error) {
	log.Println("Initializing scheduler...")

	// Setup Redis connection (same as worker)
	redisAddr := fmt.Sprintf("%s:%s",
		os.Getenv(config.ENV_KEY_REDIS_HOST),
		os.Getenv(config.ENV_KEY_REDIS_PORT),
	)
	redisPassword := os.Getenv(config.ENV_KEY_REDIS_PASSWORD)

	// Create Asynq scheduler
	asynqScheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: redisPassword,
		},
		&asynq.SchedulerOpts{
			LogLevel: asynq.InfoLevel,
		},
	)

	// Register periodic tasks
	if err := registerPeriodicTasks(asynqScheduler); err != nil {
		return nil, fmt.Errorf("failed to register periodic tasks: %w", err)
	}

	// Set up OpenTelemetry
	otelShutdown, err := telemetry.SetupOTelSDK(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to set up OpenTelemetry: %w", err)
	}

	log.Println("Scheduler initialized successfully")

	return &Scheduler{
		scheduler:   asynqScheduler,
		otelCleanup: otelShutdown,
	}, nil
}

// registerPeriodicTasks registers all scheduled tasks
func registerPeriodicTasks(scheduler *asynq.Scheduler) error {
	log.Println("Registering periodic tasks...")

	// Recurring every hour
	entryID, err := scheduler.Register(
		"@every 1h",
		asynq.NewTask(
			"notification:check-overdue",
			nil,
			asynq.TaskID("unique-notification-check-overdue-task"),
		),
		asynq.Queue("default"),
	)
	if err != nil {
		return fmt.Errorf("failed to register overdue check task: %w", err)
	}

	log.Printf("Registered overdue check task with ID: %s", entryID)

	// You can add more periodic tasks here:
	//
	// // Weekly analytics report on Mondays at 8:00 AM
	// entryID, err = scheduler.Register(
	//     "0 8 * * 1",
	//     asynq.NewTask("analytics:weekly-report", nil),
	//     asynq.Queue("low"),
	// )
	// if err != nil {
	//     return fmt.Errorf("failed to register weekly analytics task: %w", err)
	// }

	log.Println("Periodic tasks registered:")
	log.Println("  - notification:check-overdue (every hour)")

	return nil
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	log.Println("Scheduler started successfully")
	return s.scheduler.Run()
}

// Stop stops the scheduler gracefully
func (s *Scheduler) Stop() {
	log.Println("Stopping scheduler...")
	s.scheduler.Shutdown()

	// Cleanup OpenTelemetry
	if s.otelCleanup != nil {
		ctx := context.Background()
		if err := s.otelCleanup(ctx); err != nil {
			log.Printf("Error cleaning up OpenTelemetry: %v", err)
		}
	}
}
