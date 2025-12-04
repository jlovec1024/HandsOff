package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/handsoff/handsoff/internal/gitlab"
	"github.com/handsoff/handsoff/internal/llm"
	"github.com/handsoff/handsoff/internal/model"
	"github.com/handsoff/handsoff/internal/service"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// ReviewHandler handles code review tasks
type ReviewHandler struct {
	db            *gorm.DB
	log           Logger
	encryptionKey string
}

// Logger interface for handler logging
type Logger interface {
	Info(...interface{})
	Error(...interface{})
}

// NewReviewHandler creates a new review handler
func NewReviewHandler(db *gorm.DB, log Logger, encryptionKey string) *ReviewHandler {
	return &ReviewHandler{
		db:            db,
		log:           log,
		encryptionKey: encryptionKey,
	}
}

// HandleCodeReview processes code review tasks
func (h *ReviewHandler) HandleCodeReview(ctx context.Context, t *asynq.Task) error {
	// Parse task payload
	var payload CodeReviewPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		h.log.Error("Failed to unmarshal task payload", "error", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	h.log.Info("Processing code review task",
		"repository_id", payload.RepositoryID,
		"mr_id", payload.MergeRequestID,
		"task_id", t.ResultWriter().TaskID())

	// Find the repository with relationships
	var repo model.Repository
	if err := h.db.Preload("Platform").Preload("LLMModel.Provider").
		First(&repo, payload.RepositoryID).Error; err != nil {
		h.log.Error("Failed to find repository", "error", err, "repository_id", payload.RepositoryID)
		return fmt.Errorf("repository not found: %w", err)
	}

	// Verify LLM model is configured
	if repo.LLMModel == nil {
		h.log.Error("No LLM model configured", "repository_id", payload.RepositoryID)
		return fmt.Errorf("no LLM model configured for repository %d", payload.RepositoryID)
	}

	// Find or create review result record
	var reviewResult model.ReviewResult
	if err := h.db.Where("repository_id = ? AND merge_request_id = ?",
		payload.RepositoryID, payload.MergeRequestID).
		First(&reviewResult).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new review result if not exists
			reviewResult = model.ReviewResult{
				RepositoryID:   payload.RepositoryID,
				MergeRequestID: payload.MergeRequestID,
				MRTitle:        payload.MRTitle,
				MRAuthor:       payload.MRAuthor,
				SourceBranch:   payload.SourceBranch,
				TargetBranch:   payload.TargetBranch,
				MRWebURL:       payload.MRWebURL,
				LLMModelID:     *repo.LLMModelID,
				Status:         "processing",
				CommentPosted:  false,
			}
			if err := h.db.Create(&reviewResult).Error; err != nil {
				h.log.Error("Failed to create review result", "error", err)
				return fmt.Errorf("failed to create review result: %w", err)
			}
		} else {
			h.log.Error("Failed to query review result", "error", err)
			return fmt.Errorf("failed to query review result: %w", err)
		}
	}

	// Update status to processing
	if err := h.db.Model(&reviewResult).Update("status", "processing").Error; err != nil {
		h.log.Error("Failed to update review status", "error", err)
	}

	h.log.Info("Review result record found/created",
		"review_id", reviewResult.ID,
		"status", reviewResult.Status)

	// Get MR diff from GitLab
	h.log.Info("Fetching MR diff from GitLab",
		"project_id", payload.ProjectID,
		"mr_id", payload.MergeRequestID,
		"platform_url", repo.Platform.BaseURL)

	gitlabClient := gitlab.NewClient(repo.Platform.BaseURL, repo.Platform.AccessToken)
	diff, err := gitlabClient.GetMRDiff(int(payload.ProjectID), int(payload.MergeRequestID))
	if err != nil {
		// Update status to failed
		h.db.Model(&reviewResult).Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": fmt.Sprintf("Failed to get MR diff: %v", err),
		})
		h.log.Error("Failed to get MR diff from GitLab", "error", err)
		return fmt.Errorf("failed to get MR diff: %w", err)
	}

	h.log.Info("MR diff fetched successfully", "diff_size", len(diff))

	// Perform LLM code review
	h.log.Info("Starting LLM code review",
		"review_id", reviewResult.ID,
		"repository", repo.Name,
		"mr_id", payload.MergeRequestID,
		"llm_provider", repo.LLMModel.Provider.Type)

	reviewResp, err := h.performLLMReview(repo, payload, diff)
	if err != nil {
		// Use storage service to mark as failed
		storage := service.NewReviewStorageService(h.db)
		if storageErr := storage.MarkReviewFailed(&reviewResult, err.Error()); storageErr != nil {
			h.log.Error("Failed to mark review as failed", "error", storageErr)
		}
		h.log.Error("LLM review failed", "error", err, "review_id", reviewResult.ID)
		return fmt.Errorf("LLM review failed: %w", err)
	}

	// Use storage service to save review result with statistics
	h.log.Info("Saving review result with statistics", "suggestions_count", len(reviewResp.Suggestions))
	storage := service.NewReviewStorageService(h.db)
	if err := storage.SaveReviewResult(&reviewResult, reviewResp); err != nil {
		h.log.Error("Failed to save review result", "error", err)
		return fmt.Errorf("failed to save review result: %w", err)
	}

	h.log.Info("Code review completed successfully",
		"review_id", reviewResult.ID,
		"repository_id", payload.RepositoryID,
		"mr_id", payload.MergeRequestID,
		"score", reviewResp.Score,
		"suggestions_count", len(reviewResp.Suggestions))

	// Post comment to GitLab MR
	h.log.Info("Posting review comment to GitLab MR",
		"project_id", payload.ProjectID,
		"mr_id", payload.MergeRequestID)

	comment := gitlab.FormatReviewComment(reviewResp)
	if err := gitlabClient.PostMRComment(int(payload.ProjectID), int(payload.MergeRequestID), comment); err != nil {
		h.log.Error("Failed to post comment to GitLab", "error", err)
		// Don't fail the task - review is already saved
		// Just log the error and mark comment_posted as false
	} else {
		// Update comment_posted flag
		if err := h.db.Model(&reviewResult).Update("comment_posted", true).Error; err != nil {
			h.log.Error("Failed to update comment_posted flag", "error", err)
		}
		h.log.Info("Review comment posted successfully to GitLab MR")
	}

	return nil
}

// performLLMReview calls LLM to perform code review
func (h *ReviewHandler) performLLMReview(repo model.Repository, payload CodeReviewPayload, diff string) (*llm.ReviewResponse, error) {
	// Get or create LLM client (uses pool for performance)
	llmClient, err := llm.GetOrCreateClient(repo.LLMModel.Provider, repo.LLMModel, h.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM client: %w", err)
	}

	// Build prompt data
	promptData := llm.BuildPromptData(
		diff,
		payload.MRTitle,
		payload.MRAuthor,
		payload.SourceBranch,
		payload.TargetBranch,
	)

	// Render prompt using default template
	prompt := llm.RenderPrompt(llm.GetDefaultPrompt(), promptData)

	// Prepare review request
	reviewReq := llm.ReviewRequest{
		Diff:        diff,
		Prompt:      prompt,
		MaxTokens:   repo.LLMModel.MaxTokens,
		Temperature: repo.LLMModel.Temperature,
		ModelName:   repo.LLMModel.ModelName,
	}

	// Call LLM API
	h.log.Info("Calling LLM API", "provider", repo.LLMModel.Provider.Type, "model", repo.LLMModel.ModelName)
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


