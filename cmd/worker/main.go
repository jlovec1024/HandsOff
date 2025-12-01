package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/handsoff/handsoff/internal/task"
	"github.com/handsoff/handsoff/pkg/config"
	"github.com/handsoff/handsoff/pkg/database"
	"github.com/handsoff/handsoff/pkg/logger"
	"github.com/handsoff/handsoff/pkg/queue"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize logger
	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	defer log.Sync()

	log.Info("Starting HandsOff Worker...")

	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}

	// Initialize task queue
	client := queue.NewClient(cfg.Redis)

	// Create and start worker server
	srv := task.NewServer(db, cfg, log)
	
	if err := srv.Start(); err != nil {
		log.Fatal("Failed to start worker", "error", err)
	}

	log.Info("Worker started", "concurrency", cfg.Worker.Concurrency)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down worker...")
	
	srv.Shutdown()
	client.Close()

	log.Info("Worker exited")
}
