# MVP 交互流程设计

## 📋 文档概述

本文档详细说明MVP版本的核心交互流程，包括用户操作流程、系统处理流程和数据流转设计。

---

## 1. 核心业务流程总览

```mermaid
graph TB
    A[开始使用] --> B[系统初始化配置]
    B --> C[配置GitLab]
    B --> D[配置LLM]
    C --> E[导入仓库]
    D --> E
    E --> F[配置仓库Webhook和LLM]
    F --> G[开发提交MR]
    G --> H[触发自动Review]
    H --> I[查看Review结果]
    I --> J[根据建议优化代码]
    
    style B fill:#e1f5ff
    style H fill:#fff4e1
    style I fill:#e8f5e9
```

---

## 2. 用户初始化配置流程

### 2.1 完整配置流程图

```mermaid
sequenceDiagram
    participant U as 管理员
    participant F as 前端页面
    participant A as API服务
    participant DB as 数据库
    participant GL as GitLab API
    participant LLM as LLM API
    
    Note over U,LLM: Step 1: 登录系统
    U->>F: 访问 /login
    F->>U: 显示登录表单
    U->>F: 输入用户名密码
    F->>A: POST /api/auth/login
    A->>DB: 验证用户凭据
    DB-->>A: 返回用户信息
    A->>A: 生成JWT Token
    A-->>F: 返回Token和用户信息
    F->>F: 保存Token到localStorage
    F->>U: 跳转到系统设置页
    
    Note over U,LLM: Step 2: 配置GitLab平台
    U->>F: 进入"GitLab配置"Tab
    F->>A: GET /api/platform/config
    A->>DB: 查询平台配置
    DB-->>A: 返回配置（可能为空）
    A-->>F: 返回配置数据
    F->>U: 显示配置表单
    
    U->>F: 填写GitLab URL和Token
    F->>A: PUT /api/platform/config
    A->>GL: 测试连接（GET /api/v4/user）
    GL-->>A: 返回用户信息
    A->>A: AES加密AccessToken
    A->>DB: 保存平台配置
    DB-->>A: 保存成功
    A-->>F: 返回成功
    F->>U: 显示"配置成功"提示
    
    Note over U,LLM: Step 3: 配置LLM供应商
    U->>F: 进入"LLM配置"Tab
    F->>A: GET /api/llm/providers
    A->>DB: 查询供应商列表
    DB-->>A: 返回列表（可能为空）
    A-->>F: 返回供应商数据
    F->>U: 显示供应商列表
    
    U->>F: 点击"添加供应商"
    F->>U: 显示表单Modal
    U->>F: 填写供应商信息<br/>(类型、API Key、Base URL)
    F->>A: POST /api/llm/providers
    A->>LLM: 测试连接
    LLM-->>A: 返回测试响应
    A->>A: AES加密API Key
    A->>DB: 保存供应商配置
    DB-->>A: 保存成功
    A-->>F: 返回成功
    F->>U: 关闭Modal,刷新列表
    
    Note over U,LLM: Step 4: 添加LLM模型
    U->>F: 选择供应商，点击"添加模型"
    F->>U: 显示模型表单Modal
    U->>F: 填写模型名称
    F->>A: POST /api/llm/models
    A->>DB: 保存模型配置
    DB-->>A: 保存成功
    A-->>F: 返回成功
    F->>U: 关闭Modal,刷新列表
```

### 2.2 配置步骤详解

#### 步骤1: 登录系统

**页面**: `/login`

**操作流程**:
1. 用户输入用户名和密码
2. 点击"登录"按钮
3. 前端验证表单（非空、格式）
4. 发送POST请求到 `/api/auth/login`
5. 后端验证凭据
6. 生成JWT Token（24小时有效期）
7. 返回Token和用户信息
8. 前端保存Token到localStorage
9. 跳转到系统设置页

