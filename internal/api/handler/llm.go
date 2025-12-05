package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/handsoff/handsoff/internal/model"
	"github.com/handsoff/handsoff/internal/service"
	"github.com/handsoff/handsoff/pkg/logger"
	"gorm.io/gorm"
)

// LLMHandler handles LLM provider and model requests
type LLMHandler struct {
	service *service.LLMService
	db      *gorm.DB
	log     *logger.Logger
}

// NewLLMHandler creates a new LLM handler
func NewLLMHandler(service *service.LLMService, db *gorm.DB, log *logger.Logger) *LLMHandler {
	return &LLMHandler{
		service: service,
		db:      db,
		log:     log,
	}
}

// Provider handlers

// ListProviders returns all LLM providers
func (h *LLMHandler) ListProviders(c *gin.Context) {
	projectID, ok := getProjectID(c)
	if !ok {
		h.log.Error("Project ID missing from context - middleware failure")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	providers, err := h.service.ListProviders(projectID)
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

	projectID, ok := getProjectID(c)
	if !ok {
		h.log.Error("Project ID missing from context - middleware failure")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	provider, err := h.service.GetProvider(uint(id), projectID)
	if err != nil {
		h.log.Error("Failed to get provider", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	c.JSON(http.StatusOK, provider)
}

// CreateProvider creates a new LLM provider
func (h *LLMHandler) CreateProvider(c *gin.Context) {
	projectID, ok := getProjectID(c)
	if !ok {
		h.log.Error("Project ID missing from context - middleware failure")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Use inline DTO to receive api_key (model has json:"-" for security)
	var req struct {
		Name     string `json:"name" binding:"required"`
		BaseURL  string `json:"base_url" binding:"required"`
		APIKey   string `json:"api_key" binding:"required"`
		Model    string `json:"model" binding:"required"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Map DTO to model
	provider := model.LLMProvider{
		Name:      req.Name,
		BaseURL:   req.BaseURL,
		APIKey:    req.APIKey,
		Model:     req.Model,
		IsActive:  req.IsActive,
		ProjectID: projectID,
	}

	if err := h.service.CreateProvider(&provider); err != nil {
		h.log.Error("Failed to create provider", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create provider"})
		return
	}

	h.log.Info("LLM provider created", "name", provider.Name, "model", provider.Model)
	c.JSON(http.StatusCreated, provider)
}

// UpdateProvider updates an existing provider
func (h *LLMHandler) UpdateProvider(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	projectID, ok := getProjectID(c)
	if !ok {
		h.log.Error("Project ID missing from context - middleware failure")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Use inline DTO to handle optional api_key
	var req struct {
		Name     string `json:"name"`
		BaseURL  string `json:"base_url"`
		APIKey   string `json:"api_key"` // Optional: empty means keep existing
		Model    string `json:"model"`
		IsActive *bool  `json:"is_active"` // Pointer to distinguish between false and not provided
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Build update data - only include non-empty fields
	provider := &model.LLMProvider{
		ID:        uint(id),
		ProjectID: projectID,
	}

	// Update fields if provided
	if req.Name != "" {
		provider.Name = req.Name
	}
	if req.BaseURL != "" {
		provider.BaseURL = req.BaseURL
	}
	if req.APIKey != "" {
		provider.APIKey = req.APIKey
	}
	if req.Model != "" {
		provider.Model = req.Model
	}
	if req.IsActive != nil {
		provider.IsActive = *req.IsActive
	}

	if err := h.service.UpdateProvider(provider); err != nil {
		h.log.Error("Failed to update provider", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update provider"})
		return
	}

	h.log.Info("LLM provider updated", "id", id)
	
	// Fetch and return updated provider (with masked API key)
	updated, err := h.service.GetProvider(uint(id), projectID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Provider updated successfully"})
		return
	}
	c.JSON(http.StatusOK, updated)
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

	projectID, ok := getProjectID(c)
	if !ok {
		h.log.Error("Project ID missing from context - middleware failure")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if err := h.service.TestProviderConnection(uint(id), projectID); err != nil {
		h.log.Error("Provider connection test failed", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Info("Provider connection test successful", "id", id)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Connection test successful"})
}

// FetchAvailableModels fetches available models from a provider
func (h *LLMHandler) FetchAvailableModels(c *gin.Context) {
	var req struct {
		BaseURL string `json:"base_url" binding:"required"`
		APIKey  string `json:"api_key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Base URL and API key are required"})
		return
	}

	models, err := h.service.FetchAvailableModels(req.BaseURL, req.APIKey)
	if err != nil {
		h.log.Error("Failed to fetch models", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"models": models})
}

// FetchProviderModels fetches available models using stored provider configuration
func (h *LLMHandler) FetchProviderModels(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	projectID, exists := c.Get("project_id")
	if !exists {
		h.log.Error("Project ID not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Project context not found"})
		return
	}

	models, err := h.service.FetchModelsForProvider(uint(id), projectID.(uint))
	if err != nil {
		h.log.Error("Failed to fetch provider models", "id", id, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"models": models})
}

// TestTemporaryModel tests a specific model with temporary credentials (for create mode)
func (h *LLMHandler) TestTemporaryModel(c *gin.Context) {
	var req struct {
		BaseURL string `json:"base_url" binding:"required"`
		APIKey  string `json:"api_key" binding:"required"`
		Model   string `json:"model" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Base URL, API key, and model are required"})
		return
	}

	if err := h.service.TestModelConnection(req.BaseURL, req.APIKey, req.Model); err != nil {
		h.log.Error("Model test failed", "model", req.Model, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Info("Model test successful", "model", req.Model)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("Model %s test successful", req.Model)})
}

// Model handlers




