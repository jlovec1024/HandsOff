package llm

import (
	"encoding/json"
	"strings"
	"time"
)

// SchemaVersion 当前 JSON Schema 版本
const SchemaVersion = "1.0"

// ReviewOutputJSON 完整的 JSON 输出结构
type ReviewOutputJSON struct {
	SchemaVersion string         `json:"schema_version"`
	GeneratedAt   time.Time      `json:"generated_at"`
	Context       OutputContext  `json:"context"`
	Result        OutputResult   `json:"result"`
	Statistics    OutputStats    `json:"statistics"`
	Metadata      OutputMetadata `json:"metadata"`
}

// OutputContext 上下文信息
type OutputContext struct {
	Repository   ContextRepository   `json:"repository"`
	MergeRequest ContextMergeRequest `json:"merge_request"`
	Review       ContextReview       `json:"review"`
}

// ContextRepository 仓库信息
type ContextRepository struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	FullName       string `json:"full_name"`
	Platform       string `json:"platform"`
	PlatformRepoID int64  `json:"platform_repo_id"`
}

// ContextMergeRequest MR 信息
type ContextMergeRequest struct {
	ID           int64  `json:"id"`
	IID          int64  `json:"iid"`
	Title        string `json:"title"`
	Author       string `json:"author"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	WebURL       string `json:"web_url"`
}

// ContextReview Review 元信息
type ContextReview struct {
	ID          uint      `json:"id"`
	ReviewedAt  time.Time `json:"reviewed_at"`
	LLMProvider string    `json:"llm_provider"`
	LLMModel    string    `json:"llm_model"`
	TokensUsed  int       `json:"tokens_used"`
	DurationMs  int64     `json:"duration_ms"`
}

// OutputResult 审查结果（核心 LLM 输出）
type OutputResult struct {
	Summary      string             `json:"summary"`
	Score        int                `json:"score"`
	QualityLevel string             `json:"quality_level"`
	Suggestions  []OutputSuggestion `json:"suggestions"`
}

// OutputSuggestion 单个建议
type OutputSuggestion struct {
	ID          int    `json:"id"`
	FilePath    string `json:"file_path"`
	LineStart   int    `json:"line_start"`
	LineEnd     int    `json:"line_end"`
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion"`
	CodeSnippet string `json:"code_snippet,omitempty"`
}

// OutputStats 统计信息
type OutputStats struct {
	TotalIssues   int           `json:"total_issues"`
	BySeverity    SeverityCount `json:"by_severity"`
	ByCategory    CategoryCount `json:"by_category"`
	FilesAffected int           `json:"files_affected"`
	TopFiles      []FileIssue   `json:"top_files,omitempty"`
}

// SeverityCount 按严重程度统计
type SeverityCount struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
}

// CategoryCount 按类别统计
type CategoryCount struct {
	Security      int `json:"security"`
	Performance   int `json:"performance"`
	Style         int `json:"style"`
	Logic         int `json:"logic"`
	Documentation int `json:"documentation"`
	Other         int `json:"other"`
}

// FileIssue 文件问题数
type FileIssue struct {
	File   string `json:"file"`
	Issues int    `json:"issues"`
}

// OutputMetadata 元数据
type OutputMetadata struct {
	PromptTemplate     string `json:"prompt_template"`
	PromptVersion      string `json:"prompt_version,omitempty"`
	CustomPromptUsed   bool   `json:"custom_prompt_used"`
	RawResponseAvail   bool   `json:"raw_response_available"`
	ParserFallbackUsed bool   `json:"parser_fallback_used"`
}

// GetQualityLevel 根据分数返回质量等级
func GetQualityLevel(score int) string {
	switch {
	case score >= 90:
		return "excellent"
	case score >= 75:
		return "good"
	case score >= 60:
		return "acceptable"
	case score >= 40:
		return "poor"
	default:
		return "critical"
	}
}

