package task

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/handsoff/handsoff/internal/gitlab"
	"github.com/handsoff/handsoff/internal/llm"
	"github.com/handsoff/handsoff/internal/model"
	"github.com/handsoff/handsoff/internal/service"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// ReviewHandler handles code review tasks
type ReviewHandler struct {
	db              *gorm.DB
	log             Logger
	encryptionKey   string
	systemConfigSvc *service.SystemConfigService
}

// Logger interface for handler logging
type Logger interface {
	Info(...interface{})
	Error(...interface{})
}

// NewReviewHandler creates a new review handler
func NewReviewHandler(db *gorm.DB, log Logger, encryptionKey string) *ReviewHandler {
	return &ReviewHandler{
		db:              db,
		log:             log,
		encryptionKey:   encryptionKey,
		systemConfigSvc: service.NewSystemConfigService(db),
	}
}

// HandleCodeReview processes code review tasks
// REFACTORED: Split into 6 small functions with single responsibility
// FIXED: GitLab comment failure now triggers retry (was swallowed before)
func (h *ReviewHandler) HandleCodeReview(ctx context.Context, t *asynq.Task) error {
	// Step 1: Parse payload and load review result from DB
	reviewResult, err := h.loadReviewContext(t)
	if err != nil {
		return err // Already logged internally
	}

	// Step 2: Fetch MR diff from GitLab
	diff, gitlabClient, err := h.fetchMRDiff(reviewResult)
	if err != nil {
		h.markReviewFailed(reviewResult.ID, fmt.Sprintf("Failed to get MR diff: %v", err))
		return err
	}

	// Step 3: Perform LLM code review
	reviewResp, err := h.callLLMReview(reviewResult, diff)
	if err != nil {
		// Log failed usage even on error (tokens may have been consumed)
		h.logUsage(reviewResult, nil, err)
		h.markReviewFailed(reviewResult.ID, fmt.Sprintf("LLM review failed: %v", err))
		return err
	}

	// Step 3.5: Log successful LLM usage for operations analytics
	h.logUsage(reviewResult, reviewResp, nil)

	// Step 4: Save review results to database
	if err := h.saveReviewResults(reviewResult, reviewResp); err != nil {
		return err
	}

	// Step 5: Post comment to GitLab MR
	// FIXED: Now returns error to trigger Asynq retry if comment fails
	if err := h.postCommentToGitLab(reviewResult, gitlabClient, reviewResp); err != nil {
		h.log.Error("Failed to post comment, will retry", "error", err, "review_id", reviewResult.ID)
		return fmt.Errorf("failed to post comment: %w", err)
	}

	// Step 6: Update webhook event status to completed (if associated with webhook)
	if reviewResult.WebhookEventID != nil {
		h.updateWebhookEventStatus(reviewResult, model.EventStatusCompleted)
	}

	h.log.Info("Code review completed successfully",
		"review_id", reviewResult.ID,
		"score", reviewResp.Score,
		"suggestions", len(reviewResp.Suggestions))

	return nil
}

// loadReviewContext loads review result with all relationships from database
func (h *ReviewHandler) loadReviewContext(t *asynq.Task) (*model.ReviewResult, error) {
	var payload CodeReviewPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		h.log.Error("Failed to unmarshal task payload", "error", err)
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	h.log.Info("Processing code review task",
		"review_id", payload.ReviewResultID,
		"task_id", t.ResultWriter().TaskID())

	// Load ReviewResult with all relationships (Repository, Platform, LLMModel, Provider)
	var reviewResult model.ReviewResult
	err := h.db.
		Preload("Repository.Platform").
		Preload("Repository.LLMProvider").
		Preload("LLMProvider").
		First(&reviewResult, payload.ReviewResultID).Error

	if err != nil {
		h.log.Error("Failed to load review result", "error", err, "review_id", payload.ReviewResultID)
		return nil, fmt.Errorf("review result not found: %w", err)
	}

	// Verify LLM provider is configured
	if reviewResult.LLMProvider == nil {
		h.log.Error("No LLM provider configured", "review_id", reviewResult.ID)
		return nil, fmt.Errorf("no LLM provider configured")
	}

	// Update status to processing
	if err := h.db.Model(&reviewResult).Update("status", "processing").Error; err != nil {
		h.log.Error("Failed to update review status", "error", err)
	}

	return &reviewResult, nil
}

