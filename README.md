# AI-Codereview-Gitlab

> AI 驱动的代码审查自动化系统

## 📖 项目简介

> ⚠️ **项目状态**: 当前处于**设计阶段**，技术文档已完成，即将开始代码实现。

AI-Codereview-Gitlab 是一个基于 **Go + React** 架构设计的代码审查自动化平台，通过集成多种 Git 平台（GitLab/GitHub/Gitea）和 AI 大语言模型（LLM），实现智能化的代码审查、修复建议生成以及自动修复功能。

### 核心特性

✨ **多平台支持**

- 支持 GitLab（自托管/SaaS）、GitHub、Gitea
- 一键导入代码仓库
- 自动配置 Webhook

🤖 **智能 AI 审查**

- 集成多种 LLM 供应商（OpenAI、DeepSeek、ZhipuAI、Qwen、Ollama）
- 结构化审查输出
- 自定义提示词模板

🔧 **自动修复**

- 一键触发修复
- 实时日志流
- 分支管理（创建、删除、合并）

📢 **通知集成**

- 支持钉钉、企业微信、飞书
- 可配置通知内容
- 测试连通性

📊 **仓库组管理**

- 多仓库统一管理
- 组级别配置（LLM、通知、提示词）
- 灵活的权限控制

---

## 🏗️ 技术架构

### 后端技术栈

- **语言**: Go 1.21+
- **Web 框架**: Gin
- **ORM**: GORM（支持 SQLite 和 MySQL）
- **任务队列**: Asynq (Redis)
- **实时通信**: WebSocket (Gorilla)
- **Git 操作**: go-git

### 前端技术栈

- **框架**: React 18
- **UI 库**: Ant Design 5
- **语言**: TypeScript 5
- **构建工具**: Vite
- **状态管理**: Zustand
- **路由**: React Router v6

### 基础设施

- **数据库**: SQLite / MySQL 8.0+
- **缓存/队列**: Redis 7
- **容器化**: Docker + Docker Compose
- **反向代理**: Nginx

---

## 📚 文档导航

### 设计文档（docs/）

完整的技术设计文档位于 `docs/` 目录：

| 文档                                                      | 说明                     | 阅读时长 |
| --------------------------------------------------------- | ------------------------ | -------- |
| [README.md](docs/README.md)                               | 文档索引和快速导航       | 5 分钟   |
| [01-tech-stack.md](docs/01-tech-stack.md)                 | 技术栈选型论证           | 15 分钟  |
| [02-project-structure.md](docs/02-project-structure.md)   | 项目目录结构设计         | 20 分钟  |
| [03-database-design.md](docs/03-database-design.md)       | 数据模型设计（15 张表）  | 30 分钟  |
| [04-feature-list.md](docs/04-feature-list.md)             | 功能清单（118 项功能）   | 40 分钟  |
| [05-page-design.md](docs/05-page-design.md)               | 页面设计（23 个页面）    | 35 分钟  |
| [06-interaction-design.md](docs/06-interaction-design.md) | 交互逻辑设计             | 25 分钟  |
| [07-api-design.md](docs/07-api-design.md)                 | API 接口设计（80+ 接口） | 50 分钟  |

**推荐阅读路径：**

- **快速了解**: README → 功能清单 → 页面设计
- **后端开发**: 技术栈 → 目录结构 → 数据模型 → API 设计
- **前端开发**: 技术栈 → 页面设计 → 交互逻辑 → API 设计

---

## 🚀 快速开始

> ⚠️ **注意**: 项目当前处于设计阶段，以下为计划中的使用方式，暂不可用。

### 前置要求

**查看设计文档:**

- 详细的技术设计文档位于 `docs/` 目录
- 推荐从 [docs/README.md](docs/README.md) 开始阅读

**未来实现后的前置要求:**

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- Redis 7+
- SQLite 3 或 MySQL 8.0+

### 计划中的使用方式

完整的安装和使用文档将在代码实现后提供。预计的使用流程：

```bash
# 1. 克隆项目
git clone https://github.com/your-org/ai-codereview-gitlab.git
cd ai-codereview-gitlab

# 2. 使用 Docker Compose 启动（推荐）
docker-compose up -d

# 3. 访问应用
# 前端：http://localhost:3000
# 后端 API：http://localhost:8080
```

详细的部署和配置指南请参考 [设计文档](docs/README.md)。

