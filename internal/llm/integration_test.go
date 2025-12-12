package llm

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestIntegration_FullReviewPipeline 测试完整的审查流程
// 模拟 LLM 返回 → Parser 解析 → OutputFormatter 格式化
func TestIntegration_FullReviewPipeline(t *testing.T) {
	// 模拟 LLM 返回的 JSON 响应（类似真实场景）
	llmResponse := `
Here is my code review:

` + "```json" + `
{
  "summary": "代码整体质量良好，但存在一些安全和性能问题需要关注。",
  "score": 72,
  "suggestions": [
    {
      "file_path": "internal/api/handler.go",
      "line_start": 45,
      "line_end": 52,
      "severity": "Major",
      "category": "sec",
      "description": "SQL 注入风险：用户输入未经过滤直接拼接到 SQL 语句中",
      "suggestion": "使用参数化查询替代字符串拼接",
      "code_snippet": "query := \"SELECT * FROM users WHERE id = \" + userId"
    },
    {
      "file_path": "internal/service/cache.go",
      "line_start": 120,
      "line_end": 125,
      "severity": "warning",
      "category": "perf",
      "description": "缓存未设置过期时间，可能导致内存泄漏",
      "suggestion": "为缓存条目设置合理的 TTL",
      "code_snippet": "cache.Set(key, value)"
    },
    {
      "file_path": "internal/model/user.go",
      "line_start": 30,
      "line_end": 30,
      "severity": "hint",
      "category": "doc",
      "description": "公开函数缺少文档注释",
      "suggestion": "添加 godoc 风格的函数注释"
    }
  ]
}
` + "```" + `

以上是本次代码审查的结果。
`

	// Step 1: 使用 Parser 解析 LLM 响应
	reviewResp, err := parseReviewResponse(llmResponse)
	if err != nil {
		t.Fatalf("parseReviewResponse() error = %v", err)
	}

	// 验证解析结果
	t.Run("Parser_ExtractsCorrectly", func(t *testing.T) {
		if reviewResp.Summary == "" {
			t.Error("Summary should not be empty")
		}
		if !strings.Contains(reviewResp.Summary, "代码整体质量良好") {
			t.Errorf("Summary = %q, want to contain '代码整体质量良好'", reviewResp.Summary)
		}
		if reviewResp.Score != 72 {
			t.Errorf("Score = %d, want 72", reviewResp.Score)
		}
		if len(reviewResp.Suggestions) != 3 {
			t.Errorf("len(Suggestions) = %d, want 3", len(reviewResp.Suggestions))
		}
	})

	// 验证 Severity 标准化
	t.Run("Parser_NormalizesSeverity", func(t *testing.T) {
		if reviewResp.Suggestions[0].Severity != "high" {
			t.Errorf("Suggestions[0].Severity = %q, want 'high' (normalized from 'Major')",
				reviewResp.Suggestions[0].Severity)
		}
		if reviewResp.Suggestions[1].Severity != "medium" {
			t.Errorf("Suggestions[1].Severity = %q, want 'medium' (normalized from 'warning')",
				reviewResp.Suggestions[1].Severity)
		}
		if reviewResp.Suggestions[2].Severity != "low" {
			t.Errorf("Suggestions[2].Severity = %q, want 'low' (normalized from 'hint')",
				reviewResp.Suggestions[2].Severity)
		}
	})

	// 验证 Category 标准化
	t.Run("Parser_NormalizesCategory", func(t *testing.T) {
		if reviewResp.Suggestions[0].Category != "security" {
			t.Errorf("Suggestions[0].Category = %q, want 'security' (normalized from 'sec')",
				reviewResp.Suggestions[0].Category)
		}
		if reviewResp.Suggestions[1].Category != "performance" {
			t.Errorf("Suggestions[1].Category = %q, want 'performance' (normalized from 'perf')",
				reviewResp.Suggestions[1].Category)
		}
		if reviewResp.Suggestions[2].Category != "documentation" {
			t.Errorf("Suggestions[2].Category = %q, want 'documentation' (normalized from 'doc')",
				reviewResp.Suggestions[2].Category)
		}
	})

	// Step 2: 使用 OutputFormatter 生成完整 JSON 输出
	ctx := OutputContext{
		Repository: ContextRepository{
			ID:             1,
			Name:           "handsoff",
			FullName:       "company/handsoff",
			Platform:       "gitlab",
			PlatformRepoID: 12345,
		},
		MergeRequest: ContextMergeRequest{
			ID:           100,
			IID:          42,
			Title:        "feat: 添加用户认证模块",
			Author:       "developer",
			SourceBranch: "feature/auth",
			TargetBranch: "main",
			WebURL:       "https://gitlab.example.com/company/handsoff/-/merge_requests/42",
		},
		Review: ContextReview{
			ID:          1,
			ReviewedAt:  time.Now(),
			LLMProvider: "deepseek",
			LLMModel:    "deepseek-coder",
			TokensUsed:  1500,
			DurationMs:  2500,
		},
	}

	meta := OutputMetadata{
		PromptTemplate:     "default",
		PromptVersion:      "1.0",
		CustomPromptUsed:   false,
		RawResponseAvail:   true,
		ParserFallbackUsed: false,
	}

	jsonOutput, err := FormatReviewAsJSON(reviewResp, ctx, meta)
	if err != nil {
		t.Fatalf("FormatReviewAsJSON() error = %v", err)
	}

	// Step 3: 验证最终 JSON 输出
	t.Run("OutputFormatter_GeneratesValidJSON", func(t *testing.T) {
		var output ReviewOutputJSON
		if err := json.Unmarshal([]byte(jsonOutput), &output); err != nil {
			t.Fatalf("Failed to unmarshal output JSON: %v", err)
		}

		if output.SchemaVersion != "1.0" {
			t.Errorf("SchemaVersion = %q, want '1.0'", output.SchemaVersion)
		}
		if output.Result.QualityLevel != "acceptable" {
			t.Errorf("QualityLevel = %q, want 'acceptable' (score=72)", output.Result.QualityLevel)
		}
		if output.Statistics.TotalIssues != 3 {
			t.Errorf("TotalIssues = %d, want 3", output.Statistics.TotalIssues)
		}
		if output.Context.Repository.Name != "handsoff" {
			t.Errorf("Repository.Name = %q, want 'handsoff'", output.Context.Repository.Name)
		}
	})

	// 打印最终 JSON 输出供人工检查
	t.Run("OutputFormatter_PrintsSampleJSON", func(t *testing.T) {
		t.Logf("\n=== 生成的 JSON 输出示例 ===\n%s\n", jsonOutput)
	})
}

