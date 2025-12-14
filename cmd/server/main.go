package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/handsoff/handsoff/internal/api/router"
	"github.com/handsoff/handsoff/internal/task"
	"github.com/handsoff/handsoff/pkg/config"
	"github.com/handsoff/handsoff/pkg/database"
	"github.com/handsoff/handsoff/pkg/initializer"
	"github.com/handsoff/handsoff/pkg/logger"
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

	log.Info("Starting HandsOff Server (API + Worker)...")

	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}

	// Auto migrate database and create default admin user
	if err := initializer.Initialize(db, cfg, log); err != nil {
		log.Fatal("Failed to initialize database", "error", err)
	}

	// ==================== 启动 Worker ====================
	workerServer := task.NewServer(db, cfg, log)

	// Error channel for startup failures
	errChan := make(chan error, 2)

	go func() {
		log.Info("Starting Worker...", "concurrency", cfg.Worker.Concurrency)
		if err := workerServer.Start(); err != nil {
			errChan <- fmt.Errorf("worker start failed: %w", err)
		}
	}()

	// ==================== 启动 API ====================
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := router.Setup(db, cfg, log)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("API server listening", "port", cfg.Server.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server failed: %w", err)
		}
	}()

	// ==================== 等待错误或关闭信号 ====================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Fatal("Server startup failed", "error", err)
	case <-quit:
		log.Info("Shutting down server...")
	}

	// 关闭 HTTP 服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", "error", err)
	}

	// 关闭 Worker
	workerServer.Shutdown()

	log.Info("Server exited")
}
