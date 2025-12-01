package service

import (
	"testing"

	"github.com/handsoff/handsoff/internal/llm"
	"github.com/handsoff/handsoff/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&model.ReviewResult{}, &model.FixSuggestion{}, &model.Repository{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

func TestSaveReviewResult(t *testing.T) {
	db := setupTestDB(t)
	storage := NewReviewStorageService(db)

	// Create test repository
	repo := model.Repository{
		Name: "Test Repo",
	}
	if err := db.Create(&repo).Error; err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}

	// Create initial review result
	reviewResult := model.ReviewResult{
		RepositoryID:   repo.ID,
		MergeRequestID: 123,
		Status:         "processing",
	}
	if err := db.Create(&reviewResult).Error; err != nil {
		t.Fatalf("Failed to create review result: %v", err)
	}

	// Create mock LLM response
	response := &llm.ReviewResponse{
		Summary:     "This MR has some security issues",
		Score:       75,
		RawResponse: "detailed response",
		Suggestions: []llm.FixSuggestion{
			{
				FilePath:    "auth.go",
				LineStart:   10,
				LineEnd:     15,
				Severity:    "critical",
				Category:    "security",
				Description: "SQL injection vulnerability",
				Suggestion:  "Use parameterized queries",
				CodeSnippet: "query := \"SELECT * FROM users WHERE id = \" + id",
			},
			{
				FilePath:    "auth.go",
				LineStart:   20,
				LineEnd:     25,
				Severity:    "high",
				Category:    "performance",
				Description: "N+1 query problem",
				Suggestion:  "Use eager loading",
				CodeSnippet: "for user in users { fetchOrders(user.id) }",
			},
			{
				FilePath:    "utils.go",
				LineStart:   30,
				LineEnd:     32,
				Severity:    "medium",
				Category:    "code-quality",
				Description: "Unused variable",
				Suggestion:  "Remove unused variable",
				CodeSnippet: "var unused = 123",
			},
		},
	}

	// Save review result
	err := storage.SaveReviewResult(&reviewResult, response)
	if err != nil {
		t.Fatalf("SaveReviewResult failed: %v", err)
	}

	// Verify review result was updated
	var updatedReview model.ReviewResult
	if err := db.First(&updatedReview, reviewResult.ID).Error; err != nil {
		t.Fatalf("Failed to fetch updated review: %v", err)
	}

	// Check basic fields
	if updatedReview.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", updatedReview.Status)
	}
	if updatedReview.Summary != response.Summary {
		t.Errorf("Expected summary '%s', got '%s'", response.Summary, updatedReview.Summary)
	}
	if updatedReview.Score != response.Score {
		t.Errorf("Expected score %d, got %d", response.Score, updatedReview.Score)
	}
	if updatedReview.ReviewedAt == nil {
		t.Error("Expected ReviewedAt to be set")
	}

	// Check statistics fields
	if updatedReview.IssuesFound != 3 {
		t.Errorf("Expected IssuesFound 3, got %d", updatedReview.IssuesFound)
	}
	if updatedReview.CriticalIssuesCount != 1 {
		t.Errorf("Expected CriticalIssuesCount 1, got %d", updatedReview.CriticalIssuesCount)
	}
	if updatedReview.HighIssuesCount != 1 {
		t.Errorf("Expected HighIssuesCount 1, got %d", updatedReview.HighIssuesCount)
	}
	if updatedReview.MediumIssuesCount != 1 {
		t.Errorf("Expected MediumIssuesCount 1, got %d", updatedReview.MediumIssuesCount)
	}
	if updatedReview.LowIssuesCount != 0 {
		t.Errorf("Expected LowIssuesCount 0, got %d", updatedReview.LowIssuesCount)
	}
	if updatedReview.SecurityIssuesCount != 1 {
		t.Errorf("Expected SecurityIssuesCount 1, got %d", updatedReview.SecurityIssuesCount)
	}
	if updatedReview.PerformanceIssuesCount != 1 {
		t.Errorf("Expected PerformanceIssuesCount 1, got %d", updatedReview.PerformanceIssuesCount)
	}
	if updatedReview.QualityIssuesCount != 1 {
		t.Errorf("Expected QualityIssuesCount 1, got %d", updatedReview.QualityIssuesCount)
	}

	// Verify fix suggestions were created
	var suggestions []model.FixSuggestion
	if err := db.Where("review_result_id = ?", reviewResult.ID).Find(&suggestions).Error; err != nil {
		t.Fatalf("Failed to fetch suggestions: %v", err)
	}

	if len(suggestions) != 3 {
		t.Fatalf("Expected 3 suggestions, got %d", len(suggestions))
	}

	// Verify first suggestion details
	if suggestions[0].FilePath != "auth.go" {
		t.Errorf("Expected FilePath 'auth.go', got '%s'", suggestions[0].FilePath)
	}
	if suggestions[0].Severity != "critical" {
		t.Errorf("Expected Severity 'critical', got '%s'", suggestions[0].Severity)
	}
	if suggestions[0].Category != "security" {
		t.Errorf("Expected Category 'security', got '%s'", suggestions[0].Category)
	}
}

