package llm

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// parseReviewResponse parses LLM response into structured ReviewResponse
func parseReviewResponse(content string) (*ReviewResponse, error) {
	// Try to extract JSON from markdown code blocks
	content = extractJSONFromMarkdown(content)

	// Try to parse as JSON
	var reviewResp ReviewResponse
	if err := json.Unmarshal([]byte(content), &reviewResp); err != nil {
		// If JSON parsing fails, try to extract structured data from text
		return parseTextResponse(content)
	}

	// Validate parsed response
	if reviewResp.Summary == "" {
		reviewResp.Summary = "No summary provided"
	}

	// Ensure score is in valid range
	if reviewResp.Score < 0 {
		reviewResp.Score = 0
	}
	if reviewResp.Score > 100 {
		reviewResp.Score = 100
	}

	// Validate suggestions
	for i := range reviewResp.Suggestions {
		if reviewResp.Suggestions[i].Severity == "" {
			reviewResp.Suggestions[i].Severity = "medium"
		}
		if reviewResp.Suggestions[i].Category == "" {
			reviewResp.Suggestions[i].Category = "general"
		}
	}

	return &reviewResp, nil
}

// extractJSONFromMarkdown extracts JSON content from markdown code blocks
func extractJSONFromMarkdown(content string) string {
	// Pattern: ```json\n{...}\n```
	jsonBlockPattern := regexp.MustCompile("(?s)```(?:json)?\n?(.*?)\n?```")
	matches := jsonBlockPattern.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Pattern: {..."summary":...}
	jsonPattern := regexp.MustCompile(`(?s)\{.*?"summary".*?\}`)
	if jsonPattern.MatchString(content) {
		return content
	}

	return content
}

// parseTextResponse attempts to parse text response into structured format
func parseTextResponse(content string) (*ReviewResponse, error) {
	resp := &ReviewResponse{
		Summary:     extractSummary(content),
		Score:       extractScore(content),
		Suggestions: extractSuggestions(content),
	}

	if resp.Summary == "" {
		resp.Summary = "Unable to parse review summary"
	}

	return resp, nil
}

// extractSummary extracts summary from text
func extractSummary(content string) string {
	// Try to find "Summary:" or "Overall:" section
	summaryPatterns := []string{
		`(?i)(?:summary|overall):\s*(.+?)(?:\n\n|\n[A-Z]|$)`,
		`(?i)(?:##\s*summary|##\s*overall)\s*\n+(.+?)(?:\n\n|\n##|$)`,
	}

	for _, pattern := range summaryPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	// Fallback: use first paragraph
	paragraphs := strings.Split(content, "\n\n")
	if len(paragraphs) > 0 {
		return strings.TrimSpace(paragraphs[0])
	}

	return content
}

// extractScore extracts quality score from text
func extractScore(content string) int {
	// Try to find score patterns
	scorePatterns := []string{
		`(?i)score:\s*(\d+)`,
		`(?i)quality score:\s*(\d+)`,
		`(?i)rating:\s*(\d+)`,
		`(\d+)\s*/\s*100`,
	}

	for _, pattern := range scorePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			var score int
			fmt.Sscanf(matches[1], "%d", &score)
			if score >= 0 && score <= 100 {
				return score
			}
		}
	}

	// Default score based on content sentiment
	if strings.Contains(strings.ToLower(content), "excellent") ||
		strings.Contains(strings.ToLower(content), "perfect") {
		return 90
	}
	if strings.Contains(strings.ToLower(content), "good") {
		return 75
	}
	if strings.Contains(strings.ToLower(content), "issue") ||
		strings.Contains(strings.ToLower(content), "problem") {
		return 60
	}

	return 70 // Default neutral score
}

// extractSuggestions extracts fix suggestions from text
func extractSuggestions(content string) []FixSuggestion {
	suggestions := []FixSuggestion{}

	// Try to find numbered or bulleted lists
	listPattern := regexp.MustCompile(`(?m)^\s*(?:\d+\.|\-|\*)\s*(.+)$`)
	matches := listPattern.FindAllStringSubmatch(content, -1)

	for i, match := range matches {
		if len(match) > 1 {
			suggestion := FixSuggestion{
				FilePath:    "unknown",
				LineStart:   0,
				LineEnd:     0,
				Severity:    "medium",
				Category:    "general",
				Description: strings.TrimSpace(match[1]),
				Suggestion:  strings.TrimSpace(match[1]),
			}

			// Try to extract file path
			filePattern := regexp.MustCompile(`(?i)(?:file|path):\s*([^\s]+)`)
			if fileMatch := filePattern.FindStringSubmatch(match[1]); len(fileMatch) > 1 {
				suggestion.FilePath = fileMatch[1]
			}

			// Try to extract line numbers
			linePattern := regexp.MustCompile(`(?i)line[s]?\s*(\d+)(?:\s*-\s*(\d+))?`)
			if lineMatch := linePattern.FindStringSubmatch(match[1]); len(lineMatch) > 1 {
				fmt.Sscanf(lineMatch[1], "%d", &suggestion.LineStart)
				if len(lineMatch) > 2 && lineMatch[2] != "" {
					fmt.Sscanf(lineMatch[2], "%d", &suggestion.LineEnd)
				} else {
					suggestion.LineEnd = suggestion.LineStart
				}
			}

			// Determine severity based on keywords
			lowerDesc := strings.ToLower(match[1])
			if strings.Contains(lowerDesc, "critical") ||
				strings.Contains(lowerDesc, "security") ||
				strings.Contains(lowerDesc, "vulnerability") {
				suggestion.Severity = "high"
			} else if strings.Contains(lowerDesc, "minor") ||
				strings.Contains(lowerDesc, "style") ||
				strings.Contains(lowerDesc, "formatting") {
				suggestion.Severity = "low"
			}

			// Determine category
			if strings.Contains(lowerDesc, "security") {
				suggestion.Category = "security"
			} else if strings.Contains(lowerDesc, "performance") {
				suggestion.Category = "performance"
			} else if strings.Contains(lowerDesc, "style") ||
				strings.Contains(lowerDesc, "format") {
				suggestion.Category = "style"
			}

			suggestions = append(suggestions, suggestion)

			// Limit suggestions
			if i >= 19 { // Max 20 suggestions
				break
			}
		}
	}

	return suggestions
}
