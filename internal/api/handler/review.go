package handler

import (
	"net/http"
	"strconv"

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

	// Total reviews
	h.db.Model(&model.ReviewResult{}).Count(&stats.TotalReviews)

	// Reviews by status
	h.db.Model(&model.ReviewResult{}).Where("status = ?", "completed").Count(&stats.CompletedReviews)
	h.db.Model(&model.ReviewResult{}).Where("status IN ?", []string{"pending", "processing"}).Count(&stats.PendingReviews)
	h.db.Model(&model.ReviewResult{}).Where("status = ?", "failed").Count(&stats.FailedReviews)

	// Average score
	var avgScore struct {
		Avg float64
	}
	h.db.Model(&model.ReviewResult{}).
		Where("status = ? AND score > 0", "completed").
		Select("AVG(score) as avg").
		Scan(&avgScore)
	stats.AverageScore = avgScore.Avg

	// Sum statistics from completed reviews
	var sumStats struct {
		TotalIssues   int64
		Critical      int64
		High          int64
		Medium        int64
		Low           int64
		Security      int64
		Performance   int64
		Quality       int64
	}
	h.db.Model(&model.ReviewResult{}).
		Where("status = ?", "completed").
		Select(`
			SUM(issues_found) as total_issues,
			SUM(critical_issues_count) as critical,
			SUM(high_issues_count) as high,
			SUM(medium_issues_count) as medium,
			SUM(low_issues_count) as low,
			SUM(security_issues_count) as security,
			SUM(performance_issues_count) as performance,
			SUM(quality_issues_count) as quality
		`).
		Scan(&sumStats)

	stats.TotalIssuesFound = sumStats.TotalIssues
	stats.CriticalIssues = sumStats.Critical
	stats.HighIssues = sumStats.High
	stats.MediumIssues = sumStats.Medium
	stats.LowIssues = sumStats.Low
	stats.SecurityIssues = sumStats.Security
	stats.PerformanceIssues = sumStats.Performance
	stats.QualityIssues = sumStats.Quality

	c.JSON(http.StatusOK, stats)
}

// GetRecentReviews retrieves recent review results
// GET /api/dashboard/recent?limit=10
func (h *ReviewHandler) GetRecentReviews(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	var reviews []model.ReviewResult
	if err := h.db.
		Preload("Repository").
		Preload("LLMModel.Provider").
		Order("created_at DESC").
		Limit(limit).
		Find(&reviews).Error; err != nil {
		h.log.Error("Failed to get recent reviews", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recent reviews"})
		return
	}

	c.JSON(http.StatusOK, reviews)
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

// GetTrendData retrieves trend data for charts
// GET /api/dashboard/trends?days=30
func (h *ReviewHandler) GetTrendData(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	type DailyStats struct {
		Date            string  `json:"date"`
		ReviewCount     int64   `json:"review_count"`
		AverageScore    float64 `json:"average_score"`
		TotalIssues     int64   `json:"total_issues"`
		CriticalIssues  int64   `json:"critical_issues"`
	}

	var trends []DailyStats
	h.db.Model(&model.ReviewResult{}).
		Where("created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)", days).
		Where("status = ?", "completed").
		Select(`
			DATE(created_at) as date,
			COUNT(*) as review_count,
			AVG(score) as average_score,
			SUM(issues_found) as total_issues,
			SUM(critical_issues_count) as critical_issues
		`).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&trends)

	c.JSON(http.StatusOK, trends)
}
