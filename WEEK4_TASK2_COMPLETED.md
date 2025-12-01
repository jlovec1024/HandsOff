# Week 4 Task 2: Asynq任务队列实现 ✅

**完成时间**: 2025-12-01  
**状态**: 已完成

---

## 任务概览

实现Asynq Worker服务器，处理从Webhook接收的异步Review任务。

---

## 已完成功能

### 1. Review任务处理器 (`internal/task/review_handler.go`)

#### ✅ ReviewHandler结构

```go
type ReviewHandler struct {
    db  *gorm.DB
    log Logger
}
```

#### ✅ 核心方法

**`HandleCodeReview()`** - 处理Code Review任务
- 解析CodeReviewPayload
- 加载Repository及关联关系（Platform, LLMModel）
- 验证LLM模型配置
- 创建/更新review_results记录
- 更新状态为processing → pending（等待LLM集成）

**特性**:
- ✅ 完整的错误处理
- ✅ 数据库事务支持
- ✅ 详细的日志记录
- ✅ 状态管理（pending → processing → completed/failed）

---

### 2. Worker Server (`internal/task/server.go`)

#### ✅ 队列配置

```go
Queues: map[string]int{
    "critical": 6,  // 高优先级
    "default":  3,  // 默认优先级
    "low":      1,  // 低优先级
}
```

#### ✅ 任务注册

```go
mux.HandleFunc(TypeCodeReview, reviewHandler.HandleCodeReview)
```

#### ✅ 并发控制

- 并发数由配置文件控制: `cfg.Worker.Concurrency`
- 支持多队列优先级处理
- 自动重试机制（最多3次）

---

### 3. 任务中间件 (`internal/task/middleware.go`)

#### ✅ LoggingMiddleware

**功能**:
- 记录任务开始时间
- 记录任务完成/失败
- 记录执行时长

**日志输出**:
```
INFO  Task started type=code_review task_id=xxx
INFO  Task completed type=code_review task_id=xxx duration=1.5s
ERROR Task failed type=code_review task_id=xxx duration=500ms error=...
```

#### ✅ RecoveryMiddleware

**功能**:
- 捕获panic避免Worker崩溃
- 记录panic详情
- 自动跳过重试（SkipRetry）

**安全机制**:
```go
defer func() {
    if r := recover() {
        log.Error("Task panicked", "panic", r)
        err = asynq.SkipRetry
    }
}()
```

---

### 4. Worker启动流程 (`cmd/worker/main.go`)

#### ✅ 启动步骤

```
1. 加载配置（config.Load）
   ↓
2. 初始化日志（logger.New）
   ↓
3. 连接数据库（database.New）
   ↓
4. 创建Queue Client（queue.NewClient）
   ↓
5. 创建Worker Server（task.NewServer）
   ↓
6. 启动任务处理（srv.Start）
   ↓
7. 等待信号（SIGINT/SIGTERM）
   ↓
8. 优雅关闭（srv.Shutdown）
```

#### ✅ 优雅关闭

- 监听SIGINT和SIGTERM信号
- 停止接收新任务
- 等待正在执行的任务完成
- 关闭Redis连接

---

## 处理流程

### 完整Review任务流程

```
Webhook接收 (Task 1)
    ↓
入队CodeReview任务
    ↓
Worker Server接收任务
    ↓
RecoveryMiddleware（捕获panic）
    ↓
LoggingMiddleware（记录开始）
    ↓
ReviewHandler.HandleCodeReview()
    ↓
1. 解析Payload
2. 加载Repository + LLMModel
3. 验证LLM配置
4. 创建/更新review_results（processing）
5. TODO: 调用LLM进行Review（Task 3）
6. 更新状态为pending（等待LLM）
    ↓
LoggingMiddleware（记录完成）
    ↓
任务完成
```

---

## 任务状态管理

### review_results状态流转

```
pending (Webhook创建)
    ↓
processing (Worker开始处理)
    ↓
pending (等待LLM - 当前Task 2实现)
    ↓
completed (Task 3实现后: LLM Review完成)
    或
failed (错误发生)
```

---

## 文件清单

### 新增文件 (3个)

