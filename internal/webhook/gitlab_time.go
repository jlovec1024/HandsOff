package webhook

import (
	"strings"
	"time"
)

// GitLabTime is a custom time type that handles GitLab's non-standard time formats.
// GitLab sends timestamps like "2025-12-16 02:04:36 UTC" instead of RFC3339.
type GitLabTime struct {
	time.Time
}

// Supported time formats in order of priority
var gitlabTimeFormats = []string{
	time.RFC3339,              // "2006-01-02T15:04:05Z07:00"
	time.RFC3339Nano,          // "2006-01-02T15:04:05.999999999Z07:00"
	"2006-01-02 15:04:05 UTC", // GitLab's non-standard format
	"2006-01-02 15:04:05 MST", // Other timezone variants
	"2006-01-02T15:04:05.000Z", // ISO 8601 variant
}

// UnmarshalJSON implements json.Unmarshaler for GitLabTime
func (t *GitLabTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		return nil
	}

	for _, format := range gitlabTimeFormats {
		if parsed, err := time.Parse(format, s); err == nil {
			t.Time = parsed
			return nil
		}
	}

	// Final fallback to RFC3339 with explicit error
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

// MarshalJSON implements json.Marshaler for GitLabTime
func (t GitLabTime) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + t.Time.Format(time.RFC3339) + `"`), nil
}
