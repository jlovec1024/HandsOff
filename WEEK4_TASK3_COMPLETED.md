# Week 4 Task 3: LLM客户端实现 ✅

**完成时间**: 2025-12-01  
**状态**: 已完成

---

## 任务概览

实现完整的LLM客户端系统，支持OpenAI和DeepSeek API，提供统一的代码审查接口和提示词模板系统。

---

## 已完成功能

### 1. LLM核心类型定义 (`internal/llm/types.go`)

#### ✅ ReviewRequest - 审查请求

```go
type ReviewRequest struct {
    Diff         string  // Git diff content
    Prompt       string  // Rendered prompt template
    MaxTokens    int     // Maximum tokens
    Temperature  float32 // Sampling temperature
    ModelName    string  // Model identifier
}
```

#### ✅ ReviewResponse - 审查响应

```go
type ReviewResponse struct {
    Summary     string           // Overall summary
    Score       int              // Quality score 0-100
    Suggestions []FixSuggestion  // Fix suggestions list
    RawResponse string           // Original LLM response
    ModelUsed   string           // Model name
    TokensUsed  int              // Tokens consumed
    Duration    time.Duration    // Time taken
}
```

#### ✅ Client接口

```go
type Client interface {
    Review(req ReviewRequest) (*ReviewResponse, error)
    TestConnection() error
    GetProviderName() string
}
```

---

### 2. 客户端工厂 (`internal/llm/client.go`)

#### ✅ NewClient - 根据Provider类型创建客户端

**支持的Provider类型**:
- `openai` → OpenAIClient
- `deepseek` → DeepSeekClient
- `claude` → 预留（未实现）
- `gemini` → 预留（未实现）
- `ollama` → 预留（未实现）

**功能**:
- API Key自动解密
- 配置参数映射
- 类型检查和验证

---

### 3. OpenAI适配器 (`internal/llm/openai.go`)

#### ✅ API集成

**Endpoint**: `POST {baseURL}/chat/completions`

**Request结构**:
```go
{
    "model": "gpt-4",
    "messages": [
        {"role": "system", "content": "..."},
        {"role": "user", "content": "..."}
    ],
    "max_tokens": 4096,
    "temperature": 0.7
}
```

**Response解析**:
- 提取content from choices[0].message.content
- 记录tokens使用量
- 处理API错误

#### ✅ 核心方法

**Review()**: 执行代码审查
- 构造OpenAI API请求
- 发送HTTP POST
- 解析JSON响应
- 调用parseReviewResponse处理结果

**TestConnection()**: 测试连接
- 发送简单测试消息
- 验证API Key有效性
- 检查HTTP状态码

---

### 4. DeepSeek适配器 (`internal/llm/deepseek.go`)

#### ✅ OpenAI兼容实现

**说明**: DeepSeek API与OpenAI API完全兼容

**差异**:
- BaseURL: https://api.deepseek.com
- Model名称: deepseek-chat, deepseek-coder等
- 其他结构完全相同

**代码复用**: 使用相同的API结构体

---

### 5. LLM响应解析器 (`internal/llm/parser.go`)

#### ✅ parseReviewResponse - 智能解析

**支持格式**:
1. **JSON格式** (推荐)
   ```json
   {
     "summary": "...",
     "score": 75,
     "suggestions": [...]
   }
   ```

2. **Markdown JSON块**
   ````markdown
   ```json
   {...}
   ```
   ````

3. **纯文本格式** (Fallback)
   - 使用正则表达式提取摘要
   - 智能判断分数
   - 解析列表项为建议

#### ✅ 辅助函数

**extractJSONFromMarkdown()**: 提取JSON代码块
**extractSummary()**: 从文本提取摘要
**extractScore()**: 从文本提取分数
**extractSuggestions()**: 从文本提取建议列表

---

### 6. 提示词模板系统 (`internal/llm/prompt.go`)

#### ✅ DefaultPromptTemplate

