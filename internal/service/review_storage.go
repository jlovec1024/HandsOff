package service

import (
	"fmt"
	"time"

	"github.com/handsoff/handsoff/internal/llm"
	"github.com/handsoff/handsoff/internal/model"
	"gorm.io/gorm"
)

// ReviewStorageService handles review result storage operations
type ReviewStorageService struct {
	db *gorm.DB
}

// NewReviewStorageService creates a new review storage service
func NewReviewStorageService(db *gorm.DB) *ReviewStorageService {
	return &ReviewStorageService{db: db}
}

// SaveReviewResult saves the review result with statistics and suggestions in a transaction
func (s *ReviewStorageService) SaveReviewResult(
	reviewResult *model.ReviewResult,
	response *llm.ReviewResponse,
) error {
	// Use transaction to ensure atomicity
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Calculate statistics
		stats := calculateStatistics(response.Suggestions)

		// Update review result with all fields
		now := time.Now()
		updates := map[string]interface{}{
			"status":                  "completed",
			"summary":                 response.Summary,
			"score":                   response.Score,
			"raw_result":              response.RawResponse,
			"reviewed_at":             &now,
			"issues_found":            stats.TotalIssues,
			"critical_issues_count":   stats.CriticalCount,
			"high_issues_count":       stats.HighCount,
			"medium_issues_count":     stats.MediumCount,
			"low_issues_count":        stats.LowCount,
			"security_issues_count":   stats.SecurityCount,
			"performance_issues_count": stats.PerformanceCount,
			"quality_issues_count":    stats.QualityCount,
		}

		if err := tx.Model(reviewResult).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update review result: %w", err)
		}

		// Batch insert fix suggestions
		if len(response.Suggestions) > 0 {
		fixSuggestions := make([]model.FixSuggestion, 0, len(response.Suggestions))
		for _, sug := range response.Suggestions {
			fixSuggestions = append(fixSuggestions, model.FixSuggestion{
				ReviewResultID: reviewResult.ID,
				FilePath:       sug.FilePath,
				LineStart:      sug.LineStart,
				LineEnd:        sug.LineEnd,
				Severity:       sug.Severity,
				Category:       sug.Category,
				Description:    sug.Description,
				Suggestion:     sug.Suggestion,
				CodeSnippet:    sug.CodeSnippet,
			})
		}

			// Batch insert with CreateInBatches for better performance
			if err := tx.CreateInBatches(fixSuggestions, 100).Error; err != nil {
				return fmt.Errorf("failed to batch insert fix suggestions: %w", err)
			}
		}

		return nil
	})
}

// MarkReviewFailed updates review result as failed
func (s *ReviewStorageService) MarkReviewFailed(reviewResult *model.ReviewResult, errorMsg string) error {
	updates := map[string]interface{}{
		"status":        "failed",
		"error_message": errorMsg,
	}
	return s.db.Model(reviewResult).Updates(updates).Error
}

// UpdateCommentStatus updates the comment_posted flag
func (s *ReviewStorageService) UpdateCommentStatus(reviewResult *model.ReviewResult, posted bool) error {
	return s.db.Model(reviewResult).Update("comment_posted", posted).Error
}

// ReviewStatistics holds statistics about review results
type ReviewStatistics struct {
	TotalIssues        int
	CriticalCount      int
	HighCount          int
	MediumCount        int
	LowCount           int
	SecurityCount      int
	PerformanceCount   int
	QualityCount       int
	StyleCount         int
	BugCount           int
	OtherCount         int
}

// calculateStatistics calculates statistics from suggestions
func calculateStatistics(suggestions []llm.FixSuggestion) ReviewStatistics {
	stats := ReviewStatistics{
		TotalIssues: len(suggestions),
	}

	for _, sug := range suggestions {
		// Count by severity
		switch sug.Severity {
		case "critical":
			stats.CriticalCount++
		case "high":
			stats.HighCount++
		case "medium":
			stats.MediumCount++
		case "low":
			stats.LowCount++
		}

		// Count by category (support common variations)
		category := sug.Category
		switch {
		case category == "security":
			stats.SecurityCount++
		case category == "performance":
			stats.PerformanceCount++
		case category == "quality" || category == "code-quality" || category == "maintainability":
			stats.QualityCount++
		case category == "style" || category == "formatting":
			stats.StyleCount++
		case category == "bug" || category == "correctness":
			stats.BugCount++
		default:
			stats.OtherCount++
		}
	}

	return stats
}

