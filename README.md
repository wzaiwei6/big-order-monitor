# Binance 大额订单监控系统 V2.2

前后端分离的 Binance U 本位合约大额订单监控工具。后端使用 Go 订阅 Binance 深度数据、识别达到阈值的大额订单并写入 SQLite；前端使用 Vue 3 + Pinia + Vue Router 展示单币种监控和多币种汇总页面。

## V2.2 主要变化

- **SQLite 存储**：默认不再依赖 MySQL，适合轻量服务器部署。
- **Docker 部署**：补齐后端、前端 Dockerfile 和 `docker-compose.yml`，服务器上可直接构建运行。
- **数据保留可配置**：默认保留最近 `12` 小时数据，可通过环境变量调整为 `6h/12h` 等。
- **每日 VACUUM**：定时整理 SQLite 文件，减少长期运行后的磁盘膨胀。
- **端口调整**：后端默认监听 `8081`，前端容器通过 Nginx 暴露 `80`。
- **移动端适配**：支持手机访问，导航、监控卡片、汇总页和表格已做紧凑布局。
- **日夜主题**：右上角支持太阳/月亮图标切换白天/黑夜模式。
- **窗口统计修复**：修复汇总页读取 `15m` 数据时误裁剪 `30m/1h/4h` 缓存的问题。

## 功能概览

- 支持 BTC、ETH、SOL、WLD、DOGE、FIL、BNB 七个币种。
- 固定监控 Binance USDT 永续合约盘口。
- 支持大额订单阈值识别、成交记录、实时大单墙统计。
- 支持 `15m / 30m / 1h / 4h` 多时间窗口买卖对比。
- 支持 `/summary` 多币种 15 分钟汇总页面。
- 支持 Docker 单机部署，SQLite 数据挂载到宿主机目录。

## 快速部署：Docker 推荐

服务器需要先安装 Docker 和 Docker Compose。项目复制到服务器后，在项目根目录执行：

```bash
mkdir -p data
sudo docker compose up -d --build
```

访问地址：

```text
http://服务器IP/btc
http://服务器IP/summary
```

查看运行状态：

```bash
sudo docker compose ps
sudo docker compose logs -f backend
sudo docker compose logs -f frontend
```

停止服务：

```bash
sudo docker compose down
```

只重启不重新构建：

```bash
sudo docker compose up -d
```

更新代码后重新构建：

```bash
sudo docker compose up -d --build
```

## Docker 数据目录

SQLite 数据库挂载在宿主机项目目录：

```text
./data/ordermonitor.db
```

`docker-compose.yml` 中的挂载关系：

```yaml
volumes:
  - ./data:/app/data
```

只要不删除宿主机的 `data/` 目录，执行 `sudo docker compose up -d --build` 不会清空历史数据。后端启动时会从 SQLite 恢复最近最大窗口内的数据。

## 环境变量

可在服务器项目根目录创建 `.env` 覆盖默认配置：

```env
MONITOR_DATA_RETENTION_HOURS=12
MONITOR_CLEANUP_INTERVAL_MINUTES=60
SQLITE_VACUUM_ENABLED=true
SQLITE_VACUUM_HOUR=4
SQLITE_VACUUM_MINUTE=0

# 如果服务器访问 Binance 需要代理，再开启
HTTP_PROXY=
HTTPS_PROXY=
```

关键配置说明：

| 变量 | 默认值 | 说明 |
| ---- | ------ | ---- |
| `SERVER_PORT` | `8081` | 后端容器内监听端口 |
| `SQLITE_PATH` | `/app/data/ordermonitor.db` | Docker 内 SQLite 路径 |
| `MONITOR_DATA_RETENTION_HOURS` | `12` | 数据保留小时数 |
| `MONITOR_CLEANUP_INTERVAL_MINUTES` | `60` | 清理旧数据的间隔 |
| `SQLITE_VACUUM_ENABLED` | `true` | 是否每日整理数据库文件 |
| `SQLITE_VACUUM_HOUR` | `4` | 每日 VACUUM 小时 |
| `SQLITE_VACUUM_MINUTE` | `0` | 每日 VACUUM 分钟 |

