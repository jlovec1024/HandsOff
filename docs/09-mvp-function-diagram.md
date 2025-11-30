# MVP 功能关系图

## 1. 系统架构图

```mermaid
graph TB
    subgraph "用户交互层"
        A[前端 React Web]
    end
    
    subgraph "API服务层"
        B[Gin API Server]
    end
    
    subgraph "业务逻辑层"
        C1[认证服务]
        C2[GitLab服务]
        C3[仓库服务]
        C4[LLM服务]
        C5[Review服务]
        C6[Webhook服务]
    end
    
    subgraph "数据访问层"
        D[GORM ORM]
    end
    
    subgraph "数据存储层"
        E1[(SQLite/MySQL)]
        E2[(Redis)]
    end
    
    subgraph "外部服务"
        F1[GitLab API]
        F2[LLM API<br/>DeepSeek/GPT]
    end
    
    subgraph "异步任务"
        G[Asynq Worker]
    end
    
    A --> B
    B --> C1
    B --> C2
    B --> C3
    B --> C4
    B --> C5
    B --> C6
    
    C1 --> D
    C2 --> D
    C3 --> D
    C4 --> D
    C5 --> D
    C6 --> D
    
    C2 --> F1
    C4 --> F2
    C6 --> G
    
    D --> E1
    G --> E2
    G --> E1
    G --> F1
    G --> F2
```

---

## 2. 核心功能模块关系图

```mermaid
graph LR
    A[系统配置] --> B[仓库管理]
    A --> C[LLM配置]
    B --> D[Webhook配置]
    D --> E[Review任务]
    C --> E
    E --> F[Review记录]
    E --> G[修复建议]
    F --> G
    
    style A fill:#e1f5ff
    style E fill:#fff4e1
    style F fill:#e8f5e9
    style G fill:#e8f5e9
```

### 模块说明

| 模块 | 职责 | 依赖 |
|------|------|------|
| 系统配置 | GitLab配置、LLM配置、提示词模板 | 无 |
| 仓库管理 | 导入仓库、配置仓库LLM | 系统配置 |
| Webhook配置 | 配置GitLab Webhook | 仓库管理 |
| Review任务 | 触发AI审查、执行审查 | Webhook、LLM配置 |
| Review记录 | 存储审查结果 | Review任务 |
| 修复建议 | 存储修复建议 | Review记录 |

---

## 3. 用户操作流程图

```mermaid
sequenceDiagram
    participant U as 用户
    participant F as 前端
    participant A as API
    participant DB as 数据库
    participant GL as GitLab
    participant LLM as LLM API
    
    Note over U,LLM: 初始化配置流程
    U->>F: 1. 登录系统
    F->>A: POST /api/auth/login
    A->>DB: 验证用户
    A-->>F: 返回JWT Token
    
    U->>F: 2. 配置GitLab
    F->>A: PUT /api/platform/config
    A->>GL: 测试连接
    A->>DB: 保存配置
    A-->>F: 配置成功
    
    U->>F: 3. 配置LLM
    F->>A: POST /api/llm/providers
    A->>LLM: 测试连接
    A->>DB: 保存LLM配置
    A-->>F: 配置成功
    
    U->>F: 4. 导入仓库
    F->>A: GET /api/platform/repositories
    A->>GL: 获取仓库列表
    A-->>F: 返回仓库列表
    F->>A: POST /api/repositories/batch
    A->>GL: 配置Webhook
    A->>DB: 保存仓库信息
    A-->>F: 导入成功
    
    U->>F: 5. 为仓库配置LLM
    F->>A: PUT /api/repositories/:id
    A->>DB: 更新仓库LLM配置
    A-->>F: 配置成功
```

---

## 4. Webhook触发Review流程图

```mermaid
sequenceDiagram
    participant GL as GitLab
    participant WH as Webhook接收
    participant Q as Redis队列
    participant W as Worker
    participant DB as 数据库
    participant LLM as LLM API
    
    Note over GL,LLM: MR触发Review流程
    GL->>WH: 1. POST /webhook<br/>(MR事件)
    WH->>WH: 2. 解析事件数据
    WH->>Q: 3. 创建Review任务
    WH-->>GL: 4. 返回200 OK
    
    Q->>W: 5. 分发任务
    W->>GL: 6. 获取MR Diff
    W->>DB: 7. 获取仓库LLM配置
    W->>DB: 8. 获取默认提示词
    W->>W: 9. 构建Prompt
    W->>LLM: 10. 调用LLM API
    LLM-->>W: 11. 返回审查结果
    W->>W: 12. 解析结果
    W->>DB: 13. 保存Review记录
    W->>DB: 14. 保存修复建议
    W->>GL: 15. 发布评论到MR
    W-->>Q: 16. 任务完成
```

---

## 5. 数据流图