**默认管理员账号**:
- 用户名: `admin`
- 密码: `admin123`

#### 步骤2: 配置GitLab

**页面**: `/settings` (GitLab配置Tab)

**配置项**:
| 字段 | 说明 | 示例 |
|------|------|------|
| GitLab URL | GitLab实例地址 | `https://gitlab.com` |
| Access Token | 个人访问令牌 | `glpat-xxxxxxxxxxxx` |

**操作流程**:
1. 填写GitLab URL和Access Token
2. 点击"测试连接"按钮（可选）
   - 调用GitLab API `/api/v4/user`
   - 验证Token有效性
   - 显示连接成功/失败消息
3. 点击"保存"按钮
4. 后端加密保存Token
5. 显示保存成功提示

#### 步骤3: 配置LLM供应商

**页面**: `/settings` (LLM配置Tab)

**配置项**:
| 字段 | 说明 | 示例 |
|------|------|------|
| 供应商名称 | 自定义名称 | "DeepSeek生产环境" |
| 供应商类型 | deepseek/openai/qwen/ollama | deepseek |
| API Key | LLM API密钥 | `sk-xxxxxxxx` |
| Base URL | API基础地址 | `https://api.deepseek.com` |

**操作流程**:
1. 点击"添加供应商"按钮
2. 填写供应商信息
3. 点击"测试连接"（可选）
4. 点击"保存"
5. 供应商列表显示新增项

#### 步骤4: 添加LLM模型

**页面**: `/settings` (LLM配置Tab -> 模型管理)

**配置项**:
| 字段 | 说明 | 示例 |
|------|------|------|
| 模型名称 | 模型标识 | `deepseek-chat` |
| 显示名称 | 前端显示 | "DeepSeek Chat" |

**推荐模型**:
- **DeepSeek**: `deepseek-chat`
- **OpenAI**: `gpt-3.5-turbo`
- **Qwen**: `qwen-turbo`

---

## 3. 仓库导入与配置流程

### 3.1 导入仓库流程图

```mermaid
sequenceDiagram
    participant U as 用户
    participant F as 前端
    participant A as API
    participant DB as 数据库
    participant GL as GitLab API
    
    Note over U,GL: Step 1: 进入仓库管理
    U->>F: 访问 /repositories
    F->>A: GET /api/repositories
    A->>DB: 查询已导入仓库
    DB-->>A: 返回仓库列表
    A-->>F: 返回数据
    F->>U: 显示仓库列表
    
    Note over U,GL: Step 2: 获取GitLab仓库
    U->>F: 点击"导入仓库"按钮
    F->>U: 显示导入Modal
    F->>A: GET /api/platform/repositories?page=1&page_size=20
    A->>GL: GET /api/v4/projects
    GL-->>A: 返回仓库列表
    A-->>F: 返回仓库数据
    F->>U: 显示可选仓库列表（Table with Checkbox）
    
    Note over U,GL: Step 3: 选择并导入
    U->>F: 勾选要导入的仓库（支持多选）
    U->>F: 点击"导入"按钮
    F->>A: POST /api/repositories/batch<br/>{repository_ids: [1,2,3]}
    A->>DB: 批量保存仓库信息
    A->>GL: 为每个仓库配置Webhook
    GL-->>A: 返回Webhook ID
    A->>DB: 更新webhook_id
    DB-->>A: 保存成功
    A-->>F: 返回导入结果
    F->>U: 显示"成功导入X个仓库"
    F->>F: 关闭Modal,刷新列表
    
    Note over U,GL: Step 4: 配置仓库LLM
    U->>F: 点击仓库的"配置"按钮
    F->>U: 显示配置Modal
    F->>A: GET /api/llm/models
    A->>DB: 查询可用模型
    DB-->>A: 返回模型列表
    A-->>F: 返回数据
    F->>U: 显示LLM模型下拉框
    U->>F: 选择LLM模型
    U->>F: 点击"保存"
    F->>A: PUT /api/repositories/:id<br/>{llm_model_id: 1}
    A->>DB: 更新仓库配置
    DB-->>A: 更新成功
    A-->>F: 返回成功
    F->>U: 显示"配置成功"
```

