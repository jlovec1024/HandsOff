package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/handsoff/handsoff/internal/gitlab"
	"github.com/handsoff/handsoff/internal/llm"
	"github.com/handsoff/handsoff/internal/model"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	fmt.Println("==============================================")
	fmt.Println("HandsOff 组件单元测试")
	fmt.Println("==============================================\n")

	// 1. 测试数据库连接
	testDatabase()

	// 2. 测试 Redis 连接
	testRedis()

	// 3. 测试 GitLab Client
	testGitLabClient()

	// 4. 测试 LLM Client
	testLLMClient()

	fmt.Println("\n==============================================")
	fmt.Println("✅ 所有测试完成")
	fmt.Println("==============================================")
}

// testDatabase 测试数据库连接
func testDatabase() {
	fmt.Println("📦 [1/4] 测试数据库连接...")

	dbPath := getEnv("DB_PATH", "./data/handsoff.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}

	// 测试查询
	var count int64
	db.Model(&model.Repository{}).Count(&count)

	fmt.Printf("   ✅ 数据库连接成功 (路径: %s)\n", dbPath)
	fmt.Printf("   📊 Repositories 表记录数: %d\n\n", count)
}

// testRedis 测试 Redis 连接
func testRedis() {
	fmt.Println("🔴 [2/4] 测试 Redis 连接...")

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")

	// 创建 Asynq 客户端
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: redisAddr,
	})
	defer client.Close()

	// 测试创建任务
	payload := map[string]interface{}{
		"test": "ping",
	}
	payloadBytes, _ := json.Marshal(payload)

	task := asynq.NewTask("test:ping", payloadBytes)
	info, err := client.Enqueue(task)
	if err != nil {
		log.Fatalf("❌ Redis 连接失败: %v", err)
	}

	fmt.Printf("   ✅ Redis 连接成功 (地址: %s)\n", redisAddr)
	fmt.Printf("   📋 测试任务已入队: %s\n\n", info.ID)
}

// testGitLabClient 测试 GitLab Client
func testGitLabClient() {
	fmt.Println("🦊 [3/4] 测试 GitLab Client...")

	baseURL := getEnv("TEST_GITLAB_URL", "")
	accessToken := getEnv("TEST_GITLAB_TOKEN", "")
	projectID := getEnv("TEST_GITLAB_PROJECT_ID", "0")
	mrIID := getEnv("TEST_GITLAB_MR_IID", "0")

	if baseURL == "" || accessToken == "" {
		fmt.Println("   ⚠️  跳过 GitLab 测试 (未配置 TEST_GITLAB_URL 或 TEST_GITLAB_TOKEN)")
		fmt.Println("   提示: 在 .env 中设置以下变量以启用测试:")
		fmt.Println("   - TEST_GITLAB_URL=https://gitlab.com")
		fmt.Println("   - TEST_GITLAB_TOKEN=glpat-xxxxxxxxxxxx")
		fmt.Println("   - TEST_GITLAB_PROJECT_ID=12345")
		fmt.Println("   - TEST_GITLAB_MR_IID=1\n")
		return
	}

	client := gitlab.NewClient(baseURL, accessToken)

	// 测试连接
	if err := client.TestConnection(); err != nil {
		log.Printf("   ❌ GitLab 连接失败: %v\n\n", err)
		return
	}

	fmt.Printf("   ✅ GitLab 连接成功 (URL: %s)\n", baseURL)

	// 测试获取 MR Diff (如果配置了)
	if projectID != "0" && mrIID != "0" {
		var pid, iid int
		fmt.Sscanf(projectID, "%d", &pid)
		fmt.Sscanf(mrIID, "%d", &iid)

		diff, err := client.GetMRDiff(pid, iid)
		if err != nil {
			log.Printf("   ⚠️  获取 MR Diff 失败: %v\n", err)
		} else {
			fmt.Printf("   ✅ 成功获取 MR Diff (大小: %d 字节)\n", len(diff))
			if len(diff) > 200 {
				fmt.Printf("   预览: %s...\n", diff[:200])
			}
		}
	}

	fmt.Println()
}

// testLLMClient 测试 LLM Client
func testLLMClient() {
	fmt.Println("🤖 [4/4] 测试 LLM Client...")

	provider := getEnv("TEST_LLM_PROVIDER", "")
	apiKey := getEnv("TEST_LLM_API_KEY", "")
	encryptionKey := getEnv("ENCRYPTION_KEY", "")

	if provider == "" || apiKey == "" {
		fmt.Println("   ⚠️  跳过 LLM 测试 (未配置 TEST_LLM_PROVIDER 或 TEST_LLM_API_KEY)")
		fmt.Println("   提示: 在 .env 中设置以下变量以启用测试:")
		fmt.Println("   - TEST_LLM_PROVIDER=deepseek")
		fmt.Println("   - TEST_LLM_API_KEY=sk-xxxxxxxx")
		fmt.Println("   - ENCRYPTION_KEY=your-encryption-key\n")
		return
	}

	// 构造测试用的 Provider 和 Model
	testProvider := &model.LLMProvider{
		Type:    provider,
		BaseURL: getDefaultEndpoint(provider),
		APIKey:  apiKey, // 使用明文 (测试用)
	}

	testModel := &model.LLMModel{
		ModelName:   getDefaultModel(provider),
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	// 创建 LLM Client
	client, err := llm.NewClient(testProvider, testModel, encryptionKey)
	if err != nil {
		log.Printf("   ❌ LLM Client 创建失败: %v\n\n", err)
		return
	}

	fmt.Printf("   ✅ LLM Client 创建成功 (Provider: %s)\n", provider)

	// 测试连接 (发送简单请求)
	testDiff := `diff --git a/test.go b/test.go
index 1234567..abcdefg 100644
--- a/test.go
+++ b/test.go
@@ -1,3 +1,5 @@
 package main

-func main() {}
+func main() {
+    println("Hello, World!")
+}
`

	promptData := llm.BuildPromptData(testDiff, "Test MR", "tester", "feature", "main")
	prompt := llm.RenderPrompt(llm.GetDefaultPrompt(), promptData)

	req := llm.ReviewRequest{
		Diff:        testDiff,
		Prompt:      prompt,
		MaxTokens:   testModel.MaxTokens,
		Temperature: testModel.Temperature,
		ModelName:   testModel.ModelName,
	}

	fmt.Println("   🔄 发送测试请求到 LLM API...")
	start := time.Now()

	resp, err := client.Review(req)
	if err != nil {
		log.Printf("   ❌ LLM API 调用失败: %v\n\n", err)
		return
	}

	duration := time.Since(start)

	fmt.Printf("   ✅ LLM API 调用成功\n")
	fmt.Printf("   ⏱️  耗时: %.2f 秒\n", duration.Seconds())
	fmt.Printf("   📊 Tokens 使用: %d\n", resp.TokensUsed)
	fmt.Printf("   📝 Summary: %s\n", truncate(resp.Summary, 100))
	fmt.Printf("   🔍 建议数量: %d\n\n", len(resp.Suggestions))
}

// 辅助函数
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDefaultEndpoint(provider string) string {
	switch provider {
	case "openai":
		return "https://api.openai.com/v1"
	case "deepseek":
		return "https://api.deepseek.com/v1"
	default:
		return ""
	}
}

func getDefaultModel(provider string) string {
	switch provider {
	case "openai":
		return "gpt-3.5-turbo"
	case "deepseek":
		return "deepseek-chat"
	default:
		return ""
	}
}

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}
