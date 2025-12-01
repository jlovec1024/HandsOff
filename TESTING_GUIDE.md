# HandsOff 测试指南

本文档提供完整的测试步骤，帮助你验证 HandsOff 系统的各个组件和完整流程。

---

## 📋 目录

1. [测试前准备](#测试前准备)
2. [单元测试 (组件级)](#单元测试-组件级)
3. [集成测试 (完整流程)](#集成测试-完整流程)
4. [验证清单](#验证清单)
5. [常见问题排查](#常见问题排查)

---

## 测试前准备

### 1. 系统要求

✅ **必需组件**:
- Go 1.22+
- Redis 6.0+
- SQLite 3 或 MySQL 5.7+

✅ **外部服务**:
- GitLab 实例 (gitlab.com 或私有部署)
- LLM API (OpenAI 或 DeepSeek)

### 2. 环境配置

#### Step 1: 复制环境变量模板

```bash
cp .env.example .env
```

#### Step 2: 编辑 `.env` 文件

**必须配置的项**:

```bash
# 数据库
DB_TYPE=sqlite
DB_DSN=data/handsoff.db

# Redis
REDIS_URL=redis://localhost:6379/0

# 加密密钥 (生成方式: openssl rand -base64 32)
ENCRYPTION_KEY=your-generated-32-byte-base64-key

# Worker
WORKER_CONCURRENCY=5
```

**测试用配置** (可选):

```bash
# GitLab 测试配置
TEST_GITLAB_URL=https://gitlab.com
TEST_GITLAB_TOKEN=glpat-xxxxxxxxxxxxxxxxxxxx
TEST_GITLAB_PROJECT_ID=12345
TEST_GITLAB_MR_IID=1

# LLM 测试配置
TEST_LLM_PROVIDER=deepseek
TEST_LLM_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

#### Step 3: 生成加密密钥

```bash
# 方法 1: 使用 OpenSSL
openssl rand -base64 32

# 方法 2: 使用 Python
python3 -c "import base64; import os; print(base64.b64encode(os.urandom(32)).decode())"

# 方法 3: 使用 Go
go run -c 'package main; import ("crypto/rand"; "encoding/base64"; "fmt"); func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(base64.StdEncoding.EncodeToString(b)) }'
```

**复制输出的值到 `.env` 的 `ENCRYPTION_KEY`**

### 3. 启动 Redis

```bash
# macOS (Homebrew)
brew services start redis

# Linux (systemd)
sudo systemctl start redis

# Docker
docker run -d -p 6379:6379 redis:latest

# 验证 Redis 运行
redis-cli ping  # 应返回 PONG
```

### 4. 初始化数据库

```bash
# 编译并运行 API (会自动初始化数据库)
go build -o bin/api ./cmd/api
./bin/api
```

**看到以下日志说明数据库初始化成功**:
```
INFO Database connected successfully
INFO Database migrated successfully
INFO API server starting on :8080
```

**按 `Ctrl+C` 停止**

### 5. 加密并配置 API Key

#### 方法 A: 使用加密工具 (推荐)

```bash
# 加密 LLM API Key
go run tools/encrypt_apikey/main.go -key "sk-your-deepseek-api-key"

# 输出示例:
# ✅ 加密成功
# 加密后的值: ABC123XYZ...
```

**复制加密后的值**，然后插入数据库：

```sql
-- 插入 LLM Provider (使用加密后的 API Key)
INSERT INTO llm_providers (name, type, api_endpoint, api_key, enabled, created_at, updated_at)
VALUES (
    'DeepSeek',
    'deepseek',
    'https://api.deepseek.com/v1',
    'YOUR_ENCRYPTED_API_KEY_HERE',  -- 粘贴上面的加密值
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);
```

#### 方法 B: 通过 API 创建 (自动加密)

```bash
# 启动 API Server
./bin/api

# 在另一个终端调用 API
curl -X POST http://localhost:8080/api/llm/providers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "DeepSeek",
    "type": "deepseek",
    "api_endpoint": "https://api.deepseek.com/v1",
    "api_key": "sk-your-actual-api-key"
  }'
```

### 6. 配置测试数据

**编辑 `scripts/test_data.sql`** 并替换以下值:

1. **LLM Provider API Key** - 使用上面加密的值
2. **GitLab Access Token** - 从 GitLab 获取 (Settings → Access Tokens)
3. **GitLab Project ID** - 你的测试项目 ID
4. **Repository 配置** - 项目名称和 Git URL

**执行 SQL 脚本**:

```bash
# SQLite
sqlite3 data/handsoff.db < scripts/test_data.sql

# MySQL
mysql -u handsoff -p handsoff < scripts/test_data.sql
```

---

## 单元测试 (组件级)

### 运行测试脚本

```bash
go run tools/test_components/main.go
```

### 预期输出

```
==============================================
HandsOff 组件单元测试
==============================================

📦 [1/4] 测试数据库连接...
   ✅ 数据库连接成功 (路径: ./data/handsoff.db)
   📊 Repositories 表记录数: 1

🔴 [2/4] 测试 Redis 连接...
   ✅ Redis 连接成功 (地址: localhost:6379)
   📋 测试任务已入队: abc-123

🦊 [3/4] 测试 GitLab Client...
   ✅ GitLab 连接成功 (URL: https://gitlab.com)
   ✅ 成功获取 MR Diff (大小: 2450 字节)

🤖 [4/4] 测试 LLM Client...
   ✅ LLM Client 创建成功 (Provider: deepseek)
   🔄 发送测试请求到 LLM API...
   ✅ LLM API 调用成功
   ⏱️  耗时: 3.50 秒
   📊 Tokens 使用: 1200
   📝 Summary: The code looks good overall...
   🔍 建议数量: 3

==============================================
✅ 所有测试完成
==============================================
```

### ✅ 单元测试检查清单

- [ ] 数据库连接成功
- [ ] Redis 连接成功，任务可入队
- [ ] GitLab API 认证通过
- [ ] GitLab 可获取 MR Diff
- [ ] LLM API 调用成功
- [ ] LLM 响应解析正常

**如果任何测试失败，请参考 [常见问题排查](#常见问题排查)**

---

## 集成测试 (完整流程)

### 阶段 1: 准备 GitLab 测试项目

#### 1.1 获取 GitLab Access Token

1. 登录 GitLab
2. **用户设置** → **Access Tokens**
3. 创建新 Token:
   - Name: `HandsOff Test`
   - Expiration: 设置一个未来日期
   - Scopes: 选择以下权限
     - ✅ `api`
     - ✅ `read_api`
     - ✅ `read_repository`
     - ✅ `write_repository`
4. **复制生成的 Token** (glpat-xxxxxxxxxxxx)

#### 1.2 获取 Project ID

1. 访问你的测试项目
2. 在项目名称下方查看 **"Project ID: 12345"**
3. 记录这个 ID

#### 1.3 更新数据库配置

```sql
-- 更新 Git Platform 配置
UPDATE git_platform_configs
SET 
    base_url = 'https://gitlab.com',  -- 或你的 GitLab 实例地址
    access_token = 'glpat-your-actual-token',
    webhook_secret = 'my-webhook-secret'  -- 自定义密钥
WHERE id = 1;

-- 更新 Repository 配置
UPDATE repositories
SET 
    platform_project_id = 12345,  -- 你的 Project ID
    name = 'my-test-project',
    full_name = 'username/my-test-project',
    git_url = 'https://gitlab.com/username/my-test-project.git'
WHERE id = 1;
```

### 阶段 2: 配置 GitLab Webhook

#### 2.1 启动 API Server (用于接收 Webhook)

```bash
# Terminal 1
./bin/api

# 看到以下日志说明启动成功:
# INFO API server starting on :8080
# INFO Webhook endpoint: /webhook/gitlab
```

#### 2.2 暴露本地服务到公网 (如果 GitLab 无法直接访问)

**方法 A: 使用 ngrok (推荐)**

```bash
# 安装 ngrok: https://ngrok.com/download
ngrok http 8080

# 复制 Forwarding URL
# 例如: https://abc123.ngrok.io
```

**方法 B: 使用 Cloudflare Tunnel**

```bash
cloudflared tunnel --url http://localhost:8080
```

**方法 C: 如果服务器有公网 IP**

直接使用: `http://your-server-ip:8080`

#### 2.3 在 GitLab 配置 Webhook

1. 访问项目 **Settings** → **Webhooks**
2. 填写 Webhook 信息:
   - **URL**: `https://your-domain/webhook/gitlab` (ngrok URL 或公网地址)
   - **Secret Token**: 与数据库中的 `webhook_secret` 一致
   - **Trigger**: 仅勾选
     - ✅ **Merge request events**
   - **SSL verification**: 如果使用自签名证书，取消勾选
3. 点击 **Add webhook**

#### 2.4 测试 Webhook

点击刚创建的 Webhook 右侧的 **Test** → **Merge request events**

**预期 API Server 日志**:

```
INFO Webhook received: GitLab merge_request event
INFO Task enqueued successfully
      task_id=abc-123
      repository_id=1
      mr_id=42
```

**GitLab 应显示**:
```
✅ HTTP 200
Response: {"message":"Webhook processed successfully"}
```

### 阶段 3: 启动 Worker

```bash
# Terminal 2
./bin/worker

# 看到以下日志说明启动成功:
# INFO Worker server starting
# INFO Concurrency: 5
# INFO Redis: localhost:6379
# INFO Handlers registered: code_review
```

### 阶段 4: 创建测试 MR

#### 4.1 在 GitLab 项目中创建测试分支

```bash
# Clone 项目
git clone https://gitlab.com/username/my-test-project.git
cd my-test-project

# 创建测试分支
git checkout -b test/ai-review

# 修改一些代码 (例如添加一个文件)
cat > test.go << 'EOF'
package main

import "fmt"

func main() {
    // TODO: This is a test
    password := "hardcoded123"  // 安全问题: 硬编码密码
    fmt.Println("Password:", password)
}
EOF

git add test.go
git commit -m "Add test file with intentional issues"
git push origin test/ai-review
```

#### 4.2 创建 Merge Request

1. 访问 GitLab 项目
2. 点击 **Merge Requests** → **New merge request**
3. 选择:
   - Source branch: `test/ai-review`
   - Target branch: `main`
4. 填写:
   - Title: `Test AI Code Review`
   - Description: `This is a test MR to verify HandsOff functionality`
5. 点击 **Create merge request**

### 阶段 5: 观察处理流程

#### 5.1 检查 API Server 日志

**应该看到**:

```
INFO Webhook received: GitLab merge_request event
     action=open
     project_id=12345
     mr_iid=1

INFO Task enqueued successfully
     task_id=abc-123-def-456
     queue=code_review
```

#### 5.2 检查 Worker 日志

**应该看到完整处理流程**:

```
INFO Processing code review task
     repository_id=1
     mr_id=42
     task_id=abc-123

INFO Fetching MR diff from GitLab
     project_id=12345
     mr_id=1

INFO MR diff fetched successfully
     diff_size=450

INFO Starting LLM code review
     llm_provider=deepseek
     model=deepseek-chat

INFO Calling LLM API
     provider=deepseek

INFO LLM review completed
     tokens_used=1500
     duration=3.5s
     suggestions=3

INFO Saving fix suggestions
     count=3

INFO Posting review comment to GitLab MR

INFO Review comment posted successfully to GitLab MR

INFO Code review completed successfully
     score=65
     suggestions_count=3
```

#### 5.3 检查 GitLab MR 评论

**访问 GitLab MR 页面，应该看到类似以下的评论**:

```markdown
## 🤖 AI Code Review

### 📝 Summary

This code contains a critical security issue with a hardcoded password. 
The implementation is simple but needs security improvements.

**Quality Score:** 65/100

### 🔍 Issues Found (2)

#### 🔴 Critical Issues

| File | Lines | Category | Description |
|------|-------|----------|-------------|
| `test.go` | L7 | **security** | Hardcoded password detected |

<details>
<summary>📋 Detailed Suggestions</summary>

**1. Hardcoded password detected**

- **File:** `test.go`
- **Lines:** L7
- **Category:** security

**Recommendation:**
Never hardcode passwords. Use environment variables or secure configuration.

**Current Code:**
```go
password := "hardcoded123"
```

---

</details>

#### 🟡 Medium Priority

| File | Lines | Category | Description |
|------|-------|----------|-------------|
| `test.go` | L6 | **style** | TODO comment should be addressed |

---

_Generated by HandsOff AI Code Review | Model: deepseek-chat | Tokens: 1500 | Duration: 3.50s_
```

### 阶段 6: 验证数据库

```sql
-- 查看 review_results
SELECT id, repository_id, merge_request_id, status, score, summary, comment_posted
FROM review_results
ORDER BY id DESC
LIMIT 1;

-- 应该看到:
-- status = 'completed'
-- score = 65
-- comment_posted = 1
-- summary = 'This code contains...'

-- 查看 fix_suggestions
SELECT id, review_result_id, file_path, line_start, severity, category, description
FROM fix_suggestions
WHERE review_result_id = (SELECT MAX(id) FROM review_results);

-- 应该看到 2-3 条建议记录
```

---

## 验证清单

### ✅ 完整流程检查清单

#### 准备阶段
- [ ] Redis 运行正常
- [ ] 数据库初始化成功
- [ ] API Key 已加密并配置
- [ ] GitLab Access Token 有效
- [ ] 测试数据已插入

#### 单元测试
- [ ] 数据库连接成功
- [ ] Redis 任务入队成功
- [ ] GitLab API 认证通过
- [ ] GitLab 可获取 MR Diff
- [ ] LLM API 调用成功

#### 集成测试
- [ ] API Server 启动成功
- [ ] Worker 启动成功
- [ ] Webhook 配置正确
- [ ] Webhook 测试返回 200
- [ ] 创建 MR 触发 Webhook
- [ ] Worker 接收并处理任务
- [ ] GitLab 获取 Diff 成功
- [ ] LLM API 返回审查结果
- [ ] 数据库保存 review_results
- [ ] 数据库保存 fix_suggestions
- [ ] GitLab MR 收到评论
- [ ] 评论格式正确美观
- [ ] 评论内容准确有价值

---

## 常见问题排查

### 问题 1: 数据库连接失败

**错误信息**:
```
❌ 数据库连接失败: unable to open database file
```

**解决方法**:
```bash
# 确保数据目录存在
mkdir -p data

# 检查文件权限
chmod 755 data
```

### 问题 2: Redis 连接失败

**错误信息**:
```
❌ Redis 连接失败: dial tcp 127.0.0.1:6379: connect: connection refused
```

**解决方法**:
```bash
# 检查 Redis 是否运行
redis-cli ping

# 如果未运行，启动 Redis
brew services start redis  # macOS
sudo systemctl start redis # Linux
```

### 问题 3: GitLab API 认证失败

**错误信息**:
```
❌ GitLab 连接失败: GitLab API authentication failed (status 401)
```

**解决方法**:
1. 检查 Access Token 是否正确
2. 确认 Token 权限包含 `api` 和 `read_api`
3. 检查 Token 是否过期

### 问题 4: LLM API 调用失败

**错误信息**:
```
❌ LLM API 调用失败: 401 Unauthorized
```

**解决方法**:
1. 检查 API Key 是否正确
2. 验证加密/解密是否正常:
   ```bash
   go run tools/encrypt_apikey/main.go -decrypt "YOUR_ENCRYPTED_KEY"
   ```
3. 测试 API Key 直接调用:
   ```bash
   curl https://api.deepseek.com/v1/chat/completions \
     -H "Authorization: Bearer sk-your-key" \
     -H "Content-Type: application/json" \
     -d '{"model":"deepseek-chat","messages":[{"role":"user","content":"test"}]}'
   ```

### 问题 5: Webhook 未触发

**症状**: 创建 MR 后，API Server 没有日志

**解决方法**:
1. 检查 Webhook URL 是否可访问:
   ```bash
   curl http://your-domain/webhook/gitlab
   ```
2. 检查 GitLab Webhook 配置:
   - Settings → Webhooks → Recent Deliveries
   - 查看失败原因
3. 检查 Secret Token 是否一致:
   ```sql
   SELECT webhook_secret FROM git_platform_configs WHERE id = 1;
   ```

### 问题 6: Worker 处理任务失败

**错误信息**:
```
ERROR Task failed: failed to get MR diff: 404 Not Found
```

**解决方法**:
1. 检查 `platform_project_id` 是否正确
2. 确认 Access Token 有访问项目的权限
3. 验证 MR 是否存在

### 问题 7: 评论未发布到 GitLab

**症状**: review_results 保存成功，但 comment_posted = 0

**解决方法**:
1. 检查 Worker 日志中的错误信息
2. 验证 Access Token 权限 (需要 `write_repository`)
3. 测试手动发布评论:
   ```bash
   curl -X POST "https://gitlab.com/api/v4/projects/12345/merge_requests/1/notes" \
     -H "PRIVATE-TOKEN: glpat-xxxxxxxxxxxx" \
     -H "Content-Type: application/json" \
     -d '{"body":"Test comment"}'
   ```

### 问题 8: LLM 响应解析失败

**错误信息**:
```
WARN Failed to parse LLM response as JSON, trying markdown...
```

**解决方法**:
1. 这是正常的降级行为，检查是否最终解析成功
2. 如果完全失败，检查 LLM 原始响应:
   ```sql
   SELECT raw_result FROM review_results ORDER BY id DESC LIMIT 1;
   ```
3. 调整提示词模板以强制 JSON 输出

---

## 性能基准

### 正常性能指标

| 阶段 | 预期耗时 | 说明 |
|------|----------|------|
| Webhook 接收 | <50ms | 包含签名验证和任务入队 |
| 获取 MR Diff | 200-500ms | 取决于 diff 大小 |
| LLM API 调用 | 2-5秒 | DeepSeek 平均 3秒 |
| 响应解析 | <10ms | 纯内存操作 |
| 数据库保存 | 50-100ms | 包含 review_results + suggestions |
| 发布评论 | 100-300ms | GitLab API 调用 |
| **总耗时** | **3-8秒** | 从 Webhook 到评论发布 |

### 性能优化建议

如果处理时间超过 10 秒:

1. **检查 LLM API 延迟**
   - 切换到更快的模型
   - 减少 max_tokens
   - 使用流式响应 (未来优化)

2. **检查网络延迟**
   - GitLab 和 LLM API 的网络连接
   - 考虑使用 CDN 或代理

3. **优化数据库**
   - 添加索引
   - 使用批量插入

---

## 下一步

测试成功后，你可以:

1. ✅ **继续优化** - Task 5: Review 结果存储优化
2. ✅ **开发前端** - Task 6-7: React 界面
3. ✅ **性能测试** - 批量 MR 处理能力
4. ✅ **生产部署** - Docker + Kubernetes

---

**🎉 祝测试顺利！如有问题，请参考本文档或查看日志文件。**
