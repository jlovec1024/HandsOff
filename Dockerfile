# 多阶段构建 Dockerfile for HandsOff

# ============================================
# Stage 1: 构建阶段
# ============================================
FROM golang:1.22-alpine AS builder

# 安装必要的构建工具
RUN apk add --no-cache git make gcc musl-dev

WORKDIR /build

# 复制 go mod 文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建 API 服务器
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o bin/handsoff-api ./cmd/api

# 构建 Worker
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o bin/handsoff-worker ./cmd/worker

# ============================================
# Stage 2: API 运行阶段
# ============================================
FROM alpine:latest AS api

RUN apk --no-cache add ca-certificates tzdata git wget

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/bin/handsoff-api /app/handsoff-api

# 创建必要的目录
RUN mkdir -p /app/data /app/logs /app/temp/git

# 设置时区
ENV TZ=Asia/Shanghai

EXPOSE 8080

CMD ["/app/handsoff-api"]

# ============================================
# Stage 3: Worker 运行阶段
# ============================================
FROM alpine:latest AS worker

RUN apk --no-cache add ca-certificates tzdata git

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/bin/handsoff-worker /app/handsoff-worker

# 创建必要的目录
RUN mkdir -p /app/data /app/logs /app/temp/git

# 设置时区
ENV TZ=Asia/Shanghai

CMD ["/app/handsoff-worker"]

# ============================================
# Stage 4: 开发阶段（带热重载）
# ============================================
FROM golang:1.22-alpine AS dev

RUN apk add --no-cache git make gcc musl-dev

# 安装 Air for 热重载（指定兼容版本）
RUN go install github.com/air-verse/air@v1.52.3

WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 创建必要的目录
RUN mkdir -p /app/data /app/logs /app/temp/git

EXPOSE 8080

# 使用 Air 进行热重载
CMD ["air", "-c", ".air.toml"]
