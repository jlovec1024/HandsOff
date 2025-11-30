# AI-Codereview-Gitlab

> AI-powered automated code review system for GitLab, GitHub, and Gitea

## Overview

**AI-Codereview-Gitlab** is an intelligent code review automation platform that leverages Large Language Models (LLM) to perform automated code reviews during merge requests or code pushes. The system automatically triggers webhook events when developers submit code changes through GitLab merge requests or push events, invokes third-party AI models for code review, and publishes review results directly as comments on the corresponding Merge Request or Commit.

**⚠️ Current Project Status: Design Phase**

This project is currently in the **design and planning phase**. A complete redesign from the original Python/Flask system to a modern **Go + React** architecture is underway. All implementation code has been removed, and comprehensive technical design documentation has been completed.

**What's Available:**
- ✅ Complete technical design documentation (8 documents, 200+ pages)
- ✅ Detailed architecture specifications (Go backend + React frontend)
- ✅ Database schema design (15 tables, dual SQLite/MySQL support)
- ✅ API interface design (80+ RESTful endpoints)
- ✅ Frontend page design (23 pages with interaction flows)
- ✅ Feature breakdown (118 feature points across 8 modules)

**What's Next:**
- 🚧 Backend implementation (Go 1.21+, Gin framework)
- 🚧 Frontend implementation (React 18, Ant Design 5)
- 🚧 Database migrations and models
- 🚧 Docker deployment configuration

## Technology Stack

### Planned Backend Stack (Go 1.21+)

- **Language**: Go 1.21+
- **Web Framework**: Gin v1.10+
- **ORM**: GORM v1.25+ (dual SQLite/MySQL support)
- **Task Queue**: Asynq v0.24+ (Redis-based)
- **Git Operations**: go-git v5.11+
- **WebSocket**: Gorilla WebSocket v1.5+
- **Configuration**: Viper v1.18+
- **Logging**: Zap v1.26+
- **Authentication**: JWT (jwt-go v5.2+)
- **Validation**: validator v10.19+

### Planned Frontend Stack (React 18)

- **Framework**: React 18.2+
- **UI Library**: Ant Design 5.x
- **Language**: TypeScript 5.x
- **Build Tool**: Vite 5.x
- **State Management**: Zustand
- **Routing**: React Router v6
- **HTTP Client**: Axios
- **Code Editor**: Monaco Editor (for prompt templates)

### Infrastructure

- **Database**: SQLite (development) / MySQL 8.0+ (production)
- **Cache/Queue**: Redis 7+
- **Containerization**: Docker + Docker Compose
- **Reverse Proxy**: Nginx (planned)
- **Process Manager**: Supervisor (planned)

## Project Structure

**Current Structure (Design Phase):**

```
ai-codereview-gitlab/
├── .git/                  # Fresh Git repository (no history)
├── .gitignore             # Go + React gitignore rules
├── LICENSE                # Apache License 2.0
├── README.md              # Project overview and quick start
├── SNOW.md                # This file - project context
└── docs/                  # Complete technical design documentation
    ├── README.md                   # Documentation index
    ├── 01-tech-stack.md           # Technology stack selection (15 min read)
    ├── 02-project-structure.md    # Directory structure design (20 min read)
    ├── 03-database-design.md      # Database schema (15 tables, 30 min read)
    ├── 04-feature-list.md         # Feature breakdown (118 items, 40 min read)
    ├── 05-page-design.md          # Page design (23 pages, 35 min read)
    ├── 06-interaction-design.md   # Interaction logic (25 min read)
    └── 07-api-design.md           # API design (80+ endpoints, 50 min read)
```

**Planned Structure (Post-Implementation):**

