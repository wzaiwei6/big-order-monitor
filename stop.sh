#!/bin/bash

# 停止多币种订单监控系统
# Version: 2.1

echo "🛑 停止 Binance 大额订单监控系统..."

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 从 PID 文件读取
if [ -f ".backend.pid" ]; then
    BACKEND_PID=$(cat .backend.pid)
    if kill -0 $BACKEND_PID 2>/dev/null; then
        kill $BACKEND_PID
        echo -e "${GREEN}✅ 后端已停止 (PID: $BACKEND_PID)${NC}"
    else
        echo -e "${YELLOW}⚠️  后端进程不存在${NC}"
    fi
    rm -f .backend.pid
fi

if [ -f ".frontend.pid" ]; then
    FRONTEND_PID=$(cat .frontend.pid)
    if kill -0 $FRONTEND_PID 2>/dev/null; then
        kill $FRONTEND_PID
        echo -e "${GREEN}✅ 前端已停止 (PID: $FRONTEND_PID)${NC}"
    else
        echo -e "${YELLOW}⚠️  前端进程不存在${NC}"
    fi
    rm -f .frontend.pid
fi

# 清理可能残留的进程
echo "🧹 清理残留进程..."
pkill -f "bin/server" 2>/dev/null && echo -e "${GREEN}✅ 清理后端残留进程${NC}"
pkill -f "vite" 2>/dev/null && echo -e "${GREEN}✅ 清理前端残留进程${NC}"

echo ""
echo -e "${GREEN}🎉 所有服务已停止${NC}"

