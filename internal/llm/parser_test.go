package llm

import (
	"testing"
)

func TestNormalizeSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Standard values
		{"critical", "critical"},
		{"high", "high"},
		{"medium", "medium"},
		{"low", "low"},
		// Near-synonyms for critical
		{"blocker", "critical"},
		{"urgent", "critical"},
		{"fatal", "critical"},
		// Near-synonyms for high
		{"major", "high"},
		{"important", "high"},
		{"error", "high"},
		// Near-synonyms for medium
		{"moderate", "medium"},
		{"warning", "medium"},
		{"normal", "medium"},
		// Near-synonyms for low
		{"minor", "low"},
		{"trivial", "low"},
		{"info", "low"},
		{"suggestion", "low"},
		{"hint", "low"},
		// Case insensitivity
		{"CRITICAL", "critical"},
		{"High", "high"},
		{"MEDIUM", "medium"},
		{"LOW", "low"},
		// With whitespace
		{"  high  ", "high"},
		{"\tmedium\n", "medium"},
		// Unknown defaults to medium
		{"unknown", "medium"},
		{"", "medium"},
		{"something_else", "medium"},
	}

	for _, tt := range tests {
		result := NormalizeSeverity(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeSeverity(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeCategory(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Standard values
		{"security", "security"},
		{"performance", "performance"},
		{"style", "style"},
		{"logic", "logic"},
		{"documentation", "documentation"},
		{"other", "other"},
		// Near-synonyms for security
		{"sec", "security"},
		{"vulnerability", "security"},
		{"vuln", "security"},
		// Near-synonyms for performance
		{"perf", "performance"},
		{"efficiency", "performance"},
		{"optimization", "performance"},
		// Near-synonyms for style
		{"formatting", "style"},
		{"code style", "style"},
		{"lint", "style"},
		{"convention", "style"},
		// Near-synonyms for logic
		{"bug", "logic"},
		{"error", "logic"},
		{"correctness", "logic"},
		{"behavior", "logic"},
		// Near-synonyms for documentation
		{"doc", "documentation"},
		{"docs", "documentation"},
		{"comment", "documentation"},
		{"comments", "documentation"},
		// Case insensitivity
		{"SECURITY", "security"},
		{"Performance", "performance"},
		// With whitespace
		{"  style  ", "style"},
		// Unknown defaults to other
		{"unknown", "other"},
		{"", "other"},
		{"general", "other"},
	}

	for _, tt := range tests {
		result := NormalizeCategory(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeCategory(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestExtractJSONFromMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string // We check if the result contains this, since exact JSON may vary
	}{
		{
			name:     "Standard JSON block",
			input:    "```json\n{\"summary\": \"test\", \"score\": 75}\n```",
			contains: `"summary"`,
		},
		{
			name:     "JSON block without language hint",
			input:    "```\n{\"summary\": \"test\"}\n```",
			contains: `"summary"`,
		},
		{
			name:     "Raw JSON starting with {",
			input:    `{"summary": "test", "score": 80}`,
			contains: `"summary"`,
		},
		{
			name:     "JSON embedded in text",
			input:    "Here is the review:\n{\"summary\": \"test\"}\nEnd of review",
			contains: `"summary"`,
		},
		{
			name:     "Plain text (no JSON)",
			input:    "This is just plain text without any JSON",
			contains: "plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSONFromMarkdown(tt.input)
			if result == "" {
				t.Errorf("extractJSONFromMarkdown(%q) returned empty string", tt.input)
			}
			if tt.contains != "" && !contains(result, tt.contains) {
				t.Errorf("extractJSONFromMarkdown(%q) = %q, want to contain %q", tt.input, result, tt.contains)
			}
		})
	}
}

func TestParseReviewResponse(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantSummary   string
		wantScore     int
		wantSeverity  string // Expected severity for first suggestion (if any)
		wantCategory  string // Expected category for first suggestion (if any)
	}{
		{
			name: "Standard JSON response",
			input: `{
				"summary": "Code review complete",
				"score": 85,
				"suggestions": [
					{
						"severity": "Major",
						"category": "sec",
						"description": "Security issue found"
					}
				]
			}`,
			wantSummary:  "Code review complete",
			wantScore:    85,
			wantSeverity: "high",     // Major -> high
			wantCategory: "security", // sec -> security
		},
		{
			name: "JSON with non-standard severity",
			input: `{
				"summary": "Review done",
				"score": 70,
				"suggestions": [
					{
						"severity": "blocker",
						"category": "perf"
					}
				]
			}`,
			wantSummary:  "Review done",
			wantScore:    70,
			wantSeverity: "critical",    // blocker -> critical
			wantCategory: "performance", // perf -> performance
		},
		{
			name:        "Plain text fallback",
			input:       "Summary: This is a good code review.\nScore: 80\n- Minor style issue on line 10",
			wantSummary: "This is a good code review.",
			wantScore:   80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseReviewResponse(tt.input)
			if err != nil {
				t.Fatalf("parseReviewResponse() error = %v", err)
			}

			if result.Summary != tt.wantSummary && tt.wantSummary != "" {
				// Partial match is OK for text parsing
				if !contains(result.Summary, tt.wantSummary) {
					t.Errorf("Summary = %q, want %q", result.Summary, tt.wantSummary)
				}
			}

			if result.Score != tt.wantScore && tt.wantScore != 0 {
				t.Errorf("Score = %d, want %d", result.Score, tt.wantScore)
			}

			if tt.wantSeverity != "" && len(result.Suggestions) > 0 {
				if result.Suggestions[0].Severity != tt.wantSeverity {
					t.Errorf("Suggestions[0].Severity = %q, want %q", result.Suggestions[0].Severity, tt.wantSeverity)
				}
			}

			if tt.wantCategory != "" && len(result.Suggestions) > 0 {
				if result.Suggestions[0].Category != tt.wantCategory {
					t.Errorf("Suggestions[0].Category = %q, want %q", result.Suggestions[0].Category, tt.wantCategory)
				}
			}
		})
	}
}

func TestNormalizeReviewResponse(t *testing.T) {
	tests := []struct {
		name  string
		input *ReviewResponse
		check func(*ReviewResponse) bool
	}{
		{
			name: "Empty summary gets default",
			input: &ReviewResponse{
				Summary: "",
				Score:   50,
			},
			check: func(r *ReviewResponse) bool {
				return r.Summary == "No summary provided"
			},
		},
		{
			name: "Negative score clamps to 0",
			input: &ReviewResponse{
				Summary: "Test",
				Score:   -10,
			},
			check: func(r *ReviewResponse) bool {
				return r.Score == 0
			},
		},
		{
			name: "Score over 100 clamps to 100",
			input: &ReviewResponse{
				Summary: "Test",
				Score:   150,
			},
			check: func(r *ReviewResponse) bool {
				return r.Score == 100
			},
		},
		{
			name: "Empty file path gets 'unknown'",
			input: &ReviewResponse{
				Summary: "Test",
				Suggestions: []FixSuggestion{
					{FilePath: "", Severity: "high", Category: "logic"},
				},
			},
			check: func(r *ReviewResponse) bool {
				return r.Suggestions[0].FilePath == "unknown"
			},
		},
		{
			name: "LineEnd < LineStart gets corrected",
			input: &ReviewResponse{
				Summary: "Test",
				Suggestions: []FixSuggestion{
					{LineStart: 10, LineEnd: 5, Severity: "medium", Category: "style"},
				},
			},
			check: func(r *ReviewResponse) bool {
				return r.Suggestions[0].LineEnd == 10
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeReviewResponse(tt.input)
			if !tt.check(result) {
				t.Errorf("normalizeReviewResponse() did not pass check for %s", tt.name)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
