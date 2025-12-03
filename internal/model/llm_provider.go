package model

import "time"

// LLMProvider represents an LLM service provider (project-scoped)
type LLMProvider struct {
	ID              uint       `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Name            string     `gorm:"not null;size:100;index" json:"name"`        // e.g., "OpenAI", "DeepSeek"
	Type            string     `gorm:"not null;size:50" json:"type"`               // openai, deepseek, claude, gemini, ollama
	BaseURL         string     `gorm:"not null;size:255" json:"base_url"`          // API endpoint
	APIKey          string     `gorm:"not null;size:500" json:"-"`                 // Encrypted, never expose in JSON
	IsActive        bool       `gorm:"default:true;not null;index" json:"is_active"`
	LastTestedAt    *time.Time `json:"last_tested_at"`
	LastTestStatus  string     `gorm:"size:20" json:"last_test_status"`  // success, failed
	LastTestMessage string     `gorm:"size:500" json:"last_test_message"`

	// Project Relationship
	ProjectID uint    `gorm:"not null;index;constraint:OnDelete:CASCADE" json:"project_id"`
	Project   Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`

	// Relationships
	Models []LLMModel `gorm:"foreignKey:ProviderID;constraint:OnDelete:CASCADE" json:"models,omitempty"`
}

// TableName specifies the table name
func (LLMProvider) TableName() string {
	return "llm_providers"
}
