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

	"github.com/librarease/librarease/internal/config"
	"github.com/librarease/librarease/internal/server"
	"github.com/librarease/librarease/internal/telemetry"
)

func main() {
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
	app, err := server.NewApp(logger)
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