// TestIntegration_PlainTextFallback 测试纯文本 fallback 解析
func TestIntegration_PlainTextFallback(t *testing.T) {
	plainTextResponse := `
## 代码审查总结

Summary: 这段代码实现了基本功能，但有一些需要改进的地方。

Score: 65

### 问题列表：

1. 函数命名不够清晰，建议使用更具描述性的名称
2. 缺少错误处理，可能导致程序崩溃
3. 代码注释不足，难以理解业务逻辑
`

	reviewResp, err := parseReviewResponse(plainTextResponse)
	if err != nil {
		t.Fatalf("parseReviewResponse() error = %v", err)
	}

	if reviewResp.Score != 65 {
		t.Errorf("Score = %d, want 65", reviewResp.Score)
	}
	t.Logf("Fallback parsing result: Summary=%q, Score=%d, Suggestions=%d",
		reviewResp.Summary, reviewResp.Score, len(reviewResp.Suggestions))
}

// TestIntegration_EdgeCases 测试边界情况
func TestIntegration_EdgeCases(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		checkFn func(*ReviewResponse) bool
	}{
		{
			name:  "Score out of range (negative)",
			input: `{"summary": "test", "score": -50}`,
			checkFn: func(r *ReviewResponse) bool {
				return r.Score == 0
			},
		},
		{
			name:  "Score out of range (over 100)",
			input: `{"summary": "test", "score": 150}`,
			checkFn: func(r *ReviewResponse) bool {
				return r.Score == 100
			},
		},
		{
			name:  "Nested JSON in markdown",
			input: "Some text\n```json\n{\"summary\": \"nested\", \"score\": 80}\n```\nMore text",
			checkFn: func(r *ReviewResponse) bool {
				return r.Summary == "nested" && r.Score == 80
			},
		},
		{
			name:  "Chinese content",
			input: `{"summary": "代码质量优秀，无重大问题", "score": 95, "suggestions": []}`,
			checkFn: func(r *ReviewResponse) bool {
				return strings.Contains(r.Summary, "代码质量优秀") && r.Score == 95
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := parseReviewResponse(tc.input)
			if err != nil {
				t.Fatalf("parseReviewResponse() error = %v", err)
			}
			if !tc.checkFn(resp) {
				t.Errorf("Check failed: got Score=%d, Summary=%q", resp.Score, resp.Summary)
			}
		})
	}
}

// TestIntegration_StatisticsCalculation 测试统计计算
func TestIntegration_StatisticsCalculation(t *testing.T) {
	suggestions := []FixSuggestion{
		{FilePath: "main.go", Severity: "critical", Category: "security"},
		{FilePath: "main.go", Severity: "high", Category: "performance"},
		{FilePath: "utils.go", Severity: "medium", Category: "style"},
		{FilePath: "handler.go", Severity: "low", Category: "documentation"},
	}

	stats := CalculateOutputStatistics(suggestions)

	if stats.TotalIssues != 4 {
		t.Errorf("TotalIssues = %d, want 4", stats.TotalIssues)
	}
	if stats.BySeverity.Critical != 1 {
		t.Errorf("BySeverity.Critical = %d, want 1", stats.BySeverity.Critical)
	}
	if stats.BySeverity.High != 1 {
		t.Errorf("BySeverity.High = %d, want 1", stats.BySeverity.High)
	}
	if stats.BySeverity.Medium != 1 {
		t.Errorf("BySeverity.Medium = %d, want 1", stats.BySeverity.Medium)
	}
	if stats.BySeverity.Low != 1 {
		t.Errorf("BySeverity.Low = %d, want 1", stats.BySeverity.Low)
	}
	if stats.FilesAffected != 3 {
		t.Errorf("FilesAffected = %d, want 3", stats.FilesAffected)
	}

	t.Logf("Statistics: %+v", stats)
}
