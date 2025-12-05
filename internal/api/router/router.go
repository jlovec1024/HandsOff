package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/handsoff/handsoff/internal/api/handler"
	"github.com/handsoff/handsoff/internal/api/middleware"
	"github.com/handsoff/handsoff/internal/repository"
	"github.com/handsoff/handsoff/internal/service"
	"github.com/handsoff/handsoff/pkg/config"
	"github.com/handsoff/handsoff/pkg/logger"
	"github.com/handsoff/handsoff/pkg/queue"
	"gorm.io/gorm"
)

// Setup configures all routes
func Setup(db *gorm.DB, cfg *config.Config, log *logger.Logger) *gin.Engine {
	r := gin.New()

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(log))

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Initialize repositories
	platformRepo := repository.NewPlatformRepository(db)
	llmRepo := repository.NewLLMRepository(db)
	repositoryRepo := repository.NewRepositoryRepo(db)

	// Initialize services
	platformService, err := service.NewPlatformService(platformRepo, cfg)
	if err != nil {
		log.Fatal("Failed to create platform service", "error", err)
	}
	llmService, err := service.NewLLMService(llmRepo, cfg)
	if err != nil {
		log.Fatal("Failed to create LLM service", "error", err)
	}

	repositoryService, err := service.NewRepositoryService(repositoryRepo, platformRepo, cfg)
	if err != nil {
		log.Fatal("Failed to create repository service", "error", err)
	}

	// Initialize queue client
	queueClient := queue.NewClient(cfg.Redis)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(db, cfg, log)
	healthHandler := handler.NewHealthHandler(db, log)
	platformHandler := handler.NewPlatformHandler(platformService, db, log)
	llmHandler := handler.NewLLMHandler(llmService, db, log)
	repositoryHandler := handler.NewRepositoryHandler(repositoryService, db, log)
	webhookHandler := handler.NewWebhookHandler(db, log, queueClient)
	reviewHandler := handler.NewReviewHandler(db, log)

	// Public routes
	public := r.Group("/api")
	{
		public.POST("/auth/login", authHandler.Login)
		public.GET("/health", healthHandler.Check)
	}
	// Protected routes (require authentication)
	protected := r.Group("/api")
	protected.Use(middleware.Auth(cfg))
	protected.Use(middleware.ProjectContext(db)) // Add project context
	{
		// Auth routes
		protected.POST("/auth/logout", authHandler.Logout)
		protected.GET("/auth/user", authHandler.GetCurrentUser)

		// Platform routes
		protected.GET("/platform/config", platformHandler.GetConfig)
		protected.PUT("/platform/config", platformHandler.UpdateConfig)
		protected.POST("/platform/test", platformHandler.TestConnection)

	// LLM Provider routes
	protected.GET("/llm/providers", llmHandler.ListProviders)
	protected.GET("/llm/providers/:id", llmHandler.GetProvider)
	protected.POST("/llm/providers", llmHandler.CreateProvider)
	protected.PUT("/llm/providers/:id", llmHandler.UpdateProvider)
	protected.DELETE("/llm/providers/:id", llmHandler.DeleteProvider)
	protected.POST("/llm/providers/:id/test", llmHandler.TestProviderConnection)
	protected.POST("/llm/providers/models", llmHandler.FetchAvailableModels)
	protected.POST("/llm/providers/test-model", llmHandler.TestTemporaryModel) // Test temporary model config
	protected.GET("/llm/providers/:id/models", llmHandler.FetchProviderModels)

		// LLM Model routes (removed - simplified to single provider layer)

		// Repository routes
		protected.GET("/repositories/gitlab", repositoryHandler.ListFromGitLab)
		protected.GET("/repositories", repositoryHandler.List)
		protected.GET("/repositories/:id", repositoryHandler.Get)
		protected.POST("/repositories/batch", repositoryHandler.BatchImport)
		protected.PUT("/repositories/:id/llm", repositoryHandler.UpdateLLMModel)
		protected.DELETE("/repositories/:id", repositoryHandler.Delete)
		protected.GET("/repositories/:id/statistics", reviewHandler.GetRepositoryStatistics)

		// Review routes
		protected.GET("/reviews", reviewHandler.ListReviews)
		protected.GET("/reviews/:id", reviewHandler.GetReview)
		protected.GET("/reviews/:id/statistics", reviewHandler.GetReviewStatistics)

		// Dashboard routes
		protected.GET("/dashboard/statistics", reviewHandler.GetDashboardStatistics)
		protected.GET("/dashboard/recent", reviewHandler.GetRecentReviews)
		protected.GET("/dashboard/trends", reviewHandler.GetTrendData)
	}

	// Webhook routes (public, but with signature verification)
	webhook := r.Group("/api/webhook")
	{
		webhook.POST("", webhookHandler.HandleWebhook)
	}

	return r
}
