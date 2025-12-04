package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAICompatibleClient implements Client interface for OpenAI-compatible APIs
// This includes OpenAI, DeepSeek, and other providers that follow the OpenAI API specification
type OpenAICompatibleClient struct {
	providerName string
	config       Config
	client       *http.Client
}

// NewOpenAICompatibleClient creates a new OpenAI-compatible client
func NewOpenAICompatibleClient(providerName string, config Config) *OpenAICompatibleClient {
	return &OpenAICompatibleClient{
		providerName: providerName,
		config:       config,
		client: &http.Client{
			Timeout: config.Timeout * time.Second,
		},
	}
}

// OpenAI-compatible API request/response structures
type compatibleRequest struct {
	Model       string              `json:"model"`
	Messages    []compatibleMessage `json:"messages"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Temperature float32             `json:"temperature,omitempty"`
}

type compatibleMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
}

type compatibleResponse struct {
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
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// Review performs code review using OpenAI-compatible API
func (c *OpenAICompatibleClient) Review(req ReviewRequest) (*ReviewResponse, error) {
	start := time.Now()

	// Construct request
	apiReq := compatibleRequest{
		Model: c.config.ModelName,
		Messages: []compatibleMessage{
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
	reqBody, err := json.Marshal(apiReq)
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
	var apiResp compatibleResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if apiResp.Error != nil {
		return nil, fmt.Errorf("%s API error: %s (type: %s)", c.providerName, apiResp.Error.Message, apiResp.Error.Type)
	}

	// Check response validity
	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Extract content
	content := apiResp.Choices[0].Message.Content

	// Parse structured review response
	reviewResp, err := parseReviewResponse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse review response: %w", err)
	}

	// Fill metadata
	reviewResp.RawResponse = content
	reviewResp.ModelUsed = apiResp.Model
	reviewResp.TokensUsed = apiResp.Usage.TotalTokens
	reviewResp.Duration = time.Since(start)

	return reviewResp, nil
}

// TestConnection tests API connectivity
func (c *OpenAICompatibleClient) TestConnection() error {
	req := compatibleRequest{
		Model: c.config.ModelName,
		Messages: []compatibleMessage{
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
func (c *OpenAICompatibleClient) GetProviderName() string {
	return c.providerName
}
