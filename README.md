# Binance 大额订单监控系统 V2.1

前后端分离的实时大单监控工具，后端使用 Go 订阅 Binance 深度数据、执行阈值判定和聚合；前端使用 Vue 3 + Vue Router 实现多币种独立监控页面。

## ✨ 主要特性（V2.1 新增）
- 🎯 **多币种页面隔离**：BTC、ETH、SOL、WLD、DOGE、FIL、BNB 独立监控
- 🔄 **数据一致性保障**：每 60 秒自动校验内存与数据库数据
- 🚀 **智能路由系统**：Vue Router 4 实现快速币种切换
- 📊 **固定 U 本位合约**：简化配置，专注 USDT 永续合约
- 💾 **历史数据缓存**：重连时自动继承历史，避免数据丢失
- 🧹 **自动数据清理**：超过 4 小时的旧订单自动清理

## 原有特性
- 支持 USDT 合约深度订阅（固定为 U 本位）
- 阈值条件（≥ / ≤）触发的大单跟踪和成交识别
- 多时间窗口聚合（15m/30m/1h/4h）实时推送
- MySQL 持久化大单成交记录，可选 Redis 缓存
- Vue 3 + Pinia + Vue Router 实现交互式仪表盘

---

## 🚀 快速开始

### 前置依赖
- Go 1.21+
- Node.js 18+
- MySQL 8（默认连接：`47.128.154.233:32066 / order_data`）
- Redis（可选，用于缓存，默认关闭）

> 如果位于内网或需要镜像源，建议在首次执行前设置 `go env -w GOPROXY=https://goproxy.cn,direct`。

### 一键启动（推荐）
```bash
cd /Users/wang/PythonProjects/coin-quant/order-monitor

# 自动检查依赖、编译并启动前后端
./start.sh

# 停止服务
./stop.sh
```

### 手动安装依赖
```bash
cd /Users/wang/PythonProjects/coin-quant/order-monitor

# 后端依赖
cd backend
go mod tidy

# 前端依赖（包含 vue-router）
cd ../frontend
npm install
```

### 配置环境变量
默认配置已写入 `internal/config/config.go`，如需覆盖请设置环境变量或创建 `.env` 文件（systemd 可使用 `EnvironmentFile`）。

示例：
```bash
export SERVER_PORT=8080
export MYSQL_HOST=47.128.154.233
export MYSQL_PORT=32066
export MYSQL_USER=root
export MYSQL_PASSWORD=hadamysqlroot@@pass
export MYSQL_DATABASE=order_data
export MYSQL_PARAMS=parseTime=true
# 可选：关闭 Redis 缓存
export REDIS_ENABLED=false

# 本地需要代理访问 Binance 时可启用
export HTTP_PROXY=http://127.0.0.1:7890
export HTTPS_PROXY=http://127.0.0.1:7890
```

前端本地开发可在 `frontend/.env` 中设置：
```env
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

> 如果只希望前端通过自建网关代理 Binance，可以在页面配置面板中的 “自定义网关” / “网关代理” 输入框填写，例如 `ws://127.0.0.1:8765/ws` 与 `http://127.0.0.1:7890`。

### 启动与验证

#### 方式 1：使用启动脚本（推荐）
```bash
./start.sh
```
脚本会自动启动前后端，并显示访问地址。

#### 方式 2：手动启动
```bash
# 后端
cd backend
go run cmd/server/main.go

# 前端（新开终端）
cd ../frontend
npm run dev
```

访问 `http://localhost:5173`，系统会自动重定向到 `/btc` 页面。

### 支持的币种路由

- `http://localhost:5173/btc` - BTC 监控（阈值: 3）
- `http://localhost:5173/eth` - ETH 监控（阈值: 300）
- `http://localhost:5173/sol` - SOL 监控（阈值: 5000）
- `http://localhost:5173/wld` - WLD 监控（阈值: 10000）
- `http://localhost:5173/doge` - DOGE 监控（阈值: 500000）
- `http://localhost:5173/fil` - FIL 监控（阈值: 3000）
- `http://localhost:5173/bnb` - BNB 监控（阈值: 50）

