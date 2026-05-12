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
	"time"

	"github.com/librarease/librarease/internal/bootstrap"
	"github.com/librarease/librarease/internal/server"
)

func main() {
	cfg := bootstrap.LoadConfig()
	logger := bootstrap.NewLogger(cfg)

	apiService, err := bootstrap.NewAPIService(context.Background(), cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize API dependencies", slog.String("err", err.Error()))
		os.Exit(1)
	}

	app, err := server.NewApp(server.AppDeps{
		Service:     apiService.Service,
		SQL:         apiService.SQL,
		NotifyConn:  apiService.NotifyConn,
		RedisClient: apiService.RedisClient,
		QueueClient: apiService.QueueClient,
		OTELCleanup: apiService.OTELCleanup,
		Logger:      logger,
		Port:        apiService.Port,
	})
	if err != nil {
		logger.Error("Failed to create app", slog.String("err", err.Error()))
		os.Exit(1)
	}

	// Server startup
	go func() {
		logger.Info("API server starting", slog.String("addr", app.Addr()))

		if err := app.ListenAndServe(); err != nil {
			logger.Error("Server error", slog.String("err", err.Error()))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down API server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		logger.Error("Shutdown error", slog.String("err", err.Error()))
		os.Exit(1)
	}

	logger.Info("API server exited properly")
}
