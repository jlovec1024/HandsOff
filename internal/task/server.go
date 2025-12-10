package task

import (
	"github.com/handsoff/handsoff/pkg/config"
	"github.com/handsoff/handsoff/pkg/logger"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// Server wraps asynq.Server for task processing
type Server struct {
	server *asynq.Server
	mux    *asynq.ServeMux
	db     *gorm.DB
	cfg    *config.Config
	log    *logger.Logger
}

// NewServer creates a new task server
func NewServer(db *gorm.DB, cfg *config.Config, log *logger.Logger) *Server {
	opt, err := asynq.ParseRedisURI(cfg.Redis.URL)
	if err != nil {
		log.Fatal("Invalid Redis URL", "error", err)
	}
	
	srv := asynq.NewServer(
		opt,
		asynq.Config{
			Concurrency: cfg.Worker.Concurrency,
			// Queues defines queue priority (higher value = higher priority)
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			// StrictPriority: true, // Process higher priority queues first
		},
	)

	mux := asynq.NewServeMux()

	// Apply middleware
	mux.Use(RecoveryMiddleware(log))
	mux.Use(LoggingMiddleware(log))

	// Initialize task handlers
	reviewHandler := NewReviewHandler(db, log, cfg.Security.EncryptionKey)

	// Register task handlers
	mux.HandleFunc(TypeCodeReview, reviewHandler.HandleCodeReview)
	// Future: mux.HandleFunc(TypeAutoFix, autoFixHandler.HandleAutoFix)

	log.Info("Registered task handlers",
		"handlers", []string{TypeCodeReview})

	return &Server{
		server: srv,
		mux:    mux,
		db:     db,
		cfg:    cfg,
		log:    log,
	}
}

// Start starts the worker server
func (s *Server) Start() error {
	return s.server.Start(s.mux)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() {
	s.server.Shutdown()
}
