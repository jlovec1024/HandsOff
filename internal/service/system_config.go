package service

import (
	"errors"
	"fmt"

	"github.com/handsoff/handsoff/internal/model"
	"gorm.io/gorm"
)

// SystemConfigService handles system configuration operations
type SystemConfigService struct {
	db *gorm.DB
}

// NewSystemConfigService creates a new system config service
func NewSystemConfigService(db *gorm.DB) *SystemConfigService {
	return &SystemConfigService{db: db}
}

// WebhookConfig represents webhook configuration
type WebhookConfig struct {
	WebhookCallbackURL string `json:"webhook_callback_url"`
}

// GetWebhookConfig retrieves webhook configuration for a project
func (s *SystemConfigService) GetWebhookConfig(projectID uint) (*WebhookConfig, error) {
	var config model.SystemConfig
	err := s.db.Where("project_id = ? AND config_key = ?", projectID, model.ConfigKeyWebhookURL).
		First(&config).Error

	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("webhook URL not configured")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get webhook config: %w", err)
	}

	return &WebhookConfig{
		WebhookCallbackURL: config.Value,
	}, nil
}

// UpdateWebhookConfig updates webhook configuration for a project
func (s *SystemConfigService) UpdateWebhookConfig(projectID uint, webhookURL, webhookSecret string) error {
	return s.upsertConfig(s.db, projectID, model.ConfigKeyWebhookURL, webhookURL)
}

// upsertConfig creates or updates a config value
func (s *SystemConfigService) upsertConfig(tx *gorm.DB, projectID uint, key, value string) error {
	var config model.SystemConfig
	err := tx.Where("project_id = ? AND config_key = ?", projectID, key).First(&config).Error

	if err == gorm.ErrRecordNotFound {
		// Create new config
		config = model.SystemConfig{
			ProjectID: projectID,
			ConfigKey: key,
			Value:     value,
		}
		return tx.Create(&config).Error
	}

	if err != nil {
		return err
	}

	// Update existing config
	config.Value = value
	return tx.Save(&config).Error
}
