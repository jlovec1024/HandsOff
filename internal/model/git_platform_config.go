package model

import "time"

// GitPlatformConfig represents GitLab platform configuration (project-scoped)
type GitPlatformConfig struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PlatformType    string    `gorm:"not null;size:20;default:'gitlab';uniqueIndex:idx_project_platform" json:"platform_type"` // gitlab, github, gitea
	BaseURL         string    `gorm:"not null;size:255" json:"base_url"`
	AccessToken     string    `gorm:"not null;size:500" json:"-"` // Encrypted, never expose in JSON
	WebhookSecret   string    `gorm:"size:100" json:"-"`          // For webhook signature verification
	IsActive        bool      `gorm:"default:true;not null" json:"is_active"`
	LastTestedAt    *time.Time `json:"last_tested_at"`
	LastTestStatus  string     `gorm:"size:20" json:"last_test_status"` // success, failed
	LastTestMessage string     `gorm:"size:500" json:"last_test_message"`

	// Project Relationship
	ProjectID uint    `gorm:"not null;index;uniqueIndex:idx_project_platform;constraint:OnDelete:CASCADE" json:"project_id"`
	Project   Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`
}

// TableName specifies the table name
func (GitPlatformConfig) TableName() string {
	return "git_platform_configs"
}
