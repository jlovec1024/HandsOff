package llm

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGetQualityLevel(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		// Standard thresholds
		{95, "excellent"},
		{90, "excellent"},
		{89, "good"},
		{80, "good"},
		{75, "good"},
		{74, "acceptable"},
		{65, "acceptable"},
		{60, "acceptable"},
		{59, "poor"},
		{45, "poor"},
		{40, "poor"},
		{39, "critical"},
		{30, "critical"},
		{0, "critical"},
		// Edge cases
		{100, "excellent"},
		{-1, "critical"},
	}

	for _, tt := range tests {
		result := GetQualityLevel(tt.score)
		if result != tt.expected {
			t.Errorf("GetQualityLevel(%d) = %q, want %q", tt.score, result, tt.expected)
		}
	}
}

func TestCalculateOutputStatistics_Empty(t *testing.T) {
	stats := CalculateOutputStatistics([]FixSuggestion{})

	if stats.TotalIssues != 0 {
		t.Errorf("TotalIssues = %d, want 0", stats.TotalIssues)
	}
	if stats.FilesAffected != 0 {
		t.Errorf("FilesAffected = %d, want 0", stats.FilesAffected)
	}
	if stats.BySeverity.Critical != 0 || stats.BySeverity.High != 0 ||
		stats.BySeverity.Medium != 0 || stats.BySeverity.Low != 0 {
		t.Errorf("BySeverity should all be 0, got %+v", stats.BySeverity)
	}
	if stats.ByCategory.Security != 0 || stats.ByCategory.Performance != 0 ||
		stats.ByCategory.Style != 0 || stats.ByCategory.Logic != 0 ||
		stats.ByCategory.Documentation != 0 || stats.ByCategory.Other != 0 {
		t.Errorf("ByCategory should all be 0, got %+v", stats.ByCategory)
	}
	if len(stats.TopFiles) != 0 {
		t.Errorf("TopFiles should be empty, got %v", stats.TopFiles)
	}
}

func TestCalculateOutputStatistics_WithSuggestions(t *testing.T) {
	suggestions := []FixSuggestion{
		{FilePath: "main.go", Severity: "critical", Category: "security"},
		{FilePath: "main.go", Severity: "high", Category: "performance"},
		{FilePath: "main.go", Severity: "high", Category: "logic"},
		{FilePath: "utils.go", Severity: "medium", Category: "style"},
		{FilePath: "utils.go", Severity: "low", Category: "documentation"},
		{FilePath: "config.go", Severity: "medium", Category: "other"},
		{FilePath: "", Severity: "low", Category: "style"},         // Empty file path
		{FilePath: "unknown", Severity: "medium", Category: "logic"}, // "unknown" file path
	}

	stats := CalculateOutputStatistics(suggestions)

	// Total issues
	if stats.TotalIssues != 8 {
		t.Errorf("TotalIssues = %d, want 8", stats.TotalIssues)
	}

	// By severity
	if stats.BySeverity.Critical != 1 {
		t.Errorf("BySeverity.Critical = %d, want 1", stats.BySeverity.Critical)
	}
	if stats.BySeverity.High != 2 {
		t.Errorf("BySeverity.High = %d, want 2", stats.BySeverity.High)
	}
	if stats.BySeverity.Medium != 3 {
		t.Errorf("BySeverity.Medium = %d, want 3", stats.BySeverity.Medium)
	}
	if stats.BySeverity.Low != 2 {
		t.Errorf("BySeverity.Low = %d, want 2", stats.BySeverity.Low)
	}

	// By category
	if stats.ByCategory.Security != 1 {
		t.Errorf("ByCategory.Security = %d, want 1", stats.ByCategory.Security)
	}
	if stats.ByCategory.Performance != 1 {
		t.Errorf("ByCategory.Performance = %d, want 1", stats.ByCategory.Performance)
	}
	if stats.ByCategory.Style != 2 {
		t.Errorf("ByCategory.Style = %d, want 2", stats.ByCategory.Style)
	}
	if stats.ByCategory.Logic != 2 {
		t.Errorf("ByCategory.Logic = %d, want 2", stats.ByCategory.Logic)
	}
	if stats.ByCategory.Documentation != 1 {
		t.Errorf("ByCategory.Documentation = %d, want 1", stats.ByCategory.Documentation)
	}
	if stats.ByCategory.Other != 1 {
		t.Errorf("ByCategory.Other = %d, want 1", stats.ByCategory.Other)
	}

	// Files affected (empty and "unknown" should be excluded)
	if stats.FilesAffected != 3 {
		t.Errorf("FilesAffected = %d, want 3", stats.FilesAffected)
	}

	// TopFiles should be sorted by issue count descending
	if len(stats.TopFiles) != 3 {
		t.Errorf("len(TopFiles) = %d, want 3", len(stats.TopFiles))
	}
	if len(stats.TopFiles) > 0 && stats.TopFiles[0].File != "main.go" {
		t.Errorf("TopFiles[0].File = %q, want \"main.go\"", stats.TopFiles[0].File)
	}
	if len(stats.TopFiles) > 0 && stats.TopFiles[0].Issues != 3 {
		t.Errorf("TopFiles[0].Issues = %d, want 3", stats.TopFiles[0].Issues)
	}
}

