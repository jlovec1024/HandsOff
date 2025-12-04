package task

import "encoding/json"

const (
	// Task type names
	TypeCodeReview = "code_review"
	TypeAutoFix    = "auto_fix"
)

// CodeReviewPayload represents the payload for code review task
// REFACTORED: Only pass ReviewResultID instead of duplicating all MR fields
// Worker will load ReviewResult with all relationships from DB
type CodeReviewPayload struct {
	ReviewResultID uint `json:"review_result_id"` // Foreign key to review_results
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
