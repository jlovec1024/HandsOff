#!/bin/bash

# ============================================
# HandsOff 最小测试脚本 (不依赖外部工具)
# ============================================

echo "=============================================="
echo "HandsOff 最小测试"
echo "=============================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0

test_step() {
    local name="$1"
    local command="$2"
    
    echo -n "Testing $name... "
    
    if eval "$command" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}"
        ((FAILED++))
        return 1
    fi
}

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1️⃣  检查基础环境"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

test_step "Go installation" "command -v go"
test_step ".env file exists" "test -f .env"

echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2️⃣  编译项目"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

test_step "API build" "go build -o bin/api ./cmd/api"
test_step "Worker build" "go build -o bin/worker ./cmd/worker"

echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3️⃣  检查数据库"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

DB_PATH=$(grep "DB_DSN=" .env 2>/dev/null | cut -d'=' -f2 || echo "data/app.db")
test_step "Database file exists" "test -f $DB_PATH"

echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4️⃣  测试工具编译"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

test_step "encrypt_apikey tool" "go build -o bin/encrypt_apikey ./tools/encrypt_apikey"
test_step "test_components tool" "go build -o bin/test_components ./tools/test_components"

echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 测试总结"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "通过: ${GREEN}$PASSED${NC}"
echo -e "失败: ${RED}$FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ 所有基础测试通过！${NC}"
    echo ""
    echo "下一步:"
    echo "  1. 配置 Redis (可选 - 用于任务队列)"
    echo "  2. 配置 .env 中的 API Keys"
    echo "  3. 运行: ./bin/api"
    echo "  4. 运行: ./bin/worker"
    echo ""
    exit 0
else
    echo -e "${RED}❌ 部分测试失败${NC}"
    echo ""
    exit 1
fi
