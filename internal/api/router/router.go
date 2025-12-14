package router

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/handsoff/handsoff/internal/api/handler"
	"github.com/handsoff/handsoff/internal/api/middleware"
	"github.com/handsoff/handsoff/internal/repository"
	"github.com/handsoff/handsoff/internal/service"
	"github.com/handsoff/handsoff/internal/web"
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

	systemConfigService := service.NewSystemConfigService(db)

	repositoryService, err := service.NewRepositoryService(repositoryRepo, platformRepo, systemConfigService, cfg)
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
	systemConfigHandler := handler.NewSystemConfigHandler(systemConfigService, log)
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

		// System Configuration routes
		protected.GET("/system/webhook", systemConfigHandler.GetWebhookConfig)
		protected.PUT("/system/webhook", systemConfigHandler.UpdateWebhookConfig)

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
		protected.POST("/repositories/:id/webhook/test", repositoryHandler.TestWebhook)
		protected.PUT("/repositories/:id/webhook", repositoryHandler.RecreateWebhook)
		protected.GET("/repositories/:id/statistics", reviewHandler.GetRepositoryStatistics)
		protected.GET("/repositories/:id/token-usage", reviewHandler.GetRepositoryTokenUsage)

		// Review routes
		protected.GET("/reviews", reviewHandler.ListReviews)
		protected.GET("/reviews/:id", reviewHandler.GetReview)
		protected.GET("/reviews/:id/statistics", reviewHandler.GetReviewStatistics)
		protected.GET("/reviews/:id/usage-logs", reviewHandler.GetReviewUsageLogs)

		// Dashboard routes
		protected.GET("/dashboard/statistics", reviewHandler.GetDashboardStatistics)
		protected.GET("/dashboard/recent", reviewHandler.GetRecentReviews)
		protected.GET("/dashboard/trends", reviewHandler.GetTrendData)
		protected.GET("/dashboard/token-usage", reviewHandler.GetDashboardTokenUsage)
	}

	// Webhook routes (public, but with signature verification)
	webhook := r.Group("/api/webhook")
	{
		webhook.POST("", webhookHandler.HandleWebhook)
	}

	// Serve static files from embedded filesystem
	staticFS, err := fs.Sub(web.StaticFiles, "dist")
	if err != nil {
		log.Fatal("Failed to create sub filesystem", "error", err)
	}

	// Serve static assets (CSS, JS, images)
	assetsFS, err := fs.Sub(staticFS, "assets")
	if err != nil {
		log.Fatal("Failed to create assets filesystem", "error", err)
	}
	r.StaticFS("/assets", http.FS(assetsFS))

	// SPA fallback - serve index.html for all non-API routes
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		// Serve index.html for frontend routes
		data, err := web.StaticFiles.ReadFile("dist/index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load frontend")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	return r
}
