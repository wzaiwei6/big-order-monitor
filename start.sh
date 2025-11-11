#!/bin/bash

# 多币种订单监控系统启动脚本
# Version: 2.1

set -e

echo "🚀 启动 Binance 大额订单监控系统 V2.1"
echo ""

# 检查是否在项目根目录
if [ ! -f "start.sh" ]; then
    echo "❌ 错误：请在项目根目录运行此脚本"
    exit 1
fi

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 检查后端
echo "📦 检查后端..."
if [ ! -f "backend/bin/server" ]; then
    echo -e "${YELLOW}⚠️  后端未编译，正在编译...${NC}"
    cd backend
    go build -o bin/server cmd/server/main.go
    cd ..
    echo -e "${GREEN}✅ 后端编译完成${NC}"
else
    echo -e "${GREEN}✅ 后端已就绪${NC}"
fi

# 检查前端
echo "📦 检查前端..."
if [ ! -d "frontend/node_modules" ]; then
    echo -e "${YELLOW}⚠️  前端依赖未安装，正在安装...${NC}"
    cd frontend
    npm install
    cd ..
    echo -e "${GREEN}✅ 前端依赖安装完成${NC}"
else
    echo -e "${GREEN}✅ 前端依赖已就绪${NC}"
fi

# 检查环境变量
echo "🔧 检查环境配置..."
if [ -z "$MYSQL_HOST" ]; then
    echo -e "${YELLOW}⚠️  未设置 MYSQL_HOST，使用默认配置${NC}"
fi

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

# 启动后端
echo "🔥 启动后端服务..."
cd backend
./bin/server &
BACKEND_PID=$!
cd ..
echo -e "${GREEN}✅ 后端已启动 (PID: $BACKEND_PID)${NC}"
sleep 2

# 启动前端
echo "🔥 启动前端服务..."
cd frontend
npm run dev &
FRONTEND_PID=$!
cd ..
echo -e "${GREEN}✅ 前端已启动 (PID: $FRONTEND_PID)${NC}"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}🎉 系统启动成功！${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📍 访问地址："
echo "   前端: http://localhost:5173"
echo "   后端: http://localhost:8080"
echo ""
echo "🔗 快速访问："
echo "   BTC:  http://localhost:5173/btc"
echo "   ETH:  http://localhost:5173/eth"
echo "   SOL:  http://localhost:5173/sol"
echo "   WLD:  http://localhost:5173/wld"
echo "   DOGE: http://localhost:5173/doge"
echo "   FIL:  http://localhost:5173/fil"
echo "   BNB:  http://localhost:5173/bnb"
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

# 保存 PID 到文件
echo $BACKEND_PID > .backend.pid
echo $FRONTEND_PID > .frontend.pid

# 等待用户中断
trap "echo ''; echo '🛑 正在停止服务...'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; rm -f .backend.pid .frontend.pid; echo '✅ 服务已停止'; exit 0" INT TERM

# 保持脚本运行
wait

