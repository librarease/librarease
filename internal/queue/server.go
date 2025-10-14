package queue

import (
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
	server *Server
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

	// Create handlers instance
	h := handlers.NewHandlers(uc)

	// Register task handlers - one line per job type
	mux.HandleFunc("export:borrowings", h.HandleExportBorrowings)
	// Add more handlers here as needed

	log.Println("Worker registered handlers:")
	log.Println("  - export:borrowings")

	server := &Server{
		asynqServer: asynqServer,
		mux:         mux,
		gormDB:      gormDB,
		sqlDB:       sqlDB,
	}

	return &Worker{
		server: server,
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
}
