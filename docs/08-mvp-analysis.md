# MVP 文档审查与优化报告

## 📋 文档审查概述

**审查时间**: 2025-01-30  
**审查范围**: 项目全部文档（排除排期和人员安排）  
**审查标准**: 专业、简洁、完整  
**优化方向**: 基于MVP（最小可行产品）思路

---

## 1. 文档审查结果

### ✅ 完整且正确的文档

| 文档 | 状态 | 评价 |
|------|------|------|
| README.md | ✅ 优秀 | 结构清晰，内容完整，准确描述项目现状 |
| SNOW.md | ✅ 优秀 | 技术背景完整，架构设计合理 |
| 01-tech-stack.md | ✅ 优秀 | 技术选型合理，依赖清单完整 |
| 02-project-structure.md | ✅ 优秀 | 目录结构规范，分层清晰 |
| 03-database-design.md | ✅ 优秀 | 数据模型完整，支持双数据库 |
| 04-feature-list.md | ⚠️ 需优化 | 功能过于庞大，需基于MVP精简 |
| 05-page-design.md | ⚠️ 需优化 | 页面数量偏多，可合并精简 |
| 06-interaction-design.md | ✅ 良好 | 交互逻辑清晰，但可简化部分流程 |
| 07-api-design.md | ⚠️ 需优化 | API接口过多，需聚焦MVP核心功能 |

### ⚠️ 发现的主要问题

1. **功能过于庞大**: 118个功能点对于MVP来说过于复杂
2. **页面数量偏多**: 23个页面可能导致开发周期过长
3. **缺少MVP优先级**: P0/P1/P2划分存在，但未明确MVP范围
4. **缺少核心流程聚焦**: 应优先实现最核心的代码审查流程

---

## 2. MVP 核心功能定义

### 2.1 MVP 目标

**一句话描述**: 实现基于GitLab的自动代码审查，支持AI审查结果展示和简单修复建议

### 2.2 MVP 核心用户故事

1. ✅ **作为开发者，我希望提交MR后自动触发AI代码审查**
2. ✅ **作为开发者,我希望查看AI审查结果和修复建议**
3. ✅ **作为管理员,我希望配置GitLab和LLM以启用审查功能**
4. ❌ **自动修复功能** - 暂不纳入MVP
5. ❌ **多仓库组管理** - 暂不纳入MVP
6. ❌ **通知渠道** - 暂不纳入MVP

### 2.3 MVP 功能范围

#### 核心功能（必须实现）

| 模块 | 功能 | 是否纳入MVP |
|------|------|-------------|
| **认证** | 用户登录/登出 | ✅ 是 |
| **Git平台** | 配置单个GitLab实例 | ✅ 是（仅支持GitLab） |
| **仓库管理** | 导入仓库、配置Webhook | ✅ 是 |
| **LLM配置** | 配置单个LLM供应商和模型 | ✅ 是（推荐DeepSeek） |
| **Review记录** | 查看审查列表和详情 | ✅ 是 |
| **Webhook** | 接收MR事件并触发审查 | ✅ 是 |

#### 暂不纳入MVP（后续版本）

| 模块 | 理由 |
|------|------|
| 多GitLab实例支持 | MVP只需支持单个实例 |
| GitHub/Gitea支持 | 聚焦GitLab |
| 仓库组管理 | MVP直接配置单个仓库 |
| 自定义提示词 | 使用默认模板 |
| 通知渠道 | 可通过GitLab MR评论查看结果 |
| 自动修复功能 | 复杂度高，后续实现 |
| 分支管理 | 依赖自动修复功能 |
| Dashboard统计 | 数据积累后再做 |

---

## 3. MVP 数据模型优化

### 3.1 保留的核心表（7张）

```
MVP数据模型（7张表）
├── users（用户表）
├── git_platform_configs（Git平台配置）- 简化为单实例
├── repositories（代码仓库）
├── llm_providers（LLM供应商）
├── llm_models（LLM模型）
├── review_results（Review结果）
└── fix_suggestions（修复建议）
```

