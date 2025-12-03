package model

import "time"

// UserProjectPreference stores the active project for each user
type UserProjectPreference struct {
	UserID    uint      `gorm:"primarykey" json:"user_id"`
	ProjectID uint      `gorm:"not null;index" json:"project_id"`
	UpdatedAt time.Time `json:"updated_at"`

	User    User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Project Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`
}

// TableName specifies the table name
func (UserProjectPreference) TableName() string {
	return "user_project_preferences"
}
