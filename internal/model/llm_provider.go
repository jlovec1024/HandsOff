package model

import "time"

// LLMProvider represents an LLM service provider (project-scoped)
// All providers are OpenAI-compatible (unified interface)
type LLMProvider struct {
	ID              uint       `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Name            string     `gorm:"not null;size:100;index" json:"name"`          // User-defined name, e.g., "OpenAI Official", "DeepSeek China"
	BaseURL         string     `gorm:"not null;size:255" json:"base_url"`            // API endpoint
	APIKey          string     `gorm:"not null;size:500" json:"-"`                   // Encrypted, never expose in JSON
	Model           string     `gorm:"not null;size:100" json:"model"`                // Model name, e.g., "gpt-4", "deepseek-chat"
	IsActive        bool       `gorm:"default:true;not null;index" json:"is_active"`
	LastTestedAt    *time.Time `json:"last_tested_at"`
	LastTestStatus  string     `gorm:"size:20" json:"last_test_status"`  // success, failed
	LastTestMessage string     `gorm:"size:500" json:"last_test_message"`

	// Project Relationship
	ProjectID uint    `gorm:"not null;index;constraint:OnDelete:CASCADE" json:"project_id"`
	Project   Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`
}

// TableName specifies the table name
func (LLMProvider) TableName() string {
	return "llm_providers"
}
