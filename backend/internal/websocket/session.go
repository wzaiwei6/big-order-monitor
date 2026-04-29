package websocket

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"ordermonitor/internal/aggregator"
	"ordermonitor/internal/config"
	"ordermonitor/internal/tracker"
	"ordermonitor/pkg/logger"
)

type session struct {
	key           string
	req           SessionRequest
	log           *zap.Logger
	cfg           config.Config
	conn          *websocket.Conn
	db            *sql.DB
	engine        *tracker.Engine
	outbound      chan wsMessage
	history       []tracker.FilledOrder
	historyMu     sync.RWMutex
	binConn       *websocket.Conn
	lastStatsAt   time.Time
	lastAggAt     time.Time
	historyLoaded bool       // 标志位：是否已加载历史数据
	closed        bool       // 标志位：是否已关闭
	closeMu       sync.Mutex // 保护 closed 标志位
}

type wsMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

type connectionStatusPayload struct {
	Status      string  `json:"status"`
	Message     string  `json:"message"`
	Symbol      string  `json:"symbol"`
	MarketType  string  `json:"marketType"`
	Threshold   float64 `json:"threshold"`
	ThresholdOp string  `json:"thresholdOp"`
}

type heartbeatPayload struct {
	Type      string         `json:"type"`
	Timestamp int64          `json:"timestamp"`
	Session   SessionRequest `json:"session"`
}

type statsPayload struct {
	BuyWallQty    float64 `json:"buyWallQty"`
	SellWallQty   float64 `json:"sellWallQty"`
	BuyWallCount  int     `json:"buyWallCount"`
	SellWallCount int     `json:"sellWallCount"`
	BuyValue      float64 `json:"buyValue"`
	SellValue     float64 `json:"sellValue"`
}

type aggregatePayload struct {
	Windows []aggregator.Snapshot `json:"windows"`
}

type ordersPayload struct {
	Orders []tracker.FilledOrder `json:"orders"`
}

type depthPayload struct {
	Bids [][]string `json:"bids"`
	Asks [][]string `json:"asks"`
	B    [][]string `json:"b"`
	A    [][]string `json:"a"`
}

func newSession(key string, req SessionRequest, mgr *Manager, conn *websocket.Conn) *session {
	engine := tracker.NewEngine(tracker.Config{
		Threshold:   req.Threshold,
		ThresholdOp: tracker.ThresholdOp(req.ThresholdOp),
		MaxTracked:  mgr.cfg.Monitor.MaxTrackedOrders,
	})

	return &session{
		key:      key,
		req:      req,
		log:      mgr.log,
		cfg:      mgr.cfg,
		conn:     conn,
		db:       mgr.db,
		engine:   engine,
		outbound: make(chan wsMessage, 32),
		history:  make([]tracker.FilledOrder, 0, 128),
	}
}

func (s *session) run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go s.writePump(ctx, cancel)
	go s.readPump(ctx, cancel)

	s.outbound <- wsMessage{Type: "connection_status", Payload: connectionStatusPayload{
		Status:      "connecting",
		Message:     "准备连接 Binance",
		Symbol:      s.req.Symbol,
		MarketType:  s.req.MarketType,
		Threshold:   s.req.Threshold,
		ThresholdOp: s.req.ThresholdOp,
	}}

	// 加载或推送历史数据
	if err := s.loadHistory(); err != nil {
		s.log.Warn("load history failed", logger.ErrorField(err))
	}

	// 无论是首次加载还是重连，都推送当前历史和聚合数据
	if s.historyLoaded && len(s.history) > 0 {
		select {
		case s.outbound <- wsMessage{Type: "order_filled", Payload: ordersPayload{Orders: s.history}}:
			s.sendAggregates()
		case <-ctx.Done():
			s.log.Debug("context done before sending history")
			return
		}
	}

	// 启动数据一致性校验 goroutine
	go s.consistencyCheck(ctx)

	reconnectDelay := time.Second

	for {
		select {
		case <-ctx.Done():
			s.log.Info("session context done", logger.String("key", s.key))
			s.closeBinance()
			return
		default:
		}

		if err := s.connectBinance(ctx); err != nil {
			s.outbound <- wsMessage{Type: "connection_status", Payload: connectionStatusPayload{
				Status:      "error",
				Message:     err.Error(),
				Symbol:      s.req.Symbol,
				MarketType:  s.req.MarketType,
				Threshold:   s.req.Threshold,
				ThresholdOp: s.req.ThresholdOp,
			}}
			s.log.Warn("binance connect failed", logger.ErrorField(err))
			select {
			case <-time.After(reconnectDelay):
				reconnectDelay = minDuration(reconnectDelay*2, 30*time.Second)
				continue
			case <-ctx.Done():
				return
			}
		}

		reconnectDelay = time.Second
		s.outbound <- wsMessage{Type: "connection_status", Payload: connectionStatusPayload{
			Status:      "connected",
			Message:     "已连接 Binance",
			Symbol:      s.req.Symbol,
			MarketType:  s.req.MarketType,
			Threshold:   s.req.Threshold,
			ThresholdOp: s.req.ThresholdOp,
		}}

		if err := s.consume(ctx); err != nil {
			s.log.Warn("consume binance stream failed", logger.ErrorField(err))
			s.closeBinance()
			select {
			case <-time.After(reconnectDelay):
				reconnectDelay = minDuration(reconnectDelay*2, 30*time.Second)
				continue
			case <-ctx.Done():
				return
			}
		}
	}
}