### 3.2 暂不实现的表（8张）

- repository_groups
- repository_group_mappings
- notification_channels
- group_notification_mappings
- auto_fix_tasks
- auto_fix_logs
- fix_branch_management
- system_configs（部分配置写入配置文件）

### 3.3 数据模型调整建议

#### users 表（无需调整）
```sql
CREATE TABLE users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    nickname VARCHAR(100),
    role VARCHAR(20) DEFAULT 'user',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

#### git_platform_configs 表（简化）
```sql
-- MVP版本：仅支持单个GitLab实例
CREATE TABLE git_platform_configs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL DEFAULT 'GitLab',
    base_url VARCHAR(500) NOT NULL,
    access_token VARCHAR(500) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

#### repositories 表（简化）
```sql
CREATE TABLE repositories (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    platform_config_id BIGINT NOT NULL,
    repo_id VARCHAR(100) NOT NULL,
    repo_name VARCHAR(200) NOT NULL,
    repo_full_path VARCHAR(500) NOT NULL,
    repo_url VARCHAR(500) NOT NULL,
    default_branch VARCHAR(100) DEFAULT 'main',
    webhook_id VARCHAR(100),
    is_webhook_active BOOLEAN DEFAULT FALSE,
    -- MVP: 直接在仓库表配置LLM
    llm_model_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (platform_config_id) REFERENCES git_platform_configs(id),
    FOREIGN KEY (llm_model_id) REFERENCES llm_models(id)
);
```

