# 技术栈选型方案

## 1. 后端技术栈：Go 语言

### 1.1 选择理由

- ✅ **高性能**: 原生协程支持，适合处理高并发Webhook请求
- ✅ **简单部署**: 单一可执行文件，无运行时依赖
- ✅ **强类型**: 编译期类型检查，减少运行时错误
- ✅ **丰富生态**: Git操作、HTTP客户端、ORM等库成熟
- ✅ **跨平台**: 轻松编译为Linux/Windows/macOS可执行文件

### 1.2 核心框架与组件

| 组件 | 技术选型 | 版本 | 用途 |
|------|---------|------|------|
| **Web框架** | Gin | v1.10+ | HTTP路由、中间件、参数验证 |
| **ORM** | GORM | v1.25+ | 数据库操作，支持SQLite/MySQL |
| **配置管理** | Viper | v1.18+ | 配置文件读取、环境变量 |
| **日志** | Zap | v1.26+ | 结构化日志 |
| **任务队列** | Asynq (Redis) | v0.24+ | 异步任务处理 |
| **Git操作** | go-git | v5.11+ | Git仓库克隆、分支管理 |
| **HTTP客户端** | Resty | v2.11+ | 调用Git平台API |
| **JWT认证** | jwt-go | v5.2+ | 用户认证 |
| **数据验证** | validator | v10.19+ | 请求参数验证 |
| **加密** | crypto | stdlib | Token加密存储 |
| **WebSocket** | gorilla/websocket | v1.5+ | 实时日志推送 |

### 1.3 项目依赖 (go.mod)

```go
module github.com/your-org/ai-codereview

go 1.21

require (
    github.com/gin-gonic/gin v1.10.0
    gorm.io/gorm v1.25.7
    gorm.io/driver/sqlite v1.5.5
    gorm.io/driver/mysql v1.5.4
    github.com/spf13/viper v1.18.2
    go.uber.org/zap v1.26.0
    github.com/hibiken/asynq v0.24.1
    github.com/go-git/go-git/v5 v5.11.0
    github.com/go-resty/resty/v2 v2.11.0
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/go-playground/validator/v10 v10.19.0
    github.com/gorilla/websocket v1.5.1
    github.com/google/uuid v1.6.0
)
```

---

## 2. 前端技术栈：React + Ant Design

### 2.1 选择理由

经过对比 Vue 3、React、Svelte 等框架，最终选择 **React + Ant Design**：

| 框架 | 优势 | 劣势 | 评分 |
|------|------|------|------|
| **Vue 3 + Element Plus** | 上手快、文档好、国内生态好 | Element Plus文档不完善、部分组件bug多 | ⭐⭐⭐⭐ |
| **React + Ant Design** | 企业级组件库成熟、稳定性高、TypeScript支持好 | 学习曲线稍陡 | ⭐⭐⭐⭐⭐ |
| **Svelte + SvelteUI** | 性能最优、打包体积小 | 生态不成熟、企业级组件库少 | ⭐⭐⭐ |

**最终选择：React + Ant Design**

**理由：**
1. ✅ **Ant Design** 是最成熟的企业级组件库，组件丰富且稳定
2. ✅ **TypeScript** 支持极佳，类型定义完整
3. ✅ **ProComponents** 提供高级表格、表单、布局等开箱即用组件
4. ✅ **简单易维护**：组件化思想清晰，代码可读性强
5. ✅ **易迭代**：庞大的社区和丰富的第三方库

### 2.2 核心框架与组件

| 组件 | 技术选型 | 版本 | 用途 |
|------|---------|------|------|
| **框架** | React | v18.2+ | UI框架 |
| **UI库** | Ant Design | v5.15+ | 企业级组件库 |
| **高级组件** | ProComponents | v2.6+ | ProTable、ProForm等 |
| **路由** | React Router | v6.22+ | 客户端路由 |
| **状态管理** | Zustand | v4.5+ | 轻量级状态管理 |
| **请求库** | Axios | v1.6+ | HTTP请求 |
| **代码编辑器** | Monaco Editor | v0.47+ | 代码查看、Diff展示 |
| **Markdown** | react-markdown | v9.0+ | Markdown渲染 |
| **图表** | Apache ECharts | v5.5+ | 数据可视化 |
| **构建工具** | Vite | v5.1+ | 快速构建 |
| **类型检查** | TypeScript | v5.3+ | 类型安全 |

### 2.3 项目依赖 (package.json)

```json
{
  "name": "ai-codereview-web",
  "version": "1.0.0",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.22.0",
    "antd": "^5.15.0",
    "@ant-design/pro-components": "^2.6.0",
    "@ant-design/icons": "^5.3.0",
    "zustand": "^4.5.0",
    "axios": "^1.6.7",
    "@monaco-editor/react": "^4.6.0",
    "react-markdown": "^9.0.1",
    "echarts": "^5.5.0",
    "echarts-for-react": "^3.0.2",
    "dayjs": "^1.11.10"
  },
  "devDependencies": {
    "@types/react": "^18.2.56",
    "@types/react-dom": "^18.2.19",
    "@vitejs/plugin-react": "^4.2.1",
    "typescript": "^5.3.3",
    "vite": "^5.1.4"
  }
}
```

---

## 3. 数据库技术栈

### 3.1 双数据库支持方案

采用 **GORM** 作为ORM，通过驱动切换实现双数据库支持：