## 📦 项目结构

```
ai-codereview/
├── cmd/                    # 应用入口
│   ├── api/               # API 服务器
│   ├── worker/            # 异步任务 Worker
│   └── migrate/           # 数据库迁移工具
├── internal/              # 内部包（不可外部引用）
│   ├── api/               # HTTP 处理器
│   ├── service/           # 业务逻辑层
│   ├── repository/        # 数据访问层
│   ├── model/             # 数据模型
│   ├── dto/               # 数据传输对象
│   ├── middleware/        # 中间件
│   ├── webhook/           # Webhook 处理
│   ├── llm/               # LLM 客户端
│   ├── gitops/            # Git 操作
│   ├── notification/      # 通知服务
│   └── task/              # 异步任务
├── pkg/                   # 公共工具包
├── web/                   # React 前端
│   ├── src/
│   │   ├── pages/         # 页面组件
│   │   ├── components/    # 通用组件
│   │   ├── api/           # API 客户端
│   │   ├── stores/        # Zustand 状态
│   │   └── router/        # 路由配置
├── config/                # 配置文件
├── migrations/            # SQL 迁移文件
├── docs/                  # 技术设计文档
├── scripts/               # 部署脚本
├── docker-compose.yml     # Docker Compose 配置
└── Dockerfile             # Docker 镜像构建
```

---

## 🎯 核心功能

### 1. Git 平台管理

- 支持多 GitLab 实例
- GitHub 和 Gitea 集成
- 自动 Webhook 配置
- 连接测试

### 2. 代码仓库管理

- 一键导入仓库
- 批量导入
- Webhook 管理
- 仓库分组

### 3. 仓库组管理

- 多仓库统一配置
- 自定义提示词模板
- 指定 LLM 模型
- 配置通知渠道

### 4. LLM 配置

- 多供应商支持
- 动态获取可用模型
- 连接测试
- 每个供应商支持多个模型

### 5. 通知渠道

- 钉钉机器人
- 企业微信机器人
- 飞书机器人
- 通知内容自定义

### 6. Review 记录

- 结构化审查结果
- 修复建议列表
- 按严重程度分类
- 按文件、行号定位

### 7. 自动修复

- 一键触发修复
- 实时日志流（WebSocket）
- 修复分支管理
- 支持重复修复
- 可选自动合并

---

## 🔧 配置说明

### 环境变量

```bash
# 数据库配置
DB_TYPE=sqlite          # sqlite 或 mysql
DB_DSN=data/app.db      # SQLite: 文件路径，MySQL: 连接字符串

# Redis 配置
REDIS_URL=redis://localhost:6379/0

# 服务器配置
API_PORT=8080
WORKER_CONCURRENCY=10

# JWT 配置
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

# 加密密钥
ENCRYPTION_KEY=base64-encoded-key
```

### 配置文件示例

参考 `config/config.example.yaml` 进行配置。

---

## 📈 开发计划

### Phase 1: 基础框架（Week 1-2）

- [x] 设计阶段完成
- [ ] 项目代码初始化
- [ ] 数据库模型实现
- [ ] 基础 API 框架
- [ ] 前端项目初始化

### Phase 2: 核心功能（Week 3-5）

- [ ] 用户认证
- [ ] Git 平台管理
- [ ] 代码仓库管理
- [ ] LLM 配置管理
- [ ] 通知渠道管理

### Phase 3: Review 功能（Week 6-7）

- [ ] Webhook 接收
- [ ] Review 任务调度
- [ ] LLM 调用
- [ ] 结果结构化存储

### Phase 4: 自动修复（Week 8-10）

- [ ] Snow-CLI 集成
- [ ] 修复任务执行
- [ ] 实时日志流
- [ ] 分支管理

### Phase 5: 完善与部署（Week 11-12）

- [ ] 单元测试
- [ ] 集成测试
- [ ] Docker 镜像
- [ ] 部署文档

---

## 🤝 贡献指南

欢迎贡献代码、提出问题或建议！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交修改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

## 📄 许可证

本项目采用 [Apache License 2.0](LICENSE) 许可证。

---

## 💬 联系方式

- Issue Tracker: [GitHub Issues](https://github.com/your-org/ai-codereview-gitlab/issues)
- 文档: [docs/README.md](docs/README.md)

---

**Built with ❤️ by Snow AI**
