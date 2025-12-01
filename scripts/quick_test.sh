#!/bin/bash

# ============================================
# HandsOff 快速测试脚本
# ============================================
# 用途: 快速验证系统基础功能
# 使用: chmod +x scripts/quick_test.sh && ./scripts/quick_test.sh
# ============================================

# 移除 set -e，允许测试继续运行即使有失败

echo "=============================================="
echo "HandsOff 快速测试"
echo "=============================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试结果计数
PASSED=0
FAILED=0

# 测试函数
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

# ============================================
# 1. 检查依赖
# ============================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1️⃣  检查系统依赖"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

test_step "Go installation" "command -v go"
test_step "Redis CLI" "command -v redis-cli"
test_step "SQLite3" "command -v sqlite3"

echo ""

# ============================================
# 2. 检查配置文件
# ============================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2️⃣  检查配置文件"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

test_step ".env file exists" "test -f .env"

if [ -f .env ]; then
    test_step "ENCRYPTION_KEY set" "grep -q 'ENCRYPTION_KEY=' .env && ! grep -q 'ENCRYPTION_KEY=CHANGE' .env"
    test_step "DB_DSN set" "grep -q 'DB_DSN=' .env"
    test_step "REDIS_URL set" "grep -q 'REDIS_URL=' .env"
fi

echo ""

# ============================================
# 3. 检查 Redis
# ============================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3️⃣  检查 Redis 服务"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

test_step "Redis server running" "redis-cli ping | grep -q PONG"

echo ""

# ============================================
# 4. 编译项目
# ============================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4️⃣  编译项目"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

test_step "API build" "go build -o bin/api ./cmd/api"
test_step "Worker build" "go build -o bin/worker ./cmd/worker"

echo ""

# ============================================
# 5. 检查数据库
# ============================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "5️⃣  检查数据库"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 读取数据库路径
DB_PATH=$(grep "DB_DSN=" .env 2>/dev/null | cut -d'=' -f2 || echo "data/app.db")

test_step "Database file exists" "test -f $DB_PATH"

if [ -f "$DB_PATH" ]; then
    test_step "llm_providers table" "sqlite3 $DB_PATH 'SELECT count(*) FROM llm_providers;' > /dev/null"
    test_step "repositories table" "sqlite3 $DB_PATH 'SELECT count(*) FROM repositories;' > /dev/null"
    test_step "git_platform_configs table" "sqlite3 $DB_PATH 'SELECT count(*) FROM git_platform_configs;' > /dev/null"
fi

echo ""

# ============================================
# 6. 运行组件测试 (可选)
# ============================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "6️⃣  组件单元测试 (可选)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ -f "tools/test_components/main.go" ]; then
    echo -e "${YELLOW}运行单元测试...${NC}"
    if go run tools/test_components/main.go; then
        echo -e "${GREEN}✓ 单元测试通过${NC}"
        ((PASSED++))
    else
        echo -e "${RED}✗ 单元测试失败${NC}"
        ((FAILED++))
    fi
else
    echo -e "${YELLOW}⚠ 测试脚本不存在，跳过${NC}"
fi

echo ""

# ============================================
# 7. 总结
# ============================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 测试总结"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "通过: ${GREEN}$PASSED${NC}"
echo -e "失败: ${RED}$FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ 所有测试通过！系统就绪。${NC}"
    echo ""
    echo "下一步:"
    echo "  1. 启动 API Server:  ./bin/api"
    echo "  2. 启动 Worker:      ./bin/worker"
    echo "  3. 配置 GitLab Webhook"
    echo ""
    echo "详细测试步骤请参考: TESTING_GUIDE.md"
    exit 0
else
    echo -e "${RED}❌ 部分测试失败，请检查配置。${NC}"
    echo ""
    echo "常见问题:"
    echo "  - Redis 未运行: brew services start redis"
    echo "  - .env 未配置: cp .env.example .env"
    echo "  - 数据库未初始化: ./bin/api (运行一次自动初始化)"
    echo ""
    echo "详细排查请参考: TESTING_GUIDE.md → 常见问题排查"
    exit 1
fi