#### llm_providers 表（无需调整）
```sql
CREATE TABLE llm_providers (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    provider_type VARCHAR(50) NOT NULL,
    api_key VARCHAR(500),
    api_base_url VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

#### llm_models 表（无需调整）
```sql
CREATE TABLE llm_models (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    provider_id BIGINT NOT NULL,
    model_name VARCHAR(100) NOT NULL,
    model_display_name VARCHAR(200),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (provider_id) REFERENCES llm_providers(id)
);
```

#### review_results 表（简化）
```sql
CREATE TABLE review_results (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    repository_id BIGINT NOT NULL,
    llm_model_id BIGINT,
    author VARCHAR(100),
    source_branch VARCHAR(200),
    target_branch VARCHAR(200),
    mr_url VARCHAR(500),
    mr_number INT,
    commit_sha VARCHAR(100),
    raw_result LONGTEXT,
    overall_score INT DEFAULT 0,
    summary TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (repository_id) REFERENCES repositories(id)
);
```

#### fix_suggestions 表（简化）
```sql
CREATE TABLE fix_suggestions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    review_result_id BIGINT NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    line_start INT,
    line_end INT,
    severity VARCHAR(20) DEFAULT 'medium',
    description TEXT NOT NULL,
    suggestion TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (review_result_id) REFERENCES review_results(id)
);
```

---

## 4. MVP 功能清单

### 4.1 精简后的功能（35项 → 从118项）

#### 模块1: 认证（3项）

| 功能ID | 功能名称 | 说明 |
|--------|---------|------|
| AUTH-001 | 用户登录 | JWT认证 |
| AUTH-002 | 用户退出 | 清除Token |
| AUTH-003 | 获取用户信息 | 获取当前用户 |

#### 模块2: Git平台管理（4项）

| 功能ID | 功能名称 | 说明 |
|--------|---------|------|
| PLT-001 | 配置GitLab | 配置单个GitLab实例 |
| PLT-002 | 测试连接 | 验证Token有效性 |
| PLT-003 | 查看配置 | 查看当前配置 |
| PLT-004 | 更新配置 | 修改配置信息 |

#### 模块3: 仓库管理（6项）

| 功能ID | 功能名称 | 说明 |
|--------|---------|------|
| REPO-001 | 获取GitLab仓库列表 | 从GitLab API获取 |
| REPO-002 | 导入仓库 | 批量导入仓库 |
| REPO-003 | 查看仓库列表 | 展示已导入仓库 |
| REPO-004 | 配置Webhook | 自动配置Webhook |
| REPO-005 | 配置仓库LLM | 为仓库指定LLM模型 |
| REPO-006 | 删除仓库 | 移除仓库 |

#### 模块4: LLM配置（5项）

| 功能ID | 功能名称 | 说明 |
|--------|---------|------|
| LLM-001 | 添加LLM供应商 | 配置DeepSeek等 |
| LLM-002 | 测试连接 | 测试LLM连接 |
| LLM-003 | 添加模型 | 手动添加模型 |
| LLM-004 | 查看供应商列表 | 展示配置的供应商 |
| LLM-005 | 查看模型列表 | 展示可用模型 |

#### 模块5: Review记录（8项）

| 功能ID | 功能名称 | 说明 |
|--------|---------|------|
| REV-001 | 查看Review列表 | 分页展示 |
| REV-002 | 按仓库筛选 | 筛选功能 |
| REV-003 | 查看Review详情 | 详细信息 |
| REV-004 | 查看原始结果 | AI原始输出 |
| REV-005 | 查看评分 | 显示分数 |
| REV-006 | 查看总结 | 显示总结 |
| REV-007 | 查看修复建议列表 | 展示建议 |
| REV-008 | 按文件分组展示 | 建议分组 |

#### 模块6: Webhook处理（4项）

| 功能ID | 功能名称 | 说明 |
|--------|---------|------|
| WH-001 | 接收Webhook | 接收GitLab MR事件 |
| WH-002 | 解析事件 | 解析Webhook数据 |
| WH-003 | 触发Review任务 | 创建异步任务 |
| WH-004 | 发布Review结果 | 评论到GitLab MR |

#### 模块7: 系统配置（5项）

| 功能ID | 功能名称 | 说明 |
|--------|---------|------|
| SYS-001 | 查看系统配置 | 获取配置 |
| SYS-002 | 更新Webhook URL | 配置回调地址 |
| SYS-003 | 查看默认提示词 | 查看模板 |
| SYS-004 | 更新默认提示词 | 修改模板 |
| SYS-005 | 查看系统状态 | 健康检查 |

**总计**: 35个功能点（从118项精简70%）

---

## 5. MVP 页面设计优化

### 5.1 精简后的页面（9页 → 从23页）

```
MVP页面结构（9页）
├── 登录页
├── 主框架Layout
│   ├── 系统设置（4合1）
│   │   ├── GitLab配置
│   │   ├── LLM配置
│   │   ├── Webhook配置
│   │   └── 提示词模板
│   ├── 仓库管理（2合1）
│   │   ├── 仓库列表
│   │   └── 导入仓库（Modal）
│   └── Review记录（2页）
│       ├── Review列表
│       └── Review详情
└── 404页面
```

### 5.2 页面合并建议

| 原页面数量 | 优化后 | 说明 |
|-----------|--------|------|
| Git平台管理（2页） | 系统设置（1个Tab） | 单实例配置 |
| LLM配置（4页） | 系统设置（1个Tab） | 简化为列表+表单 |
| 仓库管理（3页） | 保留2页 | 列表+Modal导入 |
| 仓库组管理（3页） | ❌ 删除 | MVP不需要 |
| 通知渠道（2页） | ❌ 删除 | MVP不需要 |
| Review记录（2页） | ✅ 保留 | 核心功能 |
| 自动修复（3页） | ❌ 删除 | MVP不需要 |
| Dashboard（1页） | ❌ 删除 | 数据积累后再做 |
| 系统设置（3页） | 合并为1页多Tab | 集中配置 |

### 5.3 MVP 路由设计

```typescript
// MVP路由表
const routes = [
  { path: '/login', component: Login },
  { path: '/', component: Layout, children: [
    { path: 'settings', component: Settings }, // Tabs: GitLab, LLM, Webhook, 提示词
    { path: 'repositories', component: RepositoryList },
    { path: 'reviews', component: ReviewList },
    { path: 'reviews/:id', component: ReviewDetail },
  ]},
  { path: '/404', component: NotFound },
];
```

---

## 6. MVP API接口优化

### 6.1 精简后的API（25个 → 从80+个）

#### 认证接口（2个）
- `POST /api/auth/login`
- `POST /api/auth/logout`

#### Git平台接口（4个）
- `GET /api/platform/config` - 获取配置
- `PUT /api/platform/config` - 更新配置
- `POST /api/platform/test` - 测试连接
- `GET /api/platform/repositories` - 获取仓库列表

#### 仓库管理接口（5个）
- `GET /api/repositories` - 获取已导入仓库
- `POST /api/repositories/batch` - 批量导入
- `PUT /api/repositories/:id` - 更新仓库（配置LLM）
- `DELETE /api/repositories/:id` - 删除仓库
- `POST /api/repositories/:id/webhook` - 配置Webhook

#### LLM接口（6个）
- `GET /api/llm/providers` - 获取供应商列表
- `POST /api/llm/providers` - 创建供应商
- `PUT /api/llm/providers/:id` - 更新供应商
- `POST /api/llm/providers/:id/test` - 测试连接
- `GET /api/llm/models` - 获取模型列表
- `POST /api/llm/models` - 创建模型

#### Review接口（4个）
- `GET /api/reviews` - 获取Review列表
- `GET /api/reviews/:id` - 获取Review详情
- `GET /api/reviews/:id/suggestions` - 获取修复建议

#### Webhook接口（1个）
- `POST /api/webhook` - 接收GitLab Webhook

#### 系统配置接口（3个）
- `GET /api/system/config` - 获取系统配置
- `PUT /api/system/config` - 更新系统配置
- `GET /api/system/health` - 健康检查

**总计**: 25个API接口（从80+个精简70%）

---

## 7. MVP 交互流程

### 7.1 核心业务流程

#### Webhook触发Review流程（简化版）

```
GitLab (MR事件)
    ↓
