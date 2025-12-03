package repository

import (
	"github.com/handsoff/handsoff/internal/model"
	"gorm.io/gorm"
)

// PlatformRepository handles Git platform config database operations
type PlatformRepository struct {
	db *gorm.DB
}

// NewPlatformRepository creates a new platform repository
func NewPlatformRepository(db *gorm.DB) *PlatformRepository {
	return &PlatformRepository{db: db}
}

// GetConfig retrieves the Git platform config for a specific project
func (r *PlatformRepository) GetConfig(projectID uint) (*model.GitPlatformConfig, error) {
	var config model.GitPlatformConfig
	err := r.db.Where("project_id = ?", projectID).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// FindByProjectID retrieves all platform configs for a project
func (r *PlatformRepository) FindByProjectID(projectID uint) ([]model.GitPlatformConfig, error) {
	var configs []model.GitPlatformConfig
	err := r.db.Where("project_id = ?", projectID).Find(&configs).Error
	return configs, err
}

// CreateOrUpdateConfig creates or updates the Git platform config for a project
func (r *PlatformRepository) CreateOrUpdateConfig(config *model.GitPlatformConfig) error {
	var existing model.GitPlatformConfig
	err := r.db.Where("project_id = ? AND platform_type = ?", config.ProjectID, config.PlatformType).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new config
		return r.db.Create(config).Error
	} else if err != nil {
		return err
	}

	// Update existing config - use Updates to avoid zero-value time issues
	config.ID = existing.ID
	config.CreatedAt = existing.CreatedAt // Preserve creation time
	
	// Build update map (only update non-empty fields to preserve existing values)
	updates := map[string]interface{}{
		"platform_type": config.PlatformType,
		"base_url":      config.BaseURL,
		"access_token":  config.AccessToken,
		"is_active":     config.IsActive,
	}
	
	// Only update webhook_secret if provided (avoid overwriting with empty value)
	if config.WebhookSecret != "" {
		updates["webhook_secret"] = config.WebhookSecret
	}
	
	return r.db.Model(&existing).Updates(updates).Error
}

// UpdateTestStatus updates the test status of the platform config
func (r *PlatformRepository) UpdateTestStatus(id uint, status, message string) error {
	updates := map[string]interface{}{
		"last_test_status":  status,
		"last_test_message": message,
	}
	return r.db.Model(&model.GitPlatformConfig{}).Where("id = ?", id).Updates(updates).Error
}
