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

// UpdateProvider updates a provider
func (s *LLMService) UpdateProvider(provider *model.LLMProvider) error {
	// Handle API key encryption
	if provider.APIKey != "" && provider.APIKey != "***masked***" {
		encryptedKey, err := s.encryptor.Encrypt(provider.APIKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt API key: %w", err)
		}
		provider.APIKey = encryptedKey
	} else if provider.APIKey == "***masked***" {
		// Keep existing key
		existing, err := s.repo.GetProviderByID(provider.ID)
		if err == nil {
			provider.APIKey = existing.APIKey
		}
	}

	if err := s.repo.UpdateProvider(provider); err != nil {
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

// TestProviderConnection tests the LLM provider connection
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

	// Test connection based on provider type
	var testErr error
	switch provider.Type {
	case "openai", "deepseek":
		testErr = s.testOpenAICompatible(provider.BaseURL, decryptedKey)
	default:
		testErr = fmt.Errorf("unsupported provider type: %s", provider.Type)
	}

	if testErr != nil {
		s.repo.UpdateProviderTestStatus(id, "failed", testErr.Error())
		return testErr
	}

	// Update test status
	now := time.Now()
	provider.LastTestedAt = &now
	provider.LastTestStatus = "success"
	provider.LastTestMessage = "Connection test successful"

	if err := s.repo.UpdateProvider(provider); err != nil {
		return fmt.Errorf("failed to update test status: %w", err)
	}

	return nil
}

// testOpenAICompatible tests OpenAI-compatible API by making a real API call
func (s *LLMService) testOpenAICompatible(baseURL, apiKey string) error {
	// Validate parameters
	if baseURL == "" || apiKey == "" {
		return fmt.Errorf("base URL and API key are required")
	}

	// Create a minimal test request (5 tokens to minimize cost)
	testRequest := map[string]interface{}{
		"model": "gpt-3.5-turbo", // Default model, should work for most OpenAI-compatible APIs
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
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response to verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("invalid API response: %w", err)
	}

	return nil
}

// Model operations

// ListModels returns all models
func (s *LLMService) ListModels(providerID *uint) ([]model.LLMModel, error) {
	return s.repo.ListModels(providerID)
}

// GetModel retrieves a model
func (s *LLMService) GetModel(id uint) (*model.LLMModel, error) {
	return s.repo.GetModel(id)
}

// CreateModel creates a new model
func (s *LLMService) CreateModel(model *model.LLMModel) error {
	return s.repo.CreateModel(model)
}

// UpdateModel updates a model
func (s *LLMService) UpdateModel(model *model.LLMModel) error {
	return s.repo.UpdateModel(model)
}

// DeleteModel deletes a model
func (s *LLMService) DeleteModel(id uint) error {
	return s.repo.DeleteModel(id)
}
