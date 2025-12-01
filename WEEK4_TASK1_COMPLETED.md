# Week 4 Task 1: Webhook接收和解析 ✅

**完成时间**: 2025-12-01  
**状态**: 已完成

---

## 任务概览

实现GitLab Webhook接收接口，解析MR事件，验证签名，并创建异步Review任务。

---

## 已完成功能

### 1. Webhook事件解析 (`internal/webhook/gitlab_event.go`)

- ✅ **GitLabMergeRequestEvent**: 完整的GitLab MR事件结构体
- ✅ **ShouldTriggerReview()**: 智能判断是否触发Review（仅open/update动作）
- ✅ **辅助方法**: GetMRID, GetProjectID, GetMRTitle等

**支持的事件字段**:
```go
- ObjectKind: "merge_request"
- User: 用户信息（Name, Username, Email）
- Project: 项目信息（ID, Name, WebURL）
- ObjectAttributes: MR属性（ID, Title, SourceBranch, TargetBranch, Action, State）
- LastCommit: 最新提交信息
```

**触发条件**:
```go
- Action: "open" 或 "update"
- State: "opened"
- 跳过: merge, close, reopen
```

---

### 2. Webhook签名验证 (`internal/webhook/validator.go`)

- ✅ **ValidateGitLabSignature()**: GitLab Token验证（简单Token比对）
- ✅ **ValidateGitHubSignature()**: GitHub HMAC-SHA256验证（预留）
- ✅ **开发模式**: Secret为空时跳过验证

**安全机制**:
```go
// GitLab: 使用X-Gitlab-Token header
if receivedToken != expectedSecret {
    return error
}

// GitHub: 使用HMAC-SHA256（未来支持）
mac := hmac.New(sha256.New, []byte(secret))
```

---

### 3. 异步任务类型定义 (`internal/task/types.go`)

- ✅ **CodeReviewPayload**: Review任务载荷
- ✅ **AutoFixPayload**: 自动修复任务载荷（预留）
- ✅ **ToJSON/FromJSON**: 序列化方法

**CodeReviewPayload字段**:
```go
RepositoryID   uint
MergeRequestID int64
MRTitle        string
MRAuthor       string
SourceBranch   string
TargetBranch   string
MRWebURL       string
ProjectID      int64
```

---

### 4. Webhook Handler (`internal/api/handler/webhook.go`)

#### 核心功能

- ✅ **HandleWebhook()**: 智能路由（根据Header识别平台）
- ✅ **HandleGitLab()**: 完整的GitLab Webhook处理流程

#### 处理流程

```
1. 读取请求Body
   ↓
2. 解析事件类型（仅处理merge_request）
   ↓
3. 根据platform_repo_id查找仓库
   ↓
4. 验证Webhook Secret（如果配置）
   ↓
5. 检查是否触发Review（ShouldTriggerReview）
   ↓
6. 检查LLM模型配置
   ↓
7. 创建review_results记录（状态：pending）
   ↓
8. 构建CodeReviewPayload
   ↓
9. 入队Asynq任务（队列：default，重试：3次）
   ↓
10. 返回成功响应（含review_id和task_id）
```

#### 错误处理

- ❌ **无效Payload**: 返回400 Bad Request
- ❌ **仓库未找到**: 返回404 Not Found
- ❌ **签名验证失败**: 返回401 Unauthorized
- ❌ **LLM未配置**: 返回200（跳过处理）
- ❌ **任务入队失败**: 返回500 + 更新review_results状态为failed

---

### 5. 路由配置更新 (`internal/api/router/router.go`)

- ✅ **初始化Queue Client**: `queueClient := queue.NewClient(cfg.Redis)`
- ✅ **注册Webhook Handler**: `webhookHandler := handler.NewWebhookHandler(db, log, queueClient)`
- ✅ **添加路由**: `POST /api/webhook`（公开接口，不需要JWT认证）

```go
// Webhook routes (public, but with signature verification)
webhook := r.Group("/api/webhook")
{
    webhook.POST("", webhookHandler.HandleWebhook)
}
```

---

### 6. 数据模型更新 (`internal/model/repository.go`)

- ✅ **新增字段**: `WebhookSecret string` (size:255, json:"-")
- ✅ **隐藏敏感信息**: 不在JSON响应中暴露Secret

---

## API接口

### POST /api/webhook

**描述**: 接收GitLab Webhook事件

**请求头**:
```
X-Gitlab-Token: <webhook_secret>
X-Gitlab-Event: Merge Request Hook
Content-Type: application/json
```

**请求体**: GitLab MR Webhook Payload (JSON)

**响应示例（成功）**:
```json
{
  "message": "Webhook received and review task enqueued",
  "review_id": 123,
  "task_id": "asynq-task-uuid"
}
```

**响应示例（跳过）**:
```json
{
  "message": "Event does not trigger review"
}
```

**响应示例（错误）**:
```json
{
  "error": "Invalid webhook signature"
}
```

---

## 技术亮点

### 1. 智能事件过滤

```go
func (e *GitLabMergeRequestEvent) ShouldTriggerReview() bool {
    action := e.ObjectAttributes.Action
    state := e.ObjectAttributes.State
    return (action == "open" || action == "update") && state == "opened"
}
```

**优点**:
- 仅处理有意义的MR事件
- 避免重复Review（merge/close不触发）
- 节省LLM调用成本

---

### 2. 异步任务解耦

使用Asynq队列将Webhook接收与Review处理解耦：

