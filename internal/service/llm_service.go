package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/handsoff/handsoff/internal/llm"
	"github.com/handsoff/handsoff/internal/model"
	"github.com/handsoff/handsoff/internal/repository"
	"github.com/handsoff/handsoff/pkg/config"
	"github.com/handsoff/handsoff/pkg/crypto"
)

// LLMService handles LLM provider and model business logic
type LLMService struct {
	repo      *repository.LLMRepository
	encryptor *crypto.Encryptor
}

// NewLLMService creates a new LLM service
func NewLLMService(repo *repository.LLMRepository, cfg *config.Config) (*LLMService, error) {
	encryptor, err := crypto.NewEncryptor(cfg.Security.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	return &LLMService{
		repo:      repo,
		encryptor: encryptor,
	}, nil
}

// Provider operations

// ListProviders returns all providers with masked API keys for a specific project
func (s *LLMService) ListProviders(projectID uint) ([]model.LLMProvider, error) {
	providers, err := s.repo.ListProviders(projectID)
	if err != nil {
		return nil, err
	}

	// Mask API keys
	for i := range providers {
		if providers[i].APIKey != "" {
			providers[i].APIKey = "***masked***"
		}
	}

	return providers, nil
}

// GetProvider retrieves a provider with masked API key with project validation
func (s *LLMService) GetProvider(id uint, projectID uint) (*model.LLMProvider, error) {
	provider, err := s.repo.GetProvider(id, projectID)
	if err != nil {
		return nil, err
	}

	// Mask API key
	if provider.APIKey != "" {
		provider.APIKey = "***masked***"
	}

	return provider, nil
}

// CreateProvider creates a new provider with encrypted API key
func (s *LLMService) CreateProvider(provider *model.LLMProvider) error {
	// Encrypt API key
	if provider.APIKey != "" {
		encryptedKey, err := s.encryptor.Encrypt(provider.APIKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt API key: %w", err)
		}
		provider.APIKey = encryptedKey
	}

	return s.repo.CreateProvider(provider)
}

// UpdateProvider updates a provider with partial update support
func (s *LLMService) UpdateProvider(provider *model.LLMProvider) error {
	// Get existing provider to merge with updates
	existing, err := s.repo.GetProviderByID(provider.ID)
	if err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}

	// Merge non-empty fields from request to existing provider
	if provider.Name != "" {
		existing.Name = provider.Name
	}
	if provider.BaseURL != "" {
		existing.BaseURL = provider.BaseURL
	}
	if provider.Model != "" {
		existing.Model = provider.Model
	}

	// Handle API key: only update if a new key is provided
	if provider.APIKey != "" && provider.APIKey != "***masked***" {
		encryptedKey, err := s.encryptor.Encrypt(provider.APIKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt API key: %w", err)
		}
		existing.APIKey = encryptedKey
	}
	// If APIKey is empty or masked, keep the existing encrypted key (no change)

	// Update IsActive (always update since it's a boolean with default value)
	existing.IsActive = provider.IsActive

	// Save the merged provider
	if err := s.repo.UpdateProvider(existing); err != nil {
		return err
	}

	// Invalidate cached clients for this provider
	// This ensures subsequent requests use the updated configuration
	llm.InvalidateProvider(provider.ID)

	return nil
}

// DeleteProvider deletes a provider
func (s *LLMService) DeleteProvider(id uint) error {
	if err := s.repo.DeleteProvider(id); err != nil {
		return err
	}

	// Invalidate cached clients for this provider
	llm.InvalidateProvider(id)

	return nil
}

// TestProviderConnection tests the LLM provider connection using stored model
func (s *LLMService) TestProviderConnection(id uint, projectID uint) error {
	provider, err := s.repo.GetProvider(id, projectID)
	if err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}

	// Decrypt API key
	decryptedKey, err := s.encryptor.Decrypt(provider.APIKey)
	if err != nil {
		s.repo.UpdateProviderTestStatus(id, "failed", "Failed to decrypt API key")
		return fmt.Errorf("failed to decrypt API key: %w", err)
	}

	// Test using the model stored in Provider table (not hardcoded)
	testErr := s.TestModelConnection(provider.BaseURL, decryptedKey, provider.Model)
	if testErr != nil {
		s.repo.UpdateProviderTestStatus(id, "failed", testErr.Error())
		return testErr
	}

	// Update test status
	now := time.Now()
	provider.LastTestedAt = &now
	provider.LastTestStatus = "success"
	provider.LastTestMessage = fmt.Sprintf("Model %s test successful", provider.Model)

	if err := s.repo.UpdateProvider(provider); err != nil {
		return fmt.Errorf("failed to update test status: %w", err)
	}

	return nil
}

// FetchAvailableModels fetches the list of available models from LLM provider
func (s *LLMService) FetchAvailableModels(baseURL, apiKey string) ([]string, error) {
	// Validate parameters
	if baseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("base URL and API key are required")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// Create HTTP request to /models
	req, err := http.NewRequest("GET", baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	switch resp.StatusCode {
	case http.StatusOK:
		// Parse response
		var result struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("invalid API response: %w", err)
		}

		// Extract model IDs
		models := make([]string, 0, len(result.Data))
		for _, model := range result.Data {
			if model.ID != "" {
				models = append(models, model.ID)
			}
		}

		if len(models) == 0 {
			return nil, fmt.Errorf("no models found in response")
		}

		return models, nil

	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, fmt.Errorf("authentication failed: invalid API key or insufficient permissions")

	case http.StatusNotFound:
		return nil, fmt.Errorf("endpoint not found: this provider may not support the /models interface")

	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return nil, fmt.Errorf("provider service error (status %d): %s", resp.StatusCode, string(body))

	default:
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
}

// FetchModelsForProvider fetches available models using stored provider configuration
func (s *LLMService) FetchModelsForProvider(providerID uint, projectID uint) ([]string, error) {
	// Get provider from database
	provider, err := s.repo.GetProvider(providerID, projectID)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Decrypt API key
	decryptedKey, err := s.encryptor.Decrypt(provider.APIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt API key: %w", err)
	}

	// Use existing FetchAvailableModels logic
	return s.FetchAvailableModels(provider.BaseURL, decryptedKey)
}

// TestModelConnection tests a specific model with temporary or stored credentials
func (s *LLMService) TestModelConnection(baseURL, apiKey, model string) error {
	// Validate parameters
	if baseURL == "" || apiKey == "" || model == "" {
		return fmt.Errorf("base URL, API key, and model are required")
	}

	// Create a minimal test request (5 tokens to minimize cost)
	testRequest := map[string]interface{}{
		"model": model, // Use user-specified model
		"messages": []map[string]string{
			{"role": "user", "content": "test"},
		},
		"max_tokens": 5,
	}

	// Marshal request body
	reqBody, err := json.Marshal(testRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal test request: %w", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("model test failed (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response to verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("invalid API response: %w", err)
	}

	return nil
}