Webhook接收 (/webhook)
    ↓
异步任务入队 (Asynq)
    ↓
Worker处理
    ↓
├─ 获取MR Diff
├─ 获取仓库LLM配置
├─ 构建提示词（使用默认模板）
├─ 调用LLM API
├─ 解析结果
├─ 保存到数据库
└─ 发布评论到GitLab MR
```

**移除的步骤**:
- ❌ 获取仓库组配置
- ❌ 发送通知
- ❌ 过滤文件类型（全部审查）

### 7.2 用户配置流程（简化版）

```
1. 登录系统
    ↓
2. 配置GitLab（单个实例）
    ↓
3. 配置LLM供应商和模型
    ↓
4. 导入仓库
    ↓
5. 为仓库配置Webhook和LLM
    ↓
6. 提交MR触发审查
    ↓
7. 查看Review结果
```

---

## 8. 关键优化建议

### 8.1 数据库优化

**建议**: 
1. ✅ 移除8张非MVP表，减少复杂度
2. ✅ 在repositories表直接关联llm_model_id
3. ✅ 移除软删除（DeletedAt），直接物理删除
4. ✅ 简化索引设计，仅保留必要索引

### 8.2 功能优化

**建议**:
1. ✅ 移除仓库组管理，直接配置单个仓库
2. ✅ 移除自动修复功能，后续版本实现
3. ✅ 移除通知渠道，使用GitLab MR评论
4. ✅ 移除Dashboard，聚焦核心功能
5. ✅ 使用默认提示词模板，暂不支持自定义

### 8.3 页面优化

**建议**:
1. ✅ 合并配置页面为统一的系统设置
2. ✅ 使用Modal代替独立表单页
3. ✅ 移除不必要的详情页
4. ✅ 简化导航结构

### 8.4 API优化

**建议**:
1. ✅ 移除仓库组相关接口
2. ✅ 移除自动修复相关接口
3. ✅ 移除通知渠道相关接口
4. ✅ 简化查询参数，减少过滤条件

### 8.5 技术栈优化

**建议**:
1. ✅ 保持技术栈不变（Go + React）
2. ✅ 移除WebSocket（无自动修复功能）
3. ✅ 简化状态管理（减少Store模块）
4. ✅ 使用配置文件代替system_configs表

---

## 9. MVP 开发排期建议

### Phase 1: 基础框架（Week 1）
- 数据库初始化（7张表）
- 用户认证系统
- 基础前端框架

### Phase 2: 配置管理（Week 2）
- GitLab平台配置
- LLM配置管理
- 系统设置页面

### Phase 3: 仓库管理（Week 3）
- 仓库导入功能
- Webhook配置
- 仓库列表页面

### Phase 4: Review功能（Week 4-5）
- Webhook接收
- Review任务调度
- LLM调用
- Review列表和详情页面

### Phase 5: 测试与部署（Week 6）
- 功能测试
- Docker部署
- 文档完善

**总计**: 6周完成MVP

---

## 10. MVP 后续版本规划

### V1.0 (MVP)
- ✅ 基础代码审查功能
- ✅ GitLab集成
- ✅ LLM审查

### V1.1
- 🔄 多GitLab实例支持
- 🔄 自定义提示词模板
- 🔄 Dashboard统计

### V1.2
- 🔄 GitHub/Gitea支持
- 🔄 通知渠道集成
- 🔄 仓库组管理

### V2.0
- 🔄 自动修复功能
- 🔄 分支管理
- 🔄 高级统计分析

---

## 11. 文档修正建议

### 11.1 需要更新的文档

1. **03-database-design.md**
   - 移除8张非MVP表的设计
   - 更新ERD关系图
   - 简化SQL迁移文件

2. **04-feature-list.md**
   - 标注MVP功能范围
   - 移除非MVP功能模块
   - 更新功能统计

3. **05-page-design.md**
   - 精简页面数量至9页
   - 更新路由设计
   - 移除非MVP页面设计

4. **07-api-design.md**
   - 精简API接口至25个
   - 移除非MVP接口
   - 更新接口文档

5. **06-interaction-design.md**
   - 简化交互流程
   - 移除WebSocket相关设计
   - 更新状态管理

### 11.2 需要新增的文档

1. **MVP实施指南** (本文档)
2. **默认提示词模板文档**
3. **Docker快速部署指南**

---

## 12. 总结

### 12.1 MVP 核心价值

✅ **聚焦核心场景**: GitLab MR代码审查  
✅ **快速验证**: 6周完成可用版本  
✅ **降低复杂度**: 功能精简70%，快速迭代  
✅ **保持扩展性**: 架构设计支持后续功能扩展  

### 12.2 关键指标

| 指标 | 原设计 | MVP版本 | 精简比例 |
|------|--------|---------|----------|
| 数据库表数 | 15张 | 7张 | 53% |
| 功能点数 | 118项 | 35项 | 70% |
| 页面数量 | 23页 | 9页 | 61% |
| API接口 | 80+个 | 25个 | 69% |
| 开发周期 | 12周 | 6周 | 50% |

### 12.3 推荐技术选型（MVP）

**后端**:
- Go 1.21 + Gin + GORM
- SQLite (开发/小团队)
- Redis + Asynq (异步任务)

**前端**:
- React 18 + Ant Design 5
- TypeScript 5 + Vite
- Zustand (简化状态管理)

**LLM推荐**:
- DeepSeek Chat (性价比高，质量好)
- 备选: Qwen, GPT-3.5

---

**审查完成时间**: 2025-01-30  
**审查人**: Snow AI  
**版本**: v1.0-mvp
