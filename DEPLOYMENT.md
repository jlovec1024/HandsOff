# HandsOff 部署指南

## 🎯 概述

HandsOff 使用单镜像Docker部署方案，前端和后端打包在一个镜像中，通过docker-compose一键启动。

**架构特点**：
- ✅ 单镜像部署（前端embed到Go二进制）
- ✅ 无需Nginx（Go直接服务静态文件）
- ✅ docker-compose一键启动
- ✅ GitHub Actions自动构建镜像
- ✅ 零配置CORS（前端和后端同域名）

---

## 📦 快速开始

### 前置要求

- Docker 20.10+
- Docker Compose V2

### 1. 准备配置文件

```bash
# 复制环境变量示例
cp .env.example .env

# 编辑配置（必须修改以下两项）
vi .env
```

**关键配置项**：
```bash
# 必须修改为强随机字符串
JWT_SECRET=your_random_secret_here

# 必须修改为Base64编码的32字节密钥
# 生成方法: openssl rand -base64 32
ENCRYPTION_KEY=your_base64_encoded_32_bytes_key
```

### 2. 启动服务

```bash
# 启动所有服务
docker compose up -d

# 查看日志
docker compose logs -f app

# 查看服务状态
docker compose ps
```

### 3. 访问应用

- **Web界面**: http://localhost:8080
- **API端点**: http://localhost:8080/api/health
- **默认账号**: admin / admin123

---

## 🔄 GitHub Actions 自动部署

### 工作流程

1. **打标签触发构建**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **GitHub Actions自动执行**
   - 构建前端（Node.js）
   - 构建后端（Go + embed前端）
   - 推送镜像到 `ghcr.io/jlovec1024/handsoff:v1.0.0`

3. **首次部署额外步骤**
   - 访问 https://github.com/jlovec1024?tab=packages
   - 找到 `handsoff` 镜像包
   - Package settings → Change visibility → **Public**

### 使用指定版本

```yaml
# docker-compose.yml
services:
  app:
    image: ghcr.io/jlovec1024/handsoff:v1.0.0  # 锁定版本
```

### 镜像标签策略

- `latest` - 最新版本
- `vX.Y.Z` - 完整语义化版本（如 v1.0.0）
- `vX.Y` - 次版本（如 v1.0）
- `vX` - 主版本（如 v1）

**⚠️ 生产环境建议使用完整版本号，避免使用 `latest`**

---

## 🛠️ 本地开发构建

### 构建镜像

```bash
# 构建生产镜像
docker build -t handsoff:local --target server .

# 查看镜像大小
docker images handsoff:local
```

### 镜像结构

```
多阶段构建流程：
1. frontend (Node.js 20) → 构建前端 → internal/web/dist/
2. builder (Go 1.22)     → 构建后端 → embed前端 → 单个二进制
3. server (Alpine)       → 运行时环境 → 只包含二进制文件
```

### 验证embed

```bash
# 进入容器检查
docker run --rm -it handsoff:local ls -lh /app/

# 应该只看到：
# handsoff-server (二进制文件，包含前端静态文件)
```

---

## 🔧 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 | 是否必须 |
|--------|------|--------|---------|
| `DB_TYPE` | 数据库类型 | sqlite | 否 |
| `DB_DSN` | 数据库连接 | data/app.db | 否 |
| `REDIS_URL` | Redis地址 | redis://localhost:6379/0 | 是 |
| `API_PORT` | API端口 | 8080 | 否 |
| `JWT_SECRET` | JWT密钥 | - | **是** |
| `ENCRYPTION_KEY` | 加密密钥 | - | **是** |
| `ADMIN_INITIAL_PASSWORD` | 管理员初始密码 | admin123 | 否 |

### docker-compose配置

```yaml
services:
  app:
    image: ghcr.io/jlovec1024/handsoff:v0.1.0
    ports:
      - "8080:8080"
    env_file:
      - .env
    environment:
      REDIS_URL: redis://:handsoff_redis_pwd@redis:6379/0
    depends_on:
      redis:
        condition: service_healthy

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes --requirepass handsoff_redis_pwd
    # 无外部端口暴露（仅容器间通信）
```

---

## 📊 健康检查

### 应用健康检查

```bash
curl http://localhost:8080/api/health

# 响应示例：
{
  "status": "ok",
  "time": "2025-12-13T07:56:53Z",
  "database": "connected",
  "version": "1.0.0-mvp"
}
```

### Docker健康检查

docker-compose自动配置健康检查：
- **间隔**: 30秒
- **超时**: 10秒
- **重试**: 3次
- **启动期**: 40秒

