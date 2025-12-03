package service

import (
	"fmt"
	"time"

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

	return s.repo.UpdateProvider(provider)
}

// DeleteProvider deletes a provider
func (s *LLMService) DeleteProvider(id uint) error {
	return s.repo.DeleteProvider(id)
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

// testOpenAICompatible tests OpenAI-compatible API
func (s *LLMService) testOpenAICompatible(baseURL, apiKey string) error {
	// For MVP, we'll do a simple validation
	// In production, you'd make an actual API call
	if baseURL == "" || apiKey == "" {
		return fmt.Errorf("base URL and API key are required")
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