func TestCalculateOutputStatistics_UnknownSeverityCategory(t *testing.T) {
	suggestions := []FixSuggestion{
		{FilePath: "test.go", Severity: "unknown_severity", Category: "unknown_category"},
		{FilePath: "test.go", Severity: "", Category: ""},
	}

	stats := CalculateOutputStatistics(suggestions)

	// Unknown severity defaults to medium
	if stats.BySeverity.Medium != 2 {
		t.Errorf("BySeverity.Medium = %d, want 2 (unknown defaults to medium)", stats.BySeverity.Medium)
	}

	// Unknown category defaults to other
	if stats.ByCategory.Other != 2 {
		t.Errorf("ByCategory.Other = %d, want 2 (unknown defaults to other)", stats.ByCategory.Other)
	}
}

func TestGetTopFiles(t *testing.T) {
	t.Run("MoreThan5Files", func(t *testing.T) {
		fileCount := map[string]int{
			"a.go": 10,
			"b.go": 8,
			"c.go": 6,
			"d.go": 4,
			"e.go": 2,
			"f.go": 1,
			"g.go": 3,
		}

		result := getTopFiles(fileCount, 5)

		if len(result) != 5 {
			t.Errorf("len(result) = %d, want 5", len(result))
		}

		// Verify order (descending by issues)
		expectedOrder := []struct {
			file   string
			issues int
		}{
			{"a.go", 10},
			{"b.go", 8},
			{"c.go", 6},
			{"d.go", 4},
			{"g.go", 3},
		}

		for i, expected := range expectedOrder {
			if result[i].File != expected.file || result[i].Issues != expected.issues {
				t.Errorf("result[%d] = {%q, %d}, want {%q, %d}",
					i, result[i].File, result[i].Issues, expected.file, expected.issues)
			}
		}
	})

	t.Run("LessThan5Files", func(t *testing.T) {
		fileCount := map[string]int{
			"a.go": 5,
			"b.go": 3,
			"c.go": 1,
		}

		result := getTopFiles(fileCount, 5)

		if len(result) != 3 {
			t.Errorf("len(result) = %d, want 3", len(result))
		}

		// Verify descending order
		if result[0].Issues < result[1].Issues || result[1].Issues < result[2].Issues {
			t.Errorf("result not in descending order: %v", result)
		}
	})

	t.Run("EmptyMap", func(t *testing.T) {
		result := getTopFiles(map[string]int{}, 5)

		if len(result) != 0 {
			t.Errorf("len(result) = %d, want 0", len(result))
		}
	})
}

