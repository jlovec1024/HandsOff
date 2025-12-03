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
	// TODO: Replace with ProjectContext middleware
	projectID, err := getUserDefaultProjectID(c, h.db)
	if err != nil {
		h.log.Error("Failed to get default project", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No project found. Please create a project first."})
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

	if req.AccessToken == "" && req.AccessToken != "***masked***" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access token is required"})
		return
	}

	// TODO: Replace with ProjectContext middleware
	projectID, err := getUserDefaultProjectID(c, h.db)
	if err != nil {
		h.log.Error("Failed to get default project", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No project found. Please create a project first."})
		return
	}

	// Convert DTO to model
	config := &model.GitPlatformConfig{
		PlatformType: req.PlatformType,
		BaseURL:      req.BaseURL,
		AccessToken:  req.AccessToken,
		IsActive:     req.IsActive,
		ProjectID:    projectID,
	}

	if err := h.service.CreateOrUpdateConfig(config); err != nil {
		h.log.Error("Failed to update platform config", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
		return
	}

	h.log.Info("Platform configuration updated", "base_url", req.BaseURL)
	c.JSON(http.StatusOK, gin.H{"message": "Configuration updated successfully"})
}

// TestConnection tests the GitLab connection
func (h *PlatformHandler) TestConnection(c *gin.Context) {
	// TODO: Replace with ProjectContext middleware
	projectID, err := getUserDefaultProjectID(c, h.db)
	if err != nil {
		h.log.Error("Failed to get default project", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No project found. Please create a project first."})
		return
	}

	config, err := h.service.GetConfig(projectID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Please configure GitLab first"})
			return
		}
		h.log.Error("Failed to get platform config", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get configuration"})
		return
	}

	if err := h.service.TestConnection(projectID, config.ID); err != nil {
		h.log.Error("GitLab connection test failed", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get updated config with test results
	updatedConfig, _ := h.service.GetConfig(projectID)

	h.log.Info("GitLab connection test successful")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": updatedConfig.LastTestMessage,
		"tested_at": updatedConfig.LastTestedAt,
	})
}
