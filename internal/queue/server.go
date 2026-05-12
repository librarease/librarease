package queue

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"

	"github.com/hibiken/asynq"

	"github.com/librarease/librarease/internal/queue/handlers"
	"github.com/librarease/librarease/internal/usecase"
)

// Server wraps asynq.Server for processing tasks
type Server struct {
	asynqServer *asynq.Server
	mux         *asynq.ServeMux
	sqlDB       *sql.DB
	queueClient *Client
}

// Worker represents a worker application with all its dependencies
type Worker struct {
	logger      *slog.Logger
	server      *Server
	otelCleanup func(context.Context) error
}

// Scheduler represents a scheduler application with all its dependencies
type Scheduler struct {
	scheduler   *asynq.Scheduler
	otelCleanup func(context.Context) error
}

type WorkerDeps struct {
	Logger        *slog.Logger
	Service       usecase.Usecase
	SQL           *sql.DB
	QueueClient   *Client
	OTELCleanup   func(context.Context) error
	Concurrency   int
	RedisAddr     string
	RedisPassword string
}

type SchedulerDeps struct {
	Logger        *slog.Logger
	OTELCleanup   func(context.Context) error
	RedisAddr     string
	RedisPassword string
}

// NewWorker creates a fully configured worker with all dependencies
func NewWorker(deps WorkerDeps) (*Worker, error) {
	logger := deps.Logger
	logger.Info("Initializing worker dependencies...")

	workerConcurrency := deps.Concurrency
	if workerConcurrency <= 0 {
		workerConcurrency = 10
	}

	asynqServer := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     deps.RedisAddr,
			Password: deps.RedisPassword,
		},
		asynq.Config{
			Concurrency: workerConcurrency,
			Queues: map[string]int{
				"critical":         6,
				QueueNotifications: 4,
				QueueExports:       2,
				QueueImports:       2,
				QueueDefault:       3,
				"low":              1,
			},
		},
	)

	mux := asynq.NewServeMux()
	h := handlers.NewHandlers(deps.Service, logger)

	mux.HandleFunc(TaskExportBorrowings, h.HandleExportBorrowings)
	mux.HandleFunc(TaskNotificationCheckOverdue, h.HandleCheckOverdue)
	mux.HandleFunc(TaskNotificationCreate, h.HandleCreateNotification)
	mux.HandleFunc(TaskNotificationDeliver, h.HandleDeliverNotification)
	mux.HandleFunc(TaskImportBooks, h.HandleImportBooks)

	logger.Info("Worker registered handlers:",
		slog.String("handlers", "export:borrowings, notification:check-overdue, import:books"),
	)

	server := &Server{
		asynqServer: asynqServer,
		mux:         mux,
		sqlDB:       deps.SQL,
		queueClient: deps.QueueClient,
	}

	return &Worker{
		server:      server,
		otelCleanup: deps.OTELCleanup,
		logger:      logger,
	}, nil
}

// Start starts the worker server
func (w *Worker) Start() error {
	w.logger.Info("worker started successfully ")
	return w.server.asynqServer.Start(w.server.mux)
}

// Stop stops the worker server gracefully
func (w *Worker) Stop() {
	w.logger.Info("Stopping worker...")
	w.server.asynqServer.Shutdown()

	if w.server.queueClient != nil {
		if err := w.server.queueClient.Close(); err != nil {
			log.Printf("Error closing queue client: %v", err)
		}
	}

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
func NewScheduler(deps SchedulerDeps) (*Scheduler, error) {
	logger := deps.Logger
	logger.Info("Initializing scheduler...")

	// Create Asynq scheduler
	asynqScheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{
			Addr:     deps.RedisAddr,
			Password: deps.RedisPassword,
		},
		&asynq.SchedulerOpts{
			LogLevel: asynq.InfoLevel,
		},
	)

	// Register periodic tasks
	if err := registerPeriodicTasks(asynqScheduler, logger); err != nil {
		return nil, fmt.Errorf("failed to register periodic tasks: %w", err)
	}

	logger.Info("Scheduler initialized successfully")

	return &Scheduler{
		scheduler:   asynqScheduler,
		otelCleanup: deps.OTELCleanup,
	}, nil
}

// registerPeriodicTasks registers all scheduled tasks
func registerPeriodicTasks(scheduler *asynq.Scheduler, logger *slog.Logger) error {
	logger.Info("Registering periodic tasks...")

	// Recurring every hour
	entryID, err := scheduler.Register(
		"@every 1h",
		asynq.NewTask(
			TaskNotificationCheckOverdue,
			nil,
			asynq.TaskID("unique-notification-check-overdue-task"),
		),
		asynq.Queue(QueueNotifications),
	)
	if err != nil {
		return fmt.Errorf("failed to register overdue check task: %w", err)
	}

	logger.Info("Registered overdue check task", slog.String("entry_id", entryID))

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

	logger.Info("Periodic tasks registered:", slog.String("tasks", "notification:check-overdue (every hour)"))

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