**模板内容**:
```
Please review the following code changes and provide structured feedback.

## Code Changes (Git Diff)
{{.Diff}}

## Review Requirements
1. Analyze the code for:
   - Security vulnerabilities
   - Performance issues
   - Code quality
   - Best practices
   - Potential bugs

2. Provide feedback in JSON format
{
  "summary": "...",
  "score": 75,
  "suggestions": [...]
}

Please respond ONLY with valid JSON.
```

#### ✅ 模板渲染

**RenderPrompt()**: 替换模板变量
- `{{.Diff}}` → Git diff内容
- `{{.MRTitle}}` → MR标题
- `{{.MRAuthor}}` → 作者
- `{{.SourceBranch}}` → 源分支
- `{{.TargetBranch}}` → 目标分支

**BuildPromptData()**: 构建模板数据

---

### 7. ReviewHandler集成 (`internal/task/review_handler.go`)

#### ✅ 更新内容

**新增字段**: `encryptionKey string` - 用于解密API Key

**新增方法**:
- `performLLMReview()` - 执行LLM代码审查
- `getMRDiffPlaceholder()` - 临时Diff占位符（Task 4替换）

**HandleCodeReview更新**:
1. 获取MR Diff（占位符）
2. 调用performLLMReview()
3. 保存审查结果到review_results
4. 保存fix_suggestions到数据库
5. 记录详细日志

---

### 8. 加密工具更新 (`pkg/crypto/encrypt.go`)

#### ✅ 新增辅助函数

```go
// DecryptString - 使用Base64密钥解密
func DecryptString(ciphertext, keyBase64 string) (string, error)

// EncryptString - 使用Base64密钥加密  
func EncryptString(plaintext, keyBase64 string) (string, error)
```

**用途**: 简化LLM Client中的API Key解密

---

## 文件清单

### 新增文件 (6个)

```
internal/llm/types.go       # LLM核心类型（60行）
internal/llm/client.go      # 客户端工厂（48行）
internal/llm/openai.go      # OpenAI适配器（195行）
internal/llm/deepseek.go    # DeepSeek适配器（195行）
internal/llm/parser.go      # 响应解析器（216行）
internal/llm/prompt.go      # 提示词模板（92行）
```

**总计新增代码**: ~800行

### 修改文件 (3个)

```
internal/task/review_handler.go  # 集成LLM调用
internal/task/server.go           # 传递encryptionKey
pkg/crypto/encrypt.go             # 新增解密辅助函数
```

---

## 完整处理流程

### Webhook → Worker → LLM → Database

```
1. Webhook接收MR事件 (Task 1)
   ↓
2. 创建CodeReviewPayload任务
   ↓
3. Worker Server接收任务 (Task 2)
   ↓
4. HandleCodeReview()
   ├─ 加载Repository + LLMModel.Provider
   ├─ 创建review_results记录（status=processing）
   ├─ 获取MR Diff（占位符，Task 4实现）
   ├─ performLLMReview()
   │  ├─ 创建LLM Client（解密API Key）
   │  ├─ 渲染提示词模板
   │  ├─ 调用LLM API
   │  │  ├─ OpenAIClient.Review()
   │  │  │  ├─ 构造HTTP请求
   │  │  │  ├─ 发送到OpenAI/DeepSeek
   │  │  │  ├─ 解析JSON响应
   │  │  │  └─ parseReviewResponse()
   │  │  │     ├─ 尝试JSON解析
   │  │  │     ├─ 提取Markdown JSON块
   │  │  │     └─ Fallback文本解析
   │  │  └─ 返回ReviewResponse
   │  └─ 记录tokens和duration
   ├─ 更新review_results（status=completed, score, summary）
   ├─ 保存fix_suggestions到数据库
   └─ TODO: 发布评论到GitLab (Task 4)
```

---

## API调用示例

### OpenAI API

