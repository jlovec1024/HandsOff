package service

import (
	"fmt"
	"time"

	"github.com/handsoff/handsoff/internal/model"
	"github.com/handsoff/handsoff/internal/repository"
	"github.com/handsoff/handsoff/pkg/config"
	"github.com/handsoff/handsoff/pkg/crypto"
	"github.com/xanzy/go-gitlab"
)

// PlatformService handles Git platform business logic
type PlatformService struct {
	repo      *repository.PlatformRepository
	encryptor *crypto.Encryptor
}

// NewPlatformService creates a new platform service
func NewPlatformService(repo *repository.PlatformRepository, cfg *config.Config) (*PlatformService, error) {
	encryptor, err := crypto.NewEncryptor(cfg.Security.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	return &PlatformService{
		repo:      repo,
		encryptor: encryptor,
	}, nil
}

// GetConfig retrieves the platform config with decrypted token for a specific project
func (s *PlatformService) GetConfig(projectID uint) (*model.GitPlatformConfig, error) {
	config, err := s.repo.GetConfig(projectID)
	if err != nil {
		return nil, err
	}

	// Decrypt access token for display (masked)
	// Note: We don't return the actual token, just indicate it exists
	if config.AccessToken != "" {
		config.AccessToken = "***masked***"
	}

	return config, nil
}

// CreateOrUpdateConfig creates or updates platform config with encrypted token for a specific project
func (s *PlatformService) CreateOrUpdateConfig(config *model.GitPlatformConfig) error {
	// Handle access token based on value:
	// - Non-empty string: encrypt and save new token
	// - Empty string: keep existing token (don't update)
	if config.AccessToken != "" {
		// New token provided - encrypt it
		encryptedToken, err := s.encryptor.Encrypt(config.AccessToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt access token: %w", err)
		}
		config.AccessToken = encryptedToken
	} else {
		// Empty string means "keep existing token"
		existing, err := s.repo.GetConfig(config.ProjectID)
		if err == nil {
			// Preserve existing token
			config.AccessToken = existing.AccessToken
		} else {
			// No existing config - require token for new config
			return fmt.Errorf("access token is required for new configuration")
		}
	}

	return s.repo.CreateOrUpdateConfig(config)
}

// TestConnectionWithConfig tests GitLab connection with provided configuration (without saving)
// This allows users to validate configuration before saving
func (s *PlatformService) TestConnectionWithConfig(baseURL, accessToken string) (string, error) {
	// Create GitLab client with provided credentials
	git, err := gitlab.NewClient(accessToken, gitlab.WithBaseURL(baseURL))
	if err != nil {
		return "", fmt.Errorf("failed to create GitLab client: %w", err)
	}

	// Test connection by getting current user
	user, _, err := git.Users.CurrentUser()
	if err != nil {
		return "", fmt.Errorf("failed to connect to GitLab: %w", err)
	}

	// Return success message
	message := fmt.Sprintf("Connected successfully as %s (@%s)", user.Name, user.Username)
	return message, nil
}

// TestConnection tests the GitLab connection for a specific project (using saved config)
func (s *PlatformService) TestConnection(projectID uint, configID uint) error {
	config, err := s.repo.GetConfig(projectID)
	if err != nil {
		return fmt.Errorf("config not found: %w", err)
	}

	// Decrypt access token
	decryptedToken, err := s.encryptor.Decrypt(config.AccessToken)
	if err != nil {
		s.repo.UpdateTestStatus(configID, "failed", "Failed to decrypt access token")
		return fmt.Errorf("failed to decrypt token: %w", err)
	}

	// Create GitLab client
	git, err := gitlab.NewClient(decryptedToken, gitlab.WithBaseURL(config.BaseURL))
	if err != nil {
		s.repo.UpdateTestStatus(configID, "failed", err.Error())
		return fmt.Errorf("failed to create GitLab client: %w", err)
	}

	// Test connection by getting current user
	user, _, err := git.Users.CurrentUser()
	if err != nil {
		s.repo.UpdateTestStatus(configID, "failed", err.Error())
		return fmt.Errorf("failed to connect to GitLab: %w", err)
	}

	// Update test status
	message := fmt.Sprintf("Connected successfully as %s (@%s)", user.Name, user.Username)
	now := time.Now()
	config.LastTestedAt = &now
	config.LastTestStatus = "success"
	config.LastTestMessage = message

	if err := s.repo.CreateOrUpdateConfig(config); err != nil {
		return fmt.Errorf("failed to update test status: %w", err)
	}

	return nil
}