func TestConvertSuggestionsToOutput(t *testing.T) {
	suggestions := []FixSuggestion{
		{
			FilePath:    "main.go",
			LineStart:   10,
			LineEnd:     15,
			Severity:    "Major",      // Should be normalized to "high"
			Category:    "sec",        // Should be normalized to "security"
			Description: "Security issue",
			Suggestion:  "Fix this",
			CodeSnippet: "func foo() {}",
		},
		{
			FilePath:    "utils.go",
			LineStart:   20,
			LineEnd:     20,
			Severity:    "blocker",    // Should be normalized to "critical"
			Category:    "perf",       // Should be normalized to "performance"
			Description: "Performance issue",
			Suggestion:  "Optimize this",
			CodeSnippet: "",
		},
	}

	result := ConvertSuggestionsToOutput(suggestions)

	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}

	// Check first suggestion
	if result[0].ID != 1 {
		t.Errorf("result[0].ID = %d, want 1", result[0].ID)
	}
	if result[0].FilePath != "main.go" {
		t.Errorf("result[0].FilePath = %q, want \"main.go\"", result[0].FilePath)
	}
	if result[0].LineStart != 10 {
		t.Errorf("result[0].LineStart = %d, want 10", result[0].LineStart)
	}
	if result[0].LineEnd != 15 {
		t.Errorf("result[0].LineEnd = %d, want 15", result[0].LineEnd)
	}
	if result[0].Severity != "high" {
		t.Errorf("result[0].Severity = %q, want \"high\"", result[0].Severity)
	}
	if result[0].Category != "security" {
		t.Errorf("result[0].Category = %q, want \"security\"", result[0].Category)
	}
	if result[0].Description != "Security issue" {
		t.Errorf("result[0].Description = %q, want \"Security issue\"", result[0].Description)
	}
	if result[0].Suggestion != "Fix this" {
		t.Errorf("result[0].Suggestion = %q, want \"Fix this\"", result[0].Suggestion)
	}
	if result[0].CodeSnippet != "func foo() {}" {
		t.Errorf("result[0].CodeSnippet = %q, want \"func foo() {}\"", result[0].CodeSnippet)
	}

	// Check second suggestion
	if result[1].ID != 2 {
		t.Errorf("result[1].ID = %d, want 2", result[1].ID)
	}
	if result[1].Severity != "critical" {
		t.Errorf("result[1].Severity = %q, want \"critical\"", result[1].Severity)
	}
	if result[1].Category != "performance" {
		t.Errorf("result[1].Category = %q, want \"performance\"", result[1].Category)
	}
}

func TestConvertSuggestionsToOutput_Empty(t *testing.T) {
	result := ConvertSuggestionsToOutput([]FixSuggestion{})

	if len(result) != 0 {
		t.Errorf("len(result) = %d, want 0", len(result))
	}
}

