package repository

import (
	"github.com/handsoff/handsoff/internal/model"
	"gorm.io/gorm"
)

// LLMRepository handles LLM provider and model database operations
type LLMRepository struct {
	db *gorm.DB
}

// NewLLMRepository creates a new LLM repository
func NewLLMRepository(db *gorm.DB) *LLMRepository {
	return &LLMRepository{db: db}
}

// Provider operations

// ListProviders returns all LLM providers for a specific project
func (r *LLMRepository) ListProviders(projectID uint) ([]model.LLMProvider, error) {
	var providers []model.LLMProvider
	err := r.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&providers).Error
	return providers, err
}

// FindByProjectID retrieves all providers for a project (alias for compatibility)
func (r *LLMRepository) FindByProjectID(projectID uint) ([]model.LLMProvider, error) {
	return r.ListProviders(projectID)
}

// GetProvider retrieves a provider by ID with project validation
func (r *LLMRepository) GetProvider(id uint, projectID uint) (*model.LLMProvider, error) {
	var provider model.LLMProvider
	err := r.db.Where("id = ? AND project_id = ?", id, projectID).First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

// GetProviderByID retrieves a provider by ID WITHOUT project validation.
//
// ⚠️ SECURITY WARNING: This method bypasses project isolation!
// Only use for system-level operations where project context is not available:
//   - Internal service logic (e.g., LLMService updating provider with masked token)
//   - Background job processors
//   - Service-to-service calls
//
// ❌ DO NOT USE for user-facing API endpoints!
// ✅ For user requests, use GetProvider(id, projectID) instead to enforce project isolation.
//
// Example valid usage:
//   func (s *LLMService) UpdateProvider(provider *Provider) {
//       existing, _ := s.repo.GetProviderByID(provider.ID)  // OK: internal service logic
//       if provider.APIKey == "***masked***" {
//           provider.APIKey = existing.APIKey  // Preserve existing key
//       }
//   }
func (r *LLMRepository) GetProviderByID(id uint) (*model.LLMProvider, error) {
	var provider model.LLMProvider
	err := r.db.First(&provider, id).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

// CreateProvider creates a new LLM provider
func (r *LLMRepository) CreateProvider(provider *model.LLMProvider) error {
	return r.db.Create(provider).Error
}

// UpdateProvider updates an existing provider
func (r *LLMRepository) UpdateProvider(provider *model.LLMProvider) error {
	return r.db.Save(provider).Error
}

func (r *LLMRepository) DeleteProvider(id uint) error {
	return r.db.Delete(&model.LLMProvider{}, id).Error
}

// UpdateProviderTestStatus updates the test status
func (r *LLMRepository) UpdateProviderTestStatus(id uint, status, message string) error {
	updates := map[string]interface{}{
		"last_test_status":  status,
		"last_test_message": message,
	}
	return r.db.Model(&model.LLMProvider{}).Where("id = ?", id).Updates(updates).Error
}