// fetchMRDiff fetches MR diff from GitLab
func (h *ReviewHandler) fetchMRDiff(review *model.ReviewResult) (string, *gitlab.Client, error) {
	h.log.Info("Fetching MR diff from GitLab",
		"review_id", review.ID,
		"mr_id", review.MergeRequestID,
		"platform", review.Repository.Platform.BaseURL)

	client := gitlab.NewClient(
		review.Repository.Platform.BaseURL,
		review.Repository.Platform.AccessToken,
	)

	// Note: We need platform_project_id from Repository, not from payload
	diff, err := client.GetMRDiff(
		int(review.Repository.PlatformRepoID),
		int(review.MergeRequestID),
	)

	if err != nil {
		h.log.Error("Failed to get MR diff", "error", err, "review_id", review.ID)
		return "", nil, fmt.Errorf("failed to get MR diff: %w", err)
	}

	h.log.Info("MR diff fetched successfully", "diff_size", len(diff), "review_id", review.ID)
	return diff, client, nil
}

// callLLMReview calls LLM to perform code review
func (h *ReviewHandler) callLLMReview(review *model.ReviewResult, diff string) (*llm.ReviewResponse, error) {
	h.log.Info("Starting LLM code review",
		"review_id", review.ID,
		"repository", review.Repository.Name,
		"llm_provider", review.LLMProvider.Name)

	// Get or create LLM client (uses pool for performance)
	llmClient, err := llm.GetOrCreateClient(
		review.LLMProvider,
		h.encryptionKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM client: %w", err)
	}

	// Build prompt data
	promptData := llm.BuildPromptData(
		diff,
		review.MRTitle,
		review.MRAuthor,
		review.SourceBranch,
		review.TargetBranch,
	)

	promptTemplate := h.getPromptTemplate(review)
	prompt := llm.RenderPrompt(promptTemplate, promptData)

	// Prepare review request
	reviewReq := llm.ReviewRequest{
		Diff:        diff,
		Prompt:      prompt,
		MaxTokens:   4096,
		Temperature: 0.7,
		ModelName:   review.LLMProvider.Model,
	}

	// Call LLM API
	h.log.Info("Calling LLM API",
		"provider", review.LLMProvider.Name,
		"model", review.LLMProvider.Model)

	reviewResp, err := llmClient.Review(reviewReq)
	if err != nil {
		return nil, fmt.Errorf("LLM API call failed: %w", err)
	}

	h.log.Info("LLM review completed",
		"tokens_used", reviewResp.TokensUsed,
		"duration", reviewResp.Duration,
		"suggestions", len(reviewResp.Suggestions))

	return reviewResp, nil
}

// saveReviewResults saves review results and suggestions to database
func (h *ReviewHandler) saveReviewResults(review *model.ReviewResult, resp *llm.ReviewResponse) error {
	h.log.Info("Saving review result with statistics",
		"review_id", review.ID,
		"suggestions_count", len(resp.Suggestions))

	// Generate structured JSON output for operations analysis
	jsonOutput := h.generateReviewJSON(review, resp)
	if jsonOutput != "" {
		resp.RawResponse = jsonOutput
	}

	storage := service.NewReviewStorageService(h.db)
	if err := storage.SaveReviewResult(review, resp); err != nil {
		h.log.Error("Failed to save review result", "error", err, "review_id", review.ID)
		return fmt.Errorf("failed to save review result: %w", err)
	}

	return nil
}

