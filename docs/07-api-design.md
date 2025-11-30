# API接口设计

## 1. API设计原则

### 1.1 RESTful规范

- ✅ 使用HTTP Method语义 (GET/POST/PUT/DELETE)
- ✅ 资源路径复数形式 (`/platforms`, `/repositories`)
- ✅ 使用HTTP状态码
- ✅ 统一响应格式

### 1.2 统一响应格式

```go
// 成功响应
{
  "code": 200,
  "message": "Success",
  "data": { ... }  // 或 [ ... ]
}

// 分页响应
{
  "code": 200,
  "message": "Success",
  "data": {
    "list": [ ... ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100
    }
  }
}

// 错误响应
{
  "code": 400,
  "message": "Validation error",
  "errors": {
    "field_name": ["error message 1", "error message 2"]
  }
}
```

### 1.3 HTTP状态码

| 状态码 | 含义 | 使用场景 |
|--------|------|---------|
| 200 | OK | 成功 (GET/PUT/DELETE) |
| 201 | Created | 创建成功 (POST) |
| 400 | Bad Request | 参数错误 |
| 401 | Unauthorized | 未认证 |
| 403 | Forbidden | 无权限 |
| 404 | Not Found | 资源不存在 |
| 500 | Internal Server Error | 服务器错误 |

---

## 2. 认证接口

### 2.1 登录

**接口**: `POST /api/auth/login`