func TestFormatReviewAsJSON(t *testing.T) {
	resp := &ReviewResponse{
		Summary: "Code review complete",
		Score:   85,
		Suggestions: []FixSuggestion{
			{
				FilePath:    "main.go",
				LineStart:   10,
				LineEnd:     15,
				Severity:    "high",
				Category:    "security",
				Description: "Security issue",
				Suggestion:  "Fix this",
			},
		},
	}

	ctx := OutputContext{
		Repository: ContextRepository{
			ID:       1,
			Name:     "test-repo",
			FullName: "org/test-repo",
			Platform: "gitlab",
		},
		MergeRequest: ContextMergeRequest{
			ID:           100,
			IID:          1,
			Title:        "Test MR",
			Author:       "testuser",
			SourceBranch: "feature",
			TargetBranch: "main",
		},
		Review: ContextReview{
			ID:          1,
			ReviewedAt:  time.Now(),
			LLMProvider: "openai",
			LLMModel:    "gpt-4",
			TokensUsed:  1000,
			DurationMs:  500,
		},
	}

	meta := OutputMetadata{
		PromptTemplate:     "default",
		PromptVersion:      "1.0",
		CustomPromptUsed:   false,
		RawResponseAvail:   true,
		ParserFallbackUsed: false,
	}

	jsonStr, err := FormatReviewAsJSON(resp, ctx, meta)
	if err != nil {
		t.Fatalf("FormatReviewAsJSON() error = %v", err)
	}

	// Verify JSON is valid
	var output ReviewOutputJSON
	if err := json.Unmarshal([]byte(jsonStr), &output); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify schema_version
	if output.SchemaVersion != SchemaVersion {
		t.Errorf("SchemaVersion = %q, want %q", output.SchemaVersion, SchemaVersion)
	}

	// Verify context
	if output.Context.Repository.Name != "test-repo" {
		t.Errorf("Context.Repository.Name = %q, want \"test-repo\"", output.Context.Repository.Name)
	}
	if output.Context.MergeRequest.Title != "Test MR" {
		t.Errorf("Context.MergeRequest.Title = %q, want \"Test MR\"", output.Context.MergeRequest.Title)
	}

	// Verify result
	if output.Result.Summary != "Code review complete" {
		t.Errorf("Result.Summary = %q, want \"Code review complete\"", output.Result.Summary)
	}
	if output.Result.Score != 85 {
		t.Errorf("Result.Score = %d, want 85", output.Result.Score)
	}
	if output.Result.QualityLevel != "good" {
		t.Errorf("Result.QualityLevel = %q, want \"good\"", output.Result.QualityLevel)
	}
	if len(output.Result.Suggestions) != 1 {
		t.Errorf("len(Result.Suggestions) = %d, want 1", len(output.Result.Suggestions))
	}

	// Verify statistics
	if output.Statistics.TotalIssues != 1 {
		t.Errorf("Statistics.TotalIssues = %d, want 1", output.Statistics.TotalIssues)
	}
	if output.Statistics.FilesAffected != 1 {
		t.Errorf("Statistics.FilesAffected = %d, want 1", output.Statistics.FilesAffected)
	}

	// Verify metadata
	if output.Metadata.PromptTemplate != "default" {
		t.Errorf("Metadata.PromptTemplate = %q, want \"default\"", output.Metadata.PromptTemplate)
	}
	if output.Metadata.CustomPromptUsed != false {
		t.Errorf("Metadata.CustomPromptUsed = %v, want false", output.Metadata.CustomPromptUsed)
	}
}

func TestFormatReviewAsJSON_EmptySuggestions(t *testing.T) {
	resp := &ReviewResponse{
		Summary:     "Perfect code",
		Score:       100,
		Suggestions: []FixSuggestion{},
	}

	ctx := OutputContext{}
	meta := OutputMetadata{}

	jsonStr, err := FormatReviewAsJSON(resp, ctx, meta)
	if err != nil {
		t.Fatalf("FormatReviewAsJSON() error = %v", err)
	}

	var output ReviewOutputJSON
	if err := json.Unmarshal([]byte(jsonStr), &output); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if output.Result.QualityLevel != "excellent" {
		t.Errorf("Result.QualityLevel = %q, want \"excellent\"", output.Result.QualityLevel)
	}
	if output.Statistics.TotalIssues != 0 {
		t.Errorf("Statistics.TotalIssues = %d, want 0", output.Statistics.TotalIssues)
	}
	if len(output.Result.Suggestions) != 0 {
		t.Errorf("len(Result.Suggestions) = %d, want 0", len(output.Result.Suggestions))
	}
}

func TestFormatReviewAsJSON_GeneratedAt(t *testing.T) {
	resp := &ReviewResponse{
		Summary: "Test",
		Score:   50,
	}

	before := time.Now()
	jsonStr, err := FormatReviewAsJSON(resp, OutputContext{}, OutputMetadata{})
	after := time.Now()

	if err != nil {
		t.Fatalf("FormatReviewAsJSON() error = %v", err)
	}

	var output ReviewOutputJSON
	if err := json.Unmarshal([]byte(jsonStr), &output); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if output.GeneratedAt.Before(before) || output.GeneratedAt.After(after) {
		t.Errorf("GeneratedAt = %v, should be between %v and %v", output.GeneratedAt, before, after)
	}
}