// GetReviewResult retrieves a review result with suggestions
func (s *ReviewStorageService) GetReviewResult(id uint) (*model.ReviewResult, error) {
	var result model.ReviewResult
	if err := s.db.Preload("FixSuggestions").First(&result, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get review result: %w", err)
	}
	return &result, nil
}

// GetReviewResultByMR retrieves a review result by repository and MR ID
func (s *ReviewStorageService) GetReviewResultByMR(repositoryID uint, mrID int64) (*model.ReviewResult, error) {
	var result model.ReviewResult
	if err := s.db.Preload("FixSuggestions").
		Where("repository_id = ? AND merge_request_id = ?", repositoryID, mrID).
		First(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to get review result: %w", err)
	}
	return &result, nil
}

// ListReviewResults lists review results with pagination
func (s *ReviewStorageService) ListReviewResults(
	repositoryID uint,
	page, pageSize int,
	status string,
) ([]model.ReviewResult, int64, error) {
	var results []model.ReviewResult
	var total int64

	query := s.db.Model(&model.ReviewResult{})
	
	if repositoryID > 0 {
		query = query.Where("repository_id = ?", repositoryID)
	}
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count review results: %w", err)
	}

	// Paginate
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&results).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list review results: %w", err)
	}

	return results, total, nil
}

// GetReviewStatistics gets statistics for a single review result
func (s *ReviewStorageService) GetReviewStatistics(reviewID uint) (*ReviewStatistics, error) {
	var result model.ReviewResult
	if err := s.db.First(&result, reviewID).Error; err != nil {
		return nil, fmt.Errorf("failed to get review result: %w", err)
	}

	// Get fix suggestions count
	var suggestionsCount int64
	s.db.Model(&model.FixSuggestion{}).
		Where("review_result_id = ?", reviewID).
		Count(&suggestionsCount)

	stats := &ReviewStatistics{
		TotalIssues:      result.IssuesFound,
		CriticalCount:    result.CriticalIssuesCount,
		HighCount:        result.HighIssuesCount,
		MediumCount:      result.MediumIssuesCount,
		LowCount:         result.LowIssuesCount,
		SecurityCount:    result.SecurityIssuesCount,
		PerformanceCount: result.PerformanceIssuesCount,
		QualityCount:     result.QualityIssuesCount,
	}

	return stats, nil
}

// RepositoryStatistics holds aggregated statistics for a repository
type RepositoryStatistics struct {
	TotalReviews     int64   `json:"total_reviews"`
	CompletedReviews int64   `json:"completed_reviews"`
	FailedReviews    int64   `json:"failed_reviews"`
	AverageScore     float64 `json:"average_score"`
	TotalIssues      int     `json:"total_issues"`
	CriticalIssues   int     `json:"critical_issues"`
}

// GetRepositoryStatistics retrieves aggregated statistics for a repository
func (s *ReviewStorageService) GetRepositoryStatistics(repositoryID uint) (*RepositoryStatistics, error) {
	var stats RepositoryStatistics

	// Count total reviews
	if err := s.db.Model(&model.ReviewResult{}).
		Where("repository_id = ?", repositoryID).
		Count(&stats.TotalReviews).Error; err != nil {
		return nil, fmt.Errorf("failed to count total reviews: %w", err)
	}

	// Count completed reviews
	if err := s.db.Model(&model.ReviewResult{}).
		Where("repository_id = ? AND status = ?", repositoryID, "completed").
		Count(&stats.CompletedReviews).Error; err != nil {
		return nil, fmt.Errorf("failed to count completed reviews: %w", err)
	}

	// Count failed reviews
	if err := s.db.Model(&model.ReviewResult{}).
		Where("repository_id = ? AND status = ?", repositoryID, "failed").
		Count(&stats.FailedReviews).Error; err != nil {
		return nil, fmt.Errorf("failed to count failed reviews: %w", err)
	}

	// Calculate average score
	var avgScore struct {
		Avg float64
	}
	s.db.Model(&model.ReviewResult{}).
		Where("repository_id = ? AND status = ? AND score > 0", repositoryID, "completed").
		Select("AVG(score) as avg").
		Scan(&avgScore)
	stats.AverageScore = avgScore.Avg

	// Sum issues
	var sumIssues struct {
		Total    int
		Critical int
	}
	s.db.Model(&model.ReviewResult{}).
		Where("repository_id = ? AND status = ?", repositoryID, "completed").
		Select("SUM(issues_found) as total, SUM(critical_issues_count) as critical").
		Scan(&sumIssues)
	stats.TotalIssues = sumIssues.Total
	stats.CriticalIssues = sumIssues.Critical

	return &stats, nil
}