// CalculateOutputStatistics 从建议列表计算统计信息
func CalculateOutputStatistics(suggestions []FixSuggestion) OutputStats {
	stats := OutputStats{
		TotalIssues: len(suggestions),
	}

	fileCount := make(map[string]int)

	for _, sug := range suggestions {
		// 统计 severity
		switch strings.ToLower(sug.Severity) {
		case "critical":
			stats.BySeverity.Critical++
		case "high":
			stats.BySeverity.High++
		case "medium":
			stats.BySeverity.Medium++
		case "low":
			stats.BySeverity.Low++
		default:
			stats.BySeverity.Medium++ // 默认为 medium
		}

		// 统计 category
		switch strings.ToLower(sug.Category) {
		case "security":
			stats.ByCategory.Security++
		case "performance":
			stats.ByCategory.Performance++
		case "style":
			stats.ByCategory.Style++
		case "logic":
			stats.ByCategory.Logic++
		case "documentation":
			stats.ByCategory.Documentation++
		default:
			stats.ByCategory.Other++
		}

		// 统计文件
		if sug.FilePath != "" && sug.FilePath != "unknown" {
			fileCount[sug.FilePath]++
		}
	}

	stats.FilesAffected = len(fileCount)
	stats.TopFiles = getTopFiles(fileCount, 5)

	return stats
}

// getTopFiles 获取问题最多的 N 个文件
func getTopFiles(fileCount map[string]int, n int) []FileIssue {
	result := make([]FileIssue, 0, len(fileCount))
	for file, count := range fileCount {
		result = append(result, FileIssue{File: file, Issues: count})
	}

	// 简单排序（问题数降序）
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Issues > result[i].Issues {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	if len(result) > n {
		return result[:n]
	}
	return result
}

// ConvertSuggestionsToOutput 转换建议列表为输出格式
func ConvertSuggestionsToOutput(suggestions []FixSuggestion) []OutputSuggestion {
	result := make([]OutputSuggestion, len(suggestions))
	for i, sug := range suggestions {
		result[i] = OutputSuggestion{
			ID:          i + 1,
			FilePath:    sug.FilePath,
			LineStart:   sug.LineStart,
			LineEnd:     sug.LineEnd,
			Severity:    NormalizeSeverity(sug.Severity),
			Category:    NormalizeCategory(sug.Category),
			Description: sug.Description,
			Suggestion:  sug.Suggestion,
			CodeSnippet: sug.CodeSnippet,
		}
	}
	return result
}

// NormalizeSeverity 标准化 severity 值
// 将各种可能的输入（包括大小写变体、近义词）转换为标准值：critical, high, medium, low
func NormalizeSeverity(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "critical", "blocker", "urgent", "fatal":
		return "critical"
	case "high", "major", "important", "error":
		return "high"
	case "medium", "moderate", "warning", "normal":
		return "medium"
	case "low", "minor", "trivial", "info", "suggestion", "hint":
		return "low"
	default:
		return "medium" // 默认为 medium
	}
}

// NormalizeCategory 标准化 category 值
// 将各种可能的输入转换为标准值：security, performance, style, logic, documentation, other
func NormalizeCategory(c string) string {
	c = strings.ToLower(strings.TrimSpace(c))
	switch c {
	case "security", "sec", "vulnerability", "vuln":
		return "security"
	case "performance", "perf", "efficiency", "optimization":
		return "performance"
	case "style", "formatting", "code style", "lint", "convention":
		return "style"
	case "logic", "bug", "error", "correctness", "behavior":
		return "logic"
	case "documentation", "doc", "docs", "comment", "comments":
		return "documentation"
	default:
		return "other"
	}
}

// FormatReviewAsJSON 将 ReviewResponse 格式化为完整 JSON 输出
// 注意：context 参数需要调用方提供（因为包含数据库数据）
func FormatReviewAsJSON(resp *ReviewResponse, ctx OutputContext, meta OutputMetadata) (string, error) {
	output := ReviewOutputJSON{
		SchemaVersion: SchemaVersion,
		GeneratedAt:   time.Now(),
		Context:       ctx,
		Result: OutputResult{
			Summary:      resp.Summary,
			Score:        resp.Score,
			QualityLevel: GetQualityLevel(resp.Score),
			Suggestions:  ConvertSuggestionsToOutput(resp.Suggestions),
		},
		Statistics: CalculateOutputStatistics(resp.Suggestions),
		Metadata:   meta,
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
