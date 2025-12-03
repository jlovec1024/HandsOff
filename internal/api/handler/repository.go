package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/handsoff/handsoff/internal/service"
	"github.com/handsoff/handsoff/pkg/logger"
	"gorm.io/gorm"
)

// RepositoryHandler handles repository requests
type RepositoryHandler struct {
	service *service.RepositoryService
	db      *gorm.DB
	log     *logger.Logger
}

// NewRepositoryHandler creates a new repository handler
func NewRepositoryHandler(service *service.RepositoryService, db *gorm.DB, log *logger.Logger) *RepositoryHandler {
	return &RepositoryHandler{
		service: service,
		db:      db,
		log:     log,
	}
}

// ListFromGitLab returns repositories from GitLab
func (h *RepositoryHandler) ListFromGitLab(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	projectID, ok := getProjectID(c)
	if !ok {
		h.log.Error("Project ID missing from context - middleware failure")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	repos, totalPages, err := h.service.ListFromGitLab(projectID, page, perPage)
	if err != nil {
		h.log.Error("Failed to list GitLab repositories", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"repositories": repos,
		"page":         page,
		"per_page":     perPage,
		"total_pages":  totalPages,
	})
}

// List returns all imported repositories
func (h *RepositoryHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	projectID, ok := getProjectID(c)
	if !ok {
		h.log.Error("Project ID missing from context - middleware failure")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	repos, total, err := h.service.List(projectID, page, pageSize)
	if err != nil {
		h.log.Error("Failed to list repositories", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list repositories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"repositories": repos,
		"page":         page,
		"page_size":    pageSize,
		"total":        total,
	})
}

// Get returns a specific repository
func (h *RepositoryHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	projectID, ok := getProjectID(c)
	if !ok {
		h.log.Error("Project ID missing from context - middleware failure")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	repo, err := h.service.Get(uint(id), projectID)
	if err != nil {
		h.log.Error("Failed to get repository", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	c.JSON(http.StatusOK, repo)
}

// BatchImportRequest represents batch import request
type BatchImportRequest struct {
	RepositoryIDs      []int64 `json:"repository_ids" binding:"required"`
	WebhookCallbackURL string  `json:"webhook_callback_url" binding:"required"`
}

// BatchImport imports multiple repositories from GitLab
func (h *RepositoryHandler) BatchImport(c *gin.Context) {
	var req BatchImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if len(req.RepositoryIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No repositories selected"})
		return
	}

	projectID, ok := getProjectID(c)
	if !ok {
		h.log.Error("Project ID missing from context - middleware failure")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if err := h.service.BatchImport(projectID, req.RepositoryIDs, req.WebhookCallbackURL); err != nil {
		h.log.Error("Failed to import repositories", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.log.Info("Repositories imported", "count", len(req.RepositoryIDs))
	c.JSON(http.StatusOK, gin.H{
		"message": "Repositories imported successfully",
		"count":   len(req.RepositoryIDs),
	})
}

// UpdateLLMModelRequest represents update LLM model request
type UpdateLLMModelRequest struct {
	LLMModelID *uint `json:"llm_model_id"`
}

// UpdateLLMModel updates the LLM model for a repository
func (h *RepositoryHandler) UpdateLLMModel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	var req UpdateLLMModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.service.UpdateLLMModel(uint(id), req.LLMModelID); err != nil {
		h.log.Error("Failed to update LLM model", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update LLM model"})
		return
	}

	h.log.Info("Repository LLM model updated", "id", id)
	c.JSON(http.StatusOK, gin.H{"message": "LLM model updated successfully"})
}

// Delete deletes a repository
func (h *RepositoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	projectID, ok := getProjectID(c)
	if !ok {
		h.log.Error("Project ID missing from context - middleware failure")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if err := h.service.Delete(uint(id), projectID); err != nil {
		h.log.Error("Failed to delete repository", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete repository"})
		return
	}

	h.log.Info("Repository deleted", "id", id)
	c.JSON(http.StatusOK, gin.H{"message": "Repository deleted successfully"})
}
