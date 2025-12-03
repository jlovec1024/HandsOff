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

// DeleteProvider deletes a provider and its models
func (r *LLMRepository) DeleteProvider(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete associated models first
		if err := tx.Where("provider_id = ?", id).Delete(&model.LLMModel{}).Error; err != nil {
			return err
		}
		// Delete provider
		return tx.Delete(&model.LLMProvider{}, id).Error
	})
}

// UpdateProviderTestStatus updates the test status
func (r *LLMRepository) UpdateProviderTestStatus(id uint, status, message string) error {
	updates := map[string]interface{}{
		"last_test_status":  status,
		"last_test_message": message,
	}
	return r.db.Model(&model.LLMProvider{}).Where("id = ?", id).Updates(updates).Error
}

// Model operations

// ListModels returns all LLM models, optionally filtered by provider
func (r *LLMRepository) ListModels(providerID *uint) ([]model.LLMModel, error) {
	var models []model.LLMModel
	query := r.db.Preload("Provider").Order("created_at DESC")

	if providerID != nil {
		query = query.Where("provider_id = ?", *providerID)
	}

	err := query.Find(&models).Error
	return models, err
}

// GetModel retrieves a model by ID
func (r *LLMRepository) GetModel(id uint) (*model.LLMModel, error) {
	var model model.LLMModel
	err := r.db.Preload("Provider").First(&model, id).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// CreateModel creates a new LLM model
func (r *LLMRepository) CreateModel(model *model.LLMModel) error {
	return r.db.Create(model).Error
}

// UpdateModel updates an existing model
func (r *LLMRepository) UpdateModel(model *model.LLMModel) error {
	return r.db.Save(model).Error
}

// DeleteModel deletes a model
func (r *LLMRepository) DeleteModel(id uint) error {
	return r.db.Delete(&model.LLMModel{}, id).Error
}
