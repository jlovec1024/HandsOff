package model

import "time"

// EventType represents webhook event type
type EventType string

const (
	EventTypePush         EventType = "push"
	EventTypeMergeRequest EventType = "merge_request"
)

// EventStatus represents webhook event processing status
type EventStatus string

const (
	EventStatusPending    EventStatus = "pending"
	EventStatusProcessing EventStatus = "processing"
	EventStatusCompleted  EventStatus = "completed"
	EventStatusFailed     EventStatus = "failed"
	EventStatusIgnored    EventStatus = "ignored"
)

// MRAction represents merge request action
type MRAction string

const (
	MRActionOpen   MRAction = "open"
	MRActionUpdate MRAction = "update"
	MRActionMerge  MRAction = "merge"
	MRActionClose  MRAction = "close"
	MRActionReopen MRAction = "reopen"
)

// WebhookEvent records every received webhook event
type WebhookEvent struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	RepositoryID uint      `gorm:"not null;index:idx_webhook_repo_type" json:"repository_id"`
	EventType    EventType `gorm:"not null;size:30;index:idx_webhook_repo_type;index;type:varchar(30)" json:"event_type"` // push, merge_request
	Action       MRAction  `gorm:"size:30;index;type:varchar(30)" json:"action"`                                           // open, update, merge (for MR)

	// Event metadata
	SourceBranch string `gorm:"size:255" json:"source_branch"`
	TargetBranch string `gorm:"size:255" json:"target_branch"`                                 // for MR
	MRIID        *int64 `gorm:"index:idx_webhook_mr" json:"mr_iid"`                            // GitLab MR IID (nullable for push events)
	CommitSHA    string `gorm:"size:100;uniqueIndex:idx_repo_commit;index" json:"commit_sha"` // Unique per repository for commit-based deduplication

	// Processing status
	Status       EventStatus `gorm:"size:20;index;not null;default:'pending'" json:"status"` // pending, processing, completed, failed, ignored
	ProcessedAt  *time.Time  `json:"processed_at"`
	ErrorMessage string      `gorm:"type:text" json:"error_message"`

	// Raw data for debugging
	RawPayload string `gorm:"type:text" json:"-"` // Store complete JSON

	// Relationships
	Repository *Repository `gorm:"foreignKey:RepositoryID" json:"repository,omitempty"`
}

// TableName specifies the table name
func (WebhookEvent) TableName() string {
	return "webhook_events"
}
