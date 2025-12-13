# 本地开发指南

## 前置要求

- Go 1.22+
- Node.js 20+
- Docker & Docker Compose

---

## 首次设置

### 1. 克隆仓库
```bash
git clone https://github.com/jlovec1024/HandsOff.git
cd HandsOff
```

### 2. 安装依赖

**后端依赖**：
```bash
go mod download
```

**前端依赖**：
```bash
cd web
npm install
cd ..
```

### 3. 构建前端（必须）

**⚠️ 重要**：由于前端构建产物不在git中（正确的设计），首次clone后必须构建前端：

```bash
cd web
npm run build
cd ..
```

构建产物会生成到 `internal/web/dist/`，被Go embed使用。

### 4. 配置环境变量

```bash
cp .env.example .env
# 编辑.env，修改JWT_SECRET和ENCRYPTION_KEY
```

---

## 开发模式

### 方式1：完全本地运行（推荐开发调试）

**启动Redis**：
```bash
docker run -d -p 6379:6379 redis:7-alpine redis-server --requirepass handsoff_redis_pwd
```

**启动后端**：
```bash
go run ./cmd/server
```

**启动前端开发服务器**（另一个终端）：
```bash
cd web
npm run dev
```

访问：http://localhost:3000（前端开发服务器，支持热重载）

### 方式2：Docker Compose开发模式

使用热重载的开发容器：

```bash
docker compose -f docker-compose.dev.yml up
```

---

## 构建和测试

### 构建前端
```bash
cd web
npm run build  # 输出到 internal/web/dist/
```

### 构建后端
```bash
go build -o bin/handsoff-server ./cmd/server
```

### 构建Docker镜像
```bash
docker build -t handsoff:local --target server .
```

### 运行测试
```bash
go test ./...
```

---

## 常见问题

### Q: 为什么首次clone需要构建前端？

A: 前端构建产物（`internal/web/dist/`）不在git中，这是正确的设计。构建产物应该在构建时生成，而不是存储在版本控制中。这避免了：
- 2.3MB+ 的无意义diff
- Git仓库膨胀
- Clone和push变慢

### Q: 每次修改前端都要重新构建吗？

A: 取决于你的开发方式：
- **前端开发**：使用 `npm run dev`，不需要重新构建，有热重载
- **后端+前端集成测试**：修改前端后需要 `npm run build`，然后重启Go服务

### Q: go build报错找不到embed文件？

A: 运行 `cd web && npm run build`，确保 `internal/web/dist/` 目录存在。

### Q: Docker构建时会自动构建前端吗？

A: 是的！Dockerfile包含frontend stage，会自动构建前端，所以Docker构建不依赖本地的dist目录。

---

## Git工作流

### 提交代码

```bash
git add .
git commit -m "feat: your feature"
git push origin master
```

**注意**：`internal/web/dist/` 已在 `.gitignore` 中，不会被提交。

### 发布新版本

```bash
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions会自动构建Docker镜像并推送到 `ghcr.io/jlovec1024/handsoff:v1.0.0`

---

## 项目结构

```
HandsOff/
├── cmd/
│   └── server/          # 主程序入口
├── internal/
│   ├── api/             # API路由和处理器
│   ├── model/           # 数据模型
│   ├── repository/      # 数据访问层
│   ├── service/         # 业务逻辑层
│   ├── task/            # 异步任务
│   └── web/
│       ├── embed.go     # 前端静态文件embed
│       └── dist/        # 前端构建产物（本地，不在git）
├── pkg/                 # 可复用的包
├── web/                 # 前端源代码
│   ├── src/
│   ├── public/
│   └── package.json
├── Dockerfile           # 生产环境Docker镜像
├── docker-compose.yml   # 生产部署配置
└── .env.example         # 环境变量示例
```

---

## 更多文档

- **部署指南**：`DEPLOYMENT.md`
- **API文档**：TODO
- **贡献指南**：TODO
