#!/bin/bash

# 停止多币种订单监控系统
# Version: 2.1

set -euo pipefail

echo "🛑 停止 Binance 大额订单监控系统..."

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

BACKEND_PID_FILE=".backend.pid"
FRONTEND_PID_FILE=".frontend.pid"
BACKEND_LOG=".backend.log"
FRONTEND_LOG=".frontend.log"
PROJECT_VITE_PATH="$ROOT_DIR/frontend/node_modules/.bin/vite"

wait_for_exit() {
    local pid="$1"
    local attempts=10
    local attempt=1

    while [ "$attempt" -le "$attempts" ]; do
        if ! kill -0 "$pid" 2>/dev/null; then
            return 0
        fi
        sleep 0.3
        attempt=$((attempt + 1))
    done

    return 1
}

terminate_pid() {
    local pid="$1"
    local label="$2"

    if ! kill -0 "$pid" 2>/dev/null; then
        return 0
    fi

    kill "$pid" 2>/dev/null || true
    if wait_for_exit "$pid"; then
        return 0
    fi

    echo -e "${YELLOW}⚠️  ${label}未正常退出，正在强制结束 (PID: $pid)${NC}"
    kill -9 "$pid" 2>/dev/null || true
    wait_for_exit "$pid"
}

stop_from_pid_file() {
    local label="$1"
    local pid_file="$2"
    local pid=""

    if [ ! -f "$pid_file" ]; then
        echo -e "${YELLOW}⚠️  未找到 ${label} PID 文件${NC}"
        return 0
    fi

    pid=$(cat "$pid_file")
    if [ -z "$pid" ]; then
        echo -e "${YELLOW}⚠️  ${label} PID 文件为空${NC}"
        rm -f "$pid_file"
        return 0
    fi

    if ! kill -0 "$pid" 2>/dev/null; then
        echo -e "${YELLOW}⚠️  ${label}进程不存在 (PID: $pid)${NC}"
        rm -f "$pid_file"
        return 0
    fi

    terminate_pid "$pid" "$label"

    if kill -0 "$pid" 2>/dev/null; then
        echo -e "${YELLOW}⚠️  ${label}仍在退出中 (PID: $pid)${NC}"
    else
        echo -e "${GREEN}✅ ${label}已停止 (PID: $pid)${NC}"
    fi

    rm -f "$pid_file"
}

kill_backend_leftovers() {
    local pids=""
    pids=$(lsof -tiTCP:8081 -sTCP:LISTEN 2>/dev/null || true)

    if [ -n "$pids" ]; then
        echo "🧹 停止占用 8081 的残留进程..."
        for pid in $pids; do
            terminate_pid "$pid" "8081 残留进程"
        done
    fi
}

kill_frontend_leftovers() {
    local pids=""
    pids=$(pgrep -f "$PROJECT_VITE_PATH" 2>/dev/null || true)

    if [ -n "$pids" ]; then
        echo "🧹 停止当前项目的前端残留进程..."
        for pid in $pids; do
            terminate_pid "$pid" "前端残留进程"
        done
    fi
}

report_leftovers() {
    local has_leftover=0
    local backend_8081=""
    local frontend_proc=""

    backend_8081=$(lsof -nP -iTCP:8081 -sTCP:LISTEN 2>/dev/null || true)
    frontend_proc=$(pgrep -fl "$PROJECT_VITE_PATH" 2>/dev/null || true)

    if [ -n "$backend_8081" ]; then
        has_leftover=1
        echo -e "${YELLOW}⚠️  8081 端口仍被占用：${NC}"
        echo "$backend_8081"
    fi

    if [ -n "$frontend_proc" ]; then
        has_leftover=1
        echo -e "${YELLOW}⚠️  检测到当前项目的前端残留进程：${NC}"
        echo "$frontend_proc"
    fi

    if [ "$has_leftover" -eq 1 ]; then
        echo ""
        echo -e "${YELLOW}提示：这些进程已经超出 PID 文件管理范围，建议再执行一次 ./stop.sh，或按上面的 PID 手动 kill。${NC}"
    fi
}

stop_from_pid_file "后端" "$BACKEND_PID_FILE"
stop_from_pid_file "前端" "$FRONTEND_PID_FILE"
kill_backend_leftovers
kill_frontend_leftovers

echo "🧹 清理日志文件..."
rm -f "$BACKEND_LOG" "$FRONTEND_LOG"
echo -e "${GREEN}✅ 已清理 .backend.log / .frontend.log${NC}"

echo "🔍 检查残留进程..."
report_leftovers

echo ""
echo -e "${GREEN}🎉 停止流程已完成${NC}"
