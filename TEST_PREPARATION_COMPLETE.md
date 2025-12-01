# 🎉 测试准备完成！

**创建时间**: 2025-12-01  
**状态**: ✅ 所有测试工具和文档已就绪

---

## ✅ 已完成的工作

### 1. 测试配置文件

- ✅ `.env.example` - 环境变量模板 (已存在)
- ✅ `scripts/test_data.sql` - 测试数据 SQL 脚本 (新建)

### 2. 测试工具 (Go)

- ✅ `tools/test_components/main.go` - 组件单元测试
  - 测试数据库连接
  - 测试 Redis 队列
  - 测试 GitLab API
  - 测试 LLM API

- ✅ `tools/encrypt_apikey/main.go` - API Key 加密/解密工具
  - 加密模式: `go run tools/encrypt_apikey/main.go -key "sk-xxx"`
  - 解密模式: `go run tools/encrypt_apikey/main.go -decrypt "encrypted"`

### 3. 测试脚本 (Shell)

- ✅ `scripts/quick_test.sh` - 快速测试脚本
  - 检查系统依赖
  - 验证配置文件
  - 测试 Redis 连接
  - 编译项目
  - 检查数据库
  - 运行单元测试

### 4. 测试文档

- ✅ `TESTING_GUIDE.md` (17KB) - 完整测试指南
  - 测试前准备
  - 单元测试步骤
  - 集成测试步骤
  - 验证清单
  - 常见问题排查

- ✅ `TESTING_READY.md` (7.7KB) - 快速开始指南
  - 快速开始步骤
  - 测试检查清单
  - 常用命令
  - 文档索引

---

## 🚀 现在可以开始测试！

### 方式 1: 快速测试 (推荐)

```bash
# 运行快速测试脚本
./scripts/quick_test.sh
```

### 方式 2: 手动测试

```bash
# 1. 单元测试
go run tools/test_components/main.go

# 2. 编译项目
go build -o bin/api ./cmd/api
go build -o bin/worker ./cmd/worker

# 3. 启动服务
./bin/api       # Terminal 1
./bin/worker    # Terminal 2

# 4. 配置 GitLab Webhook 并创建测试 MR
```

---

## 📋 测试前检查

在开始测试前，请确保：

### ✅ 环境依赖

- [ ] Go 1.22+ 已安装
- [ ] Redis 已安装并运行
- [ ] SQLite3 或 MySQL 已安装

```bash
# 检查
go version
redis-cli ping
sqlite3 --version
```

### ✅ 配置文件

- [ ] 已复制 `.env.example` 为 `.env`
- [ ] 已生成并配置 `ENCRYPTION_KEY`
- [ ] 已配置 `DB_DSN`
- [ ] 已配置 `REDIS_URL`

```bash
# 生成加密密钥
openssl rand -base64 32

# 编辑 .env
vim .env  # 或使用你喜欢的编辑器
```

### ✅ 外部服务

- [ ] 已获取 GitLab Access Token
- [ ] 已获取 LLM API Key (DeepSeek 或 OpenAI)
- [ ] 已知道测试项目的 GitLab Project ID

### ✅ 数据库准备

- [ ] 已运行一次 API Server (自动初始化数据库)
- [ ] 已加密 LLM API Key
- [ ] 已编辑并执行 `scripts/test_data.sql`

```bash
# 加密 API Key
go run tools/encrypt_apikey/main.go -key "sk-your-api-key"

# 执行测试数据 SQL
sqlite3 data/handsoff.db < scripts/test_data.sql
```

---

## 📖 详细文档索引

| 文档 | 说明 | 何时阅读 |
|------|------|----------|
| **TESTING_READY.md** | 快速开始 | **现在** - 快速了解测试流程 |
| **TESTING_GUIDE.md** | 完整指南 | **测试时** - 详细步骤和排查 |
| `WEEK4_PROGRESS_SUMMARY.md` | 进度总结 | 了解已完成的功能 |
| `WEEK4_TASK4_COMPLETED.md` | Task 4 文档 | 了解 GitLab 集成详情 |

---

## 🎯 测试目标

### 阶段 1: 单元测试 (预计 30 分钟)

**目标**: 验证各组件独立功能

- [ ] 数据库连接成功
- [ ] Redis 队列工作正常
- [ ] GitLab API 可访问
- [ ] LLM API 可调用

**运行**: `./scripts/quick_test.sh` 或 `go run tools/test_components/main.go`

---

### 阶段 2: 集成测试 (预计 1-2 小时)

**目标**: 验证完整的 AI 代码审查流程