```mermaid
graph TD
    subgraph "数据输入"
        A1[用户配置数据]
        A2[GitLab MR事件]
    end
    
    subgraph "数据处理"
        B1[配置管理]
        B2[Webhook解析]
        B3[LLM调用]
        B4[结果解析]
    end
    
    subgraph "数据存储"
        C1[(平台配置)]
        C2[(仓库信息)]
        C3[(LLM配置)]
        C4[(Review记录)]
        C5[(修复建议)]
    end
    
    subgraph "数据输出"
        D1[Review列表页面]
        D2[Review详情页面]
        D3[GitLab MR评论]
    end
    
    A1 --> B1
    A2 --> B2
    B1 --> C1
    B1 --> C2
    B1 --> C3
    B2 --> B3
    B3 --> B4
    B4 --> C4
    B4 --> C5
    
    C1 --> D1
    C2 --> D1
    C4 --> D1
    C4 --> D2
    C5 --> D2
    C4 --> D3
    C5 --> D3
```

---

## 6. 页面导航关系图

```mermaid
graph TD
    A[登录页] --> B[系统设置]
    B --> C[仓库管理]
    B --> D[Review记录]
    C --> E[导入仓库Modal]
    D --> F[Review详情]
    
    subgraph "系统设置Tabs"
        B1[GitLab配置]
        B2[LLM配置]
        B3[Webhook配置]
        B4[提示词模板]
    end
    
    B --> B1
    B --> B2
    B --> B3
    B --> B4
    
    style A fill:#ffe1e1
    style B fill:#e1f5ff
    style C fill:#e1f5ff
    style D fill:#e8f5e9
    style F fill:#e8f5e9
```

---

## 7. API接口关系图

```mermaid
graph TB
    subgraph "认证接口"
        A1[POST /api/auth/login]
        A2[POST /api/auth/logout]
    end
    
    subgraph "平台配置接口"
        B1[GET /api/platform/config]
        B2[PUT /api/platform/config]
        B3[POST /api/platform/test]
        B4[GET /api/platform/repositories]
    end
    
    subgraph "仓库管理接口"
        C1[GET /api/repositories]
        C2[POST /api/repositories/batch]
        C3[PUT /api/repositories/:id]
        C4[DELETE /api/repositories/:id]
        C5[POST /api/repositories/:id/webhook]
    end
    
    subgraph "LLM配置接口"
        D1[GET /api/llm/providers]
        D2[POST /api/llm/providers]
        D3[PUT /api/llm/providers/:id]
        D4[POST /api/llm/providers/:id/test]
        D5[GET /api/llm/models]
        D6[POST /api/llm/models]
    end
    
    subgraph "Review记录接口"
        E1[GET /api/reviews]
        E2[GET /api/reviews/:id]
        E3[GET /api/reviews/:id/suggestions]
    end
    
    subgraph "Webhook接口"
        F1[POST /api/webhook]
    end
    
    subgraph "系统配置接口"
        G1[GET /api/system/config]
        G2[PUT /api/system/config]
        G3[GET /api/system/health]
    end
    
    A1 --> B1
    B2 --> C2
    C2 --> C5
    C3 --> D1
    F1 --> E1
    E1 --> E2
    E2 --> E3
```

---

## 8. 状态管理关系图（前端）

```mermaid
graph TB
    subgraph "Zustand Stores"
        S1[authStore<br/>用户认证]
        S2[platformStore<br/>平台配置]
        S3[repositoryStore<br/>仓库管理]
        S4[llmStore<br/>LLM配置]
        S5[reviewStore<br/>Review记录]
    end
    
    subgraph "页面组件"
        P1[系统设置]
        P2[仓库管理]
        P3[Review列表]
        P4[Review详情]
    end
    
    S1 --> P1
    S1 --> P2
    S1 --> P3
    
    S2 --> P1
    S2 --> P2
    
    S3 --> P2
    S3 --> P3
    
    S4 --> P1
    S4 --> P2
    
    S5 --> P3
    S5 --> P4
```

---

## 9. 数据库表关系图（MVP简化版）

```mermaid
erDiagram
    users ||--o{ git_platform_configs : "creates"
    git_platform_configs ||--o{ repositories : "contains"
    llm_providers ||--o{ llm_models : "has"
    repositories }o--|| llm_models : "uses"
    repositories ||--o{ review_results : "generates"
    review_results ||--o{ fix_suggestions : "contains"

    users {
        bigint id PK
        string username UK
        string password
        string role
        timestamp created_at
    }
    
    git_platform_configs {
        bigint id PK
        string name
        string base_url
        string access_token
        boolean is_active
    }
    
    repositories {
        bigint id PK
        bigint platform_config_id FK
        bigint llm_model_id FK
        string repo_name
        string webhook_id
        boolean is_webhook_active
    }
    
    llm_providers {
        bigint id PK
        string provider_type
        string api_key
        string api_base_url
    }
    
    llm_models {
        bigint id PK
        bigint provider_id FK
        string model_name
        boolean is_active
    }
    
    review_results {
        bigint id PK
        bigint repository_id FK
        bigint llm_model_id FK
        string mr_url
        text raw_result
        int overall_score
        text summary
    }
    
    fix_suggestions {
        bigint id PK
        bigint review_result_id FK
        string file_path
        int line_start
        int line_end
        string severity
        text description
        text suggestion
    }
```

