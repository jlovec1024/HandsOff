package llm

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// llmSuggestion 中间结构体，用于解析 LLM 返回的原始 JSON
// 支持 LLM 常见的字段命名（file, line, message）
type llmSuggestion struct {
	File       string `json:"file"`        // LLM 可能返回 "file"
	FilePath   string `json:"file_path"`   // 或 "file_path"
	Line       int    `json:"line"`        // LLM 可能返回 "line"
	LineStart  int    `json:"line_start"`  // 或 "line_start"
	LineEnd    int    `json:"line_end"`    // 结束行
	Severity   string `json:"severity"`    // 严重程度
	Category   string `json:"category"`    // 类别
	Message    string `json:"message"`     // LLM 可能返回 "message"
	Description string `json:"description"` // 或 "description"
	Suggestion string `json:"suggestion"`  // 修复建议
	CodeSnippet string `json:"code_snippet"` // 代码片段
}

// toFixSuggestion 转换为标准的 FixSuggestion
func (ls *llmSuggestion) toFixSuggestion() FixSuggestion {
	// 优先使用更具体的字段名，回退到通用字段名
	filePath := ls.FilePath
	if filePath == "" {
		filePath = ls.File
	}
	
	lineStart := ls.LineStart
	if lineStart == 0 {
		lineStart = ls.Line
	}
	
	description := ls.Description
	if description == "" {
		description = ls.Message
	}
	
	return FixSuggestion{
		FilePath:    filePath,
		LineStart:   lineStart,
		LineEnd:     ls.LineEnd,
		Severity:    ls.Severity,
		Category:    ls.Category,
		Description: description,
		Suggestion:  ls.Suggestion,
		CodeSnippet: ls.CodeSnippet,
	}
}

// llmReviewResponse 中间结构体，用于解析 LLM 返回的完整响应
type llmReviewResponse struct {
	Summary     string          `json:"summary"`
	Score       int             `json:"score"`
	Suggestions []llmSuggestion `json:"suggestions"`
}

// parseReviewResponse parses LLM response into structured ReviewResponse
func parseReviewResponse(content string) (*ReviewResponse, error) {
	// Try to extract JSON from markdown code blocks
	jsonContent := extractJSONFromMarkdown(content)

	// Try to parse as JSON with intermediate structure
	var llmResp llmReviewResponse
	if err := json.Unmarshal([]byte(jsonContent), &llmResp); err == nil {
		// JSON 解析成功，转换为标准格式
		if llmResp.Summary != "" || llmResp.Score > 0 || len(llmResp.Suggestions) > 0 {
			reviewResp := &ReviewResponse{
				Summary:     llmResp.Summary,
				Score:       llmResp.Score,
				Suggestions: make([]FixSuggestion, len(llmResp.Suggestions)),
			}
			
			// 转换 suggestions
			for i, llmSug := range llmResp.Suggestions {
				reviewResp.Suggestions[i] = llmSug.toFixSuggestion()
			}
			
			return normalizeReviewResponse(reviewResp), nil
		}
	}

	// If JSON parsing fails, try to extract structured data from text
	return parseTextResponse(content)
}

// normalizeReviewResponse 标准化审查响应
func normalizeReviewResponse(resp *ReviewResponse) *ReviewResponse {
	// 1. 处理 summary
	if resp.Summary == "" {
		resp.Summary = "No summary provided"
	}

	// 2. 规范化 score (0-100)
	if resp.Score < 0 {
		resp.Score = 0
	}
	if resp.Score > 100 {
		resp.Score = 100
	}

	// 3. 标准化每个 suggestion
	for i := range resp.Suggestions {
		// 标准化 severity
		resp.Suggestions[i].Severity = NormalizeSeverity(resp.Suggestions[i].Severity)

		// 标准化 category
		resp.Suggestions[i].Category = NormalizeCategory(resp.Suggestions[i].Category)

		// 确保 FilePath 不为空
		if resp.Suggestions[i].FilePath == "" {
			resp.Suggestions[i].FilePath = "unknown"
		}

		// 确保 LineEnd >= LineStart
		if resp.Suggestions[i].LineEnd < resp.Suggestions[i].LineStart {
			resp.Suggestions[i].LineEnd = resp.Suggestions[i].LineStart
		}
	}

	return resp
}

// extractJSONFromMarkdown extracts JSON content from markdown code blocks or raw JSON
func extractJSONFromMarkdown(content string) string {
	content = strings.TrimSpace(content)

	// Pattern 1: ```json\n{...}\n``` or ```\n{...}\n```
	jsonBlockPattern := regexp.MustCompile("(?s)```(?:json)?\\s*\\n?(\\{.*?\\})\\s*\\n?```")
	if matches := jsonBlockPattern.FindStringSubmatch(content); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Pattern 2: 直接以 { 开头的 JSON
	if strings.HasPrefix(content, "{") {
		// 找到匹配的右括号
		braceCount := 0
		for i, ch := range content {
			if ch == '{' {
				braceCount++
			} else if ch == '}' {
				braceCount--
				if braceCount == 0 {
					return content[:i+1]
				}
			}
		}
	}

	// Pattern 3: 在文本中查找包含 "summary" 或 "score" 的 JSON 对象
	// 更宽松的匹配，支持嵌套结构
	startIdx := strings.Index(content, "{")
	if startIdx >= 0 {
		endIdx := strings.LastIndex(content, "}")
		if endIdx > startIdx {
			potential := content[startIdx : endIdx+1]
			// 验证是否是有效 JSON
			var test map[string]interface{}
			if json.Unmarshal([]byte(potential), &test) == nil {
				return potential
			}
		}
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
			suggestion.Severity = detectSeverityFromText(lowerDesc)
			suggestion.Category = detectCategoryFromText(lowerDesc)

			suggestions = append(suggestions, suggestion)

			// Limit suggestions
			if i >= 19 { // Max 20 suggestions
				break
			}
		}
	}

	return suggestions
}

// detectSeverityFromText 从文本中检测 severity
func detectSeverityFromText(text string) string {
	if strings.Contains(text, "critical") ||
		strings.Contains(text, "blocker") ||
		strings.Contains(text, "fatal") {
		return "critical"
	}
	if strings.Contains(text, "security") ||
		strings.Contains(text, "vulnerability") ||
		strings.Contains(text, "major") ||
		strings.Contains(text, "important") {
		return "high"
	}
	if strings.Contains(text, "minor") ||
		strings.Contains(text, "trivial") ||
		strings.Contains(text, "style") ||
		strings.Contains(text, "formatting") ||
		strings.Contains(text, "hint") {
		return "low"
	}
	return "medium"
}

// detectCategoryFromText 从文本中检测 category
func detectCategoryFromText(text string) string {
	if strings.Contains(text, "security") ||
		strings.Contains(text, "vulnerability") ||
		strings.Contains(text, "injection") ||
		strings.Contains(text, "xss") {
		return "security"
	}
	if strings.Contains(text, "performance") ||
		strings.Contains(text, "slow") ||
		strings.Contains(text, "optimize") {
		return "performance"
	}
	if strings.Contains(text, "style") ||
		strings.Contains(text, "format") ||
		strings.Contains(text, "naming") ||
		strings.Contains(text, "convention") {
		return "style"
	}
	if strings.Contains(text, "logic") ||
		strings.Contains(text, "bug") ||
		strings.Contains(text, "error") ||
		strings.Contains(text, "incorrect") {
		return "logic"
	}
	if strings.Contains(text, "documentation") ||
		strings.Contains(text, "comment") ||
		strings.Contains(text, "doc") {
		return "documentation"
	}
	return "other"
}
