package model

import "time"

// ReviewResult represents a code review result
type ReviewResult struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	RepositoryID   uint      `gorm:"not null;uniqueIndex:idx_repo_mr" json:"repository_id"`   // Foreign key to repositories
	MergeRequestID int64     `gorm:"not null;uniqueIndex:idx_repo_mr" json:"merge_request_id"` // GitLab MR ID (unique per repository)
	MRTitle        string    `gorm:"size:500" json:"mr_title"`
	MRAuthor       string    `gorm:"size:100;index" json:"mr_author"`
	SourceBranch   string    `gorm:"size:255" json:"source_branch"`
	TargetBranch   string    `gorm:"size:255" json:"target_branch"`
	MRWebURL       string    `gorm:"size:500" json:"mr_web_url"`
	LLMProviderID  uint      `gorm:"not null;index" json:"llm_provider_id"`  // Foreign key to llm_providers
	Score          int       `gorm:"index" json:"score"`                  // 0-100
	Summary        string    `gorm:"type:text" json:"summary"`            // AI summary
	RawResult      string    `gorm:"type:text" json:"raw_result"`         // Raw AI response (JSON)
	Status         string    `gorm:"size:20;index" json:"status"`         // pending, processing, completed, failed
	ErrorMessage   string    `gorm:"size:1000" json:"error_message"`
	ReviewedAt     *time.Time `json:"reviewed_at"`
	CommentPosted  bool       `gorm:"default:false;not null" json:"comment_posted"` // Whether comment was posted to GitLab
	CommentURL     string     `gorm:"size:500" json:"comment_url"`

	// Webhook event relationship (optional, for tracing which webhook triggered this review)
	WebhookEventID *uint `gorm:"index" json:"webhook_event_id"` // Foreign key to webhook_events

	// Statistics fields
	IssuesFound            int `gorm:"default:0" json:"issues_found"`              // Total number of issues found
	CriticalIssuesCount    int `gorm:"default:0;index" json:"critical_issues_count"` // Number of critical severity issues
	HighIssuesCount        int `gorm:"default:0" json:"high_issues_count"`         // Number of high severity issues
	MediumIssuesCount      int `gorm:"default:0" json:"medium_issues_count"`       // Number of medium severity issues
	LowIssuesCount         int `gorm:"default:0" json:"low_issues_count"`          // Number of low severity issues
	SecurityIssuesCount    int `gorm:"default:0;index" json:"security_issues_count"`  // Number of security issues
	PerformanceIssuesCount int `gorm:"default:0" json:"performance_issues_count"` // Number of performance issues
	QualityIssuesCount     int `gorm:"default:0" json:"quality_issues_count"`      // Number of quality issues

	// Relationships
	Repository     *Repository      `gorm:"foreignKey:RepositoryID" json:"repository,omitempty"`
	LLMProvider    *LLMProvider     `gorm:"foreignKey:LLMProviderID" json:"llm_provider,omitempty"`
	FixSuggestions []FixSuggestion  `gorm:"foreignKey:ReviewResultID" json:"fix_suggestions,omitempty"`
}

// TableName specifies the table name
func (ReviewResult) TableName() string {
	return "review_results"
}
