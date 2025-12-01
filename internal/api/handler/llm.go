package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/handsoff/handsoff/internal/model"
	"github.com/handsoff/handsoff/internal/service"
	"github.com/handsoff/handsoff/pkg/logger"
)

// LLMHandler handles LLM provider and model requests
type LLMHandler struct {
	service *service.LLMService
	log     *logger.Logger
}

// NewLLMHandler creates a new LLM handler
func NewLLMHandler(service *service.LLMService, log *logger.Logger) *LLMHandler {
	return &LLMHandler{
		service: service,
		log:     log,
	}
}

// Provider handlers

// ListProviders returns all LLM providers
func (h *LLMHandler) ListProviders(c *gin.Context) {
	providers, err := h.service.ListProviders()
	if err != nil {
		h.log.Error("Failed to list providers", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list providers"})
		return
	}

	c.JSON(http.StatusOK, providers)
}

// GetProvider returns a specific provider
func (h *LLMHandler) GetProvider(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	provider, err := h.service.GetProvider(uint(id))
	if err != nil {
		h.log.Error("Failed to get provider", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	c.JSON(http.StatusOK, provider)
}

// CreateProvider creates a new LLM provider
func (h *LLMHandler) CreateProvider(c *gin.Context) {
	var req model.LLMProvider
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate required fields
	if req.Name == "" || req.Type == "" || req.BaseURL == "" || req.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name, type, base URL, and API key are required"})
		return
	}

	if err := h.service.CreateProvider(&req); err != nil {
		h.log.Error("Failed to create provider", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create provider"})
		return
	}

	h.log.Info("LLM provider created", "name", req.Name, "type", req.Type)
	c.JSON(http.StatusCreated, req)
}

// UpdateProvider updates an existing provider
func (h *LLMHandler) UpdateProvider(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	var req model.LLMProvider
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	req.ID = uint(id)
	if err := h.service.UpdateProvider(&req); err != nil {
		h.log.Error("Failed to update provider", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update provider"})
		return
	}

	h.log.Info("LLM provider updated", "id", id)
	c.JSON(http.StatusOK, req)
}

// DeleteProvider deletes a provider
func (h *LLMHandler) DeleteProvider(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	if err := h.service.DeleteProvider(uint(id)); err != nil {
		h.log.Error("Failed to delete provider", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete provider"})
		return
	}

	h.log.Info("LLM provider deleted", "id", id)
	c.JSON(http.StatusOK, gin.H{"message": "Provider deleted successfully"})
}

// TestProviderConnection tests the LLM provider connection
func (h *LLMHandler) TestProviderConnection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	if err := h.service.TestProviderConnection(uint(id)); err != nil {
		h.log.Error("Provider connection test failed", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Info("Provider connection test successful", "id", id)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Connection test successful"})
}

// Model handlers

// ListModels returns all LLM models
func (h *LLMHandler) ListModels(c *gin.Context) {
	var providerID *uint
	if pidStr := c.Query("provider_id"); pidStr != "" {
		pid, err := strconv.ParseUint(pidStr, 10, 32)
		if err == nil {
			pidUint := uint(pid)
			providerID = &pidUint
		}
	}

	models, err := h.service.ListModels(providerID)
	if err != nil {
		h.log.Error("Failed to list models", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list models"})
		return
	}

	c.JSON(http.StatusOK, models)
}

// CreateModel creates a new LLM model
func (h *LLMHandler) CreateModel(c *gin.Context) {
	var req model.LLMModel
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate required fields
	if req.ProviderID == 0 || req.ModelName == "" || req.DisplayName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider ID, model name, and display name are required"})
		return
	}

	if err := h.service.CreateModel(&req); err != nil {
		h.log.Error("Failed to create model", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create model"})
		return
	}

	h.log.Info("LLM model created", "name", req.ModelName)
	c.JSON(http.StatusCreated, req)
}

// UpdateModel updates an existing model
func (h *LLMHandler) UpdateModel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid model ID"})
		return
	}

	var req model.LLMModel
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	req.ID = uint(id)
	if err := h.service.UpdateModel(&req); err != nil {
		h.log.Error("Failed to update model", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update model"})
		return
	}

	h.log.Info("LLM model updated", "id", id)
	c.JSON(http.StatusOK, req)
}

// DeleteModel deletes a model
func (h *LLMHandler) DeleteModel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid model ID"})
		return
	}

	if err := h.service.DeleteModel(uint(id)); err != nil {
		h.log.Error("Failed to delete model", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete model"})
		return
	}

	h.log.Info("LLM model deleted", "id", id)
	c.JSON(http.StatusOK, gin.H{"message": "Model deleted successfully"})
}
