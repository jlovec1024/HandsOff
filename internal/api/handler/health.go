package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/handsoff/handsoff/pkg/logger"
	"gorm.io/gorm"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db  *gorm.DB
	log *logger.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *gorm.DB, log *logger.Logger) *HealthHandler {
	return &HealthHandler{
		db:  db,
		log: log,
	}
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status   string    `json:"status"`
	Time     time.Time `json:"time"`
	Database string    `json:"database"`
	Version  string    `json:"version"`
}

// Check performs health check
func (h *HealthHandler) Check(c *gin.Context) {
	response := HealthResponse{
		Status:  "ok",
		Time:    time.Now().UTC(),
		Version: "1.0.0-mvp",
	}

	// Check database connection
	sqlDB, err := h.db.DB()
	if err != nil {
		response.Status = "error"
		response.Database = "disconnected"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	if err := sqlDB.Ping(); err != nil {
		response.Status = "error"
		response.Database = "unreachable"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	response.Database = "connected"
	c.JSON(http.StatusOK, response)
}