```
internal/task/review_handler.go   # Review任务处理器（167行）
internal/task/middleware.go        # 任务中间件（60行）
internal/task/types.go             # 任务类型定义（48行 - Task 1创建）
```

### 修改文件 (2个)

```
internal/task/server.go            # 注册任务Handler和中间件
cmd/worker/main.go                 # Worker启动逻辑（已有，无需修改）
```

**总计新增代码**: ~275行（含中间件和处理器）

---

## 配置说明

### .env配置（Worker相关）

```bash
# Redis配置（Asynq依赖）
REDIS_URL=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Worker配置
WORKER_CONCURRENCY=10  # 并发处理任务数
```

### Worker并发数建议

| 场景 | 并发数 | 说明 |
|------|--------|------|
| 开发环境 | 5-10 | 便于调试 |
| 生产环境（小） | 10-20 | 单机部署 |
| 生产环境（大） | 20-50 | 多核服务器 |

---

## 启动命令

### 方式1: 直接运行

```bash
# 启动API服务器（Terminal 1）
go run cmd/api/main.go

# 启动Worker（Terminal 2）
go run cmd/worker/main.go

# 启动Redis（Terminal 3，如果未运行）
redis-server
```

### 方式2: 使用Makefile

```bash
# Terminal 1
make run-api

# Terminal 2
make run-worker
```

### 方式3: 编译后运行

```bash
# 编译
make build

# 运行
./bin/handsoff-api
./bin/handsoff-worker
```

---

## 日志示例

### Worker启动日志

```
INFO  Starting HandsOff Worker...
INFO  Registered task handlers handlers=[code_review]
INFO  Worker started concurrency=10
```

### 任务处理日志

```
INFO  Task started type=code_review task_id=abc-123
INFO  Processing code review task repository_id=5 mr_id=42 task_id=abc-123
INFO  Review result record found/created review_id=10 status=processing
INFO  Review task queued successfully (LLM integration pending) review_id=10
INFO  Code review task completed (placeholder) review_id=10 repository_id=5 mr_id=42
INFO  Task completed type=code_review task_id=abc-123 duration=150ms
```

### 错误日志

```
ERROR Task failed type=code_review task_id=abc-123 duration=50ms error=repository not found
```

### Panic恢复日志

```
ERROR Task panicked type=code_review task_id=abc-123 panic=runtime error: invalid memory address
```

---

## 性能特性

### 1. 异步处理

- Webhook接收 < 200ms
- 任务处理异步化，不阻塞HTTP请求

### 2. 并发控制

```go
Concurrency: cfg.Worker.Concurrency  // 同时处理N个任务
```

### 3. 队列优先级

```go
"critical": 6,  // 紧急任务
"default":  3,  // 普通任务
"low":      1,  // 低优先级任务
```

### 4. 重试机制

- 最多重试3次（在Task 1的Webhook Handler中配置）
- 指数退避策略
- Panic任务不重试（asynq.SkipRetry）

---

## 监控与调试

### Asynq Web UI（可选）

```bash
# 安装asynqmon
go install github.com/hibiken/asynqmon@latest

# 启动监控界面
asynqmon --redis-addr=localhost:6379

# 访问 http://localhost:8080
```

**功能**:
- 查看队列状态
- 查看任务详情
- 手动重试失败任务
- 查看任务执行历史

### 手动检查Redis队列

```bash
# 连接Redis
redis-cli

# 查看队列长度
LLEN asynq:queues:default

# 查看任务详情
LRANGE asynq:queues:default 0 -1
```

---

## 错误处理

### 1. 任务解析失败

```go
if err := json.Unmarshal(t.Payload(), &payload); err != nil {
    log.Error("Failed to unmarshal task payload", "error", err)
    return fmt.Errorf("failed to unmarshal payload: %w", err)
}
```

**处理**: 返回错误，任务进入重试队列

---

### 2. 仓库未找到

```go
if err := h.db.First(&repo, payload.RepositoryID).Error; err != nil {
    return fmt.Errorf("repository not found: %w", err)
}
```

**处理**: 返回错误，任务失败（可能是数据不一致）

---

### 3. LLM未配置

```go
if repo.LLMModel == nil {
    return fmt.Errorf("no LLM model configured for repository %d")
}
```

**处理**: 返回错误，提示配置LLM模型

---

### 4. 数据库错误