```
ai-codereview-gitlab/
├── cmd/                   # Application entry points
│   ├── api/              # API server (main)
│   ├── worker/           # Async task worker
│   └── migrate/          # Database migration tool
├── internal/             # Internal packages (not importable)
│   ├── api/              # HTTP handlers (controllers)
│   ├── service/          # Business logic layer
│   ├── repository/       # Data access layer (DAO)
│   ├── model/            # Database entities (GORM models)
│   ├── dto/              # Data Transfer Objects
│   ├── middleware/       # HTTP middleware (auth, logging)
│   ├── webhook/          # Webhook event handlers
│   ├── llm/              # LLM client abstraction
│   ├── gitops/           # Git operations (clone, branch)
│   ├── notification/     # IM notifications (DingTalk, WeCom, Feishu)
│   └── task/             # Async task definitions
├── pkg/                  # Shared utility packages
│   ├── config/           # Configuration management
│   ├── logger/           # Logging utilities
│   ├── crypto/           # Encryption/decryption
│   └── validator/        # Custom validators
├── web/                  # React frontend
│   ├── public/           # Static assets
│   ├── src/
│   │   ├── pages/        # Page components
│   │   ├── components/   # Reusable components
│   │   ├── api/          # API client layer
│   │   ├── stores/       # Zustand state stores
│   │   ├── router/       # Routing configuration
│   │   ├── hooks/        # Custom React hooks
│   │   └── utils/        # Utility functions
│   ├── package.json
│   └── vite.config.ts
├── config/               # Configuration files
│   ├── config.yaml       # Main config (gitignored)
│   └── config.example.yaml
├── migrations/           # SQL migration files
│   └── 001_initial_schema.sql
├── scripts/              # Deployment/build scripts
├── docs/                 # Technical documentation
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

## Key Features (Planned)

### 🔐 1. Multi-Platform Git Integration
- Support for multiple GitLab instances (self-hosted + SaaS)
- GitHub and Gitea integration
- One-click repository import
- Automatic webhook configuration
- Custom webhook callback URLs

### 🤖 2. LLM Configuration Management
- Support for multiple LLM providers:
  - OpenAI (GPT-4, GPT-3.5)
  - DeepSeek
  - ZhipuAI (GLM-4-Flash)
  - Qwen (Alibaba Cloud)
  - Ollama (local deployment)
- Dynamic model fetching from provider APIs
- Connection testing from UI
- Multiple models per provider

### 📦 3. Repository Group Management
- Group multiple repositories for unified configuration
- Custom prompt templates per group
- Dedicated LLM models per group
- Group-level notification channels

### 📢 4. Notification Channels
- DingTalk (钉钉) robot integration
- WeCom (企业微信) robot integration
- Feishu (飞书) robot integration
- Configurable notification content
- Connection testing from UI

### 📝 5. Structured Review Results
- JSON-formatted review output
- Severity classification (high/medium/low)
- Category tagging (security, performance, style, etc.)
- File and line number mapping
- Fix suggestion list

### 🔧 6. Auto-Fix Capabilities
- One-click fix trigger for each suggestion
- Automated workflow:
  1. Clone repository
  2. Create fix branch
  3. Execute Snow-CLI for code modification
  4. Commit changes
  5. Push to remote
- Real-time progress tracking via WebSocket
- Execution log streaming
- Fix branch management (view, delete, merge)
- Support for re-fixing the same issue

### 👥 7. User Management
- Role-based access control (Admin/User)
- JWT authentication
- Session management

### 📊 8. Dashboard & Statistics
- Review statistics (planned)
- Project and developer metrics (planned)
- Data visualization charts (planned)

## Design Documentation

The project has comprehensive technical design documentation in the `docs/` directory:

### Quick Navigation

**For Product Managers:**
- [Feature List](docs/04-feature-list.md) → [Page Design](docs/05-page-design.md)

**For Backend Developers:**
- [Tech Stack](docs/01-tech-stack.md) → [Project Structure](docs/02-project-structure.md) → [Database Design](docs/03-database-design.md) → [API Design](docs/07-api-design.md)

**For Frontend Developers:**
- [Tech Stack](docs/01-tech-stack.md) → [Page Design](docs/05-page-design.md) → [Interaction Design](docs/06-interaction-design.md) → [API Design](docs/07-api-design.md)

### Document Summary

| Document | Description | Status |
|----------|-------------|--------|
| [README.md](docs/README.md) | Documentation index and navigation | ✅ Complete |
| [01-tech-stack.md](docs/01-tech-stack.md) | Technology selection rationale (Go vs Python/Node.js, React vs Vue) | ✅ Complete |
| [02-project-structure.md](docs/02-project-structure.md) | Go standard layout, React structure, layered architecture | ✅ Complete |
| [03-database-design.md](docs/03-database-design.md) | 15 tables with GORM models, SQLite/MySQL compatibility | ✅ Complete |
| [04-feature-list.md](docs/04-feature-list.md) | 118 feature points across 8 modules (P0/P1/P2 priority) | ✅ Complete |
| [05-page-design.md](docs/05-page-design.md) | 23 pages with layouts, routing, and component design | ✅ Complete |
| [06-interaction-design.md](docs/06-interaction-design.md) | State management (Zustand), data flows, WebSocket design | ✅ Complete |
| [07-api-design.md](docs/07-api-design.md) | 80+ RESTful endpoints, request/response formats, error codes | ✅ Complete |

**Total Reading Time:** ~3.5 hours for complete understanding

## Getting Started

### Prerequisites

Since the project is in the design phase, you'll need:

- **For Reading Documentation:**
  - Markdown viewer or IDE with Markdown support
  - Web browser for viewing diagrams

- **For Future Implementation:**
  - Go 1.21+
  - Node.js 18+
  - Docker & Docker Compose
  - Redis 7+
  - SQLite 3 or MySQL 8.0+

### Installation (Not Yet Available)

The implementation phase has not started. Please refer to the design documents to understand the planned architecture.

**Roadmap for implementation:**

1. **Week 1-2: Foundation**
   - Initialize Go module
   - Initialize React project with Vite
   - Setup Docker Compose for development
   - Implement database models and migrations

2. **Week 3-5: Core Features**
   - User authentication system
   - Git platform management
   - Repository management
   - LLM configuration management
   - Notification channel management

3. **Week 6-7: Review Engine**
   - Webhook receiver
   - Review task scheduler
   - LLM invocation
   - Structured result storage

4. **Week 8-10: Auto-Fix**
   - Snow-CLI integration
   - Fix task execution
   - Real-time log streaming
   - Branch management

5. **Week 11-12: Polish & Deploy**
   - Unit tests
   - Integration tests
   - Docker images
   - Deployment documentation

### Usage (Placeholder)

```bash
# Backend (once implemented)
go mod tidy
go run cmd/api/main.go
go run cmd/worker/main.go

