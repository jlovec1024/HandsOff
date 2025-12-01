package model

import "time"

// FixSuggestion represents a specific fix suggestion
type FixSuggestion struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	ReviewResultID uint      `gorm:"not null;index" json:"review_result_id"` // Foreign key to review_results
	FilePath       string    `gorm:"not null;size:500;index" json:"file_path"`
	LineStart      int       `json:"line_start"`
	LineEnd        int       `json:"line_end"`
	Severity       string    `gorm:"not null;size:20;index" json:"severity"` // high, medium, low
	Category       string    `gorm:"size:100;index" json:"category"`         // e.g., "security", "performance", "style"
	Description    string    `gorm:"type:text;not null" json:"description"`
	Suggestion     string    `gorm:"type:text" json:"suggestion"`
	CodeSnippet    string    `gorm:"type:text" json:"code_snippet"` // Original code snippet

	// Relationships
	ReviewResult *ReviewResult `gorm:"foreignKey:ReviewResultID" json:"review_result,omitempty"`
}

// TableName specifies the table name
func (FixSuggestion) TableName() string {
	return "fix_suggestions"
}