```go
taskInfo, err := h.queue.Enqueue(
    asynq.NewTask(task.TypeCodeReview, payloadBytes),
    asynq.Queue("default"),
    asynq.MaxRetry(3),
)
```

**优点**:
- Webhook快速响应（200ms内）
- Review处理可靠（支持重试）
- 支持并发控制（Worker数量）

---

### 3. 安全设计

- **签名验证**: 防止伪造Webhook请求
- **Secret隐藏**: `json:"-"` 不暴露WebhookSecret
- **开发模式**: Secret为空时跳过验证（便于测试）

---

## 文件清单

### 新增文件 (4个)

```
internal/webhook/gitlab_event.go    # GitLab事件结构体（156行）
internal/webhook/validator.go       # Webhook签名验证（63行）
internal/task/types.go               # 任务载荷定义（48行）
internal/api/handler/webhook.go     # Webhook处理器（224行）
```

### 修改文件 (2个)

```
internal/api/router/router.go       # 添加Queue Client和Webhook路由
internal/model/repository.go        # 新增WebhookSecret字段
```

**总计新增代码**: ~500行

---

## 依赖包

### 已有依赖（无需新增）

```
github.com/gin-gonic/gin            # HTTP框架
github.com/hibiken/asynq            # 异步任务队列
gorm.io/gorm                        # ORM
```

---

## 测试场景

### 场景1: 接收GitLab MR Open事件

**前置条件**:
- 仓库已导入并配置LLM模型
- Redis服务运行中

**步骤**:
1. 创建GitLab MR
2. GitLab自动发送Webhook到 `/api/webhook`
3. 系统接收并解析事件
4. 验证Webhook Secret
5. 创建review_results记录
6. 入队Asynq任务

**预期结果**:
- 返回200 OK
- `review_results`表新增记录（status=pending）
- Asynq队列新增任务
- 日志记录task_id

---

### 场景2: 拒绝未配置LLM的仓库

**步骤**:
1. 导入仓库但不配置LLM模型
2. 发送MR Webhook

**预期结果**:
- 返回200 OK
- 消息: "No LLM model configured"
- 不创建review记录
- 不入队任务

---

### 场景3: 验证Webhook签名

**步骤**:
1. 发送Webhook但使用错误的X-Gitlab-Token

**预期结果**:
- 返回401 Unauthorized
- 错误: "Invalid webhook signature"
- 日志记录验证失败

---

### 场景4: 忽略非MR事件

**步骤**:
1. 发送Push Event Webhook

**预期结果**:
- 返回200 OK
- 消息: "Event type not supported"
- 日志记录跳过事件

---

## 日志示例

### 成功处理

```
INFO  Enqueued code review task
      task_id=asynq-12345
      review_id=123
      repository_id=5
      mr_id=42
      queue=default
```

### 签名验证失败

```
WARN  Webhook signature validation failed
      error=invalid webhook token
      repository_id=5
      mr_id=42
```

### 跳过事件

```
INFO  Event does not trigger review
      action=merge
      state=merged
      mr_id=42
```

---

## 数据库变更

### repositories表新增字段

```sql
ALTER TABLE repositories ADD COLUMN webhook_secret VARCHAR(255);
```

**说明**: GORM会在启动时自动迁移

---

## 性能考虑

### 1. Webhook响应时间

- **目标**: < 200ms
- **实现**: 
  - 仅创建数据库记录
  - 任务处理异步化
  - 无阻塞操作

### 2. 数据库查询优化

```go
h.db.Preload("LLMModel").
    Where("platform_repo_id = ? AND is_active = ?", projectID, true).
    First(&repo)
```

**优化点**:
- 使用索引（platform_repo_id, is_active）
- Preload避免N+1查询

### 3. 任务队列配置

```go
asynq.Queue("default")     // 队列名称
asynq.MaxRetry(3)          // 最大重试次数
```

---

## 安全检查清单

- [x] Webhook签名验证
- [x] Secret不暴露在JSON中
- [x] 仓库状态检查（is_active）
- [x] 事件类型白名单（仅merge_request）
- [x] 输入验证（JSON解析）
- [x] 错误日志记录

---

## 已知限制（MVP版本）

1. **仅支持GitLab**: GitHub/Gitea支持预留但未实现
2. **固定队列**: 所有任务使用"default"队列
3. **无速率限制**: 未实现Webhook调用频率限制
4. **Secret未加密**: WebhookSecret明文存储（建议后续加密）

---

## 与其他任务的集成

### Task 2: Asynq任务队列（下一步）

- 需要实现Worker处理`TypeCodeReview`任务
- 从队列中获取`CodeReviewPayload`
- 调用Review处理逻辑

### Task 3: LLM客户端（下一步）

- Worker调用LLM Client执行Review
- 解析LLM响应并保存结果

---

## 代码质量

- ✅ **分层架构**: Handler → Task → Worker（解耦）
- ✅ **错误处理**: 完善的错误日志和响应
- ✅ **类型安全**: 完整的结构体定义
- ✅ **日志记录**: 关键操作日志
- ✅ **代码风格**: 与现有代码一致

---

## 下一步（Task 2）

### Asynq Worker实现

- [ ] 创建`internal/task/server.go` - Worker服务器
- [ ] 创建`internal/task/review_handler.go` - Review任务处理器
- [ ] 更新`cmd/worker/main.go` - Worker启动逻辑
- [ ] 实现任务处理流程（获取Diff → 调用LLM → 保存结果）

---

**✅ Task 1 完成！准备进入Task 2：Asynq任务队列实现**

**编译状态**: ✅ 通过 (`go build -o bin/api ./cmd/api`)