**请求**:
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "admin",
      "nickname": "管理员",
      "role": "admin"
    }
  }
}
```

### 2.2 退出登录

**接口**: `POST /api/auth/logout`

**响应**:
```json
{
  "code": 200,
  "message": "Logout successful"
}
```

---

## 3. Git平台管理接口

### 3.1 获取平台列表

**接口**: `GET /api/platforms`

**Query参数**:
- `page`: int, 页码 (default: 1)
- `page_size`: int, 每页数量 (default: 20)
- `search`: string, 搜索关键词
- `platform_type`: string, 平台类型 (gitlab/github/gitea)
- `is_active`: bool, 是否启用

**响应**:
```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "name": "公司GitLab",
        "platform_type": "gitlab",
        "base_url": "https://gitlab.company.com",
        "is_active": true,
        "created_at": "2023-12-01T10:00:00Z",
        "updated_at": "2023-12-01T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 5
    }
  }
}
```

### 3.2 创建平台配置

**接口**: `POST /api/platforms`

**请求**:
```json
{
  "name": "公司GitLab",
  "platform_type": "gitlab",
  "base_url": "https://gitlab.company.com",
  "access_token": "glpat-xxxxxxxxxxxx"
}
```

**响应**:
```json
{
  "code": 201,
  "message": "Platform created successfully",
  "data": {
    "id": 1
  }
}
```

### 3.3 更新平台配置

**接口**: `PUT /api/platforms/:id`

**请求**:
```json
{
  "name": "公司GitLab2",
  "base_url": "https://gitlab2.company.com",
  "access_token": "glpat-yyyyyyyyyyyy",
  "is_active": true
}
```

### 3.4 删除平台配置

**接口**: `DELETE /api/platforms/:id`

**响应**:
```json
{
  "code": 200,
  "message": "Platform deleted successfully"
}
```

### 3.5 测试平台连接

**接口**: `POST /api/platforms/:id/test`

**响应**:
```json
{
  "code": 200,
  "message": "Connection successful",
  "data": {
    "username": "admin",
    "name": "Administrator",
    "version": "15.5.0"
  }
}
```

---

## 4. 代码仓库管理接口

### 4.1 从平台获取仓库列表

**接口**: `GET /api/platforms/:platform_id/repositories/fetch`

**Query参数**:
- `page`: int, 页码
- `page_size`: int, 每页数量
- `search`: string, 搜索关键词

**响应**:
```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "repo_id": "123",
        "repo_name": "my-project",
        "repo_full_path": "group/my-project",
        "repo_url": "https://gitlab.com/group/my-project",
        "default_branch": "main",
        "description": "项目描述",
        "last_activity_at": "2023-12-01T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100
    }
  }
}
```

### 4.2 批量导入仓库

**接口**: `POST /api/repositories/batch`

**请求**:
```json
{
  "platform_config_id": 1,
  "repositories": [
    {
      "repo_id": "123",
      "repo_name": "my-project",
      "repo_full_path": "group/my-project",
      "repo_url": "https://gitlab.com/group/my-project",
      "default_branch": "main"
    }
  ]
}
```

**响应**:
```json
{
  "code": 201,
  "message": "Successfully imported 10 repositories",
  "data": {
    "imported_count": 10
  }
}
```

### 4.3 获取仓库列表

**接口**: `GET /api/repositories`

**Query参数**:
- `page`, `page_size`
- `platform_config_id`: int, 平台ID
- `search`: string
- `is_webhook_active`: bool

**响应**: (同3.1格式)

### 4.4 配置仓库Webhook

**接口**: `POST /api/repositories/:id/webhook`

**请求**:
```json
{
  "webhook_url": "https://your-server.com/api/webhook",
  "events": ["push", "merge_request"]
}
```

**响应**:
```json
{
  "code": 200,
  "message": "Webhook configured successfully",
  "data": {
    "webhook_id": "456",
    "webhook_url": "https://your-server.com/api/webhook"
  }
}
```

### 4.5 删除仓库Webhook

**接口**: `DELETE /api/repositories/:id/webhook`

### 4.6 测试Webhook

**接口**: `POST /api/repositories/:id/webhook/test`

---

## 5. 仓库组管理接口

### 5.1 获取仓库组列表

**接口**: `GET /api/groups`

**响应**:
```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "name": "后端服务组",
        "description": "所有后端微服务",
        "llm_model_id": 1,
        "llm_model": {
          "id": 1,
          "model_name": "deepseek-chat",
          "model_display_name": "DeepSeek Chat"
        },
        "repositories_count": 10,
        "notification_channels_count": 2,
        "created_at": "2023-12-01T10:00:00Z"
      }
    ],
    "pagination": { ... }
  }
}
```

### 5.2 获取仓库组详情

**接口**: `GET /api/groups/:id`

**响应**:
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "name": "后端服务组",
    "description": "所有后端微服务",
    "llm_model_id": 1,
    "prompt_template": {
      "system_prompt": "你是一个资深的Go代码审查专家...",
      "user_prompt": "请审查以下代码:\\n{{diffs_text}}"
    },
    "notification_config": {
      "show_commits": true,
      "show_score": true,
      "show_suggestions": true
    },
    "repositories": [
      {
        "id": 1,
        "repo_name": "user-service",
        "repo_full_path": "backend/user-service"
      }
    ],
    "notification_channels": [
      {
        "id": 1,
        "name": "开发组钉钉",
        "channel_type": "dingtalk"
      }
    ]
  }
}
```

### 5.3 创建仓库组

**接口**: `POST /api/groups`

**请求**:
```json
{
  "name": "后端服务组",
  "description": "所有后端微服务",
  "llm_model_id": 1,
  "repository_ids": [1, 2, 3],
  "notification_channel_ids": [1, 2],
  "prompt_template": {
    "system_prompt": "...",
    "user_prompt": "..."
  },
  "notification_config": {
    "show_commits": true,
    "show_score": true
  }
}
```

### 5.4 更新仓库组

**接口**: `PUT /api/groups/:id`

**请求**: (同创建)

### 5.5 更新提示词模板

**接口**: `PUT /api/groups/:id/prompt`

**请求**:
```json
{
  "system_prompt": "你是一个资深的{{language}}代码审查专家...",
  "user_prompt": "请审查以下代码:\\n{{diffs_text}}"
}
```

### 5.6 获取提示词模板

**接口**: `GET /api/groups/:id/prompt`

**响应**:
```json
{
  "code": 200,
  "data": {
    "system_prompt": "...",
    "user_prompt": "..."
  }
}
```

---

## 6. LLM配置管理接口

### 6.1 获取LLM供应商列表

**接口**: `GET /api/llm/providers`

**响应**:
```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "name": "DeepSeek生产",
        "provider_type": "deepseek",
        "api_base_url": "https://api.deepseek.com",
        "is_active": true,
        "models_count": 3
      }
    ]
  }
}
```

