# 🚀 HandsOff 快速启动指南

本指南帮助您在5分钟内启动HandsOff AI代码审查系统。

---

## 📋 前置要求

- **Go** 1.21+ ([下载](https://go.dev/dl/))
- **Node.js** 18+ 和 npm ([下载](https://nodejs.org/))
- **Redis** (可选，Week 4+需要) ([下载](https://redis.io/download))

---

## ⚡ 快速启动（3步）

### 1️⃣ 启动后端API

```bash
# 克隆项目（如果还没有）
cd /path/to/handsoff

# 安装Go依赖
go mod download

# 初始化数据库（创建表+管理员用户）
go run scripts/init_db.go

# 启动API服务器
go run cmd/api/main.go
```

**输出：**
```
Starting HandsOff API server...
API server listening on port 8080
```

✅ 后端API运行在 `http://localhost:8080`

---

### 2️⃣ 启动前端（新终端）

```bash
cd web

# 安装依赖（首次）
npm install

# 启动开发服务器
npm run dev
```

**输出：**
```
  VITE v7.2.4  ready in 500 ms

  ➜  Local:   http://localhost:5173/
  ➜  Network: use --host to expose
```

✅ 前端运行在 `http://localhost:5173`

---

### 3️⃣ 登录系统

1. 打开浏览器访问 `http://localhost:5173`
2. 使用默认账号登录：
   - **用户名**: `admin`
   - **密码**: `admin123`

✅ 成功登录后进入Dashboard！

---

## 🔧 使用Makefile（推荐）

项目提供了Makefile简化操作：

```bash
# 查看所有命令
make help

# 完整开发环境设置（首次）
make dev

# 后续启动
make run-api    # 启动API服务器
```

---

## 📁 项目结构速览

```
handsoff/
├── cmd/                 # 应用入口
│   ├── api/            # API服务器
│   └── worker/         # 异步Worker（Week 4+）
├── internal/           # 后端代码
│   ├── api/handler/    # HTTP处理器
│   ├── model/          # 数据库模型（7张表）
│   └── ...
├── pkg/                # 公共工具包
├── web/                # 前端React项目
│   ├── src/
│   │   ├── pages/      # 页面组件
│   │   ├── api/        # API客户端
│   │   └── ...
├── scripts/            # 初始化脚本
├── .env                # 环境变量配置
└── Makefile           # 构建脚本
```

---

## 🧪 验证安装

### 测试后端API

```bash
# 健康检查
curl http://localhost:8080/api/health

# 登录测试
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### 测试前端

访问 `http://localhost:5173`，应该看到登录页面。

---

## 📚 下一步

Week 1（当前）完成后，按照以下顺序实现功能：

1. **Week 2**: 配置管理（GitLab + LLM）
2. **Week 3**: 仓库管理（导入仓库 + Webhook）
3. **Week 4-5**: Review核心功能（自动代码审查）
4. **Week 6**: 测试与部署

详见：[WEEK1_COMPLETED.md](./WEEK1_COMPLETED.md)

---

## 🆘 常见问题

### Q1: 数据库文件在哪里？
**A**: 默认SQLite数据库位于 `data/app.db`

### Q2: 如何重置数据库？
**A**: 
```bash
rm data/app.db
go run scripts/init_db.go
```

### Q3: 端口被占用怎么办？
**A**: 修改 `.env` 文件中的 `API_PORT=8080` 为其他端口

### Q4: 前端无法连接后端？
**A**: 检查 `web/.env` 中的 `VITE_API_BASE_URL` 是否正确

### Q5: 如何修改管理员密码？
**A**: 目前需要直接修改数据库，后续版本会添加密码修改功能

---

## 📖 完整文档

- [Week 1 完成总结](./WEEK1_COMPLETED.md) - 详细的实现说明
- [项目概览](./SNOW.md) - 技术栈和架构
- [设计文档](./docs/) - 完整的MVP设计

---

## 🐛 遇到问题？

1. 检查Go和Node.js版本
2. 确保所有依赖安装成功
3. 查看日志文件 `logs/api.log`
4. 确认Redis启动（Week 4+需要）

---

**🎉 祝您使用愉快！**

如有问题，请查看详细文档或提交Issue。