# Frontend (once implemented)
cd web
npm install
npm run dev

# Docker Compose (once implemented)
docker-compose up -d
```

## Development

### Current Phase: Design Complete ✅

All design documents have been completed. The next step is to begin implementation.

### Recommended Reading Order

1. Start with [docs/README.md](docs/README.md) for an overview
2. Review [01-tech-stack.md](docs/01-tech-stack.md) to understand technology choices
3. Read [03-database-design.md](docs/03-database-design.md) for data model
4. Study [07-api-design.md](docs/07-api-design.md) for API contracts
5. Review [02-project-structure.md](docs/02-project-structure.md) before coding

### Contributing

Contributions are welcome! Since the project is in the design phase:

1. **For Design Feedback:**
   - Review the design documents
   - Open an issue with suggestions
   - Propose improvements via PR

2. **For Implementation:**
   - Wait for the initial implementation framework
   - Check the project roadmap
   - Coordinate with maintainers

## Configuration (Planned)

### Environment Variables

The following environment variables will be required (not yet implemented):

```bash
# Database
DB_TYPE=sqlite                    # sqlite or mysql
DB_DSN=data/app.db               # SQLite path or MySQL connection string

# Redis
REDIS_URL=redis://localhost:6379/0

# Server
API_PORT=8080
WORKER_CONCURRENCY=10

# JWT Authentication
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

# Encryption
ENCRYPTION_KEY=base64-encoded-key

# LLM Providers (at least one required)
OPENAI_API_KEY=sk-...
DEEPSEEK_API_KEY=...
ZHIPUAI_API_KEY=...
QWEN_API_KEY=...
OLLAMA_BASE_URL=http://localhost:11434

# Git Platforms
GITLAB_DEFAULT_URL=https://gitlab.com
GITHUB_DEFAULT_URL=https://github.com

# Notification (optional)
DINGTALK_ENABLED=false
WECOM_ENABLED=false
FEISHU_ENABLED=false
```

### Configuration Files

Planned configuration structure:

```
config/
├── config.yaml              # Main configuration (gitignored)
├── config.example.yaml      # Template for users
└── prompt_templates/        # Default LLM prompts
    ├── professional.md
    ├── concise.md
    └── detailed.md
```

## Architecture (Planned)

### High-Level System Flow

```
┌─────────────────────────────────────────────────────────────┐
│  Git Platform (GitLab/GitHub/Gitea)                         │
│  - Push Event / Merge Request Event                         │
└────────────────────┬────────────────────────────────────────┘
                     │ Webhook POST
                     ▼