### 3.2 导入步骤详解

#### 步骤1: 查看已导入仓库

**页面**: `/repositories`

**显示内容**:
- 仓库列表（Table）
- 列: 仓库名、完整路径、默认分支、Webhook状态、LLM模型、操作
- 操作按钮: 配置、删除

#### 步骤2: 从GitLab获取仓库

**触发**: 点击"导入仓库"按钮

**显示**: Modal对话框

**内容**:
- 搜索框（可选）
- 仓库列表（带复选框）
- 分页控件
- 批量导入按钮

**GitLab API调用**:
```
GET /api/v4/projects?per_page=20&page=1&owned=true
```

#### 步骤3: 批量导入

**操作**:
1. 勾选要导入的仓库
2. 点击"导入"按钮
3. 后端处理:
   - 保存仓库信息到数据库
   - 为每个仓库配置GitLab Webhook
   - 设置Webhook事件: `merge_request_events`
   - 保存Webhook ID
4. 显示导入结果

**Webhook配置**:
```json
{
  "url": "http://your-server.com/api/webhook",
  "merge_request_events": true,
  "enable_ssl_verification": false
}
```

#### 步骤4: 配置仓库LLM

**操作**:
1. 点击仓库的"配置"按钮
2. 选择LLM模型
3. 保存配置

---

## 4. Webhook触发Review流程

### 4.1 完整Review流程图

```mermaid
sequenceDiagram
    participant Dev as 开发者
    participant GL as GitLab
    participant WH as Webhook接收
    participant Q as Redis队列
    participant W as Worker
    participant DB as 数据库
    participant LLM as LLM API
    
    Note over Dev,LLM: 开发者提交MR
    Dev->>GL: 创建Merge Request
    GL->>GL: 触发MR事件
    GL->>WH: POST /api/webhook<br/>(MR Webhook Payload)
    
    Note over Dev,LLM: Webhook接收处理
    WH->>WH: 验证Webhook签名
    WH->>WH: 解析MR事件数据
    WH->>DB: 查询仓库信息
    DB-->>WH: 返回仓库配置
    WH->>Q: 创建Review任务<br/>(Asynq.Enqueue)
    Q-->>WH: 任务ID
    WH-->>GL: 返回200 OK
    
    Note over Dev,LLM: Worker异步处理
    Q->>W: 分发Review任务
    W->>GL: GET /api/v4/projects/:id/merge_requests/:mr_number/changes
    GL-->>W: 返回MR Diff数据
    
    W->>DB: 获取仓库LLM配置
    DB-->>W: 返回llm_model_id
    W->>DB: 获取LLM模型信息
    DB-->>W: 返回API配置
    
    W->>W: 加载默认提示词模板
    W->>W: 构建Prompt<br/>(填充repo_name, author, diff等)
    
    W->>LLM: POST /chat/completions<br/>(发送Prompt)
    LLM-->>W: 返回Review结果
    
    W->>W: 解析AI返回的结果<br/>(提取score, summary, suggestions)
    
    W->>DB: 保存review_results
    DB-->>W: 返回review_id
    W->>DB: 批量保存fix_suggestions
    DB-->>W: 保存成功
    
    W->>GL: POST /api/v4/projects/:id/merge_requests/:mr_number/notes<br/>(发布Review评论)
    GL-->>W: 评论成功
    
    W-->>Q: 任务完成
    
    Note over Dev,LLM: 开发者查看结果
    Dev->>GL: 访问MR页面
    GL->>Dev: 显示AI Review评论
```

### 4.2 Webhook Payload示例

