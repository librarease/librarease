package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/librarease/librarease/internal/queue"
)

func main() {
	var mode = flag.String("mode", "worker", "Mode to run: 'worker', 'scheduler'")
	flag.Parse()

	switch *mode {
	case "worker":
		runWorker()
	case "scheduler":
		runScheduler()
	default:
		log.Fatalf("Invalid mode: %s. Use 'worker' or 'scheduler'", *mode)
	}
}

func runWorker() {
	log.Println("Starting in WORKER mode...")

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

func runScheduler() {
	log.Println("Starting in SCHEDULER mode...")

	scheduler, err := queue.NewScheduler()
	if err != nil {
		log.Fatalf("Failed to create scheduler: %v", err)
	}

	// Start scheduler in goroutine
	go func() {
		log.Println("Starting Asynq scheduler...")
		if err := scheduler.Start(); err != nil {
			log.Printf("Scheduler error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down scheduler...")
	scheduler.Stop()
	log.Println("Scheduler exited properly")
}
