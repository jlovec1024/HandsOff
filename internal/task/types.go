package task

import "encoding/json"

const (
	// Task type names
	TypeCodeReview = "code_review"
	TypeAutoFix    = "auto_fix"
)

// CodeReviewPayload represents the payload for code review task
type CodeReviewPayload struct {
	RepositoryID   uint   `json:"repository_id"`
	MergeRequestID int64  `json:"merge_request_id"`
	MRTitle        string `json:"mr_title"`
	MRAuthor       string `json:"mr_author"`
	SourceBranch   string `json:"source_branch"`
	TargetBranch   string `json:"target_branch"`
	MRWebURL       string `json:"mr_web_url"`
	ProjectID      int64  `json:"project_id"`
}

// ToJSON converts payload to JSON
func (p *CodeReviewPayload) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// FromJSON parses JSON to payload
func (p *CodeReviewPayload) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

// AutoFixPayload represents the payload for auto-fix task
type AutoFixPayload struct {
	SuggestionID uint `json:"suggestion_id"`
	RepositoryID uint `json:"repository_id"`
}

// ToJSON converts payload to JSON
func (p *AutoFixPayload) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// FromJSON parses JSON to payload
func (p *AutoFixPayload) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}