```json
{
  "object_kind": "merge_request",
  "user": {
    "name": "张三",
    "username": "zhangsan"
  },
  "project": {
    "id": 123,
    "name": "my-project",
    "web_url": "https://gitlab.com/group/my-project"
  },
  "object_attributes": {
    "id": 456,
    "iid": 10,
    "title": "feat: add login feature",
    "description": "实现用户登录功能",
    "source_branch": "feature/login",
    "target_branch": "main",
    "state": "opened",
    "action": "open",
    "url": "https://gitlab.com/group/my-project/-/merge_requests/10"
  }
}
```

### 4.3 提示词模板示例

```markdown
You are an experienced code reviewer. Please analyze the following code changes and provide constructive feedback.

**Repository**: {{repo_name}}
**Author**: {{author}}
**Merge Request**: {{source_branch}} -> {{target_branch}}
**MR URL**: {{mr_url}}

**Code Changes**:
```diff
{{diff_content}}
```

Please provide a structured review with:

1. **Overall Score** (0-100): Rate the code quality
2. **Summary**: Brief summary of the code quality
3. **Issues**: List specific issues with:
   - File path
   - Line numbers
   - Severity (critical/high/medium/low)
   - Description
   - Suggestion for improvement

**Output Format** (JSON):
```json
{
  "overall_score": 85,
  "summary": "Code quality is good overall...",
  "suggestions": [
    {
      "file_path": "src/auth.go",
      "line_start": 10,
      "line_end": 15,
      "severity": "high",
      "description": "Password is stored in plain text",
      "suggestion": "Use bcrypt to hash the password"
    }
  ]
}
```

### 4.4 AI返回结果解析

**原始返回**:
```json
{
  "overall_score": 85,
  "summary": "代码整体质量良好，但存在一些安全隐患需要修复...",
  "suggestions": [
    {
      "file_path": "src/login.go",
      "line_start": 25,
      "line_end": 30,
      "severity": "high",
      "description": "密码未加密直接存储到数据库",
      "suggestion": "使用bcrypt对密码进行加密后再存储"
    },
    {
      "file_path": "src/api/user.go",
      "line_start": 45,
      "line_end": 50,
      "severity": "medium",
      "description": "SQL查询存在注入风险",
      "suggestion": "使用参数化查询或ORM"
    }
  ]
}
```

**数据库存储**:

**review_results表**:
| 字段 | 值 |
|------|------|
| repository_id | 1 |
| llm_model_id | 1 |
| author | "zhangsan" |
| source_branch | "feature/login" |
| target_branch | "main" |
| mr_url | "https://gitlab.com/..." |
| mr_number | 10 |
| raw_result | "{整个AI返回的JSON}" |
| overall_score | 85 |
| summary | "代码整体质量良好..." |

**fix_suggestions表** (2条记录):
| 字段 | 记录1 | 记录2 |
|------|-------|-------|
| review_result_id | 1 | 1 |
| file_path | "src/login.go" | "src/api/user.go" |
| line_start | 25 | 45 |
| line_end | 30 | 50 |
| severity | "high" | "medium" |
| description | "密码未加密..." | "SQL查询..." |
| suggestion | "使用bcrypt..." | "使用参数化..." |

### 4.5 GitLab评论格式

```markdown
## 🤖 AI Code Review

**Overall Score**: 85/100  
**Summary**: 代码整体质量良好，但存在一些安全隐患需要修复

---

### 🔴 High Severity Issues (1)

#### 1. src/login.go:25-30
**Description**: 密码未加密直接存储到数据库  
**Suggestion**: 使用bcrypt对密码进行加密后再存储

---

### 🟡 Medium Severity Issues (1)

#### 2. src/api/user.go:45-50
**Description**: SQL查询存在注入风险  
**Suggestion**: 使用参数化查询或ORM

---

