package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DeepSeekClient implements Client interface for DeepSeek API
// DeepSeek API is OpenAI-compatible, so we reuse similar structures
type DeepSeekClient struct {
	config Config
	client *http.Client
}

// NewDeepSeekClient creates a new DeepSeek client
func NewDeepSeekClient(config Config) *DeepSeekClient {
	return &DeepSeekClient{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout * time.Second,
		},
	}
}

// DeepSeek API request/response structures (OpenAI-compatible)
type deepSeekRequest struct {
	Model       string             `json:"model"`
	Messages    []deepSeekMessage  `json:"messages"`
	MaxTokens   int                `json:"max_tokens,omitempty"`
	Temperature float32            `json:"temperature,omitempty"`
}

type deepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepSeekResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// Review performs code review using DeepSeek API
func (c *DeepSeekClient) Review(req ReviewRequest) (*ReviewResponse, error) {
	start := time.Now()

	// Construct request
	deepseekReq := deepSeekRequest{
		Model: c.config.ModelName,
		Messages: []deepSeekMessage{
			{
				Role:    "system",
				Content: "You are an expert code reviewer. Analyze the code changes and provide structured feedback in JSON format.",
			},
			{
				Role:    "user",
				Content: req.Prompt,
			},
		},
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
	}

	// Marshal request
	reqBody, err := json.Marshal(deepseekReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	// Send request
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var deepseekResp deepSeekResponse
	if err := json.Unmarshal(body, &deepseekResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if deepseekResp.Error != nil {
		return nil, fmt.Errorf("DeepSeek API error: %s (type: %s)", deepseekResp.Error.Message, deepseekResp.Error.Type)
	}

	// Check response validity
	if len(deepseekResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Extract content
	content := deepseekResp.Choices[0].Message.Content

	// Parse structured review response
	reviewResp, err := parseReviewResponse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse review response: %w", err)
	}

	// Fill metadata
	reviewResp.RawResponse = content
	reviewResp.ModelUsed = deepseekResp.Model
	reviewResp.TokensUsed = deepseekResp.Usage.TotalTokens
	reviewResp.Duration = time.Since(start)

	return reviewResp, nil
}

// TestConnection tests DeepSeek API connectivity
func (c *DeepSeekClient) TestConnection() error {
	req := deepSeekRequest{
		Model: c.config.ModelName,
		Messages: []deepSeekMessage{
			{
				Role:    "user",
				Content: "Hello, this is a test message.",
			},
		},
		MaxTokens: 10,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetProviderName returns the provider name
func (c *DeepSeekClient) GetProviderName() string {
	return "DeepSeek"
}