┌─────────────────────────────────────────────────────────────┐
│  Gin API Server (:8080)                                      │
│  - /api/v1/webhooks/receive                                  │
│  - Validate webhook signature                                │
│  - Enqueue review task to Redis                              │
└────────────────────┬────────────────────────────────────────┘
                     │ Async Task (Asynq)
                     ▼
┌─────────────────────────────────────────────────────────────┐
│  Worker Process (Background)                                 │
│  1. Fetch diff from Git platform API                         │
│  2. Filter files by extension                                │
│  3. Prepare prompt with code changes                         │
│  4. Call LLM API (OpenAI/DeepSeek/etc.)                      │
│  5. Parse structured JSON response                           │
│  6. Post review comments to Git platform                     │
│  7. Send IM notification (optional)                          │
│  8. Store results in database                                │
└────────────────────┬────────────────────────────────────────┘
                     │ Review Results
                     ▼
┌─────────────────────────────────────────────────────────────┐
│  Database (SQLite/MySQL)                                     │
│  - review_results table                                      │
│  - fix_suggestions table                                     │
└─────────────────────────────────────────────────────────────┘
                     │ Query Results
                     ▼
┌─────────────────────────────────────────────────────────────┐
│  React Frontend (:3000)                                      │
│  - Dashboard: Review statistics                              │
│  - Detail Page: Fix suggestions list                         │
│  - Auto-Fix: Trigger fix and watch logs (WebSocket)          │
└─────────────────────────────────────────────────────────────┘
```

### Core Components

1. **API Layer** (`internal/api/`)
   - Gin HTTP handlers
   - JWT middleware
   - Request validation
   - WebSocket endpoints for real-time logs

2. **Service Layer** (`internal/service/`)
   - Business logic
   - Review orchestration
   - Fix execution workflow
   - Notification dispatch

3. **Repository Layer** (`internal/repository/`)
   - GORM database operations
   - Transaction management
   - Query builders

4. **LLM Abstraction** (`internal/llm/`)
   - Factory pattern for provider selection
   - Unified interface for different LLM APIs
   - Prompt template management

5. **Git Operations** (`internal/gitops/`)
   - Repository cloning
   - Branch creation/deletion
   - Commit and push

6. **Async Tasks** (`internal/task/`)
   - Asynq task definitions
   - Worker handlers
   - Retry logic

## Database Schema

**15 Tables Covering All Features:**

### Core Tables
- `users` - User accounts with roles
- `git_platform_configs` - GitLab/GitHub/Gitea instances
- `repositories` - Imported code repositories
- `webhooks` - Webhook configurations

### Repository Organization
- `repository_groups` - Repository grouping
- `group_repositories` - Many-to-many relation

### LLM Configuration
- `llm_providers` - LLM vendor configs (OpenAI, DeepSeek, etc.)
- `llm_models` - Available models per provider

### Notifications
- `notification_channels` - DingTalk/WeCom/Feishu configs

### Templates
- `prompt_templates` - Custom prompts per repository group

### Review Results
- `review_results` - Review history and metadata
- `fix_suggestions` - Structured fix recommendations

### Auto-Fix
- `auto_fix_tasks` - Fix execution tasks
- `auto_fix_logs` - Real-time execution logs
- `fix_branch_management` - Fix branch tracking

**See [docs/03-database-design.md](docs/03-database-design.md) for complete schema with SQL.**

## License

This project is licensed under the **Apache License 2.0** - see the [LICENSE](LICENSE) file for details.

## Project History

- **Original Version:** Python-based system with Flask API and Streamlit UI
- **v2.0 (Current):** Complete redesign to Go + React architecture
  - Old code removed (2025-01-30)
  - Design phase completed (2025-01-30)
  - Implementation pending

## Contact & Resources

- **Design Documentation:** [docs/README.md](docs/README.md)
- **Original Repository:** https://github.com/sunmh207/AI-Codereview-Gitlab
- **Issue Tracker:** [GitHub Issues](https://github.com/your-org/ai-codereview-gitlab/issues)

---

**Current Version:** v2.0-design  
**Last Updated:** 2025-01-30  
**Status:** 🎨 Design Phase Complete → 🚧 Implementation Pending
