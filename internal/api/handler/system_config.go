package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/handsoff/handsoff/internal/service"
	"github.com/handsoff/handsoff/pkg/logger"
)

// SystemConfigHandler handles system configuration requests
type SystemConfigHandler struct {
	service *service.SystemConfigService
	log     *logger.Logger
}

// NewSystemConfigHandler creates a new system config handler
func NewSystemConfigHandler(service *service.SystemConfigService, log *logger.Logger) *SystemConfigHandler {
	return &SystemConfigHandler{
		service: service,
		log:     log,
	}
}

// GetWebhookConfig returns webhook configuration
func (h *SystemConfigHandler) GetWebhookConfig(c *gin.Context) {
	projectID, ok := getProjectID(c)
	if !ok {
		h.log.Error("Project ID missing from context - middleware failure")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	config, err := h.service.GetWebhookConfig(projectID)
	if err != nil {
		h.log.Error("Failed to get webhook config", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"webhook_callback_url": "",
			"webhook_secret":       "",
		})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateWebhookConfigRequest represents update webhook config request
type UpdateWebhookConfigRequest struct {
	WebhookCallbackURL string `json:"webhook_callback_url" binding:"required"`
	WebhookSecret      string `json:"webhook_secret"`
}

// UpdateWebhookConfig updates webhook configuration
func (h *SystemConfigHandler) UpdateWebhookConfig(c *gin.Context) {
	projectID, ok := getProjectID(c)
	if !ok {
		h.log.Error("Project ID missing from context - middleware failure")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var req UpdateWebhookConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.service.UpdateWebhookConfig(projectID, req.WebhookCallbackURL, req.WebhookSecret); err != nil {
		h.log.Error("Failed to update webhook config", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update webhook configuration"})
		return
	}

	h.log.Info("Webhook configuration updated", "project_id", projectID)
	c.JSON(http.StatusOK, gin.H{"message": "Webhook configuration updated successfully"})
}
