package service

import (
	"time"

	"github.com/handsoff/handsoff/internal/model"
	"gorm.io/gorm"
)

// UsageService handles LLM usage logging and token statistics
type UsageService struct {
	db *gorm.DB
}

// NewUsageService creates a new usage service
func NewUsageService(db *gorm.DB) *UsageService {
	return &UsageService{db: db}
}

// UsageContext contains all business context needed to log a usage record
type UsageContext struct {
	ReviewResultID *uint
	RepositoryID   uint
	ProjectID      uint
	LLMProviderID  uint
	ModelName      string
	RequestType    model.UsageRequestType
}

// UsageMetrics contains the actual metrics from an LLM API call
type UsageMetrics struct {
	Status           model.UsageStatus
	ErrorCode        string
	ErrorMsg         string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	DurationMs       int64
	RequestSizeB     int
	ResponseSizeB    int
}

// LogUsage logs an LLM API call to the usage log table
func (s *UsageService) LogUsage(ctx UsageContext, metrics UsageMetrics) (*model.LLMUsageLog, error) {
	log := &model.LLMUsageLog{
		CreatedAt:        time.Now(),
		ReviewResultID:   ctx.ReviewResultID,
		RepositoryID:     ctx.RepositoryID,
		ProjectID:        ctx.ProjectID,
		LLMProviderID:    ctx.LLMProviderID,
		ModelName:        ctx.ModelName,
		RequestType:      ctx.RequestType,
		Status:           metrics.Status,
		ErrorCode:        metrics.ErrorCode,
		ErrorMsg:         metrics.ErrorMsg,
		PromptTokens:     metrics.PromptTokens,
		CompletionTokens: metrics.CompletionTokens,
		TotalTokens:      metrics.TotalTokens,
		DurationMs:       metrics.DurationMs,
		RequestSizeB:     metrics.RequestSizeB,
		ResponseSizeB:    metrics.ResponseSizeB,
	}

	if err := s.db.Create(log).Error; err != nil {
		return nil, err
	}

	return log, nil
}

// UpdateReviewTokens updates the denormalized token fields in ReviewResult
func (s *UsageService) UpdateReviewTokens(reviewResultID uint, promptTokens, completionTokens, totalTokens int, durationMs int64) error {
	updates := map[string]interface{}{
		"prompt_tokens":     promptTokens,
		"completion_tokens": completionTokens,
		"total_tokens":      totalTokens,
		"llm_duration_ms":   durationMs,
	}
	return s.db.Model(&model.ReviewResult{}).Where("id = ?", reviewResultID).Updates(updates).Error
}

// TokenStats represents aggregated token statistics
type TokenStats struct {
	TotalCalls       int64   `json:"total_calls"`
	SuccessfulCalls  int64   `json:"successful_calls"`
	FailedCalls      int64   `json:"failed_calls"`
	TotalTokens      int64   `json:"total_tokens"`
	PromptTokens     int64   `json:"prompt_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	AvgDurationMs    float64 `json:"avg_duration_ms"`
	SuccessRate      float64 `json:"success_rate"`
}

// GetProjectTokenStats returns token statistics for a project
func (s *UsageService) GetProjectTokenStats(projectID uint, startDate, endDate time.Time) (*TokenStats, error) {
	var stats TokenStats

	err := s.db.Model(&model.LLMUsageLog{}).
		Where("project_id = ? AND created_at >= ? AND created_at <= ?", projectID, startDate, endDate).
		Select(`
			COUNT(*) as total_calls,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as successful_calls,
			SUM(CASE WHEN status != 'success' THEN 1 ELSE 0 END) as failed_calls,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(prompt_tokens), 0) as prompt_tokens,
			COALESCE(SUM(completion_tokens), 0) as completion_tokens,
			COALESCE(AVG(duration_ms), 0) as avg_duration_ms
		`).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	if stats.TotalCalls > 0 {
		stats.SuccessRate = float64(stats.SuccessfulCalls) / float64(stats.TotalCalls) * 100
	}

	return &stats, nil
}

// RepositoryTokenStats represents token statistics for a repository
type RepositoryTokenStats struct {
	RepositoryID   uint   `json:"repository_id"`
	RepositoryName string `json:"repository_name"`
	TotalTokens    int64  `json:"total_tokens"`
	ReviewCount    int64  `json:"review_count"`
	AvgTokens      int64  `json:"avg_tokens"`
}

// GetTopRepositoriesByTokens returns top N repositories by token consumption
func (s *UsageService) GetTopRepositoriesByTokens(projectID uint, limit int, startDate, endDate time.Time) ([]RepositoryTokenStats, error) {
	var results []RepositoryTokenStats

	err := s.db.Model(&model.LLMUsageLog{}).
		Select(`
			llm_usage_logs.repository_id,
			repositories.name as repository_name,
			COALESCE(SUM(llm_usage_logs.total_tokens), 0) as total_tokens,
			COUNT(DISTINCT llm_usage_logs.review_result_id) as review_count,
			COALESCE(AVG(llm_usage_logs.total_tokens), 0) as avg_tokens
		`).
		Joins("LEFT JOIN repositories ON llm_usage_logs.repository_id = repositories.id").
		Where("llm_usage_logs.project_id = ? AND llm_usage_logs.created_at >= ? AND llm_usage_logs.created_at <= ?",
			projectID, startDate, endDate).
		Group("llm_usage_logs.repository_id, repositories.name").
		Order("total_tokens DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// DailyTokenStats represents daily token statistics
type DailyTokenStats struct {
	Date         string `json:"date"`
	TotalTokens  int64  `json:"total_tokens"`
	ReviewCount  int64  `json:"review_count"`
	AvgDuration  int64  `json:"avg_duration_ms"`
	SuccessRate  float64 `json:"success_rate"`
}

// GetDailyTokenStats returns daily token statistics for trend analysis
func (s *UsageService) GetDailyTokenStats(projectID uint, days int) ([]DailyTokenStats, error) {
	var results []DailyTokenStats

	startDate := time.Now().AddDate(0, 0, -days)

	err := s.db.Model(&model.LLMUsageLog{}).
		Select(`
			DATE(created_at) as date,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COUNT(DISTINCT review_result_id) as review_count,
			COALESCE(AVG(duration_ms), 0) as avg_duration,
			CASE WHEN COUNT(*) > 0 
				THEN SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) * 100.0 / COUNT(*) 
				ELSE 0 
			END as success_rate
		`).
		Where("project_id = ? AND created_at >= ?", projectID, startDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetUsageLogsByReview returns all usage logs for a specific review
func (s *UsageService) GetUsageLogsByReview(reviewResultID uint) ([]model.LLMUsageLog, error) {
	var logs []model.LLMUsageLog

	err := s.db.
		Where("review_result_id = ?", reviewResultID).
		Order("created_at ASC").
		Find(&logs).Error

	if err != nil {
		return nil, err
	}

	return logs, nil
}