func TestMarkReviewFailed(t *testing.T) {
	db := setupTestDB(t)
	storage := NewReviewStorageService(db)

	// Create test repository
	repo := model.Repository{
		Name: "Test Repo",
	}
	if err := db.Create(&repo).Error; err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}

	// Create initial review result
	reviewResult := model.ReviewResult{
		RepositoryID:   repo.ID,
		MergeRequestID: 123,
		Status:         "processing",
	}
	if err := db.Create(&reviewResult).Error; err != nil {
		t.Fatalf("Failed to create review result: %v", err)
	}

	// Mark as failed
	err := storage.MarkReviewFailed(&reviewResult, "LLM API timeout")
	if err != nil {
		t.Fatalf("MarkReviewFailed failed: %v", err)
	}

	// Verify review was marked as failed
	var updatedReview model.ReviewResult
	if err := db.First(&updatedReview, reviewResult.ID).Error; err != nil {
		t.Fatalf("Failed to fetch updated review: %v", err)
	}

	if updatedReview.Status != "failed" {
		t.Errorf("Expected status 'failed', got '%s'", updatedReview.Status)
	}
	if updatedReview.ErrorMessage != "LLM API timeout" {
		t.Errorf("Expected error message 'LLM API timeout', got '%s'", updatedReview.ErrorMessage)
	}
}

func TestCalculateStatistics(t *testing.T) {
	suggestions := []llm.FixSuggestion{
		{Severity: "critical", Category: "security"},
		{Severity: "critical", Category: "security"},
		{Severity: "high", Category: "performance"},
		{Severity: "medium", Category: "code-quality"},
		{Severity: "low", Category: "style"},
	}

	stats := calculateStatistics(suggestions)

	if stats.TotalIssues != 5 {
		t.Errorf("Expected TotalIssues 5, got %d", stats.TotalIssues)
	}
	if stats.CriticalCount != 2 {
		t.Errorf("Expected CriticalCount 2, got %d", stats.CriticalCount)
	}
	if stats.HighCount != 1 {
		t.Errorf("Expected HighCount 1, got %d", stats.HighCount)
	}
	if stats.MediumCount != 1 {
		t.Errorf("Expected MediumCount 1, got %d", stats.MediumCount)
	}
	if stats.LowCount != 1 {
		t.Errorf("Expected LowCount 1, got %d", stats.LowCount)
	}
	if stats.SecurityCount != 2 {
		t.Errorf("Expected SecurityCount 2, got %d", stats.SecurityCount)
	}
	if stats.PerformanceCount != 1 {
		t.Errorf("Expected PerformanceCount 1, got %d", stats.PerformanceCount)
	}
	if stats.QualityCount != 1 {
		t.Errorf("Expected QualityCount 1, got %d", stats.QualityCount)
	}
}

func TestSaveReviewResultEmptySuggestions(t *testing.T) {
	db := setupTestDB(t)
	storage := NewReviewStorageService(db)

	// Create test repository
	repo := model.Repository{
		Name: "Test Repo",
	}
	if err := db.Create(&repo).Error; err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}

	// Create initial review result
	reviewResult := model.ReviewResult{
		RepositoryID:   repo.ID,
		MergeRequestID: 123,
		Status:         "processing",
	}
	if err := db.Create(&reviewResult).Error; err != nil {
		t.Fatalf("Failed to create review result: %v", err)
	}

	// Create response with no suggestions
	response := &llm.ReviewResponse{
		Summary:     "Perfect code, no issues found",
		Score:       100,
		RawResponse: "detailed response",
		Suggestions: []llm.FixSuggestion{},
	}

	// Save review result
	err := storage.SaveReviewResult(&reviewResult, response)
	if err != nil {
		t.Fatalf("SaveReviewResult failed: %v", err)
	}

	// Verify statistics are all zero
	var updatedReview model.ReviewResult
	if err := db.First(&updatedReview, reviewResult.ID).Error; err != nil {
		t.Fatalf("Failed to fetch updated review: %v", err)
	}

	if updatedReview.IssuesFound != 0 {
		t.Errorf("Expected IssuesFound 0, got %d", updatedReview.IssuesFound)
	}
	if updatedReview.CriticalIssuesCount != 0 {
		t.Errorf("Expected CriticalIssuesCount 0, got %d", updatedReview.CriticalIssuesCount)
	}
}

func BenchmarkSaveReviewResult(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.ReviewResult{}, &model.FixSuggestion{}, &model.Repository{})

	storage := NewReviewStorageService(db)

	// Create test repository
	repo := model.Repository{Name: "Test Repo"}
	db.Create(&repo)

	// Create many suggestions for stress testing
	suggestions := make([]llm.FixSuggestion, 100)
	for i := 0; i < 100; i++ {
		suggestions[i] = llm.FixSuggestion{
			FilePath:    "test.go",
			LineStart:   i,
			LineEnd:     i + 5,
			Severity:    "medium",
			Category:    "code-quality",
			Description: "Test issue",
			Suggestion:  "Fix it",
			CodeSnippet: "code",
		}
	}

	response := &llm.ReviewResponse{
		Summary:     "Test summary",
		Score:       75,
		RawResponse: "raw",
		Suggestions: suggestions,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reviewResult := model.ReviewResult{
			RepositoryID:   repo.ID,
			MergeRequestID: int64(i),
			Status:         "processing",
		}
		db.Create(&reviewResult)
		storage.SaveReviewResult(&reviewResult, response)
	}
}