点击顶部导航栏可快速切换币种。终端中后端日志会输出 Binance 连接、阈值命中、数据一致性检查与 MySQL 入库情况。

---

## 🔌 WebSocket 消息格式
前端通过 `ws://<host>:<port>/ws` 建立连接，后端会推送以下事件：

| type | payload | 说明 |
| ---- | ------- | ---- |
| `connection_status` | `{status,message,symbol,marketType,threshold,thresholdOp}` | 连接状态（connecting/connected/error） |
| `heartbeat` | `{timestamp, session}` | 每 20 秒发送一次心跳 |
| `stats_update` | `{buyWallQty,sellWallQty,buyWallCount,sellWallCount,buyValue,sellValue}` | 当前大单墙汇总 |
| `aggregate_update` | `{windows:[{window,buyQty,sellQty,buyCnt,sellCnt}]}` | 15m/30m/1h/4h 聚合数据（V2.1 更新） |
| `order_filled` | `{orders:[{side,price,quantity,firstSeen,filledTime,durationSec}]}` | 新增的大单成交记录 |

前端已实现对上述消息的自动解析与状态更新。

---

## 🧪 测试建议
- **后端单元测试**：针对 `tracker` 与 `aggregator` 编写测试，模拟深度消息验证阈值命中与聚合结果。
- **集成测试**：在测试环境配置 Binance 测试网或录制的深度数据，通过 `websocket.SessionRequest` 注入，验证数据库写入与消息推送。
- **前端联调**：使用浏览器 DevTools 观察 WebSocket 流、检查 Pinia 状态是否与服务器推送一致。必要时可在 `useWebSocket` 中开启 `console.debug`。
- **性能监控**：部署后建议使用 `pprof` 或自定义指标（Prometheus）跟踪 session 数量、消息频率与 MySQL 写入情况。
- **多页面测试**：同时打开多个币种页面，验证数据隔离和一致性。

---

## 📚 V2.1 相关文档

- **[MULTI_COIN_DEPLOYMENT.md](MULTI_COIN_DEPLOYMENT.md)** - 多币种页面详细部署指南
- **[CHANGELOG_V2.1.md](CHANGELOG_V2.1.md)** - V2.1 完整更新日志
- **[BUGFIX_MULTI_WINDOW.md](BUGFIX_MULTI_WINDOW.md)** - 多窗口问题修复文档
- **[BUGFIX_PANIC_ON_RELOAD.md](BUGFIX_PANIC_ON_RELOAD.md)** - 数据一致性检查 Panic 修复（V2.1.1）
- **[FIXES_SUMMARY.md](FIXES_SUMMARY.md)** - 所有修复总结
- **[PROJECT_ARCHITECTURE.md](PROJECT_ARCHITECTURE.md)** - 项目架构说明

---

## 🎉 V2.1 主要改进

### 解决的问题

1. ✅ **买单量异常**：数据会从 1034 减到 688.94
   - 原因：内存数据未按时间清理
   - 解决：自动清理超过 4 小时的旧订单

2. ✅ **多窗口数据混乱**：打开多个标签页后数据乱闪乱跳
   - 原因：不同配置共享历史数据
   - 解决：多页面隔离 + sessionRecord 缓存

3. ✅ **重连后数据归零**：WebSocket 重连后统计数据归零
   - 原因：重连时未加载历史数据
   - 解决：sessionRecord 持久化 + 自动继承

### 新增功能

- 🎯 7 个币种独立监控页面
- 🔄 每 60 秒数据一致性校验
- 💾 历史数据缓存机制
- 🧹 自动数据清理
- 🚀 Vue Router 快速切换

---