**Powered by HandsOff（甩手掌柜）**
```

---

## 5. 用户查看Review记录流程

### 5.1 查看Review列表

```mermaid
sequenceDiagram
    participant U as 用户
    participant F as 前端
    participant A as API
    participant DB as 数据库
    
    U->>F: 访问 /reviews
    F->>A: GET /api/reviews?page=1&page_size=20
    A->>DB: 查询review_results
    DB-->>A: 返回列表数据
    A-->>F: 返回JSON数据
    F->>U: 显示Review列表（Table）
    
    U->>F: 筛选仓库
    F->>A: GET /api/reviews?repository_id=1
    A->>DB: 按条件查询
    DB-->>A: 返回过滤后的数据
    A-->>F: 返回数据
    F->>U: 更新列表
```

**列表字段**:
- 仓库名
- 作者
- 分支（source -> target）
- 评分
- 总结摘要
- MR链接
- 创建时间
- 操作（查看详情）

### 5.2 查看Review详情

```mermaid
sequenceDiagram
    participant U as 用户
    participant F as 前端
    participant A as API
    participant DB as 数据库
    
    U->>F: 点击"查看详情"
    F->>U: 跳转到 /reviews/:id
    F->>A: GET /api/reviews/:id
    A->>DB: 查询review_results（包含suggestions）
    DB-->>A: 返回完整数据
    A-->>F: 返回JSON
    F->>U: 显示详情页面
    
    Note over U,DB: 显示内容
    F->>U: 基本信息卡片<br/>(仓库、作者、分支、时间)
    F->>U: 评分展示<br/>(Progress Bar)
    F->>U: 总结卡片
    F->>U: 修复建议列表<br/>(按严重程度排序)
    
    U->>F: 点击"查看原始结果"Tab
    F->>U: 显示AI原始返回的JSON
```

**详情页布局**:

```
┌─────────────────────────────────────────┐
│ Review详情 #123                         │
├─────────────────────────────────────────┤
│ 仓库: my-project                        │
│ 作者: zhangsan                          │
│ 分支: feature/login -> main             │
│ MR: #10                                 │
│ 时间: 2025-01-30 10:30                 │
├─────────────────────────────────────────┤
│ 评分: ████████▓░ 85/100                │
├─────────────────────────────────────────┤
│ 总结:                                   │
│ 代码整体质量良好，但存在一些安全隐患... │
├─────────────────────────────────────────┤
│ 修复建议 (2条)                          │
│ ┌─────────────────────────────────────┐│
│ │ 🔴 HIGH src/login.go:25-30         ││
│ │ 密码未加密直接存储                  ││
│ │ 建议: 使用bcrypt加密               ││
│ └─────────────────────────────────────┘│
│ ┌─────────────────────────────────────┐│
│ │ 🟡 MEDIUM src/api/user.go:45-50    ││
│ │ SQL查询存在注入风险                 ││
│ │ 建议: 使用参数化查询               ││
│ └─────────────────────────────────────┘│
└─────────────────────────────────────────┘
```

---

## 6. 异常处理流程

### 6.1 配置错误处理

```mermaid
flowchart TD
    A[用户配置GitLab] --> B{连接测试}
    B -->|成功| C[保存配置]
    B -->|失败| D[显示错误消息]
    D --> E[检查URL格式]
    D --> F[检查Token有效性]
    D --> G[检查网络连接]
    E --> A
    F --> A
    G --> A
    
    C --> H[配置完成]
```

**常见错误**:
- URL格式错误: "请输入有效的GitLab URL"
- Token无效: "Access Token无效，请检查"
- 网络错误: "无法连接到GitLab，请检查网络"

### 6.2 Webhook错误处理

```mermaid
flowchart TD
    A[接收Webhook] --> B{验证签名}
    B -->|失败| C[返回401]
    B -->|成功| D{解析Payload}
    D -->|失败| E[返回400]
    D -->|成功| F{查询仓库}
    F -->|不存在| G[返回404]
    F -->|存在| H[创建任务]
    H --> I[返回200]
