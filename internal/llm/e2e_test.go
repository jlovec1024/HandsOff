//go:build e2e
// +build e2e

// 端到端测试：真实 LLM 调用 + 真实 GitLab Diff
// 运行方式: go test -tags=e2e ./internal/llm/... -v -run E2E
//
// 需要设置环境变量:
//   export LLM_BASE_URL=https://api.deepseek.com/v1
//   export LLM_API_KEY=your-api-key
//   export LLM_MODEL=deepseek-coder
//   export GITLAB_URL=https://gitlab.example.com
//   export GITLAB_TOKEN=your-gitlab-token
//   export GITLAB_PROJECT_ID=123
//   export GITLAB_MR_IID=1

package llm

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/handsoff/handsoff/internal/gitlab"
)

func TestE2E_RealLLMWithRealGitLabDiff(t *testing.T) {
	// 检查必要的环境变量
	llmBaseURL := os.Getenv("LLM_BASE_URL")
	llmAPIKey := os.Getenv("LLM_API_KEY")
	llmModel := os.Getenv("LLM_MODEL")
	gitlabURL := os.Getenv("GITLAB_URL")
	gitlabToken := os.Getenv("GITLAB_TOKEN")
	gitlabProjectID := os.Getenv("GITLAB_PROJECT_ID")
	gitlabMRIID := os.Getenv("GITLAB_MR_IID")

	if llmBaseURL == "" || llmAPIKey == "" {
		t.Skip("跳过: 未设置 LLM_BASE_URL 或 LLM_API_KEY 环境变量")
	}
	if gitlabURL == "" || gitlabToken == "" {
		t.Skip("跳过: 未设置 GITLAB_URL 或 GITLAB_TOKEN 环境变量")
	}

	// 默认值
	if llmModel == "" {
		llmModel = "deepseek-coder"
	}

	t.Logf("=== 端到端测试配置 ===")
	t.Logf("LLM Provider: %s", llmBaseURL)
	t.Logf("LLM Model: %s", llmModel)
	t.Logf("GitLab URL: %s", gitlabURL)

	// Step 1: 获取真实的 GitLab MR Diff
	var diff string
	if gitlabProjectID != "" && gitlabMRIID != "" {
		t.Run("Step1_FetchRealGitLabDiff", func(t *testing.T) {
			var projectID, mrIID int
			fmt.Sscanf(gitlabProjectID, "%d", &projectID)
			fmt.Sscanf(gitlabMRIID, "%d", &mrIID)

			gitlabClient := gitlab.NewClient(gitlabURL, gitlabToken)
			var err error
			diff, err = gitlabClient.GetMRDiff(projectID, mrIID)
			if err != nil {
				t.Fatalf("获取 GitLab MR Diff 失败: %v", err)
			}

			t.Logf("✅ 成功获取 GitLab Diff")
			t.Logf("Diff 长度: %d 字符", len(diff))
			// 打印 diff 的前 500 字符预览
			preview := diff
			if len(preview) > 500 {
				preview = preview[:500] + "...(truncated)"
			}
			t.Logf("Diff 预览:\n%s", preview)
		})
	} else {
		// 使用示例 diff
		diff = `--- a/main.go
+++ b/main.go
@@ -10,6 +10,15 @@ import (
 )
 
 func main() {
+    // 新增用户认证
+    userId := r.URL.Query().Get("user_id")
+    query := "SELECT * FROM users WHERE id = " + userId  // 潜在 SQL 注入
+    
+    // 缓存处理
+    cache.Set(userId, userData)  // 未设置过期时间
+    
+    // 处理请求
     fmt.Println("Hello, World!")
 }
`
		t.Logf("⚠️ 使用示例 Diff (未设置 GITLAB_PROJECT_ID/GITLAB_MR_IID)")
	}

	// Step 2: 构建 Prompt
	var prompt string
	t.Run("Step2_BuildPrompt", func(t *testing.T) {
		prompt = buildReviewPrompt(diff)
		t.Logf("✅ Prompt 构建完成")
		t.Logf("Prompt 长度: %d 字符", len(prompt))
	})

	// Step 3: 调用真实 LLM
	var llmResponse string
	t.Run("Step3_CallRealLLM", func(t *testing.T) {
		config := Config{
			BaseURL:     llmBaseURL,
			APIKey:      llmAPIKey,
			ModelName:   llmModel,
			MaxTokens:   4096,
			Temperature: 0.3,
			Timeout:     120, // 120 秒超时
		}

		client := NewOpenAICompatibleClient("deepseek", config)

		t.Logf("正在调用 LLM API...")
		start := time.Now()

		resp, err := client.Review(ReviewRequest{
			Diff:        diff,
			Prompt:      prompt,
			MaxTokens:   4096,
			Temperature: 0.3,
			ModelName:   llmModel,
		})

		duration := time.Since(start)

		if err != nil {
			t.Fatalf("LLM 调用失败: %v", err)
		}

		llmResponse = resp.RawResponse
		t.Logf("✅ LLM 调用成功")
		t.Logf("耗时: %v", duration)
		t.Logf("Token 使用: %d", resp.TokensUsed)
		t.Logf("原始响应长度: %d 字符", len(llmResponse))
	})

	// Step 4: 解析 LLM 响应
	var reviewResp *ReviewResponse
	t.Run("Step4_ParseLLMResponse", func(t *testing.T) {
		var err error
		reviewResp, err = parseReviewResponse(llmResponse)
		if err != nil {
			t.Fatalf("解析 LLM 响应失败: %v", err)
		}

		t.Logf("✅ 响应解析成功")
		t.Logf("Summary: %s", reviewResp.Summary)
		t.Logf("Score: %d", reviewResp.Score)
		t.Logf("Suggestions 数量: %d", len(reviewResp.Suggestions))

		for i, sug := range reviewResp.Suggestions {
			t.Logf("  [%d] %s (%s/%s): %s",
				i+1, sug.FilePath, sug.Severity, sug.Category, sug.Description)
		}
	})

	// Step 5: 生成完整 JSON 输出
	t.Run("Step5_FormatAsJSON", func(t *testing.T) {
		ctx := OutputContext{
			Repository: ContextRepository{
				ID:       1,
				Name:     "test-repo",
				FullName: "company/test-repo",
				Platform: "gitlab",
			},
			MergeRequest: ContextMergeRequest{
				ID:           1,
				IID:          1,
				Title:        "Test MR",
				Author:       "developer",
				SourceBranch: "feature",
				TargetBranch: "main",
			},
			Review: ContextReview{
				ID:          1,
				ReviewedAt:  time.Now(),
				LLMProvider: "deepseek",
				LLMModel:    llmModel,
			},
		}

		meta := OutputMetadata{
			PromptTemplate:     "default",
			CustomPromptUsed:   false,
			RawResponseAvail:   true,
			ParserFallbackUsed: false,
		}

		jsonOutput, err := FormatReviewAsJSON(reviewResp, ctx, meta)
		if err != nil {
			t.Fatalf("JSON 格式化失败: %v", err)
		}

		// 验证 JSON 有效性
		var output ReviewOutputJSON
		if err := json.Unmarshal([]byte(jsonOutput), &output); err != nil {
			t.Fatalf("JSON 验证失败: %v", err)
		}

		t.Logf("✅ JSON 输出生成成功")
		t.Logf("Schema Version: %s", output.SchemaVersion)
		t.Logf("Quality Level: %s", output.Result.QualityLevel)
		t.Logf("Total Issues: %d", output.Statistics.TotalIssues)

		t.Logf("\n=== 完整 JSON 输出 ===\n%s", jsonOutput)
	})
}

// buildReviewPrompt 构建代码审查 prompt
func buildReviewPrompt(diff string) string {
	return fmt.Sprintf(`请对以下代码变更进行专业的代码审查。

## 代码变更 (Git Diff)
%s

## 输出要求
请以 JSON 格式返回审查结果，包含以下字段：

{
  "summary": "审查总结（中文，2-3句话）",
  "score": 0-100 的质量分数,
  "suggestions": [
    {
      "file_path": "文件路径",
      "line_start": 起始行号,
      "line_end": 结束行号,
      "severity": "critical/high/medium/low",
      "category": "security/performance/style/logic/documentation",
      "description": "问题描述",
      "suggestion": "改进建议",
      "code_snippet": "相关代码片段（可选）"
    }
  ]
}

## 审查重点
1. 安全问题（SQL注入、XSS、敏感信息泄露等）
2. 性能问题（N+1查询、内存泄漏、缓存问题等）
3. 代码逻辑错误
4. 代码风格和可维护性
5. 文档和注释

请直接返回 JSON，不要添加其他文字说明。`, diff)
}