```go
if err := h.db.Create(&reviewResult).Error; err != nil {
    return fmt.Errorf("failed to create review result: %w", err)
}
```

**处理**: 返回错误，任务重试

---

## 安全考虑

### 1. Panic恢复

```go
defer func() {
    if r := recover() {
        log.Error("Task panicked", "panic", r)
        err = asynq.SkipRetry
    }
}()
```

**防止**: 单个任务panic导致整个Worker进程崩溃

---

### 2. 优雅关闭

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
srv.Shutdown()
```

**保证**: 正在处理的任务完成后再关闭

---

### 3. 资源限制

- 并发数控制避免过载
- 数据库连接池管理
- Redis连接复用

---

## 测试场景

### 场景1: 正常任务处理

**前置条件**:
- Redis运行中
- Worker启动
- 仓库已配置LLM模型

**步骤**:
1. 发送Webhook触发Review
2. 观察Worker日志
3. 检查数据库review_results表

**预期结果**:
- Worker接收任务
- 创建review_results记录
- 状态为pending（等待LLM）
- 任务完成无错误

---

### 场景2: LLM未配置

**步骤**:
1. 导入仓库但不配置LLM
2. 发送Webhook

**预期结果**:
- Webhook返回"No LLM model configured"
- 不创建任务
- Worker不处理

---

### 场景3: 任务重试

**模拟步骤**:
1. 暂停数据库
2. 发送Webhook（创建任务）
3. Worker尝试处理（失败）
4. 恢复数据库
5. Worker自动重试

**预期结果**:
- 第1-3次失败
- 第4次成功处理

---

### 场景4: Worker重启

**步骤**:
1. 创建10个任务
2. 关闭Worker（Ctrl+C）
3. 重新启动Worker

**预期结果**:
- Worker继续处理未完成任务
- 队列中的任务不丢失

---

## 与其他任务的集成

### Task 1: Webhook接收（已完成）

✅ **已集成**:
- Webhook创建CodeReviewPayload
- 入队到Asynq
- Worker接收并处理

---

### Task 3: LLM客户端（下一步）

🔜 **待集成**:
```go
// 在HandleCodeReview中添加
llmClient := llm.NewClient(repo.LLMModel.Provider)
result, err := llmClient.Review(diff, prompt)
```

---

### Task 4: GitLab集成（下一步）

🔜 **待集成**:
```go
// 获取MR Diff
diff, err := gitlabClient.GetMRDiff(projectID, mrID)

// 发布评论
err = gitlabClient.PostComment(projectID, mrID, comment)
```

---

### Task 5: Review结果存储（下一步）

🔜 **待集成**:
```go
// 保存fix_suggestions
for _, suggestion := range result.Suggestions {
    h.db.Create(&model.FixSuggestion{...})
}
```

---

## 已知限制（当前阶段）

1. **LLM未集成**: 当前仅创建review_results，不调用LLM
2. **Diff未获取**: 未从GitLab获取MR差异
3. **评论未发布**: 未向GitLab发布Review评论
4. **修复建议未保存**: 未创建fix_suggestions记录

**这些功能将在Task 3-5中实现**

---

## 代码质量

- ✅ **错误处理**: 完善的错误捕获和日志
- ✅ **中间件模式**: 可扩展的任务处理链
- ✅ **优雅关闭**: 信号处理和资源清理
- ✅ **日志记录**: 详细的执行日志
- ✅ **类型安全**: 完整的结构体定义
- ✅ **代码风格**: 与现有代码一致

---

## 下一步（Task 3）

### LLM客户端实现

需要实现：
- [ ] 创建`internal/llm/client.go` - LLM客户端接口
- [ ] 创建`internal/llm/deepseek.go` - DeepSeek适配器
- [ ] 创建`internal/llm/openai.go` - OpenAI适配器
- [ ] 创建`internal/llm/prompt.go` - 提示词模板渲染
- [ ] 在ReviewHandler中调用LLM Client
- [ ] 解析LLM响应并保存结果

---

**✅ Task 2 完成！准备进入Task 3：LLM客户端实现**

**编译状态**: ✅ 通过
- API: `go build -o bin/api ./cmd/api`
- Worker: `go build -o bin/worker ./cmd/worker`