func (s *session) connectBinance(ctx context.Context) error {
	urlStr := buildBinanceURL(s.req)
	dialer := websocket.Dialer{
		Proxy: http.ProxyFromEnvironment,
	}

	if s.req.GatewayProxy != "" {
		if proxyURL, err := url.Parse(strings.TrimSpace(s.req.GatewayProxy)); err == nil {
			dialer.Proxy = http.ProxyURL(proxyURL)
		}
	}

	conn, _, err := dialer.DialContext(ctx, urlStr, nil)
	if err != nil {
		return err
	}
	s.binConn = conn
	return nil
}

func (s *session) consume(ctx context.Context) error {
	if s.binConn == nil {
		return errors.New("binance connection nil")
	}

	heartbeatTicker := time.NewTicker(20 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		s.binConn.SetReadDeadline(time.Now().Add(60 * time.Second))
		_, data, err := s.binConn.ReadMessage()
		if err != nil {
			return err
		}

		var payload depthPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			s.log.Debug("unmarshal depth payload failed", logger.ErrorField(err))
			continue
		}

		bids := payload.Bids
		asks := payload.Asks
		if len(bids) == 0 && len(asks) == 0 {
			bids = payload.B
			asks = payload.A
		}

		now := time.Now()
		result, err := s.engine.Process(bids, asks, now.Unix())
		if err != nil {
			s.log.Warn("process depth failed", logger.ErrorField(err))
			continue
		}

		if len(result.Filled) > 0 {
			s.appendHistory(result.Filled)
			s.persistFilled(result.Filled)
			s.outbound <- wsMessage{Type: "order_filled", Payload: ordersPayload{Orders: result.Filled}}
		}

		if len(result.ActiveBids) > 0 || len(result.ActiveAsks) > 0 || now.Sub(s.lastStatsAt) > time.Second {
			s.sendStats(result)
			s.lastStatsAt = now
		}

		if len(result.Filled) > 0 || now.Sub(s.lastAggAt) > 2*time.Second {
			s.sendAggregates()
			s.lastAggAt = now
		}

		select {
		case <-heartbeatTicker.C:
			s.outbound <- wsMessage{Type: "heartbeat", Payload: heartbeatPayload{
				Type:      "heartbeat",
				Timestamp: time.Now().Unix(),
				Session:   s.req,
			}}
		default:
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
}

func (s *session) sendStats(result tracker.ProcessResult) {
	var buyQty, buyValue float64
	for _, o := range result.ActiveBids {
		buyQty += o.Quantity
		buyValue += o.Quantity * o.Price
	}
	var sellQty, sellValue float64
	for _, o := range result.ActiveAsks {
		sellQty += o.Quantity
		sellValue += o.Quantity * o.Price
	}

	payload := statsPayload{
		BuyWallQty:    buyQty,
		SellWallQty:   sellQty,
		BuyWallCount:  len(result.ActiveBids),
		SellWallCount: len(result.ActiveAsks),
		BuyValue:      buyValue,
		SellValue:     sellValue,
	}

	s.outbound <- wsMessage{Type: "stats_update", Payload: payload}
}

func (s *session) sendAggregates() {
	history := s.snapshotHistory()
	snapshots := aggregator.Aggregate(historyToInterface(history), aggregator.DefaultWindows, time.Now())
	s.outbound <- wsMessage{Type: "aggregate_update", Payload: aggregatePayload{Windows: snapshots}}
}

func (s *session) appendHistory(orders []tracker.FilledOrder) {
	s.historyMu.Lock()
	defer s.historyMu.Unlock()

	if len(orders) == 0 {
		return
	}

	// 追加新订单
	s.history = append(orders, s.history...)

	// 清理超过 4 小时的旧订单
	cutoff := time.Now().Add(-4 * time.Hour).Unix()
	validHistory := make([]tracker.FilledOrder, 0, len(s.history))
	for _, order := range s.history {
		if order.FilledTime >= cutoff {
			validHistory = append(validHistory, order)
		}
	}
	s.history = validHistory

	// 限制最大数量
	if len(s.history) > s.cfg.Monitor.MaxTrackedOrders {
		s.history = s.history[:s.cfg.Monitor.MaxTrackedOrders]
	}
}

func (s *session) snapshotHistory() []tracker.FilledOrder {
	s.historyMu.RLock()
	defer s.historyMu.RUnlock()

	out := make([]tracker.FilledOrder, len(s.history))
	copy(out, s.history)
	return out
}

func (s *session) persistFilled(orders []tracker.FilledOrder) {
	if s.db == nil {
		return
	}

	stmt, err := s.db.Prepare(`INSERT OR IGNORE INTO orders_filled (symbol, market_type, side, price, quantity, first_seen, filled_time, duration_seconds, threshold, threshold_op) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		s.log.Warn("prepare insert failed", logger.ErrorField(err))
		return
	}
	defer stmt.Close()

	for _, order := range orders {
		if _, err := stmt.Exec(
			s.req.Symbol,
			s.req.MarketType,
			order.Side,
			order.Price,
			order.Quantity,
			order.FirstSeen,
			order.FilledTime,
			order.DurationSec,
			s.req.Threshold,
			s.req.ThresholdOp,
		); err != nil {
			s.log.Warn("insert order failed", logger.ErrorField(err))
		}
	}
}

func (s *session) loadHistory() error {
	// 防止重复加载
	if s.historyLoaded {
		s.log.Debug("history already loaded, skipping")
		return nil
	}

	if s.db == nil {
		return nil
	}

	cutoff := time.Now().Add(-4 * time.Hour).Unix()
	rows, err := s.db.Query(`SELECT side, price, quantity, first_seen, filled_time, duration_seconds FROM orders_filled WHERE symbol = ? AND market_type = ? AND threshold = ? AND threshold_op = ? AND filled_time >= ? ORDER BY filled_time ASC`, s.req.Symbol, s.req.MarketType, s.req.Threshold, s.req.ThresholdOp, cutoff)
	if err != nil {
		return err
	}
	defer rows.Close()

	loaded := make([]tracker.FilledOrder, 0)
	for rows.Next() {
		var side string
		var price, qty float64
		var firstSeen, filledTime int64
		var duration int64
		if err := rows.Scan(&side, &price, &qty, &firstSeen, &filledTime, &duration); err != nil {
			return err
		}
		loaded = append(loaded, tracker.FilledOrder{
			Side:        tracker.OrderSide(side),
			Price:       price,
			Quantity:    qty,
			FirstSeen:   firstSeen,
			FilledTime:  filledTime,
			DurationSec: duration,
		})
	}

	s.historyMu.Lock()
	s.history = loaded // 直接赋值而不是 append，确保不重复
	s.historyLoaded = true
	s.historyMu.Unlock()

	s.log.Info("history loaded from database", logger.Int("count", len(loaded)))
	return nil
}

func (s *session) consistencyCheck(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.checkAndReloadIfNeeded(); err != nil {
				s.log.Warn("consistency check failed", logger.ErrorField(err))
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *session) checkAndReloadIfNeeded() error {
	if s.db == nil {
		return nil
	}

	// 查询数据库中的订单数量（4小时内）
	var dbCount int
	cutoff := time.Now().Add(-4 * time.Hour).Unix()
	err := s.db.QueryRow(`SELECT COUNT(*) FROM orders_filled WHERE symbol = ? AND market_type = ? AND threshold = ? AND threshold_op = ? AND filled_time >= ?`, s.req.Symbol, s.req.MarketType, s.req.Threshold, s.req.ThresholdOp, cutoff).Scan(&dbCount)
	if err != nil {
		return err
	}

	// 获取内存中的订单数量
	s.historyMu.RLock()
	memCount := len(s.history)
	s.historyMu.RUnlock()

	// 如果数据库记录数超过 MaxTrackedOrders，则只比较 MaxTrackedOrders 范围内的差异
	// 因为内存中最多只保留 MaxTrackedOrders 条记录
	expectedMemCount := dbCount
	if dbCount > s.cfg.Monitor.MaxTrackedOrders {
		expectedMemCount = s.cfg.Monitor.MaxTrackedOrders
	}

	// 计算差异百分比（基于期望的内存数量）
	var diff float64
	if expectedMemCount > 0 {
		diff = float64(expectedMemCount-memCount) / float64(expectedMemCount) * 100
		if diff < 0 {
			diff = -diff
		}
	}

	s.log.Debug("consistency check", logger.Int("db_count", dbCount), logger.Int("mem_count", memCount), logger.Int("expected_mem", expectedMemCount), zap.Float64("diff_pct", diff))

	// 如果差异超过 20% 且内存数量明显不足，重新加载
	// 只有当内存数量小于期望值的 80% 时才重新加载
	// 并且避免频繁重新加载（至少间隔 5 分钟）
	if diff > 20 && memCount < expectedMemCount*8/10 && dbCount > 100 {
		s.log.Warn("data inconsistency detected, reloading", logger.Int("db_count", dbCount), logger.Int("mem_count", memCount), logger.Int("expected_mem", expectedMemCount), zap.Float64("diff_pct", diff))
		return s.reloadHistoryFromDB()
	}

	return nil
}

func (s *session) reloadHistoryFromDB() error {
	if s.db == nil {
		return nil
	}

	cutoff := time.Now().Add(-4 * time.Hour).Unix()
	rows, err := s.db.Query(`SELECT side, price, quantity, first_seen, filled_time, duration_seconds FROM orders_filled WHERE symbol = ? AND market_type = ? AND threshold = ? AND threshold_op = ? AND filled_time >= ? ORDER BY filled_time DESC`, s.req.Symbol, s.req.MarketType, s.req.Threshold, s.req.ThresholdOp, cutoff)
	if err != nil {
		return err
	}
	defer rows.Close()

	loaded := make([]tracker.FilledOrder, 0)
	for rows.Next() {
		var side string
		var price, qty float64
		var firstSeen, filledTime int64
		var duration int64
		if err := rows.Scan(&side, &price, &qty, &firstSeen, &filledTime, &duration); err != nil {
			return err
		}
		loaded = append(loaded, tracker.FilledOrder{
			Side:        tracker.OrderSide(side),
			Price:       price,
			Quantity:    qty,
			FirstSeen:   firstSeen,
			FilledTime:  filledTime,
			DurationSec: duration,
		})

		// 限制加载数量，避免内存过大
		if len(loaded) >= s.cfg.Monitor.MaxTrackedOrders {
			break
		}
	}

	s.historyMu.Lock()
	s.history = loaded
	s.historyMu.Unlock()

	s.log.Info("history reloaded from database", logger.Int("count", len(loaded)))

	// 注意：不在这里推送数据给前端
	// 数据会通过正常的定时聚合推送机制发送
	// 这样可以避免与 sessionRecord 的同步问题

	return nil
}

func (s *session) closeBinance() {
	if s.binConn != nil {
		s.binConn.Close()
		s.binConn = nil
	}
}

func (s *session) writePump(ctx context.Context, cancel context.CancelFunc) {
	for {
		select {
		case msg := <-s.outbound:
			if err := s.conn.WriteJSON(msg); err != nil {
				s.log.Warn("write ws message failed", logger.ErrorField(err))
				cancel()
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *session) readPump(ctx context.Context, cancel context.CancelFunc) {
	s.conn.SetReadLimit(1024)
	s.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	s.conn.SetPongHandler(func(string) error {
		s.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := s.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.log.Warn("websocket read error", logger.ErrorField(err))
			}
			cancel()
			return
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (s *session) close() {
	s.closeMu.Lock()
	defer s.closeMu.Unlock()

	if s.closed {
		return
	}
	s.closed = true

	s.closeBinance()
	close(s.outbound)
	s.conn.Close()
}

func buildBinanceURL(req SessionRequest) string {
	symbol := strings.ToLower(strings.TrimSpace(req.Symbol))
	depth := req.Depth
	if depth == 0 {
		depth = 20
	}

	marketType := strings.ToLower(req.MarketType)
	var base string
	switch marketType {
	case "usdtm":
		base = "wss://fstream.binance.com/ws"
	case "coinm":
		base = "wss://dstream.binance.com/ws"
	default:
		base = "wss://stream.binance.com:9443/ws"
	}

	var stream string
	if depth == 5 || depth == 10 || depth == 20 {
		stream = fmt.Sprintf("%s@depth%d@100ms", symbol, depth)
	} else {
		stream = fmt.Sprintf("%s@depth@100ms", symbol)
	}

	binanceURL := fmt.Sprintf("%s/%s", base, stream)

	if req.Gateway != "" {
		target := url.QueryEscape(binanceURL)
		remote := strings.TrimSuffix(req.Gateway, "?")
		final := fmt.Sprintf("%s?target=%s", remote, target)
		if req.GatewayProxy != "" {
			final = fmt.Sprintf("%s&proxy=%s", final, url.QueryEscape(strings.TrimSpace(req.GatewayProxy)))
		}
		return final
	}

	return binanceURL
}

func historyToInterface(orders []tracker.FilledOrder) []aggregator.FilledOrder {
	res := make([]aggregator.FilledOrder, len(orders))
	for i := range orders {
		res[i] = orders[i]
	}
	return res
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
