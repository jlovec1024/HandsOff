package llm

import (
	"fmt"
	"strings"
)

// DefaultPromptTemplate is the default code review prompt
const DefaultPromptTemplate = `Please review the following code changes and provide structured feedback.

## Code Changes (Git Diff)
{{.Diff}}

## Review Requirements
1. Analyze the code for:
   - Security vulnerabilities
   - Performance issues
   - Code quality and maintainability
   - Best practices violations
   - Potential bugs

2. Provide feedback in the following JSON format:
{
  "summary": "Overall review summary (2-3 sentences)",
  "score": 75,  // Quality score from 0-100
  "suggestions": [
    {
      "file_path": "path/to/file.go",
      "line_start": 10,
      "line_end": 15,
      "severity": "high",  // high, medium, low
      "category": "security",  // security, performance, style, logic, etc.
      "description": "Detailed description of the issue",
      "suggestion": "Recommended fix or improvement",
      "code_snippet": "Original problematic code"
    }
  ]
}

## Guidelines
- Be constructive and specific
- Prioritize critical issues (security, bugs)
- Include line numbers when possible
- Limit to top 10 most important issues
- Provide actionable suggestions

Please respond ONLY with valid JSON.`

// PromptData holds data for prompt template rendering
type PromptData struct {
	Diff          string
	MRTitle       string
	MRAuthor      string
	SourceBranch  string
	TargetBranch  string
	CommitMessage string
}

// RenderPrompt renders the prompt template with data
func RenderPrompt(template string, data PromptData) string {
	if template == "" {
		template = DefaultPromptTemplate
	}

	// Simple template variable replacement
	result := template
	result = strings.ReplaceAll(result, "{{.Diff}}", data.Diff)
	result = strings.ReplaceAll(result, "{{.MRTitle}}", data.MRTitle)
	result = strings.ReplaceAll(result, "{{.MRAuthor}}", data.MRAuthor)
	result = strings.ReplaceAll(result, "{{.SourceBranch}}", data.SourceBranch)
	result = strings.ReplaceAll(result, "{{.TargetBranch}}", data.TargetBranch)
	result = strings.ReplaceAll(result, "{{.CommitMessage}}", data.CommitMessage)

	return result
}

// ValidatePromptTemplate validates a custom prompt template
func ValidatePromptTemplate(template string) error {
	if template == "" {
		return fmt.Errorf("template cannot be empty")
	}

	// Check for required placeholder
	if !strings.Contains(template, "{{.Diff}}") {
		return fmt.Errorf("template must contain {{.Diff}} placeholder")
	}

	return nil
}

// GetDefaultPrompt returns the default prompt template
func GetDefaultPrompt() string {
	return DefaultPromptTemplate
}

// BuildPromptData builds prompt data from review context
func BuildPromptData(diff, mrTitle, mrAuthor, sourceBranch, targetBranch string) PromptData {
	return PromptData{
		Diff:         diff,
		MRTitle:      mrTitle,
		MRAuthor:     mrAuthor,
		SourceBranch: sourceBranch,
		TargetBranch: targetBranch,
	}
}