---

## 10. 技术栈关系图

```mermaid
graph TB
    subgraph "前端技术栈"
        FE1[React 18]
        FE2[Ant Design 5]
        FE3[TypeScript 5]
        FE4[Vite]
        FE5[Zustand]
        FE6[Axios]
    end
    
    subgraph "后端技术栈"
        BE1[Go 1.21]
        BE2[Gin]
        BE3[GORM]
        BE4[Asynq]
        BE5[JWT]
    end
    
    subgraph "数据存储"
        DS1[SQLite/MySQL]
        DS2[Redis]
    end
    
    subgraph "外部服务"
        EX1[GitLab API]
        EX2[LLM API]
    end
    
    FE1 --> FE2
    FE1 --> FE3
    FE1 --> FE5
    FE1 --> FE6
    FE4 --> FE1
    
    BE1 --> BE2
    BE1 --> BE3
    BE1 --> BE4
    BE1 --> BE5
    
    BE3 --> DS1
    BE4 --> DS2
    
    BE2 --> EX1
    BE4 --> EX1
    BE4 --> EX2
    
    FE6 --> BE2
```

---

## 11. 功能优先级矩阵

```mermaid
quadrantChart
    title MVP功能优先级矩阵
    x-axis 低实现难度 --> 高实现难度
    y-axis 低业务价值 --> 高业务价值
    quadrant-1 高价值高难度（后续）
    quadrant-2 高价值低难度（MVP）
    quadrant-3 低价值低难度（可选）
    quadrant-4 低价值高难度（放弃）
    
    用户登录: [0.2, 0.9]
    GitLab配置: [0.3, 0.95]
    仓库导入: [0.4, 0.9]
    Webhook配置: [0.5, 0.85]
    LLM配置: [0.4, 0.95]
    Review记录: [0.5, 1.0]
    修复建议: [0.6, 0.9]
    自动修复: [0.9, 0.8]
    仓库组: [0.6, 0.4]
    通知渠道: [0.5, 0.5]
    Dashboard: [0.4, 0.3]
    多平台支持: [0.7, 0.6]
```

**说明**:
- **象限2（左上）**: MVP核心功能，优先实现
- **象限1（右上）**: 高价值但复杂，后续版本实现
- **象限3（左下）**: 低优先级，可选
- **象限4（右下）**: 暂不实现

---

## 12. 开发阶段关系图

```mermaid
gantt
    title MVP开发阶段（6周）
    dateFormat  YYYY-MM-DD
    section 基础框架
    数据库设计           :done, des1, 2025-02-01, 2d
    用户认证             :done, des2, 2025-02-03, 3d
    前端框架搭建         :done, des3, 2025-02-03, 3d
    
    section 配置管理
    GitLab配置          :active, des4, 2025-02-06, 4d
    LLM配置管理         :active, des5, 2025-02-08, 3d
    系统设置页面        :des6, 2025-02-10, 2d
    
    section 仓库管理
    仓库导入功能        :des7, 2025-02-12, 4d
    Webhook配置         :des8, 2025-02-14, 3d
    仓库列表页面        :des9, 2025-02-16, 2d
    
    section Review功能
    Webhook接收         :des10, 2025-02-18, 3d
    Review任务调度      :des11, 2025-02-20, 4d
    LLM调用集成         :des12, 2025-02-22, 3d
    Review页面          :des13, 2025-02-24, 4d
    
    section 测试部署
    功能测试            :des14, 2025-02-28, 3d
    Docker部署          :des15, 2025-03-03, 2d
    文档完善            :des16, 2025-03-05, 2d
```

---

## 总结

本文档通过12个关系图全面展示了MVP版本的：

1. ✅ **系统架构**: 展示整体分层结构
2. ✅ **功能模块**: 展示模块间依赖关系
3. ✅ **用户流程**: 展示用户操作全流程
4. ✅ **Review流程**: 展示核心业务流程
5. ✅ **数据流**: 展示数据流转过程
6. ✅ **页面导航**: 展示前端页面结构
7. ✅ **API关系**: 展示接口调用关系
8. ✅ **状态管理**: 展示前端状态设计
9. ✅ **数据模型**: 展示表结构关系
10. ✅ **技术栈**: 展示技术选型关系
11. ✅ **优先级矩阵**: 展示功能价值分析
12. ✅ **开发阶段**: 展示实施计划

这些关系图能够帮助团队快速理解MVP的整体设计和实施路径。
