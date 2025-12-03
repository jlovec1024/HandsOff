package model

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Project represents a user's project with independent configs
type Project struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Basic Info
	Name        string `gorm:"not null;size:100;uniqueIndex:idx_user_project" json:"name"`
	Description string `gorm:"size:500" json:"description"`

	// User Relationship (每个项目属于一个用户)
	UserID uint `gorm:"not null;index;uniqueIndex:idx_user_project" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Relationships
	GitConfigs   []GitPlatformConfig `gorm:"foreignKey:ProjectID" json:"git_configs,omitempty"`
	LLMProviders []LLMProvider       `gorm:"foreignKey:ProjectID" json:"llm_providers,omitempty"`
	Repositories []Repository        `gorm:"foreignKey:ProjectID" json:"repositories,omitempty"`
}

// TableName specifies the table name
func (Project) TableName() string {
	return "projects"
}

// BeforeCreate hook to validate project
func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.Name == "" {
		return fmt.Errorf("project name is required")
	}
	return nil
}
