// Copyright (c) 2025 [LibrarEase]
//
// This software is licensed under the PolyForm Noncommercial License 1.0.0
// See LICENSE file in the project root for full license terms.
//
// For commercial licensing inquiries, contact: solidifyarmor@gmail.com
// https://github.com/librarease/librarease

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/librarease/librarease/internal/bootstrap"
	"github.com/librarease/librarease/internal/queue"
)

func main() {
	cfg := bootstrap.LoadConfig()
	logger := bootstrap.NewLogger(cfg)

	logger.Info("Starting in SCHEDULER mode...")

	schedulerService, err := bootstrap.NewSchedulerService(context.Background(), cfg)
	if err != nil {
		logger.Error("Failed to initialize scheduler dependencies", slog.String("err", err.Error()))
		os.Exit(1)
	}

	scheduler, err := queue.NewScheduler(queue.SchedulerDeps{
		Logger:        logger,
		OTELCleanup:   schedulerService.OTELCleanup,
		RedisAddr:     schedulerService.RedisAddr,
		RedisPassword: schedulerService.RedisPassword,
	})
	if err != nil {
		logger.Error("Failed to create scheduler", slog.String("err", err.Error()))
		os.Exit(1)
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("Starting Asynq scheduler...")
		errCh <- scheduler.Start()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info("Shutting down scheduler...")
		scheduler.Stop()
		logger.Info("Scheduler exited properly")
	case err := <-errCh:
		if err != nil {
			logger.Error("Scheduler error", slog.String("err", err.Error()))
		}
		scheduler.Stop()
		os.Exit(1)
	}
}
