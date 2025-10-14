package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/librarease/librarease/internal/server"
)

func main() {
	app, err := server.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	// Server startup
	go func() {
		log.Printf("API server starting on %s", app.Addr())
		if err := app.ListenAndServe(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down API server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
		os.Exit(1)
	}

	log.Println("API server exited properly")
}
