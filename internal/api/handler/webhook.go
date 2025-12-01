package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/handsoff/handsoff/internal/model"
	"github.com/handsoff/handsoff/internal/task"
	"github.com/handsoff/handsoff/internal/webhook"
	"github.com/handsoff/handsoff/pkg/logger"
	"github.com/handsoff/handsoff/pkg/queue"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// WebhookHandler handles webhook requests
type WebhookHandler struct {
	db        *gorm.DB
	log       *logger.Logger
	validator *webhook.Validator
	queue     *queue.Client
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(db *gorm.DB, log *logger.Logger, queueClient *queue.Client) *WebhookHandler {
	return &WebhookHandler{
		db:        db,
		log:       log,
		validator: webhook.NewValidator(),
		queue:     queueClient,
	}
}

// HandleGitLab handles GitLab webhook events
func (h *WebhookHandler) HandleGitLab(c *gin.Context) {
	// Read raw body for signature validation
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.log.Error("Failed to read webhook body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Get webhook token from header
	receivedToken := c.GetHeader("X-Gitlab-Token")
	
	// Parse event to get project ID
	var baseEvent webhook.GitLabWebhookEvent
	if err := json.Unmarshal(body, &baseEvent); err != nil {
		h.log.Error("Failed to parse webhook event", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook payload"})
		return
	}

	// Only handle merge_request events
	if baseEvent.ObjectKind != "merge_request" {
		h.log.Info("Ignoring non-merge-request event", "event_type", baseEvent.ObjectKind)
		c.JSON(http.StatusOK, gin.H{"message": "Event type not supported"})
		return
	}

	// Parse full MR event
	var mrEvent webhook.GitLabMergeRequestEvent
	if err := json.Unmarshal(body, &mrEvent); err != nil {
		h.log.Error("Failed to parse MR event", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merge request payload"})
		return
	}

	// Find repository by platform_repo_id (GitLab project ID)
	var repo model.Repository
	if err := h.db.Preload("LLMModel").
		Where("platform_repo_id = ? AND is_active = ?", mrEvent.GetProjectID(), true).
		First(&repo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			h.log.Warn("Repository not found or inactive", 
				"project_id", mrEvent.GetProjectID(),
				"mr_id", mrEvent.GetMRID())
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found or inactive"})
			return
		}
		h.log.Error("Failed to query repository", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Validate webhook secret (if configured)
	if repo.WebhookSecret != "" {
		if err := h.validator.ValidateGitLabSignature(body, receivedToken, repo.WebhookSecret); err != nil {
			h.log.Warn("Webhook signature validation failed", 
				"error", err,
				"repository_id", repo.ID,
				"mr_id", mrEvent.GetMRID())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid webhook signature"})
			return
		}
	}

	// Check if this event should trigger a review
	if !mrEvent.ShouldTriggerReview() {
		h.log.Info("Event does not trigger review", 
			"action", mrEvent.ObjectAttributes.Action,
			"state", mrEvent.ObjectAttributes.State,
			"mr_id", mrEvent.GetMRID())
		c.JSON(http.StatusOK, gin.H{"message": "Event does not trigger review"})
		return
	}

	// Check if LLM model is configured
	if repo.LLMModelID == nil {
		h.log.Warn("No LLM model configured for repository", 
			"repository_id", repo.ID,
			"repository_name", repo.Name)
		c.JSON(http.StatusOK, gin.H{"message": "No LLM model configured"})
		return
	}

	// Create review result record with pending status
	reviewResult := model.ReviewResult{
		RepositoryID:   repo.ID,
		MergeRequestID: mrEvent.GetMRID(),
		MRTitle:        mrEvent.GetMRTitle(),
		MRAuthor:       mrEvent.GetMRAuthor(),
		SourceBranch:   mrEvent.GetSourceBranch(),
		TargetBranch:   mrEvent.GetTargetBranch(),
		MRWebURL:       mrEvent.GetMRWebURL(),
		LLMModelID:     *repo.LLMModelID,
		Status:         "pending",
		CommentPosted:  false,
	}

	if err := h.db.Create(&reviewResult).Error; err != nil {
		h.log.Error("Failed to create review result", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review record"})
		return
	}

	h.log.Info("Created review result record", 
		"review_id", reviewResult.ID,
		"repository_id", repo.ID,
		"mr_id", mrEvent.GetMRID())

	// Create async task for code review
	payload := task.CodeReviewPayload{
		RepositoryID:   repo.ID,
		MergeRequestID: mrEvent.GetMRID(),
		MRTitle:        mrEvent.GetMRTitle(),
		MRAuthor:       mrEvent.GetMRAuthor(),
		SourceBranch:   mrEvent.GetSourceBranch(),
		TargetBranch:   mrEvent.GetTargetBranch(),
		MRWebURL:       mrEvent.GetMRWebURL(),
		ProjectID:      mrEvent.GetProjectID(),
	}

	payloadBytes, err := payload.ToJSON()
	if err != nil {
		h.log.Error("Failed to marshal task payload", "error", err)
		// Update review result status to failed
		h.db.Model(&reviewResult).Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": fmt.Sprintf("Failed to create task: %v", err),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	// Enqueue task
	taskInfo, err := h.queue.Enqueue(
		asynq.NewTask(task.TypeCodeReview, payloadBytes),
		asynq.Queue("default"),
		asynq.MaxRetry(3),
	)
	if err != nil {
		h.log.Error("Failed to enqueue task", "error", err)
		// Update review result status to failed
		h.db.Model(&reviewResult).Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": fmt.Sprintf("Failed to enqueue task: %v", err),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue review task"})
		return
	}

	h.log.Info("Enqueued code review task", 
		"task_id", taskInfo.ID,
		"review_id", reviewResult.ID,
		"repository_id", repo.ID,
		"mr_id", mrEvent.GetMRID(),
		"queue", taskInfo.Queue)

	c.JSON(http.StatusOK, gin.H{
		"message":   "Webhook received and review task enqueued",
		"review_id": reviewResult.ID,
		"task_id":   taskInfo.ID,
	})
}

// HandleWebhook is a generic webhook handler that routes to platform-specific handlers
func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	// Detect platform from headers
	// GitLab sends X-Gitlab-Event header
	// GitHub sends X-GitHub-Event header
	
	if c.GetHeader("X-Gitlab-Event") != "" || c.GetHeader("X-Gitlab-Token") != "" {
		h.HandleGitLab(c)
		return
	}

	if c.GetHeader("X-GitHub-Event") != "" {
		// GitHub support (future implementation)
		c.JSON(http.StatusNotImplemented, gin.H{"error": "GitHub webhooks not yet implemented"})
		return
	}

	// Unknown webhook source
	h.log.Warn("Unknown webhook source", "headers", c.Request.Header)
	c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown webhook source"})
}
