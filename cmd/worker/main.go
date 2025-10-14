package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/librarease/librarease/internal/queue"
)

func main() {
	worker, err := queue.NewWorker()
	if err != nil {
		log.Fatalf("Failed to create worker: %v", err)
	}

	// Start worker in goroutine
	go func() {
		log.Println("Starting Asynq worker...")
		if err := worker.Start(); err != nil {
			log.Printf("Worker error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down worker...")
	worker.Stop()

	log.Println("Worker exited properly")
}