## 本地开发启动

前置依赖：

- Go 1.21+
- Node.js 18+

一键启动：

```bash
./start.sh
```

停止：

```bash
./stop.sh
```

手动启动：

```bash
cd backend
go run cmd/server/main.go
```

```bash
cd frontend
npm install
npm run dev
```

本地默认访问：

```text
http://localhost:5173/btc
```

如果 Vite 自动切换端口，以终端输出的 `Local:` 地址为准。

## 币种阈值

前端展示阈值配置：

```text
frontend/src/config/coins.ts
```

后端汇总采集阈值配置：

```text
backend/internal/summary/worker.go
```

当前前端展示阈值：

| 币种 | 路由 | 阈值 |
| ---- | ---- | ---- |
| BTC | `/btc` | `3 BTC` |
| ETH | `/eth` | `300 ETH` |
| SOL | `/sol` | `800 SOL` |
| WLD | `/wld` | `10000 WLD` |
| DOGE | `/doge` | `200000 DOGE` |
| FIL | `/fil` | `10000 FIL` |
| BNB | `/bnb` | `50 BNB` |

## 路由

```text
/btc
/eth
/sol
/wld
/doge
/fil
/bnb
/summary
```

## 公网访问与多项目部署

当前 `docker-compose.yml` 将前端映射到服务器 `80` 端口：

```yaml
ports:
  - "80:80"
```

这意味着同一台服务器上只能有一个项目直接占用 `80`。如果后续还有别的项目，建议使用域名和反向代理：

- `monitor.example.com` 指向本项目。
- `app2.example.com` 指向另一个项目。
- 服务器最外层只开放 `80/443`，由 Caddy、Nginx 或 Traefik 按域名转发到不同容器。

临时方案也可以给不同项目使用不同端口，例如 `8082:80`、`8083:80`，但长期访问体验不如域名清晰。

## API 与 WebSocket

前端通过同源 Nginx 代理访问后端：

```text
GET /api/health
GET /api/monitor/:coinId
GET /api/summary/15m
GET /ws
```

WebSocket 主要消息：

| type | payload | 说明 |
| ---- | ------- | ---- |
| `connection_status` | `{status,message,symbol,marketType,threshold,thresholdOp}` | 连接状态 |
| `heartbeat` | `{timestamp, session}` | 心跳 |
| `stats_update` | `{buyWallQty,sellWallQty,buyWallCount,sellWallCount,buyValue,sellValue}` | 当前盘口大单墙统计 |
| `aggregate_update` | `{windows:[{window,buyQty,sellQty,buyCnt,sellCnt}]}` | `15m/30m/1h/4h` 聚合数据 |
| `order_filled` | `{orders:[{side,price,quantity,firstSeen,filledTime,durationSec}]}` | 新增大单成交记录 |

## 验证命令

后端：

```bash
cd backend
go test ./...
go build ./...
```

前端：

```bash
cd frontend
npm run build
```

Docker：

```bash
sudo docker compose up -d --build
sudo docker compose ps
```

## 相关文档

- [MULTI_COIN_DEPLOYMENT.md](MULTI_COIN_DEPLOYMENT.md)
- [CHANGELOG_V2.1.md](CHANGELOG_V2.1.md)
- [BUGFIX_MULTI_WINDOW.md](BUGFIX_MULTI_WINDOW.md)
- [BUGFIX_PANIC_ON_RELOAD.md](BUGFIX_PANIC_ON_RELOAD.md)
- [FIXES_SUMMARY.md](FIXES_SUMMARY.md)
- [PROJECT_ARCHITECTURE.md](PROJECT_ARCHITECTURE.md)