```bash
# 查看健康状态
docker compose ps

# 应该看到：
# app    Up (healthy)
# redis  Up (healthy)
```

---

## 🐛 故障排查

### 容器启动失败

**现象**: 容器不断重启
```bash
docker compose logs app
```

**常见原因**：
1. `.env`文件中 `JWT_SECRET` 或 `ENCRYPTION_KEY` 使用默认值
2. Redis连接失败
3. 数据库路径权限问题

**解决方案**：
```bash
# 检查配置
cat .env | grep -E "(JWT_SECRET|ENCRYPTION_KEY)"

# 重新生成密钥
openssl rand -hex 32  # JWT_SECRET
openssl rand -base64 32  # ENCRYPTION_KEY

# 重启服务
docker compose down
docker compose up -d
```

### 前端404错误

**现象**: 访问 http://localhost:8080 返回404

**检查步骤**：
```bash
# 1. 确认容器正在运行
docker compose ps

# 2. 查看应用日志
docker compose logs app

# 3. 测试API是否正常
curl http://localhost:8080/api/health

# 4. 进入容器检查文件
docker exec -it handsoff-app ls -la /app/
# 应该只有 handsoff-server 一个文件
```

### Redis连接失败

**现象**: 日志显示 "Failed to connect to Redis"

**解决方案**：
```bash
# 检查Redis健康状态
docker compose ps redis

# 如果不健康，查看Redis日志
docker compose logs redis

# 重启Redis
docker compose restart redis
```

### 权限问题

**现象**: "Permission denied" 错误

**解决方案**：
```bash
# 修复数据目录权限
sudo chown -R 1000:1000 data/ logs/ temp/

# 或在docker-compose中指定用户
services:
  app:
    user: "${UID}:${GID}"
```

---

## 📈 性能优化

### 镜像大小优化

当前镜像大小：~50MB（Alpine + Go二进制）

**优化建议**：
- ✅ 已使用Alpine Linux（最小化基础镜像）
- ✅ 已使用多阶段构建（分离构建和运行时）
- ✅ 已使用embed（无需额外文件）
- ⚠️ 可选：启用UPX压缩Go二进制（权衡：启动速度 vs 镜像大小）

### 启动速度优化

**当前启动时间**：~5秒（含健康检查等待）

**优化建议**：
- 减少健康检查 `start_period`（当前40秒，可降到20秒）
- 使用内存数据库（SQLite in-memory mode）用于测试环境

---

## 🔐 安全最佳实践

### 生产环境检查清单

- [ ] **强密钥**: 修改 `JWT_SECRET` 和 `ENCRYPTION_KEY` 为强随机值
- [ ] **修改默认密码**: 首次登录后立即修改 admin 密码
- [ ] **Redis密码**: 修改 docker-compose 中的 Redis 密码
- [ ] **防火墙**: 仅暴露必要端口（8080）
- [ ] **HTTPS**: 在前端代理（如Nginx）配置SSL证书
- [ ] **日志审计**: 定期检查 `logs/app.log`
- [ ] **数据备份**: 定期备份 `data/` 目录和Redis数据

### 推荐的生产环境架构

```
Internet
    ↓
  Nginx (SSL终止 + 反向代理)
    ↓
  HandsOff容器 (端口不暴露到外网)
    ↓
  Redis (内部网络，无外部端口)
```

---

## 📝 维护命令

### 查看日志
```bash
docker compose logs -f app        # 实时日志
docker compose logs --tail=100 app # 最后100行
```

### 重启服务
```bash
docker compose restart app        # 仅重启应用
docker compose restart            # 重启所有服务
```

### 停止服务
```bash
docker compose stop               # 停止（保留数据）
docker compose down               # 停止并删除容器
docker compose down -v            # 停止、删除容器和卷（⚠️ 数据会丢失）
```

### 更新镜像
```bash
docker compose pull               # 拉取最新镜像
docker compose up -d              # 重新创建容器
```

### 备份数据
```bash
# 备份SQLite数据库
docker compose exec app cp /app/data/app.db /app/data/app.db.backup

# 导出到宿主机
docker cp handsoff-app:/app/data/app.db ./backup-$(date +%Y%m%d).db

# 备份Redis数据
docker compose exec redis redis-cli -a handsoff_redis_pwd SAVE
docker cp handsoff-redis:/data/dump.rdb ./redis-backup-$(date +%Y%m%d).rdb
```

---

## 🆘 支持

- **GitHub Issues**: https://github.com/jlovec1024/HandsOff/issues
- **文档**: 查看项目 README.md

---

## 📄 License

Apache License 2.0
