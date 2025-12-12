package model

import "time"

// SystemConfig represents system-level configuration stored in database
type SystemConfig struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ConfigKey string    `gorm:"uniqueIndex;size:100;not null" json:"config_key"` // Configuration key
	Value     string    `gorm:"type:text" json:"value"`                          // Configuration value

	// Project Relationship
	ProjectID uint    `gorm:"not null;index;constraint:OnDelete:CASCADE" json:"project_id"`
	Project   Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`
}

// TableName specifies the table name
func (SystemConfig) TableName() string {
	return "system_configs"
}

// Predefined configuration keys
const (
	ConfigKeyWebhookURL           = "webhook_callback_url"   // System Webhook URL
	ConfigKeyReviewPromptTemplate = "review_prompt_template" // Review Prompt template
	ConfigKeyReviewPromptVersion  = "review_prompt_version"  // Review Prompt version
)