```http
POST https://api.openai.com/v1/chat/completions
Authorization: Bearer sk-xxx
Content-Type: application/json

{
  "model": "gpt-4",
  "messages": [
    {
      "role": "system",
      "content": "You are an expert code reviewer..."
    },
    {
      "role": "user",
      "content": "Please review the following code changes..."
    }
  ],
  "max_tokens": 4096,
  "temperature": 0.7
}
```

### DeepSeek API

```http
POST https://api.deepseek.com/v1/chat/completions
Authorization: Bearer sk-xxx
Content-Type: application/json

{
  "model": "deepseek-chat",
  "messages": [...],
  "max_tokens": 4096,
  "temperature": 0.7
}
```

---

## LLM响应格式

### 标准JSON响应

```json
{
  "summary": "Overall, the code quality is good. However, there are a few security concerns that need to be addressed.",
  "score": 75,
  "suggestions": [
    {
      "file_path": "example.go",
      "line_start": 10,
      "line_end": 15,
      "severity": "high",
      "category": "security",
      "description": "Potential SQL injection vulnerability",
      "suggestion": "Use parameterized queries instead of string concatenation",
      "code_snippet": "query := \"SELECT * FROM users WHERE id = \" + userID"
    },
    {
      "file_path": "example.go",
      "line_start": 20,
      "line_end": 20,
      "severity": "medium",
      "category": "performance",
      "description": "Inefficient loop implementation",
      "suggestion": "Consider using a map for O(1) lookup instead of O(n) iteration",
      "code_snippet": "for _, item := range items { ... }"
    }
  ]
}
```

---

## 错误处理

### 1. API Key解密失败

```go
if err := crypto.DecryptString(provider.APIKey, encryptionKey); err != nil {
    return fmt.Errorf("failed to decrypt API key: %w", err)
}
```

### 2. LLM API调用失败

```go
if err := llmClient.Review(reviewReq); err != nil {
    // Update review_results.status = "failed"
    // Update review_results.error_message
    return fmt.Errorf("LLM API call failed: %w", err)
}
```

### 3. 响应解析失败

```go
// JSON解析失败 → Fallback到文本解析
if err := json.Unmarshal(content, &reviewResp); err != nil {
    return parseTextResponse(content)
}
```

### 4. HTTP请求超时

```go
client: &http.Client{
    Timeout: config.Timeout * time.Second,  // 默认60秒
}
```

---

## 日志示例

### 成功流程

```
INFO  Starting LLM code review
      review_id=10
      repository=my-project
      mr_id=42
      llm_provider=deepseek

INFO  Calling LLM API provider=deepseek model=deepseek-chat

INFO  LLM review completed
      tokens_used=1500
      duration=3.5s
      suggestions=5

INFO  Saving fix suggestions count=5

INFO  Code review completed successfully
      review_id=10
      score=75
      suggestions_count=5
```

### 失败流程

```
ERROR LLM review failed
      error=failed to decrypt API key: invalid base64
      review_id=10

ERROR Task failed
      type=code_review
      task_id=abc-123
      duration=100ms
      error=LLM API call failed: 401 Unauthorized
```

---

## 性能指标

| 指标 | OpenAI | DeepSeek | 说明 |
|------|--------|----------|------|
| API延迟 | 2-5秒 | 1-3秒 | 取决于diff大小 |
| Tokens消耗 | 1000-3000 | 1000-3000 | 取决于代码量 |
| 超时设置 | 60秒 | 60秒 | 可配置 |
| 重试次数 | 3次 | 3次 | Asynq配置 |

---

## 安全考虑

### 1. API Key保护

- ✅ 数据库存储：AES-256加密
- ✅ 传输：HTTPS
- ✅ 使用：内存中临时解密
- ✅ 日志：不记录明文Key

### 2. 提示词注入防护

- ✅ 固定system prompt
- ✅ 用户输入仅在user message
- ✅ 模板变量转义

### 3. API限流

- ⚠️ 当前未实现
- 🔜 建议：添加rate limiting

