package model

import "time"

// UsageStatus represents LLM API call status
type UsageStatus string

const (
	UsageStatusSuccess UsageStatus = "success"
	UsageStatusFailed  UsageStatus = "failed"
	UsageStatusTimeout UsageStatus = "timeout"
)

// UsageRequestType represents LLM API request type
type UsageRequestType string

const (
	UsageTypeCodeReview     UsageRequestType = "code_review"
	UsageTypeTestConnection UsageRequestType = "test_connection"
)

// LLMUsageLog records each LLM API request for token tracking and cost analysis
// This is the core table for operations metrics - every API call is logged here
type LLMUsageLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`

	// Business Context (关联到业务实体)
	ReviewResultID *uint `gorm:"index" json:"review_result_id"` // nullable, 失败时可能还没有 review
	RepositoryID   uint  `gorm:"not null;index" json:"repository_id"`
	ProjectID      uint  `gorm:"not null;index" json:"project_id"`

	// LLM Context (关联到 LLM 配置)
	LLMProviderID uint   `gorm:"not null;index" json:"llm_provider_id"`
	ModelName     string `gorm:"not null;size:100;index" json:"model_name"` // 实际使用的模型名

	// Request Details
	RequestType UsageRequestType `gorm:"not null;size:50;index;type:varchar(50)" json:"request_type"` // code_review, test_connection
	Status      UsageStatus      `gorm:"not null;size:20;index;type:varchar(20)" json:"status"`       // success, failed, timeout
	ErrorCode   string           `gorm:"size:50" json:"error_code"`                                   // API error code if failed
	ErrorMsg    string           `gorm:"size:1000" json:"error_msg"`                                  // Error message if failed

	// Token Usage (核心数据)
	PromptTokens     int `gorm:"not null;default:0" json:"prompt_tokens"`
	CompletionTokens int `gorm:"not null;default:0" json:"completion_tokens"`
	TotalTokens      int `gorm:"not null;default:0;index" json:"total_tokens"`

	// Performance Metrics
	DurationMs    int64 `gorm:"not null;default:0" json:"duration_ms"` // API 调用耗时(毫秒)
	RequestSizeB  int   `gorm:"default:0" json:"request_size_b"`       // 请求体大小（字节）
	ResponseSizeB int   `gorm:"default:0" json:"response_size_b"`      // 响应体大小（字节）

	// Relationships
	ReviewResult *ReviewResult `gorm:"foreignKey:ReviewResultID" json:"review_result,omitempty"`
	Repository   *Repository   `gorm:"foreignKey:RepositoryID" json:"repository,omitempty"`
	Project      *Project      `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	LLMProvider  *LLMProvider  `gorm:"foreignKey:LLMProviderID" json:"llm_provider,omitempty"`
}

// TableName specifies the table name
func (LLMUsageLog) TableName() string {
	return "llm_usage_logs"
}