1. [ ] Webhook 接收 MR 事件
2. [ ] 任务成功入队到 Redis
3. [ ] Worker 接收并处理任务
4. [ ] 成功获取 GitLab MR Diff
5. [ ] LLM 返回审查结果
6. [ ] 结果保存到数据库
7. [ ] 评论发布到 GitLab MR
8. [ ] 评论格式正确美观

**步骤**: 参考 `TESTING_GUIDE.md` → 集成测试

---

## 🛠️ 常用命令速查

### 环境管理

```bash
# 启动 Redis
brew services start redis           # macOS
sudo systemctl start redis          # Linux
docker run -d -p 6379:6379 redis    # Docker

# 生成加密密钥
openssl rand -base64 32

# 初始化数据库
./bin/api  # 运行一次，看到日志后 Ctrl+C
```

### 测试工具

```bash
# 快速测试
./scripts/quick_test.sh

# 单元测试
go run tools/test_components/main.go

# 加密 API Key
go run tools/encrypt_apikey/main.go -key "sk-xxx"

# 解密验证
go run tools/encrypt_apikey/main.go -decrypt "encrypted-value"
```

### 编译运行

```bash
# 编译
go build -o bin/api ./cmd/api
go build -o bin/worker ./cmd/worker

# 运行
./bin/api       # Terminal 1
./bin/worker    # Terminal 2
```

### 数据库操作

```bash
# 连接 SQLite
sqlite3 data/handsoff.db

# 执行 SQL
sqlite3 data/handsoff.db < scripts/test_data.sql

# 查看数据
sqlite3 data/handsoff.db "SELECT * FROM repositories;"
```

---

## ⚡ 3 分钟快速测试

如果你已经完成所有配置，最快的测试流程：

```bash
# 1. 快速检查 (30秒)
./scripts/quick_test.sh

# 2. 启动服务 (10秒)
./bin/api &
./bin/worker &

# 3. 在 GitLab 中测试 Webhook (1分钟)
# Settings → Webhooks → Test → Merge request events

# 4. 创建测试 MR (1分钟)
# 创建一个简单的 MR

# 5. 检查 GitLab MR 评论 (30秒)
# 应该看到 AI 生成的评论

# 6. 停止服务
pkill api worker
```

---

## ❓ 遇到问题？

### 步骤 1: 运行快速测试

```bash
./scripts/quick_test.sh
```

**根据输出的 ✗ FAIL 项进行修复**

### 步骤 2: 查看详细文档

打开 **`TESTING_GUIDE.md`** → 常见问题排查

包含以下问题的解决方案：
- 数据库连接失败
- Redis 连接失败
- GitLab API 认证失败
- LLM API 调用失败
- Webhook 未触发
- Worker 处理失败
- 评论未发布

### 步骤 3: 查看日志

```bash
# API Server 日志 (直接在 Terminal 查看)
# Worker 日志 (直接在 Terminal 查看)

# Redis 队列状态
redis-cli
> LLEN asynq:queues:code_review
> LRANGE asynq:queues:code_review 0 -1
```

---

## 🎉 测试成功后

恭喜！你已经拥有一个完整可用的 AI 代码审查系统！

### 下一步选择

1. **优化现有功能** - Task 5: Review 结果存储优化
2. **开发前端界面** - Task 6-7: React 界面
3. **生产部署** - Docker + Kubernetes
4. **扩展功能** - 支持 GitHub, Claude 等

---

## 📊 测试工具总览

```
handsoff/
├── scripts/
│   ├── quick_test.sh          # 快速测试 (主要入口)
│   └── test_data.sql          # 测试数据脚本
├── tools/
│   ├── test_components/       # 单元测试工具
│   │   └── main.go
│   └── encrypt_apikey/        # 加密工具
│       └── main.go
└── docs/
    ├── TESTING_GUIDE.md       # 完整测试指南 ⭐
    ├── TESTING_READY.md       # 快速开始指南 ⭐
    └── TEST_PREPARATION_COMPLETE.md  # 本文件
```

---

## 🏆 准备工作总结

✅ **4 个新文件创建**
- scripts/test_data.sql
- tools/test_components/main.go
- tools/encrypt_apikey/main.go
- scripts/quick_test.sh

✅ **2 个详细文档编写**
- TESTING_GUIDE.md (17KB)
- TESTING_READY.md (7.7KB)

✅ **所有测试工具编译通过**

---

**🚀 开始测试吧！运行: `./scripts/quick_test.sh`**

**📖 需要帮助？阅读: `TESTING_GUIDE.md`**