### 6.2 创建LLM供应商

**接口**: `POST /api/llm/providers`

**请求**:
```json
{
  "name": "DeepSeek生产",
  "provider_type": "deepseek",
  "api_key": "sk-xxxxxxxx",
  "api_base_url": "https://api.deepseek.com"
}
```

### 6.3 测试LLM连接

**接口**: `POST /api/llm/providers/:id/test`

**响应**:
```json
{
  "code": 200,
  "message": "Connection successful",
  "data": {
    "response": "连接成功",
    "latency_ms": 1250
  }
}
```

### 6.4 动态获取模型列表 (需求12)

**接口**: `POST /api/llm/providers/:id/models/fetch`

**响应**:
```json
{
  "code": 200,
  "message": "Fetched 5 models",
  "data": {
    "models": [
      {
        "model_name": "deepseek-chat",
        "model_display_name": "DeepSeek Chat",
        "max_tokens": 10000
      }
    ],
    "fetched_count": 5
  }
}
```

### 6.5 获取模型列表

**接口**: `GET /api/llm/providers/:provider_id/models`

**响应**:
```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "provider_id": 1,
        "model_name": "deepseek-chat",
        "model_display_name": "DeepSeek Chat",
        "max_tokens": 10000,
        "is_active": true,
        "is_from_api": true
      }
    ]
  }
}
```

### 6.6 创建/更新/删除模型

**接口**: 
- `POST /api/llm/models`
- `PUT /api/llm/models/:id`
- `DELETE /api/llm/models/:id`

---

## 7. 通知渠道管理接口

### 7.1 获取通知渠道列表

**接口**: `GET /api/notifications`

**响应**:
```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "name": "开发组钉钉",
        "channel_type": "dingtalk",
        "webhook_url": "https://oapi.dingtalk.com/robot/send?access_token=xxx",
        "is_active": true
      }
    ]
  }
}
```

### 7.2 创建通知渠道

**接口**: `POST /api/notifications`

**请求**:
```json
{
  "name": "开发组钉钉",
  "channel_type": "dingtalk",
  "webhook_url": "https://oapi.dingtalk.com/robot/send?access_token=xxx",
  "secret": "SECxxxxxxxx",
  "config_json": "{\"keywords\": [\"code review\"]}"
}
```

### 7.3 测试通知渠道

**接口**: `POST /api/notifications/:id/test`

**请求**:
```json
{
  "test_message": "这是一条测试消息"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "Test message sent successfully"
}
```

---

## 8. Review记录查询接口

### 8.1 获取Review列表

**接口**: `GET /api/reviews`

**Query参数**:
- `page`, `page_size`
- `review_type`: string, mr/push
- `repository_id`: int
- `group_id`: int
- `author`: string
- `start_date`, `end_date`: string (ISO 8601)
- `min_score`, `max_score`: int

**响应**:
```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "review_type": "mr",
        "repository": {
          "id": 1,
          "repo_name": "my-project"
        },
        "author": "zhangsan",
        "source_branch": "feature/login",
        "target_branch": "main",
        "overall_score": 85,
        "summary": "代码整体质量良好",
        "suggestions_count": 5,
        "mr_url": "https://gitlab.com/group/project/-/merge_requests/123",
        "created_at": "2023-12-01T10:00:00Z"
      }
    ],
    "pagination": { ... }
  }
}
```

### 8.2 获取Review详情

**接口**: `GET /api/reviews/:id`

