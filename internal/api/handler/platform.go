package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/handsoff/handsoff/internal/api/dto"
	"github.com/handsoff/handsoff/internal/model"
	"github.com/handsoff/handsoff/internal/service"
	"github.com/handsoff/handsoff/pkg/logger"
	"gorm.io/gorm"
)

// PlatformHandler handles Git platform configuration requests
type PlatformHandler struct {
	service *service.PlatformService
	db      *gorm.DB
	log     *logger.Logger
}

// NewPlatformHandler creates a new platform handler
func NewPlatformHandler(service *service.PlatformService, db *gorm.DB, log *logger.Logger) *PlatformHandler {
	return &PlatformHandler{
		service: service,
		db:      db,
		log:     log,
	}
}

// GetConfig returns the current platform configuration
func (h *PlatformHandler) GetConfig(c *gin.Context) {
	projectID, ok := getProjectID(c)
	if !ok {
		h.log.Error("Project ID missing from context - middleware failure")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	config, err := h.service.GetConfig(projectID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"exists": false,
				"message": "No platform configured yet",
			})
			return
		}
		h.log.Error("Failed to get platform config", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateConfig creates or updates the platform configuration
func (h *PlatformHandler) UpdateConfig(c *gin.Context) {
	var req dto.UpdatePlatformConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate required fields
	if req.BaseURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Base URL is required"})
		return
	}

	projectID, ok := getProjectID(c)
	if !ok {
		h.log.Error("Project ID missing from context - middleware failure")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Convert DTO to model
	config := &model.GitPlatformConfig{
		PlatformType: req.PlatformType,
		BaseURL:      req.BaseURL,
		AccessToken:  "", // Will be set by service layer based on req.AccessToken
		IsActive:     req.IsActive,
		ProjectID:    projectID,
	}

	// Handle access token: nil = keep existing, non-nil = update
	if req.AccessToken != nil {
		config.AccessToken = *req.AccessToken
	}

	if err := h.service.CreateOrUpdateConfig(config); err != nil {
		h.log.Error("Failed to update platform config", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
		return
	}

	// Get updated config to return to frontend (avoids extra API call)
	updatedConfig, err := h.service.GetConfig(projectID)
	if err != nil {
		// If fetching updated config fails, fall back to simple message
		h.log.Warn("Failed to fetch updated config", "error", err)
		h.log.Info("Platform configuration updated", "base_url", req.BaseURL)
		c.JSON(http.StatusOK, gin.H{"message": "Configuration updated successfully"})
		return
	}

	h.log.Info("Platform configuration updated", "base_url", req.BaseURL)
	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration updated successfully",
		"config":  updatedConfig,
	})
}

// TestConnection tests the GitLab connection
// Accepts test configuration in request body (allows testing before saving)
func (h *PlatformHandler) TestConnection(c *gin.Context) {
	var req dto.TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate required fields
	if req.BaseURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Base URL is required"})
		return
	}

	if req.AccessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access token is required"})
		return
	}

	// Test connection with provided configuration (no saving)
	message, err := h.service.TestConnectionWithConfig(req.BaseURL, req.AccessToken)
	if err != nil {
		h.log.Error("GitLab connection test failed", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Info("GitLab connection test successful")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}
