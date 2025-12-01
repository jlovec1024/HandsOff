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

// GetConfig retrieves the Git platform config (single instance)
func (r *PlatformRepository) GetConfig() (*model.GitPlatformConfig, error) {
	var config model.GitPlatformConfig
	err := r.db.First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// CreateOrUpdateConfig creates or updates the Git platform config
func (r *PlatformRepository) CreateOrUpdateConfig(config *model.GitPlatformConfig) error {
	var existing model.GitPlatformConfig
	err := r.db.First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new config
		return r.db.Create(config).Error
	} else if err != nil {
		return err
	}

	// Update existing config
	config.ID = existing.ID
	return r.db.Save(config).Error
}

// UpdateTestStatus updates the test status of the platform config
func (r *PlatformRepository) UpdateTestStatus(id uint, status, message string) error {
	updates := map[string]interface{}{
		"last_test_status":  status,
		"last_test_message": message,
	}
	return r.db.Model(&model.GitPlatformConfig{}).Where("id = ?", id).Updates(updates).Error
}