**响应**:
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "review_type": "mr",
    "repository": { ... },
    "group": { ... },
    "llm_model": { ... },
    "author": "zhangsan",
    "source_branch": "feature/login",
    "target_branch": "main",
    "mr_url": "https://gitlab.com/...",
    "commit_messages": "feat: add login page",
    "last_commit_id": "abc123",
    "additions": 150,
    "deletions": 30,
    "raw_result": "# Code Review Result\\n...",
    "structured_result": {
      "overall_score": 85,
      "summary": "代码整体质量良好",
      "suggestions": [ ... ]
    },
    "overall_score": 85,
    "summary": "代码整体质量良好",
    "suggestions": [
      {
        "id": 1,
        "file_path": "src/login.go",
        "line_start": 10,
        "line_end": 15,
        "issue_type": "security",
        "severity": "high",
        "description": "密码未加密存储",
        "suggestion": "使用bcrypt加密密码",
        "code_snippet": "password := req.Password\\nuser.Password = password",
        "fix_tasks": [
          {
            "id": 1,
            "status": "success",
            "fix_branch": "ai-fix/suggestion-1-xxx"
          }
        ]
      }
    ],
    "created_at": "2023-12-01T10:00:00Z"
  }
}
```

---

## 9. 自动修复管理接口

### 9.1 创建修复任务

**接口**: `POST /api/fix/tasks`

**请求**:
```json
{
  "suggestion_id": 1,
  "repository_id": 1
}
```

**响应**:
```json
{
  "code": 201,
  "message": "Fix task created and queued",
  "data": {
    "task_id": 1
  }
}
```

### 9.2 获取修复任务列表

**接口**: `GET /api/fix/tasks`

**Query参数**:
- `page`, `page_size`
- `repository_id`: int
- `status`: string (pending/running/success/failed/cancelled)
- `start_date`, `end_date`

**响应**:
```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "suggestion": {
          "id": 1,
          "file_path": "src/login.go",
          "description": "密码未加密"
        },
        "repository": {
          "id": 1,
          "repo_name": "my-project"
        },
        "base_branch": "main",
        "fix_branch": "ai-fix/suggestion-1-xxx",
        "status": "success",
        "fix_commit_sha": "def456",
        "is_ignored": false,
        "started_at": "2023-12-01T10:30:00Z",
        "completed_at": "2023-12-01T10:32:45Z",
        "created_at": "2023-12-01T10:30:00Z"
      }
    ],
    "pagination": { ... }
  }
}
```

### 9.3 获取修复任务详情

**接口**: `GET /api/fix/tasks/:id`

**响应**:
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "suggestion": { ... },
    "repository": { ... },
    "base_branch": "main",
    "fix_branch": "ai-fix/suggestion-1-xxx",
    "status": "success",
    "fix_commit_sha": "def456",
    "fix_commit_message": "🤖 AI自动修复: security issue\\n\\n文件: src/login.go\\n问题: 密码未加密",
    "error_message": null,
    "execution_log": "完整日志内容...",
    "is_ignored": false,
    "ignore_category": null,
    "started_at": "2023-12-01T10:30:00Z",
    "completed_at": "2023-12-01T10:32:45Z",
    "logs": [
      {
        "id": 1,
        "log_level": "info",
        "message": "开始执行自动修复任务",
        "timestamp": "2023-12-01T10:30:00Z"
      }
    ]
  }
}
```

### 9.4 获取修复任务日志

**接口**: `GET /api/fix/tasks/:id/logs`

**Query参数**:
- `since`: int (timestamp), 获取该时间之后的日志

**响应**:
```json
{
  "code": 200,
  "data": {
    "logs": [
      {
        "id": 1,
        "log_level": "info",
        "message": "开始执行自动修复任务",
        "timestamp": "2023-12-01T10:30:00Z"
      }
    ]
  }
}
```

### 9.5 实时日志流 (WebSocket)

**接口**: `WS /api/fix/tasks/:id/logs/stream`

**消息格式**:
```json
{
  "type": "log",
  "data": {
    "log_level": "info",
    "message": "克隆仓库到本地...",
    "timestamp": "2023-12-01T10:30:05Z"
  }
}

// 任务完成消息
{
  "type": "done",
  "data": {
    "status": "success"
  }
}
```

### 9.6 重新执行修复任务

**接口**: `POST /api/fix/tasks/:id/retry`

**响应**:
```json
{
  "code": 200,
  "message": "Fix task queued for retry"
}
```

### 9.7 标记任务忽略状态

**接口**: `PUT /api/fix/tasks/:id/ignore`

**请求**:
```json
{
  "is_ignored": true,
  "ignore_category": "auto-fix"
}
```

---

## 10. 修复分支管理接口

