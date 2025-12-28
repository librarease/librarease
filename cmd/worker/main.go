// Copyright (c) 2025 [LibrarEase]
//
// This software is licensed under the PolyForm Noncommercial License 1.0.0
// See LICENSE file in the project root for full license terms.
//
// For commercial licensing inquiries, contact: solidifyarmor@gmail.com
// https://github.com/librarease/librarease

package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/librarease/librarease/internal/config"
	"github.com/librarease/librarease/internal/queue"
	"github.com/librarease/librarease/internal/telemetry"
)

func main() {
	var mode = flag.String("mode", "worker", "Mode to run: 'worker', 'scheduler'")
	flag.Parse()

	level := slog.LevelInfo
	if lvl := os.Getenv(config.ENV_KEY_LOG_LEVEL); lvl != "" {
		switch lvl {
		case "DEBUG":
			level = slog.LevelDebug
		case "INFO":
			level = slog.LevelInfo
		case "WARN":
			level = slog.LevelWarn
		case "ERROR":
			level = slog.LevelError
		default:
			level = slog.LevelInfo
		}
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	logger := slog.New(telemetry.NewTraceHandler(jsonHandler))

	switch *mode {
	case "worker":
		runWorker(logger)
	case "scheduler":
		runScheduler(logger)
	default:
		logger.Error("Invalid mode. Use 'worker' or 'scheduler'", slog.String("mode", *mode))
		os.Exit(1)
	}
}

func runWorker(logger *slog.Logger) {
	logger.Info("Starting in WORKER mode...")

	worker, err := queue.NewWorker(logger)
	if err != nil {
		logger.Error("Failed to create worker", slog.String("err", err.Error()))
		os.Exit(1)
	}

	// Start worker in goroutine
	go func() {
		logger.Info("Starting Asynq worker...")
		if err := worker.Start(); err != nil {
			logger.Error("Worker error", slog.String("err", err.Error()))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down worker...")
	worker.Stop()
	logger.Info("Worker exited properly")
}

func runScheduler(logger *slog.Logger) {
	logger.Info("Starting in SCHEDULER mode...")

	scheduler, err := queue.NewScheduler(logger)
	if err != nil {
		logger.Error("Failed to create scheduler", slog.String("err", err.Error()))
		os.Exit(1)
	}

	// Start scheduler in goroutine
	go func() {
		logger.Info("Starting Asynq scheduler...")
		if err := scheduler.Start(); err != nil {
			logger.Error("Scheduler error", slog.String("err", err.Error()))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down scheduler...")
	scheduler.Stop()
	logger.Info("Scheduler exited properly")
}
