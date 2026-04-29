#!/bin/bash

# 多币种订单监控系统启动脚本
# Version: 2.1

set -euo pipefail

echo "🚀 启动 Binance 大额订单监控系统 V2.1"
echo ""

if [ ! -f "start.sh" ]; then
    echo "❌ 错误：请在项目根目录运行此脚本"
    exit 1
fi

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

BACKEND_LOG=".backend.log"
FRONTEND_LOG=".frontend.log"
BACKEND_PID_FILE=".backend.pid"
FRONTEND_PID_FILE=".frontend.pid"
BACKEND_URL="http://localhost:8081"
FRONTEND_URL=""
BACKEND_PID=""
FRONTEND_PID=""

cleanup() {
    if [ -n "${BACKEND_PID:-}" ] && kill -0 "$BACKEND_PID" 2>/dev/null; then
        kill "$BACKEND_PID" 2>/dev/null || true
    fi
    if [ -n "${FRONTEND_PID:-}" ] && kill -0 "$FRONTEND_PID" 2>/dev/null; then
        kill "$FRONTEND_PID" 2>/dev/null || true
    fi
    rm -f "$BACKEND_PID_FILE" "$FRONTEND_PID_FILE"
}

abort() {
    echo -e "${RED}❌ $1${NC}"
    cleanup
    exit 1
}

check_backend_port() {
    local port_info=""

    port_info=$(lsof -nP -iTCP:8081 -sTCP:LISTEN 2>/dev/null || true)
    if [ -n "$port_info" ]; then
        echo -e "${RED}❌ 8081 端口已被占用，无法启动后端${NC}"
        echo "$port_info"
        echo ""
        echo "请先执行 ./stop.sh，或手动停止占用 8081 的进程后再重试。"
        exit 1
    fi
}

wait_for_backend() {
    local attempts=20
    local attempt=1

    while [ "$attempt" -le "$attempts" ]; do
        if ! kill -0 "$BACKEND_PID" 2>/dev/null; then
            echo ""
            echo "后端日志："
            tail -n 20 "$BACKEND_LOG" 2>/dev/null || true
            abort "后端启动失败"
        fi

        if grep -q 'server starting' "$BACKEND_LOG"; then
            return 0
        fi

        sleep 0.5
        attempt=$((attempt + 1))
    done

    echo ""
    echo "后端日志："
    tail -n 20 "$BACKEND_LOG" 2>/dev/null || true
    abort "后端启动超时"
}

wait_for_frontend() {
    local attempts=40
    local attempt=1

    while [ "$attempt" -le "$attempts" ]; do
        if ! kill -0 "$FRONTEND_PID" 2>/dev/null; then
            echo ""
            echo "前端日志："
            tail -n 40 "$FRONTEND_LOG" 2>/dev/null || true
            abort "前端启动失败"
        fi

        FRONTEND_URL=$(sed -n 's/.*Local:[[:space:]]*\(http:\/\/[^[:space:]]*\).*/\1/p' "$FRONTEND_LOG" | tail -n 1)
        if [ -n "$FRONTEND_URL" ]; then
            return 0
        fi

        sleep 0.5
        attempt=$((attempt + 1))
    done

    echo ""
    echo "前端日志："
    tail -n 40 "$FRONTEND_LOG" 2>/dev/null || true
    abort "前端端口解析失败"
}

echo "📦 编译后端..."
mkdir -p backend/bin
(
    cd backend
    go build -o bin/server cmd/server/main.go
)
echo -e "${GREEN}✅ 后端编译完成${NC}"

echo "📦 检查前端..."
if [ ! -d "frontend/node_modules" ]; then
    echo -e "${YELLOW}⚠️  前端依赖未安装，正在安装...${NC}"
    (
        cd frontend
        npm install
    )
    echo -e "${GREEN}✅ 前端依赖安装完成${NC}"
else
    echo -e "${GREEN}✅ 前端依赖已就绪${NC}"
fi

echo "🔧 检查环境配置..."
echo "   SQLite: ${SQLITE_PATH:-./backend/data/ordermonitor.db}"
echo "   数据保留: ${MONITOR_DATA_RETENTION_HOURS:-12} 小时"
echo "   每日 VACUUM: ${SQLITE_VACUUM_ENABLED:-true}"

echo ""
echo "🎯 支持的币种："
echo "  - BTC (阈值: 3)"
echo "  - ETH (阈值: 300)"
echo "  - SOL (阈值: 800)"
echo "  - WLD (阈值: 10000)"
echo "  - DOGE (阈值: 200000)"
echo "  - FIL (阈值: 10000)"
echo "  - BNB (阈值: 50)"
echo ""

rm -f "$BACKEND_LOG" "$FRONTEND_LOG"
check_backend_port

echo "🔥 启动后端服务..."
(
    cd backend
    exec ./bin/server > "../$BACKEND_LOG" 2>&1
) &
BACKEND_PID=$!
echo "$BACKEND_PID" > "$BACKEND_PID_FILE"
wait_for_backend
echo -e "${GREEN}✅ 后端已启动 (PID: $BACKEND_PID)${NC}"

echo "🔥 启动前端服务..."
(
    cd frontend
    exec npm run dev > "../$FRONTEND_LOG" 2>&1
) &
FRONTEND_PID=$!
echo "$FRONTEND_PID" > "$FRONTEND_PID_FILE"
wait_for_frontend
echo -e "${GREEN}✅ 前端已启动 (PID: $FRONTEND_PID)${NC}"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}🎉 系统启动成功！${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📍 访问地址："
echo "   前端: $FRONTEND_URL"
echo "   后端: $BACKEND_URL"
echo ""
echo "🔗 快速访问："
echo "   BTC:  $FRONTEND_URL/btc"
echo "   ETH:  $FRONTEND_URL/eth"
echo "   SOL:  $FRONTEND_URL/sol"
echo "   WLD:  $FRONTEND_URL/wld"
echo "   DOGE: $FRONTEND_URL/doge"
echo "   FIL:  $FRONTEND_URL/fil"
echo "   BNB:  $FRONTEND_URL/bnb"
echo "   汇总: $FRONTEND_URL/summary"
echo ""
echo "📝 日志文件："
echo "   后端: $BACKEND_LOG"
echo "   前端: $FRONTEND_LOG"
echo ""
echo "⚙️  进程 ID："
echo "   后端: $BACKEND_PID"
echo "   前端: $FRONTEND_PID"
echo ""
echo "🛑 停止服务："
echo "   kill $BACKEND_PID $FRONTEND_PID"
echo "   或按 Ctrl+C 然后运行: ./stop.sh"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

trap "echo ''; echo '🛑 正在停止服务...'; cleanup; echo '✅ 服务已停止'; exit 0" INT TERM

wait