// generateReviewJSON generates structured JSON output for operations analysis
// Returns empty string if generation fails (non-fatal)
func (h *ReviewHandler) generateReviewJSON(review *model.ReviewResult, resp *llm.ReviewResponse) string {
	// Build context from review data
	ctx := llm.OutputContext{
		Repository: llm.ContextRepository{
			ID:             review.Repository.ID,
			Name:           review.Repository.Name,
			FullName:       review.Repository.FullPath,
			Platform:       review.Repository.Platform.PlatformType,
			PlatformRepoID: review.Repository.PlatformRepoID,
		},
		MergeRequest: llm.ContextMergeRequest{
			ID:           0, // System internal ID (not tracked separately)
			IID:          review.MergeRequestID,
			Title:        review.MRTitle,
			Author:       review.MRAuthor,
			SourceBranch: review.SourceBranch,
			TargetBranch: review.TargetBranch,
			WebURL:       review.MRWebURL,
		},
		Review: llm.ContextReview{
			ID:          review.ID,
			ReviewedAt:  time.Now(),
			LLMProvider: review.LLMProvider.Name,
			LLMModel:    review.LLMProvider.Model,
			TokensUsed:  resp.TokensUsed,
			DurationMs:  resp.Duration.Milliseconds(),
		},
	}

	// Build metadata
	meta := llm.OutputMetadata{
		PromptTemplate:     h.getPromptSource(review),
		CustomPromptUsed:   h.isCustomPromptUsed(review),
		RawResponseAvail:   true,
		ParserFallbackUsed: false,
	}

	jsonOutput, err := llm.FormatReviewAsJSON(resp, ctx, meta)
	if err != nil {
		h.log.Error("Failed to format review as JSON", "error", err, "review_id", review.ID)
		return ""
	}

	return jsonOutput
}

// getPromptSource returns the prompt source name for metadata
func (h *ReviewHandler) getPromptSource(review *model.ReviewResult) string {
	if review.Repository != nil &&
		review.Repository.CustomReviewPrompt != nil &&
		*review.Repository.CustomReviewPrompt != "" {
		return "repository"
	}
	if review.Repository != nil && review.Repository.ProjectID != 0 {
		globalPrompt := h.systemConfigSvc.GetReviewPrompt(review.Repository.ProjectID)
		if globalPrompt != llm.GetDefaultPrompt() {
			return "custom"
		}
	}
	return "default"
}

// isCustomPromptUsed returns whether a custom prompt is used
func (h *ReviewHandler) isCustomPromptUsed(review *model.ReviewResult) bool {
	return h.getPromptSource(review) != "default"
}

// postCommentToGitLab posts review comment to GitLab MR
// FIXED: Now returns error to trigger retry (was swallowing error before)
func (h *ReviewHandler) postCommentToGitLab(review *model.ReviewResult, client *gitlab.Client, resp *llm.ReviewResponse) error {
	h.log.Info("Posting review comment to GitLab MR",
		"review_id", review.ID,
		"mr_id", review.MergeRequestID)

	comment := gitlab.FormatReviewComment(resp)
	err := client.PostMRComment(
		int(review.Repository.PlatformRepoID),
		int(review.MergeRequestID),
		comment,
	)

	if err != nil {
		h.log.Error("Failed to post comment to GitLab", "error", err, "review_id", review.ID)
		return fmt.Errorf("failed to post comment: %w", err)
	}

	// Update comment_posted flag
	if err := h.db.Model(review).Update("comment_posted", true).Error; err != nil {
		h.log.Error("Failed to update comment_posted flag", "error", err)
		// Don't fail the task for this minor error
	}

	h.log.Info("Review comment posted successfully", "review_id", review.ID)
	return nil
}

// markReviewFailed marks review as failed in database
func (h *ReviewHandler) markReviewFailed(reviewID uint, errorMsg string) {
	storage := service.NewReviewStorageService(h.db)
	if err := storage.MarkReviewFailed(&model.ReviewResult{ID: reviewID}, errorMsg); err != nil {
		h.log.Error("Failed to mark review as failed", "error", err, "review_id", reviewID)
	}
}

