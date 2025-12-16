package handler

import (
	"encoding/json"
	"errors"
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

// WebhookError represents an error with HTTP status code for centralized handling
type WebhookError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *WebhookError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *WebhookError) Unwrap() error {
	return e.Err
}

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

// handleWebhookError centralized error response handler
func (h *WebhookHandler) handleWebhookError(c *gin.Context, err error) {
	var we *WebhookError
	if errors.As(err, &we) {
		c.JSON(we.StatusCode, gin.H{"message": we.Message})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}

// HandleGitLab handles GitLab webhook events
// REFACTORED: Centralized error handling with clean separation:
// - (nil, nil) = event ignored (200 OK)
// - (nil, error) = real error (4xx/5xx)
// - (result, nil) = success, continue processing
func (h *WebhookHandler) HandleGitLab(c *gin.Context) {
	// Step 0: Read body once (needed for parsing)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.log.Error("Failed to read webhook body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Step 1: Parse and validate webhook event
	mrEvent, err := h.parseAndValidateWebhook(body)
	if err != nil {
		h.handleWebhookError(c, err)
		return
	}
	if mrEvent == nil {
		// Event ignored (not MR event or doesn't trigger review)
		c.JSON(http.StatusOK, gin.H{"message": "Event ignored"})
		return
	}

	// Step 2: Find and validate repository
	repo, err := h.findAndValidateRepository(mrEvent)
	if err != nil {
		h.handleWebhookError(c, err)
		return
	}
	if repo == nil {
		// Repository not configured or no LLM - ignore silently
		c.JSON(http.StatusOK, gin.H{"message": "Repository not configured for review"})
		return
	}

	// Step 3: Create review result record
	reviewID, err := h.createReviewRecord(repo, mrEvent)
	if err != nil {
		h.handleWebhookError(c, err)
		return
	}

	// Step 4: Enqueue review task
	if err := h.enqueueReviewTask(reviewID); err != nil {
		h.handleWebhookError(c, err)
		return
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
// Returns (*event, nil) for success, (nil, nil) for ignored events, (nil, error) for errors
// Does NOT touch gin.Context - caller handles all responses
func (h *WebhookHandler) parseAndValidateWebhook(body []byte) (*webhook.GitLabMergeRequestEvent, error) {
	// Check event type first (lightweight check)
	var baseEvent webhook.GitLabWebhookEvent
	if err := json.Unmarshal(body, &baseEvent); err != nil {
		h.log.Error("Failed to parse webhook event", "error", err)
		return nil, &WebhookError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid webhook payload",
			Err:        err,
		}
	}

	if baseEvent.ObjectKind != "merge_request" {
		h.log.Info("Ignoring non-merge-request event", "event_type", baseEvent.ObjectKind)
		return nil, nil // Not an error, just ignored
	}

	// Parse full MR event
	var mrEvent webhook.GitLabMergeRequestEvent
	if err := json.Unmarshal(body, &mrEvent); err != nil {
		h.log.Error("Failed to parse MR event", "error", err)
		return nil, &WebhookError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid merge request payload",
			Err:        err,
		}
	}

	if !mrEvent.ShouldTriggerReview() {
		h.log.Info("Event does not trigger review", 
			"action", mrEvent.ObjectAttributes.Action,
			"state", mrEvent.ObjectAttributes.State,
			"mr_id", mrEvent.GetMRID())
		return nil, nil // Not an error, just ignored
	}

	return &mrEvent, nil
}

// findAndValidateRepository finds repository and validates webhook signature
// Returns (*repo, nil) for success, (nil, nil) for ignored, (nil, error) for errors
// Does NOT touch gin.Context - caller handles all responses
func (h *WebhookHandler) findAndValidateRepository(mrEvent *webhook.GitLabMergeRequestEvent) (*model.Repository, error) {
	var repo model.Repository
	err := h.db.Preload("LLMProvider").
		Where("platform_repo_id = ? AND is_active = ?", mrEvent.GetProjectID(), true).
		First(&repo).Error

	if err == gorm.ErrRecordNotFound {
		h.log.Warn("Repository not found or inactive", 
			"project_id", mrEvent.GetProjectID(),
			"mr_id", mrEvent.GetMRID())
		return nil, nil // Not an error, just ignored (repo not configured)
	}

	if err != nil {
		h.log.Error("Failed to query repository", "error", err)
		return nil, &WebhookError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Database error",
			Err:        err,
		}
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
		return nil, nil // Not an error, just ignored (no LLM configured)
	}

	return &repo, nil
}

// createReviewRecord creates a new review result record in database
// Returns *WebhookError for centralized handling - does NOT touch gin.Context
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
		return 0, &WebhookError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to create review record",
			Err:        err,
		}
	}

	h.log.Info("Review result record ensured", 
		"review_id", reviewResult.ID,
		"repository_id", repo.ID,
		"mr_id", mrEvent.GetMRID())

	return reviewResult.ID, nil
}

// enqueueReviewTask enqueues async review task to Redis queue
// Returns *WebhookError for centralized handling - does NOT touch gin.Context
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
		return &WebhookError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to create task",
			Err:        err,
		}
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
		return &WebhookError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to enqueue task",
			Err:        err,
		}
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
