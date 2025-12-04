package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/handsoff/handsoff/internal/model"
	"github.com/handsoff/handsoff/internal/service"
	"github.com/handsoff/handsoff/pkg/logger"
	"gorm.io/gorm"
)

// ReviewHandler handles review-related HTTP requests
type ReviewHandler struct {
	db  *gorm.DB
	log *logger.Logger
}

// NewReviewHandler creates a new review handler
func NewReviewHandler(db *gorm.DB, log *logger.Logger) *ReviewHandler {
	return &ReviewHandler{
		db:  db,
		log: log,
	}
}

// ListReviews lists all review results with pagination and filtering
// GET /api/reviews?page=1&page_size=20&status=completed&repository_id=1
func (h *ReviewHandler) ListReviews(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Parse filter parameters
	status := c.Query("status")
	repositoryIDStr := c.Query("repository_id")
	author := c.Query("author")

	// Build query
	query := h.db.Model(&model.ReviewResult{}).
		Preload("Repository").
		Preload("LLMModel.Provider")

	// Apply filters
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if repositoryIDStr != "" {
		repositoryID, _ := strconv.ParseUint(repositoryIDStr, 10, 32)
		if repositoryID > 0 {
			query = query.Where("repository_id = ?", repositoryID)
		}
	}
	if author != "" {
		query = query.Where("mr_author LIKE ?", "%"+author+"%")
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		h.log.Error("Failed to count reviews", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count reviews"})
		return
	}

	// Get reviews with pagination
	var reviews []model.ReviewResult
	if err := query.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&reviews).Error; err != nil {
		h.log.Error("Failed to list reviews", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list reviews"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": reviews,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetReview retrieves a single review result by ID
// GET /api/reviews/:id
func (h *ReviewHandler) GetReview(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	storage := service.NewReviewStorageService(h.db)
	review, err := storage.GetReviewResult(uint(id))
	if err != nil {
		h.log.Error("Failed to get review", "error", err, "id", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	c.JSON(http.StatusOK, review)
}

// GetReviewStatistics retrieves statistics for a specific review
// GET /api/reviews/:id/statistics
func (h *ReviewHandler) GetReviewStatistics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	storage := service.NewReviewStorageService(h.db)
	stats, err := storage.GetReviewStatistics(uint(id))
	if err != nil {
		h.log.Error("Failed to get review statistics", "error", err, "id", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Statistics not found"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetDashboardStatistics retrieves overall dashboard statistics
// GET /api/dashboard/statistics
func (h *ReviewHandler) GetDashboardStatistics(c *gin.Context) {
	// Get user's project ID for data isolation
	projectID, err := getUserDefaultProjectID(c, h.db)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get project context"})
		return
	}

	var stats struct {
		TotalReviews       int64   `json:"total_reviews"`
		CompletedReviews   int64   `json:"completed_reviews"`
		PendingReviews     int64   `json:"pending_reviews"`
		FailedReviews      int64   `json:"failed_reviews"`
		AverageScore       float64 `json:"average_score"`
		TotalIssuesFound   int64   `json:"total_issues_found"`
		CriticalIssues     int64   `json:"critical_issues"`
		HighIssues         int64   `json:"high_issues"`
		MediumIssues       int64   `json:"medium_issues"`
		LowIssues          int64   `json:"low_issues"`
		SecurityIssues     int64   `json:"security_issues"`
		PerformanceIssues  int64   `json:"performance_issues"`
		QualityIssues      int64   `json:"quality_issues"`
	}

	// Single aggregated query with project isolation (6 queries → 1)
	err = h.db.Table("review_results").
		Joins("INNER JOIN repositories ON review_results.repository_id = repositories.id").
		Where("repositories.project_id = ?", projectID).
		Select(`
			COUNT(*) as total_reviews,
			SUM(CASE WHEN review_results.status = 'completed' THEN 1 ELSE 0 END) as completed_reviews,
			SUM(CASE WHEN review_results.status IN ('pending', 'processing') THEN 1 ELSE 0 END) as pending_reviews,
			SUM(CASE WHEN review_results.status = 'failed' THEN 1 ELSE 0 END) as failed_reviews,
			AVG(CASE WHEN review_results.status = 'completed' AND review_results.score > 0 THEN review_results.score ELSE NULL END) as average_score,
			SUM(CASE WHEN review_results.status = 'completed' THEN review_results.issues_found ELSE 0 END) as total_issues_found,
			SUM(CASE WHEN review_results.status = 'completed' THEN review_results.critical_issues_count ELSE 0 END) as critical_issues,
			SUM(CASE WHEN review_results.status = 'completed' THEN review_results.high_issues_count ELSE 0 END) as high_issues,
			SUM(CASE WHEN review_results.status = 'completed' THEN review_results.medium_issues_count ELSE 0 END) as medium_issues,
			SUM(CASE WHEN review_results.status = 'completed' THEN review_results.low_issues_count ELSE 0 END) as low_issues,
			SUM(CASE WHEN review_results.status = 'completed' THEN review_results.security_issues_count ELSE 0 END) as security_issues,
			SUM(CASE WHEN review_results.status = 'completed' THEN review_results.performance_issues_count ELSE 0 END) as performance_issues,
			SUM(CASE WHEN review_results.status = 'completed' THEN review_results.quality_issues_count ELSE 0 END) as quality_issues
		`).
		Scan(&stats).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetRecentReviews returns recent review results
// GET /api/dashboard/recent?limit=10
func (h *ReviewHandler) GetRecentReviews(c *gin.Context) {
	limit := h.parseLimitParam(c)

	projectID, ok := getProjectID(c)
	if !ok {
		h.log.Error(ErrMsgProjectIDMissing)
		RespondInternalError(c, ErrMsgInternalServer)
		return
	}

	reviews, err := h.fetchRecentReviews(projectID, limit)
	if err != nil {
		h.log.Error("Failed to get recent reviews", "error", err)
		RespondInternalError(c, "Failed to fetch recent reviews")
		return
	}

	RespondSuccess(c, reviews)
}

// parseLimitParam extracts and validates the limit query parameter
func (h *ReviewHandler) parseLimitParam(c *gin.Context) int {
	limitStr := c.DefaultQuery("limit", strconv.Itoa(DefaultRecentReviewsLimit))
	limit, _ := strconv.Atoi(limitStr)
	
	if limit < MinRecentReviewsLimit || limit > MaxRecentReviewsLimit {
		return DefaultRecentReviewsLimit
	}
	
	return limit
}

// fetchRecentReviews retrieves recent reviews from database
func (h *ReviewHandler) fetchRecentReviews(projectID uint, limit int) ([]model.ReviewResult, error) {
	var reviews []model.ReviewResult
	err := h.db.
		Where("project_id = ?", projectID).
		Preload("Repository", "project_id = ?", projectID).
		Preload("LLMModel.Provider", "project_id = ?", projectID).
		Order("created_at DESC").
		Limit(limit).
		Find(&reviews).Error
	
	return reviews, err
}

// parseDaysParam extracts and validates the days query parameter
func (h *ReviewHandler) parseDaysParam(c *gin.Context) int {
	daysStr := c.DefaultQuery("days", strconv.Itoa(DefaultTrendDays))
	days, _ := strconv.Atoi(daysStr)
	
	if days < MinTrendDays || days > MaxTrendDays {
		return DefaultTrendDaysOnError
	}
	
	return days
}

// GetRepositoryStatistics retrieves statistics for a specific repository
// GET /api/repositories/:id/statistics
func (h *ReviewHandler) GetRepositoryStatistics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	storage := service.NewReviewStorageService(h.db)
	stats, err := storage.GetRepositoryStatistics(uint(id))
	if err != nil {
		h.log.Error("Failed to get repository statistics", "error", err, "id", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Statistics not found"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetTrendData returns review trend data
// GET /api/dashboard/trend?days=7
func (h *ReviewHandler) GetTrendData(c *gin.Context) {
	days := h.parseDaysParam(c)

	projectID, ok := getProjectID(c)
	if !ok {
		h.log.Error(ErrMsgProjectIDMissing)
		RespondInternalError(c, ErrMsgInternalServer)
		return
	}

	type DailyStats struct {
		Date           string  `json:"date"`
		ReviewCount    int64   `json:"review_count"`
		AvgScore       float64 `json:"avg_score"`
		TotalIssues    int64   `json:"total_issues"`
		CriticalIssues int64   `json:"critical_issues"`
	}

	var trends []DailyStats
	startDate := time.Now().AddDate(0, 0, -days)

	if err := h.db.Model(&model.ReviewResult{}).
		Where("project_id = ?", projectID).
		Where("created_at >= ?", startDate).
		Where("status = ?", "completed").
		Select(`
			DATE(created_at) as date,
			COUNT(*) as review_count,
			AVG(score) as avg_score,
			COUNT(*) as total_issues,
			SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as critical_issues
		`, "completed").
		Group("DATE(created_at)").
		Order("date ASC").
		Find(&trends).Error; err != nil {
		h.log.Error("Failed to fetch trend data", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trend statistics"})
		return
	}

	c.JSON(http.StatusOK, trends)
}
