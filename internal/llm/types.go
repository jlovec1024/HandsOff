package llm

import "time"

// ReviewRequest represents a code review request
type ReviewRequest struct {
	Diff         string  // Git diff content
	Prompt       string  // Rendered prompt template
	MaxTokens    int     // Maximum tokens for response
	Temperature  float32 // Sampling temperature
	ModelName    string  // Model identifier
}

// ReviewResponse represents LLM review response
type ReviewResponse struct {
	Summary     string           `json:"summary"`      // Overall review summary
	Score       int              `json:"score"`        // Quality score 0-100
	Suggestions []FixSuggestion  `json:"suggestions"`  // List of fix suggestions
	RawResponse string           `json:"-"`            // Original LLM response
	ModelUsed   string           `json:"model_used"`   // Model that generated this
	TokensUsed  int              `json:"tokens_used"`  // Tokens consumed (deprecated, use TokenUsage)
	Duration    time.Duration    `json:"duration"`     // Time taken

	// Detailed Token Usage (for operations analytics)
	TokenUsage TokenUsage `json:"token_usage"`
}

// TokenUsage represents detailed token consumption from LLM API
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// FixSuggestion represents a single code fix suggestion
type FixSuggestion struct {
	FilePath    string `json:"file_path"`    // File path
	LineStart   int    `json:"line_start"`   // Starting line number
	LineEnd     int    `json:"line_end"`     // Ending line number
	Severity    string `json:"severity"`     // high, medium, low
	Category    string `json:"category"`     // security, performance, style, etc.
	Description string `json:"description"`  // Issue description
	Suggestion  string `json:"suggestion"`   // Fix recommendation
	CodeSnippet string `json:"code_snippet"` // Original code snippet
}

// Client interface for LLM providers
type Client interface {
	// Review performs code review using LLM
	Review(req ReviewRequest) (*ReviewResponse, error)
	
	// TestConnection tests the LLM API connection
	TestConnection() error
	
	// GetProviderName returns the provider name
	GetProviderName() string
}

// Config holds LLM client configuration
type Config struct {
	BaseURL     string
	APIKey      string
	ModelName   string
	MaxTokens   int
	Temperature float32
	Timeout     time.Duration
}