### 10.1 获取修复分支列表

**接口**: `GET /api/fix/branches`

**Query参数**:
- `repository_id`: int
- `is_merged`: bool
- `is_deleted`: bool

**响应**:
```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "repository": {
          "id": 1,
          "repo_name": "my-project"
        },
        "branch_name": "ai-fix/suggestion-1-xxx",
        "base_branch": "main",
        "related_mr_url": "https://gitlab.com/group/project/-/merge_requests/456",
        "is_merged": false,
        "is_deleted": false,
        "created_at": "2023-12-01T10:30:00Z"
      }
    ]
  }
}
```

### 10.2 删除修复分支

**接口**: `DELETE /api/fix/branches/:id`

**响应**:
```json
{
  "code": 200,
  "message": "Branch deleted successfully"
}
```

### 10.3 合并修复分支

**接口**: `POST /api/fix/branches/:id/merge`

**请求**:
```json
{
  "merge_method": "squash",
  "delete_after_merge": true
}
```

**响应**:
```json
{
  "code": 200,
  "message": "Branch merged successfully",
  "data": {
    "mr_url": "https://gitlab.com/group/project/-/merge_requests/456"
  }
}
```

---

## 11. Webhook接收接口

### 11.1 统一Webhook接收

**接口**: `POST /api/webhook`

**请求**: (GitLab/GitHub/Gitea Webhook Payload)

**响应**:
```json
{
  "code": 200,
  "message": "Webhook received, processing asynchronously"
}
```

**处理逻辑**:
1. 识别平台类型 (通过Header)
2. 解析Webhook数据
3. 创建异步任务
4. 立即返回200

---

## 12. 系统配置接口

### 12.1 获取系统配置

**接口**: `GET /api/system/config`

**Query参数**:
- `keys`: string[], 配置键列表

**响应**:
```json
{
  "code": 200,
  "data": {
    "webhook_base_url": "http://localhost:8080",
    "default_fix_branch_prefix": "ai-fix/",
    "notification_show_commits": "true"
  }
}
```

### 12.2 更新系统配置

**接口**: `PUT /api/system/config`

**请求**:
```json
{
  "configs": {
    "webhook_base_url": "https://your-server.com",
    "max_concurrent_fix_tasks": "5"
  }
}
```

---

## 13. 统计与仪表盘接口

### 13.1 获取统计数据

**接口**: `GET /api/dashboard/stats`

**响应**:
```json
{
  "code": 200,
  "data": {
    "total_repositories": 120,
    "active_repositories": 80,
    "total_reviews": 1245,
    "total_fix_tasks": 45,
    "notification_channels": 5,
    "llm_models": 8
  }
}
```

### 13.2 获取Review趋势

**接口**: `GET /api/dashboard/review-trend`

**Query参数**:
- `days`: int, 最近天数 (default: 30)

**响应**:
```json
{
  "code": 200,
  "data": {
    "dates": ["2023-12-01", "2023-12-02", ...],
    "counts": [10, 15, 8, ...]
  }
}
```

---

## 14. 通用接口规范

### 14.1 认证Header

所有需要认证的接口都需要携带:

```
Authorization: Bearer <JWT_TOKEN>
```

### 14.2 分页参数

**Query参数**:
- `page`: int, 页码 (从1开始)
- `page_size`: int, 每页数量 (default: 20, max: 100)

**响应**:
```json
{
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

### 14.3 排序参数

**Query参数**:
- `order_by`: string, 排序字段
- `order`: string, asc/desc

示例: `/api/reviews?order_by=created_at&order=desc`

### 14.4 错误码定义

| Code | 说明 | 示例 |
|------|------|------|
| 10001 | 参数验证失败 | 必填字段缺失 |
| 10002 | 资源不存在 | Platform not found |
| 10003 | 重复资源 | Platform already exists |
| 10004 | 操作失败 | Failed to connect to GitLab |
| 20001 | 认证失败 | Invalid credentials |
| 20002 | Token过期 | Token expired |
| 20003 | 无权限 | Permission denied |
| 50000 | 服务器错误 | Internal server error |

---

**下一步**: 生成完整技术设计文档汇总
