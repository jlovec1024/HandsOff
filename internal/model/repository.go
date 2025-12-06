package model

import "time"

// Webhook test status constants
const (
	WebhookTestStatusNever   = "never"
	WebhookTestStatusSuccess = "success"
	WebhookTestStatusFailed  = "failed"
)

// Repository represents a Git repository (project-scoped)
type Repository struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	PlatformID     uint      `gorm:"not null;index" json:"platform_id"`           // Foreign key to git_platform_configs
	PlatformRepoID int64     `gorm:"not null;index" json:"platform_repo_id"`      // GitLab/GitHub repository ID
	Name           string    `gorm:"not null;size:255;index" json:"name"`         // Repository name
	FullPath       string    `gorm:"not null;size:500" json:"full_path"`          // e.g., "group/subgroup/repo"
	HTTPURL        string    `gorm:"size:500" json:"http_url"`                    // HTTP clone URL
	SSHURL         string    `gorm:"size:500" json:"ssh_url"`                     // SSH clone URL
	DefaultBranch  string    `gorm:"size:100" json:"default_branch"`              // e.g., "main", "master"
	LLMProviderID  *uint     `gorm:"index" json:"llm_provider_id"`                // Foreign key to llm_providers (nullable)
	WebhookID      *int64    `json:"webhook_id"`                                  // GitLab webhook ID
	WebhookURL     string    `gorm:"size:500" json:"webhook_url"`                 // Webhook callback URL
	WebhookSecret  string    `gorm:"size:255" json:"-"`                           // Webhook secret token (not exposed in JSON)
	IsActive       bool      `gorm:"default:true;not null;index" json:"is_active"`

	// Webhook status tracking
	LastWebhookTestAt     *time.Time `json:"last_webhook_test_at"`                           // Last webhook test time
	LastWebhookTestStatus string     `gorm:"size:20;default:'never'" json:"last_webhook_test_status"` // success, failed, never
	LastWebhookTestError  string     `gorm:"type:text" json:"last_webhook_test_error"`       // Error message if failed

	// Project Relationship
	ProjectID uint    `gorm:"not null;index;constraint:OnDelete:CASCADE" json:"project_id"`
	Project   Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`

	// Relationships
	Platform GitPlatformConfig `gorm:"foreignKey:PlatformID;constraint:OnDelete:CASCADE" json:"platform,omitempty"`
	LLMProvider *LLMProvider      `gorm:"foreignKey:LLMProviderID" json:"llm_provider,omitempty"`
}

// TableName specifies the table name
func (Repository) TableName() string {
	return "repositories"
}
