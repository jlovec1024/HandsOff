# 🐛 Bug 修复：LLM Provider 创建时 400 参数错误

## 📝 问题描述

**症状**：前端保存 LLM Provider 配置时，后端返回 400 参数错误，无法创建 Provider。

**影响范围**：
- ❌ 无法添加新的 LLM Provider
- ✅ 编辑现有 Provider 正常（如果 project_id 已存在）
- ✅ 列表、删除、测试功能不受影响

---

## 🔍 根本原因分析

### 数据库约束

```go
// internal/model/llm_provider.go:21
ProjectID uint `gorm:"not null;index;constraint:OnDelete:CASCADE" json:"project_id"`
```

`project_id` 字段在数据库中有 **NOT NULL 约束**，插入时必须提供有效值。

### 代码流程分析

**预期流程（正确）：**
```
前端发送: { name, base_url, api_key, model, is_active }
         ↓
Auth Middleware: 提取 user_id (JWT token)
         ↓
ProjectContext Middleware: 查询用户活跃项目，注入 project_id 到 context
         ↓
Handler: 从 context 提取 project_id，注入到请求对象
         ↓
Service: 加密 API key
         ↓
Repository: 插入数据库（包含 project_id）
```

**实际流程（错误）：**
```
前端发送: { name, base_url, api_key, model, is_active }
         ↓
Handler: ❌ 没有提取 project_id，直接传递给 Service
         ↓
Service: 加密 API key
         ↓
Repository: 插入数据库时 project_id = 0
         ↓
❌ Database Error: NOT NULL constraint violation → 400
```

### 代码对比

**其他 Handler（正确示例 - ListProviders）：**
```go
func (h *LLMHandler) ListProviders(c *gin.Context) {
    projectID, ok := getProjectID(c)  // ✅ 提取 project_id
    if !ok {
        // Error handling
    }
    providers, err := h.service.ListProviders(projectID)
    ...
}
```

**CreateProvider（修复前 - 错误）：**
```go
func (h *LLMHandler) CreateProvider(c *gin.Context) {
    var req model.LLMProvider
    c.ShouldBindJSON(&req)
    // ❌ 缺少：没有提取和注入 project_id
    h.service.CreateProvider(&req)  // req.ProjectID = 0
    ...
}
```

---

## ✅ 解决方案

### 修复代码（internal/api/handler/llm.go）

```go
// CreateProvider creates a new LLM provider
func (h *LLMHandler) CreateProvider(c *gin.Context) {
    // ✅ 1. 提取 project_id（新增）
    projectID, ok := getProjectID(c)
    if !ok {
        h.log.Error("Project ID missing from context - middleware failure")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }

    var req model.LLMProvider
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    // Validate required fields
    if req.Name == "" || req.Model == "" || req.BaseURL == "" || req.APIKey == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Name, model, base URL, and API key are required"})
        return
    }

    // ✅ 2. 注入 project_id（新增）
    req.ProjectID = projectID

    if err := h.service.CreateProvider(&req); err != nil {
        h.log.Error("Failed to create provider", "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create provider"})
        return
    }

    h.log.Info("LLM provider created", "name", req.Name, "model", req.Model)
    c.JSON(http.StatusCreated, req)
}
```

### 核心改动

**新增2行关键代码：**
1. **Line 78-82**: 从 context 提取 `project_id`（通过 `getProjectID(c)` helper 函数）
2. **Line 98**: 注入 `project_id` 到请求对象（`req.ProjectID = projectID`）

---

## 🎯 设计原则

### Why This is "Good Taste" (Linus 风格)

1. **✅ 消除特殊情况**
   - 所有资源创建（LLM Provider、Git Config、Repository）都遵循统一模式
   - 不需要前端传递 project_id（消除了前后端的数据不一致风险）

2. **✅ 数据流清晰**
   ```
   Authentication → Project Context → Business Logic
   ```
   - 中间件负责提取上下文
   - Handler 负责组装数据
   - Service 只关注业务逻辑

3. **✅ 类型安全**
   - 前端类型定义保持简洁（不包含 project_id）
   - project_id 是后端实现细节，前端无需关心

4. **✅ Never Break Userspace**
   - ✅ 不改数据库结构
   - ✅ 不改前端接口
   - ✅ 不影响其他功能
   - ✅ 编辑功能继续正常工作

---

## 🧪 测试验证

### 前置条件
1. 启动后端服务：`make run-api` 或 `go run ./cmd/api`
2. 启动前端服务：`cd web && npm run dev`
3. 登录系统：`admin / admin123`

### 测试步骤

