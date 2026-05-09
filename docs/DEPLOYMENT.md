# 部署指南

本文档说明如何将「Binance 大额订单监控系统」发布到生产或准生产环境。

## 1. 准备工作

### 1.1 服务器要求
- 操作系统：Linux (Ubuntu 20.04+/CentOS 8+) 或 macOS
- CPU：2 核以上，内存 ≥ 2GB，磁盘空间 ≥ 10GB
- 网络：可以访问 Binance WebSocket，以及 MySQL/Redis 服务

### 1.2 依赖软件
- Go 1.21+
- Node.js 18+ / pnpm (可选)
- MySQL 8.0+（保存成交数据）
- Redis 6+（可选，加速缓存和共享状态）
- 反向代理（Nginx、Caddy、Traefik 等，用于 HTTPS 和端口代理）

### 1.3 数据库准备
```sql
CREATE DATABASE IF NOT EXISTS order_data
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;
```

#### 建议的数据表
> 可根据实际业务扩展字段，以下示例仅作为占位：

```sql
CREATE TABLE IF NOT EXISTS orders_filled (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  symbol VARCHAR(32) NOT NULL,
  market_type VARCHAR(16) NOT NULL,
  side ENUM('buy','sell') NOT NULL,
  price DECIMAL(20, 8) NOT NULL,
  quantity DECIMAL(20, 8) NOT NULL,
  first_seen TIMESTAMP NOT NULL,
  filled_time TIMESTAMP NOT NULL,
  duration_seconds INT NOT NULL,
  threshold DECIMAL(20, 8) NOT NULL,
  threshold_op ENUM('gt','lt') NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

```sql
CREATE TABLE IF NOT EXISTS window_aggregates (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  symbol VARCHAR(32) NOT NULL,
  market_type VARCHAR(16) NOT NULL,
  window VARCHAR(8) NOT NULL,
  buy_qty DECIMAL(20, 8) NOT NULL,
  sell_qty DECIMAL(20, 8) NOT NULL,
  buy_cnt INT NOT NULL,
  sell_cnt INT NOT NULL,
  generated_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_window(symbol, market_type, window, generated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 2. 配置环境变量
编辑 `/etc/profile` 或 systemd Service 文件，设置以下变量：

```bash
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080
export MYSQL_HOST=47.128.154.233
export MYSQL_PORT=32066
export MYSQL_USER=root
export MYSQL_PASSWORD=hadamysqlroot@@pass
export MYSQL_DATABASE=order_data
export MYSQL_PARAMS='parseTime=true&loc=Asia%2FShanghai'

# 如果服务器需要通过代理访问 Binance，可按需开启
export HTTP_PROXY=http://127.0.0.1:7890
export HTTPS_PROXY=http://127.0.0.1:7890
```

## 4. 前端部署

### 4.1 构建静态资源
```bash
cd /path/to/order-monitor/frontend
npm install
npm run build
```

## 5. 验证
1. 打开浏览器访问 `http://monitor.example.com`。
2. 填写交易对并点击“连接”，确认界面出现心跳、实时统计与成交数据；若暂无大单，可降低阈值进行测试。
3. SSH 登录服务器，查看后端日志是否出现 `session registered`、`order_filled` 插入日志。
4. 检查 MySQL `orders_filled` 表是否有新增记录，必要时使用 `SELECT * FROM orders_filled ORDER BY id DESC LIMIT 20;`。
5. 可使用浏览器 DevTools 的 WebSocket 面板确认推送类型：
   - `connection_status`
   - `heartbeat`
   - `stats_update`
   - `aggregate_update`
   - `order_filled`

## 6. 运行与维护建议
- 使用 `systemctl status order-monitor` 定期检查服务状态；如需滚动升级，先 `stop` 再 `start`。
- 结合 `journalctl -u order-monitor -f` 观察 Binance 连接或数据库异常。
- 建议为 `orders_filled` 增加定期归档任务，避免表无限增长。
- 如需性能监控，可加入 `pprof` 或 Prometheus 指标（后续可扩展到 `/metrics`）。
- 备份 `order_data` 数据库与前端静态资源，确保灾难恢复。