// updateWebhookEventStatus updates webhook event status
// Note: This function assumes WebhookEventID is NOT nil - caller must check before calling
// Do not add defensive nil checks here - let it fail fast if misused
func (h *ReviewHandler) updateWebhookEventStatus(review *model.ReviewResult, status model.EventStatus) {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       status,
		"processed_at": now,
	}

	if err := h.db.Model(&model.WebhookEvent{}).Where("id = ?", *review.WebhookEventID).Updates(updates).Error; err != nil {
		h.log.Error("Failed to update webhook event status", "error", err, "webhook_event_id", *review.WebhookEventID)
		// Don't fail the task for this minor error - webhook status is for tracking only
	}
}

// getPromptTemplate returns the prompt template by priority
// Priority: Repository-level > Global config > Hardcoded default
func (h *ReviewHandler) getPromptTemplate(review *model.ReviewResult) string {
	// 1. Check repository-level custom prompt (highest priority)
	if review.Repository != nil &&
		review.Repository.CustomReviewPrompt != nil &&
		*review.Repository.CustomReviewPrompt != "" {
		h.log.Info("Using repository-level custom prompt", "repository_id", review.Repository.ID)
		return *review.Repository.CustomReviewPrompt
	}

	// 2. Check global config (requires Repository with ProjectID)
	if review.Repository != nil && review.Repository.ProjectID != 0 {
		globalPrompt := h.systemConfigSvc.GetReviewPrompt(review.Repository.ProjectID)
		if globalPrompt != llm.GetDefaultPrompt() {
			h.log.Info("Using global configured prompt", "project_id", review.Repository.ProjectID)
			return globalPrompt
		}
	}

	// 3. Fallback to hardcoded default
	h.log.Info("Using default hardcoded prompt")
	return llm.GetDefaultPrompt()
}

// logUsage logs LLM API usage to the database for operations analytics
// This function is best-effort - it won't fail the review if logging fails
func (h *ReviewHandler) logUsage(review *model.ReviewResult, resp *llm.ReviewResponse, apiErr error) {
	usageSvc := service.NewUsageService(h.db)

	// Build usage context from review
	ctx := service.UsageContext{
		ReviewResultID: &review.ID,
		RepositoryID:   review.RepositoryID,
		ProjectID:      review.Repository.ProjectID,
		LLMProviderID:  review.LLMProviderID,
		ModelName:      review.LLMProvider.Model,
		RequestType:    model.UsageTypeCodeReview,
	}

	// Build metrics based on success or failure
	var metrics service.UsageMetrics

	if apiErr != nil {
		// Failed request
		metrics = service.UsageMetrics{
			Status:   model.UsageStatusFailed,
			ErrorMsg: apiErr.Error(),
		}
	} else if resp != nil {
		// Successful request
		metrics = service.UsageMetrics{
			Status:           model.UsageStatusSuccess,
			PromptTokens:     resp.TokenUsage.PromptTokens,
			CompletionTokens: resp.TokenUsage.CompletionTokens,
			TotalTokens:      resp.TokenUsage.TotalTokens,
			DurationMs:       resp.Duration.Milliseconds(),
		}
	}

	// Log usage (best-effort, don't fail the review)
	if _, err := usageSvc.LogUsage(ctx, metrics); err != nil {
		h.log.Error("Failed to log LLM usage", "error", err, "review_id", review.ID)
		// Don't return error - usage logging is non-critical
	}

	// Update denormalized token fields in ReviewResult (if successful)
	if resp != nil {
		if err := usageSvc.UpdateReviewTokens(
			review.ID,
			resp.TokenUsage.PromptTokens,
			resp.TokenUsage.CompletionTokens,
			resp.TokenUsage.TotalTokens,
			resp.Duration.Milliseconds(),
		); err != nil {
			h.log.Error("Failed to update review tokens", "error", err, "review_id", review.ID)
			// Don't return error - this is also non-critical
		}
	}
}
