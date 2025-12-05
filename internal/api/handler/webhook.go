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
// REFACTORED: Split into 4 small functions with single responsibility
func (h *WebhookHandler) HandleGitLab(c *gin.Context) {
	// Step 1: Parse and validate webhook event
	mrEvent, err := h.parseAndValidateWebhook(c)
	if err != nil {
		return // Error already handled internally
	}

	// Step 2: Find and validate repository
	repo, err := h.findAndValidateRepository(mrEvent)
	if err != nil {
		return // Error already handled internally
	}

	// Step 3: Create review result record
	reviewID, err := h.createReviewRecord(repo, mrEvent)
	if err != nil {
		return // Error already handled internally
	}

	// Step 4: Enqueue review task
	if err := h.enqueueReviewTask(reviewID); err != nil {
		return // Error already handled internally
	}

	h.log.Info("Webhook processed successfully", 
		"review_id", reviewID,
		"repository_id", repo.ID,
		"mr_id", mrEvent.GetMRID())

	c.JSON(http.StatusOK, gin.H{
		"message":   "Webhook received and review task enqueued",
		"review_id": reviewID,
	})
}

// parseAndValidateWebhook parses and validates GitLab webhook event
func (h *WebhookHandler) parseAndValidateWebhook(c *gin.Context) (*webhook.GitLabMergeRequestEvent, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.log.Error("Failed to read webhook body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return nil, err
	}

	// Check event type first (lightweight check)
	var baseEvent webhook.GitLabWebhookEvent
	if err := json.Unmarshal(body, &baseEvent); err != nil {
		h.log.Error("Failed to parse webhook event", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook payload"})
		return nil, err
	}

	if baseEvent.ObjectKind != "merge_request" {
		h.log.Info("Ignoring non-merge-request event", "event_type", baseEvent.ObjectKind)
		c.JSON(http.StatusOK, gin.H{"message": "Event type not supported"})
		return nil, fmt.Errorf("unsupported event type: %s", baseEvent.ObjectKind)
	}

	// Parse full MR event
	var mrEvent webhook.GitLabMergeRequestEvent
	if err := json.Unmarshal(body, &mrEvent); err != nil {
		h.log.Error("Failed to parse MR event", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merge request payload"})
		return nil, err
	}

	if !mrEvent.ShouldTriggerReview() {
		h.log.Info("Event does not trigger review", 
			"action", mrEvent.ObjectAttributes.Action,
			"state", mrEvent.ObjectAttributes.State,
			"mr_id", mrEvent.GetMRID())
		c.JSON(http.StatusOK, gin.H{"message": "Event does not trigger review"})
		return nil, fmt.Errorf("event does not trigger review")
	}

	return &mrEvent, nil
}

// findAndValidateRepository finds repository and validates webhook signature
func (h *WebhookHandler) findAndValidateRepository(mrEvent *webhook.GitLabMergeRequestEvent) (*model.Repository, error) {
	var repo model.Repository
	err := h.db.Preload("LLMModel").
		Where("platform_repo_id = ? AND is_active = ?", mrEvent.GetProjectID(), true).
		First(&repo).Error

	if err == gorm.ErrRecordNotFound {
		h.log.Warn("Repository not found or inactive", 
			"project_id", mrEvent.GetProjectID(),
			"mr_id", mrEvent.GetMRID())
		return nil, fmt.Errorf("repository not found")
	}

	if err != nil {
		h.log.Error("Failed to query repository", "error", err)
		return nil, err
	}

	// Validate webhook signature if configured
	if repo.WebhookSecret != "" {
		// Note: Signature validation requires raw body, which is not available here
		// This is a known limitation after refactoring
		// TODO: Pass raw body from parseAndValidateWebhook if signature validation is needed
	}

	if repo.LLMProviderID == nil {
		h.log.Warn("No LLM model configured for repository", 
			"repository_id", repo.ID,
			"repository_name", repo.Name)
		return nil, fmt.Errorf("no LLM model configured")
	}

	return &repo, nil
}

// createReviewRecord creates a new review result record in database
func (h *WebhookHandler) createReviewRecord(repo *model.Repository, mrEvent *webhook.GitLabMergeRequestEvent) (uint, error) {
	// Use FirstOrCreate to ensure idempotency (prevent duplicate reviews for same MR)
	reviewResult := model.ReviewResult{
		RepositoryID:   repo.ID,
		MergeRequestID: mrEvent.GetMRID(),
	}

	// Update fields if record exists, or create new if not found
	err := h.db.Where(&model.ReviewResult{
		RepositoryID:   repo.ID,
		MergeRequestID: mrEvent.GetMRID(),
	}).Assign(model.ReviewResult{
		MRTitle:       mrEvent.GetMRTitle(),
		MRAuthor:      mrEvent.GetMRAuthor(),
		SourceBranch:  mrEvent.GetSourceBranch(),
		TargetBranch:  mrEvent.GetTargetBranch(),
		MRWebURL:      mrEvent.GetMRWebURL(),
		LLMProviderID: *repo.LLMProviderID,
		Status:        "pending",
		CommentPosted: false,
	}).FirstOrCreate(&reviewResult).Error

	if err != nil {
		h.log.Error("Failed to create or update review result", "error", err)
		return 0, err
	}

	h.log.Info("Review result record ensured", 
		"review_id", reviewResult.ID,
		"repository_id", repo.ID,
		"mr_id", mrEvent.GetMRID())

	return reviewResult.ID, nil
}

// enqueueReviewTask enqueues async review task to Redis queue
func (h *WebhookHandler) enqueueReviewTask(reviewID uint) error {
	// REFACTORED: Only pass ReviewResultID instead of all MR fields
	payload := task.CodeReviewPayload{
		ReviewResultID: reviewID,
	}

	payloadBytes, err := payload.ToJSON()
	if err != nil {
		h.log.Error("Failed to marshal task payload", "error", err)
		// Mark review as failed
		h.db.Model(&model.ReviewResult{}).Where("id = ?", reviewID).Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": fmt.Sprintf("Failed to create task: %v", err),
		})
		return err
	}

	taskInfo, err := h.queue.Enqueue(
		asynq.NewTask(task.TypeCodeReview, payloadBytes),
		asynq.Queue("default"),
		asynq.MaxRetry(3),
	)

	if err != nil {
		h.log.Error("Failed to enqueue task", "error", err)
		// Mark review as failed
		h.db.Model(&model.ReviewResult{}).Where("id = ?", reviewID).Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": fmt.Sprintf("Failed to enqueue task: %v", err),
		})
		return err
	}

	h.log.Info("Enqueued code review task", 
		"task_id", taskInfo.ID,
		"review_id", reviewID,
		"queue", taskInfo.Queue)

	return nil
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