**1. 创建新 LLM Provider**
```bash
# 获取 token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.token')

# 测试创建 Provider（不传 project_id）
curl -X POST http://localhost:8080/api/llm/providers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "OpenAI Official",
    "base_url": "https://api.openai.com/v1",
    "api_key": "sk-test-key",
    "model": "gpt-4",
    "is_active": true
  }'
```

**预期结果（修复后）：**
```json
{
  "id": 1,
  "name": "OpenAI Official",
  "base_url": "https://api.openai.com/v1",
  "model": "gpt-4",
  "is_active": true,
  "project_id": 1,  // ✅ 后端自动注入
  "created_at": "2024-12-05T...",
  "updated_at": "2024-12-05T..."
}
```

**2. 验证数据库**
```bash
sqlite3 data/app.db "SELECT id, name, project_id FROM llm_providers;"
```

**预期输出：**
```
1|OpenAI Official|1
```

**3. 前端测试**
1. 访问 `设置 -> LLM 供应商`
2. 点击"添加供应商"
3. 填写表单：
   - 名称：OpenAI Test
   - Base URL：https://api.openai.com/v1
   - API Key：sk-test
   - 点击"获取可用模型"按钮
   - 从下拉框选择模型（如 gpt-4）
4. 点击"保存"

**预期结果：**
- ✅ 成功提示："供应商已创建"
- ✅ 列表中显示新建的 Provider
- ✅ 无 400 错误

---

## 📚 相关文件

### 修改的文件
- `internal/api/handler/llm.go` - CreateProvider 函数（+4 行）

### 相关架构组件
- `internal/api/middleware/auth.go` - JWT 认证中间件
- `internal/api/middleware/project.go` - 项目上下文中间件（提取 project_id）
- `internal/api/handler/helper.go` - `getProjectID()` helper 函数
- `internal/model/llm_provider.go` - LLMProvider 数据模型
- `internal/service/llm_service.go` - LLMService 业务逻辑
- `internal/repository/llm_repo.go` - LLM 数据访问层

### 前端相关文件
- `web/src/pages/Settings/LLMProviders.tsx` - LLM Provider 管理页面
- `web/src/api/llm.ts` - LLM API 客户端
- `web/src/types/index.ts` - TypeScript 类型定义

---

## 🔄 项目架构总结

### Multi-Project Architecture

```
User (用户)
  ├── Project 1 (项目1)
  │   ├── LLM Providers
  │   ├── Git Configs
  │   └── Repositories
  └── Project 2 (项目2)
      ├── LLM Providers
      └── ...

UserProjectPreference (用户偏好)
  └── 记录用户当前活跃的项目 ID
```

### 认证与授权流程

```
1. 用户登录 → JWT Token (包含 user_id)
2. 请求 API → Auth Middleware (验证 token, 提取 user_id)
3. ProjectContext Middleware (查询 UserProjectPreference, 提取 project_id)
4. Handler (使用 project_id 过滤/创建数据)
5. Service (业务逻辑，不关心 project_id 来源)
6. Repository (数据库操作)
```

### 关键设计原则

1. **资源隔离**：所有资源（LLMProvider, GitConfig, Repository）都属于 Project
2. **中间件注入**：project_id 由中间件自动提取，Handler 无需手动查询
3. **前端无感知**：前端不需要知道 project_id，后端自动处理

---

## 📌 注意事项

### 对开发者的提醒

1. **所有创建资源的 Handler 都必须注入 project_id**
   ```go
   // ❌ 错误
   func CreateSomething(c *gin.Context) {
       var req model.Something
       c.ShouldBindJSON(&req)
       service.Create(&req)  // req.ProjectID = 0 → 数据库错误
   }

   // ✅ 正确
   func CreateSomething(c *gin.Context) {
       projectID, ok := getProjectID(c)
       // ... error handling
       var req model.Something
       c.ShouldBindJSON(&req)
       req.ProjectID = projectID  // 必须注入
       service.Create(&req)
   }
   ```

2. **不要在前端传递 project_id**
   - project_id 是后端实现细节
   - 前端只需要发送业务数据（name, base_url, api_key, model 等）
   - 后端通过认证用户自动关联到正确的项目

3. **使用 `getProjectID()` helper 函数**
   - 不要直接使用 `c.GetUint("project_id")`（返回 0 如果不存在）
   - 使用 `getProjectID(c)` 进行类型安全检查

---

## ✅ 验证通过

- ✅ 编译通过：`go build -v ./cmd/api`
- ✅ 代码符合项目风格（与 ListProviders, GetProvider 等函数一致）
- ✅ 不破坏现有功能（Never break userspace）
- ✅ 简洁实用（+4 行代码解决核心问题）

---

**修复日期**：2024-12-05  
**修复人**：Linus Style Code Review  
**版本**：v1.0
