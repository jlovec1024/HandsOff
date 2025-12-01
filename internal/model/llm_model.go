package model

import "time"

// LLMModel represents a specific LLM model
type LLMModel struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ProviderID   uint      `gorm:"not null;index" json:"provider_id"`         // Foreign key to llm_providers
	ModelName    string    `gorm:"not null;size:100;index" json:"model_name"` // e.g., "gpt-4", "deepseek-chat"
	DisplayName  string    `gorm:"not null;size:100" json:"display_name"`     // User-friendly name
	Description  string    `gorm:"size:500" json:"description"`
	MaxTokens    int       `gorm:"default:4096" json:"max_tokens"`
	Temperature  float32   `gorm:"default:0.7" json:"temperature"`
	IsActive     bool      `gorm:"default:true;not null;index" json:"is_active"`

	// Relationships
	Provider *LLMProvider `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
}

// TableName specifies the table name
func (LLMModel) TableName() string {
	return "llm_models"
}
