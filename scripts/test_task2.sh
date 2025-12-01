#!/bin/bash

# Week 4 Task 2: Asynq任务队列测试脚本

echo "================================"
echo "HandsOff Task 2 测试"
echo "================================"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查Redis是否运行
echo "检查Redis状态..."
if redis-cli ping > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Redis运行中${NC}"
else
    echo -e "${RED}✗ Redis未运行${NC}"
    echo "  启动Redis: redis-server"
    exit 1
fi

# 检查编译
echo ""
echo "检查编译状态..."
if [ -f "bin/handsoff-worker" ]; then
    echo -e "${GREEN}✓ Worker已编译${NC}"
else
    echo -e "${YELLOW}→ 编译Worker...${NC}"
    go build -o bin/handsoff-worker ./cmd/worker
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Worker编译成功${NC}"
    else
        echo -e "${RED}✗ Worker编译失败${NC}"
        exit 1
    fi
fi

if [ -f "bin/handsoff-api" ]; then
    echo -e "${GREEN}✓ API已编译${NC}"
else
    echo -e "${YELLOW}→ 编译API...${NC}"
    go build -o bin/handsoff-api ./cmd/api
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ API编译成功${NC}"
    else
        echo -e "${RED}✗ API编译失败${NC}"
        exit 1
    fi
fi

echo ""
echo "================================"
echo "测试环境准备完成"
echo "================================"
echo ""
echo "下一步："
echo "1. Terminal 1: ${GREEN}make run-api${NC}     # 启动API服务器"
echo "2. Terminal 2: ${GREEN}make run-worker${NC}  # 启动Worker"
echo "3. Terminal 3: ${GREEN}./scripts/test_webhook.sh${NC}  # 发送测试Webhook"
echo ""
echo "监控队列："
echo "  ${YELLOW}redis-cli LLEN asynq:queues:default${NC}  # 查看队列长度"
echo "  ${YELLOW}redis-cli LRANGE asynq:queues:default 0 -1${NC}  # 查看队列内容"
echo ""
echo "查看日志："
echo "  ${YELLOW}tail -f logs/api.log${NC}     # API日志"
echo "  ${YELLOW}tail -f logs/worker.log${NC}  # Worker日志（如果配置）"
echo ""
echo "Asynq监控（可选）："
echo "  ${YELLOW}go install github.com/hibiken/asynqmon@latest${NC}"
echo "  ${YELLOW}asynqmon --redis-addr=localhost:6379${NC}"
echo "  访问: http://localhost:8080"
echo ""