```

### 6.3 Review任务错误处理

```mermaid
flowchart TD
    A[Worker执行任务] --> B{获取MR Diff}
    B -->|失败| C[记录错误日志]
    B -->|成功| D{调用LLM API}
    D -->|超时| E[重试3次]
    D -->|失败| C
    D -->|成功| F{解析结果}
    F -->|失败| C
    F -->|成功| G[保存数据库]
    G --> H{发布评论}
    H -->|失败| I[记录但不阻断]
    H -->|成功| J[任务完成]
    
    E -->|3次都失败| C
    C --> K[任务标记为失败]
```

---

## 7. 性能优化设计

### 7.1 异步任务处理

**优势**:
- ✅ Webhook立即返回200，不阻塞GitLab
- ✅ Worker并发处理多个Review任务
- ✅ 支持任务重试机制

**配置**:
```go
// Asynq配置
config := asynq.Config{
    Concurrency: 10,  // 并发10个Worker
    Queues: map[string]int{
        "critical": 6,  // 高优先级
        "default":  3,  // 默认优先级
        "low":      1,  // 低优先级
    },
}
```

### 7.2 数据库查询优化

**索引设计**:
- `repositories`: idx_webhook_active, idx_llm_model
- `review_results`: idx_repository, idx_created_at
- `fix_suggestions`: idx_review_result, idx_severity

**分页查询**:
```go
// 限制单次查询数量
func ListReviews(page, pageSize int) ([]ReviewResult, int64, error) {
    var results []ReviewResult
    var total int64
    
    db.Model(&ReviewResult{}).Count(&total)
    
    offset := (page - 1) * pageSize
    db.Limit(pageSize).Offset(offset).
        Order("created_at DESC").
        Preload("Repository").
        Preload("Suggestions").
        Find(&results)
    
    return results, total, nil
}
```

---

## 8. 安全性设计

### 8.1 认证与授权

```mermaid
flowchart TD
    A[前端请求] --> B{携带Token?}
    B -->|否| C[返回401]
    B -->|是| D{验证Token}
    D -->|无效| C
    D -->|有效| E{检查权限}
    E -->|无权限| F[返回403]
    E -->|有权限| G[执行请求]
```

**JWT Token设计**:
```go
type Claims struct {
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.StandardClaims
}

// 生成Token
token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
    UserID:   user.ID,
    Username: user.Username,
    Role:     user.Role,
    StandardClaims: jwt.StandardClaims{
        ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
    },
})
```

### 8.2 敏感数据加密

**加密字段**:
- GitLab Access Token
- LLM API Key

**加密方式**:
```go
// AES-256-GCM加密
func Encrypt(plaintext string, key []byte) (string, error) {
    block, _ := aes.NewCipher(key)
    gcm, _ := cipher.NewGCM(block)
    nonce := make([]byte, gcm.NonceSize())
    rand.Read(nonce)
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}
```

### 8.3 Webhook签名验证

```go
// GitLab Webhook签名验证
func VerifyWebhookSignature(payload []byte, signature string, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expectedMAC := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expectedMAC))
}
```

---

## 9. 总结

### 9.1 MVP交互流程特点

✅ **简化配置**: 仅3步完成系统初始化  
✅ **自动化**: Webhook自动触发Review  
✅ **异步处理**: 不阻塞用户操作  
✅ **实时反馈**: GitLab MR直接显示结果  
✅ **容错处理**: 完善的错误处理机制  

### 9.2 用户体验优化

1. **快速上手**: 默认配置即可使用
2. **即时反馈**: 操作后立即显示结果
3. **错误提示**: 明确的错误信息和解决方案
4. **批量操作**: 支持批量导入仓库
5. **筛选查询**: 方便查找Review记录

---

**设计版本**: v1.0-mvp  
**最后更新**: 2025-01-30  
**设计人**: Snow AI