```go
// SQLite 驱动
import "gorm.io/driver/sqlite"

// MySQL 驱动
import "gorm.io/driver/mysql"
```

### 3.2 配置切换

```yaml
# config.yaml
database:
  driver: "sqlite"  # 或 "mysql"
  
  # SQLite配置
  sqlite:
    path: "./data/ai-codereview.db"
  
  # MySQL配置
  mysql:
    host: "localhost"
    port: 3306
    user: "root"
    password: "password"
    dbname: "ai_codereview"
    charset: "utf8mb4"
    parse_time: true
    loc: "Local"
```

### 3.3 数据库初始化代码设计

```go
func InitDB(config *Config) (*gorm.DB, error) {
    var dialector gorm.Dialector
    
    switch config.Database.Driver {
    case "sqlite":
        dialector = sqlite.Open(config.Database.SQLite.Path)
    case "mysql":
        dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
            config.Database.MySQL.User,
            config.Database.MySQL.Password,
            config.Database.MySQL.Host,
            config.Database.MySQL.Port,
            config.Database.MySQL.DBName,
            config.Database.MySQL.Charset,
            config.Database.MySQL.ParseTime,
            config.Database.MySQL.Loc,
        )
        dialector = mysql.Open(dsn)
    default:
        return nil, fmt.Errorf("unsupported database driver: %s", config.Database.Driver)
    }
    
    db, err := gorm.Open(dialector, &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    
    if err != nil {
        return nil, err
    }
    
    // 自动迁移
    if err := AutoMigrate(db); err != nil {
        return nil, err
    }
    
    return db, nil
}
```

---

## 4. 任务队列与异步处理

### 4.1 选型：Asynq (Redis-based)

**为什么选择 Asynq？**

- ✅ Go原生支持，API简洁
- ✅ 基于Redis，部署简单
- ✅ 支持任务优先级、重试、延迟执行
- ✅ 内置Web UI监控
- ✅ 支持任务取消和暂停

**替代方案对比：**

| 方案 | 优势 | 劣势 | 选择 |
|------|------|------|------|
| **Asynq** | 简单、轻量、监控好 | 依赖Redis | ✅ **推荐** |
| Machinery | 功能丰富 | 配置复杂 | ❌ |
| RabbitMQ | 企业级 | 运维成本高 | ❌ |

### 4.2 任务定义示例

```go
// 代码审查任务
type CodeReviewTask struct {
    WebhookID   string
    RepositoryID uint
    MRNumber    int
}

// 自动修复任务
type AutoFixTask struct {
    SuggestionID uint
    RepositoryID uint
}
```

---

## 5. 其他技术选型

### 5.1 缓存

- **Redis**: 用于Asynq任务队列、Session存储、API限流

### 5.2 文件存储

- **本地文件系统**: 临时Git仓库克隆目录
- **可选 MinIO**: 存储大量日志文件（未来扩展）

### 5.3 监控与日志

| 组件 | 技术 | 用途 |
|------|------|------|
| **应用日志** | Zap | 结构化日志输出 |
| **访问日志** | Gin中间件 | HTTP请求日志 |
| **性能监控** | pprof | Go性能分析 |
| **任务监控** | Asynq Web UI | 任务队列监控 |

---

## 6. 开发工具链

### 6.1 Go工具

- **golangci-lint**: 代码静态检查
- **gofmt/goimports**: 代码格式化
- **air**: 热重载开发

### 6.2 前端工具

- **ESLint**: 代码规范检查
- **Prettier**: 代码格式化
- **TypeScript**: 类型检查

### 6.3 版本控制

- **Git**: 源码管理
- **GitHub/GitLab**: 代码托管

---

## 7. 部署方案

### 7.1 Docker部署

```dockerfile
# 后端Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o ai-codereview ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/ai-codereview .
COPY configs ./configs
EXPOSE 8080
CMD ["./ai-codereview"]
```

### 7.2 Docker Compose

```yaml
version: '3.8'

services:
  backend:
    build: ./backend
    ports:
      - "8080:8080"
    depends_on:
      - redis
      - mysql
    environment:
      - DB_DRIVER=mysql
      - REDIS_URL=redis://redis:6379
  
  frontend:
    build: ./web
    ports:
      - "3000:80"
  
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
  
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: ai_codereview
    ports:
      - "3306:3306"
  
  asynq-worker:
    build: ./backend
    command: ["./ai-codereview", "worker"]
    depends_on:
      - redis
```

---

## 8. 技术栈总结

### 8.1 核心技术

```
后端：Go 1.21 + Gin + GORM + Asynq
前端：React 18 + Ant Design 5 + TypeScript 5
数据库：SQLite / MySQL (GORM双支持)
缓存/队列：Redis 7
```

### 8.2 技术栈优势

1. ✅ **高性能**: Go协程 + Redis队列
2. ✅ **易部署**: 单一可执行文件 + Docker
3. ✅ **易维护**: 强类型 + 清晰架构
4. ✅ **易扩展**: 模块化设计 + 中间件机制
5. ✅ **企业级**: Ant Design + 成熟组件库

### 8.3 开发体验

- **后端**: 简洁的Go语法 + 丰富的标准库
- **前端**: React生态 + TypeScript类型安全
- **调试**: 热重载 + 详细日志 + pprof性能分析
- **测试**: Go原生测试 + React Testing Library

---

**下一步**: 基于此技术栈设计项目目录结构