---

## 已知限制（当前阶段）

1. **Diff来源**: 使用占位符，Task 4实现真实GitLab API
2. **提示词模板**: 仅支持默认模板，未来支持自定义
3. **支持的Provider**: 仅OpenAI和DeepSeek，其他预留
4. **评论发布**: Task 4实现
5. **错误重试**: 依赖Asynq，无LLM特定重试策略

---

## 测试场景

### 场景1: OpenAI成功Review

**前置条件**:
- LLM Provider: OpenAI
- Model: gpt-4
- API Key: 已加密配置

**步骤**:
1. 发送Webhook触发MR Review
2. Worker接收任务
3. 调用OpenAI API
4. 解析JSON响应
5. 保存到数据库

**预期结果**:
- review_results.status = "completed"
- review_results.score = 70-90
- fix_suggestions表有5-10条记录
- raw_result包含LLM原始响应

---

### 场景2: DeepSeek成功Review

**差异**: BaseURL和Model名称不同，其他流程相同

---

### 场景3: API Key解密失败

**模拟**: 修改.env的ENCRYPTION_KEY

**预期结果**:
- LLM review failed: failed to decrypt API key
- review_results.status = "failed"
- review_results.error_message记录错误

---

### 场景4: LLM返回非JSON格式

**模拟**: LLM返回纯文本

**预期结果**:
- parseTextResponse() fallback解析
- 提取summary, score, suggestions
- 成功保存（可能不完整）

---

### 场景5: API超时

**模拟**: 网络延迟>60秒

**预期结果**:
- HTTP client timeout
- 任务失败并重试（最多3次）

---

## 与其他Task集成

### ✅ Task 1集成

- Webhook创建任务 → Worker接收 → **LLM处理** ✅

### ✅ Task 2集成

- Worker调度任务 → **调用LLM Client** ✅

### 🔜 Task 4集成（下一步）

需要添加：
```go
// 1. 真实获取MR Diff
diff, err := gitlabClient.GetMRDiff(projectID, mrID)

// 2. 发布评论到GitLab
comment := formatReviewComment(reviewResp)
err = gitlabClient.PostComment(projectID, mrID, comment)
```

### ✅ Task 5集成（部分完成）

- 保存review_results ✅
- 保存fix_suggestions ✅
- （Task 5将优化存储逻辑）

---

## 代码质量

- ✅ **接口设计**: 统一Client接口，易扩展
- ✅ **错误处理**: 完善的错误包装和日志
- ✅ **解析鲁棒性**: JSON和文本双重解析
- ✅ **类型安全**: 完整的结构体定义
- ✅ **安全性**: API Key加密解密
- ✅ **可测试性**: 接口设计便于Mock

---

## 下一步（Task 4）

### GitLab集成实现

需要实现：
- [ ] 创建`internal/gitlab/client.go` - GitLab API客户端
- [ ] 实现`GetMRDiff()` - 获取MR差异
- [ ] 实现`PostComment()` - 发布评论到MR
- [ ] 格式化评论内容（Markdown表格）
- [ ] 在ReviewHandler中替换占位符
- [ ] 处理GitLab API错误和重试

---

## 总结

### Task 3成就

✅ **6个新文件** (~800行)  
✅ **统一LLM接口**  
✅ **OpenAI + DeepSeek适配器**  
✅ **智能响应解析**  
✅ **提示词模板系统**  
✅ **完整集成到ReviewHandler**  
✅ **编译成功无错误**  

### 进度更新

- Task 1: Webhook接收 ✅
- Task 2: Asynq队列 ✅
- Task 3: LLM客户端 ✅
- **总体进度: 3/8 Tasks完成 (37.5%)**

---

**✅ Task 3完成！LLM系统已就绪，准备进入Task 4：GitLab集成**

**编译状态**: ✅ 通过
- API: `go build -o bin/api ./cmd/api`
- Worker: `go build -o bin/worker ./cmd/worker`
