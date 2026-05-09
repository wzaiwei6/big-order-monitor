package summary

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	orderagg "ordermonitor/internal/aggregator"
	"ordermonitor/internal/config"
	"ordermonitor/internal/tracker"
	"ordermonitor/pkg/logger"
)

type CoinSpec struct {
	ID          string
	Symbol      string
	DisplayName string
	Threshold   float64
	ThresholdOp string
	MarketType  string
	Depth       int
}

var DefaultSpecs = []CoinSpec{
	{ID: "btc", Symbol: "btcusdt", DisplayName: "BTC", Threshold: 3, ThresholdOp: "gt", MarketType: "usdtm", Depth: 20},
	{ID: "eth", Symbol: "ethusdt", DisplayName: "ETH", Threshold: 50, ThresholdOp: "gt", MarketType: "usdtm", Depth: 20},
	{ID: "sol", Symbol: "solusdt", DisplayName: "SOL", Threshold: 1000, ThresholdOp: "gt", MarketType: "usdtm", Depth: 20},
	{ID: "wld", Symbol: "wldusdt", DisplayName: "WLD", Threshold: 100000, ThresholdOp: "gt", MarketType: "usdtm", Depth: 20},
	{ID: "doge", Symbol: "dogeusdt", DisplayName: "DOGE", Threshold: 1000000, ThresholdOp: "gt", MarketType: "usdtm", Depth: 20},
	{ID: "fil", Symbol: "filusdt", DisplayName: "FIL", Threshold: 50000, ThresholdOp: "gt", MarketType: "usdtm", Depth: 20},
	{ID: "bnb", Symbol: "bnbusdt", DisplayName: "BNB", Threshold: 100, ThresholdOp: "gt", MarketType: "usdtm", Depth: 20},
}

type worker struct {
	cfg     config.Config
	log     *zap.Logger
	db      *sql.DB
	spec    CoinSpec
	engine  *tracker.Engine
	window  *windowAggregator
	statsMu sync.RWMutex
	stats   StatsSnapshot
}

func newWorker(cfg config.Config, log *zap.Logger, db *sql.DB, spec CoinSpec, window *windowAggregator) *worker {
	engine := tracker.NewEngine(tracker.Config{
		Threshold:   spec.Threshold,
		ThresholdOp: tracker.ThresholdOp(spec.ThresholdOp),
		MaxTracked:  cfg.Monitor.MaxTrackedOrders,
	})

	w := &worker{
		cfg:    cfg,
		log:    log.With(logger.String("coin", spec.ID)),
		db:     db,
		spec:   spec,
		engine: engine,
		window: window,
	}

	w.restoreRecentFilled()

	return w
}

type depthPayload struct {
	Bids [][]string `json:"bids"`
	Asks [][]string `json:"asks"`
	B    [][]string `json:"b"`
	A    [][]string `json:"a"`
}

type StatsSnapshot struct {
	BuyWallQty    float64 `json:"buyWallQty"`
	SellWallQty   float64 `json:"sellWallQty"`
	BuyWallCount  int     `json:"buyWallCount"`
	SellWallCount int     `json:"sellWallCount"`
	BuyValue      float64 `json:"buyValue"`
	SellValue     float64 `json:"sellValue"`
	UpdatedAt     int64   `json:"updatedAt"`
}

type CoinRuntimeSnapshot struct {
	CoinID      string              `json:"coinId"`
	Symbol      string              `json:"symbol"`
	DisplayName string              `json:"displayName"`
	Stats       StatsSnapshot       `json:"stats"`
	Windows     []orderagg.Snapshot `json:"windows"`
}

func (w *worker) run(ctx context.Context) {
	reconnectDelay := time.Second

	for {
		select {
		case <-ctx.Done():
			w.log.Info("summary worker context done")
			return
		default:
		}

		if err := w.consume(ctx); err != nil && !errors.Is(err, context.Canceled) {
			w.log.Warn("summary worker consume failed", logger.ErrorField(err))
			select {
			case <-time.After(reconnectDelay):
			case <-ctx.Done():
				return
			}
			reconnectDelay = minDuration(reconnectDelay*2, 30*time.Second)
		} else {
			reconnectDelay = time.Second
		}
	}
}

func (w *worker) consume(ctx context.Context) error {
	conn, err := w.openConn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	for {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		_, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var payload depthPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			w.log.Debug("summary worker unmarshal failed", logger.ErrorField(err))
			continue
		}

		bids := payload.Bids
		asks := payload.Asks
		if len(bids) == 0 && len(asks) == 0 {
			bids = payload.B
			asks = payload.A
		}

		now := time.Now()
		result, err := w.engine.Process(bids, asks, now.Unix())
		if err != nil {
			w.log.Warn("summary worker process failed", logger.ErrorField(err))
			continue
		}

		if len(result.Filled) > 0 {
			w.persistFilled(result.Filled)
		}

		w.updateStats(result, now)

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
}

func (w *worker) openConn(ctx context.Context) (*websocket.Conn, error) {
	urlStr := buildBinanceURL(w.spec)
	dialer := websocket.Dialer{Proxy: http.ProxyFromEnvironment}

	conn, _, err := dialer.DialContext(ctx, urlStr, nil)
	if err != nil {
		return nil, err
	}
	w.log.Info("summary worker connected", logger.String("url", urlStr))
	return conn, nil
}

func (w *worker) persistFilled(orders []tracker.FilledOrder) {
	if w.db == nil || len(orders) == 0 {
		return
	}

	if w.window != nil {
		w.window.add(orders)
	}

	const insertSQL = `INSERT OR IGNORE INTO orders_filled (symbol, market_type, side, price, quantity, first_seen, filled_time, duration_seconds, threshold, threshold_op)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := w.db.Prepare(insertSQL)
	if err != nil {
		w.log.Warn("summary worker prepare insert failed", logger.ErrorField(err))
		return
	}
	defer stmt.Close()

	for _, order := range orders {
		if _, err := stmt.Exec(
			w.spec.Symbol,
			w.spec.MarketType,
			order.Side,
			order.Price,
			order.Quantity,
			order.FirstSeen,
			order.FilledTime,
			order.DurationSec,
			w.spec.Threshold,
			w.spec.ThresholdOp,
		); err != nil {
			w.log.Warn("summary worker insert failed", logger.ErrorField(err))
		}
	}
}

func (w *worker) restoreRecentFilled() {
	if w.db == nil || w.window == nil {
		return
	}

	maxWindow := time.Duration(0)
	for _, spec := range orderagg.DefaultWindows {
		if spec.Duration > maxWindow {
			maxWindow = spec.Duration
		}
	}
	if maxWindow <= 0 {
		maxWindow = 4 * time.Hour
	}

	const query = `SELECT side, price, quantity, first_seen, filled_time, duration_seconds
FROM orders_filled
WHERE symbol = ? AND market_type = ? AND threshold = ? AND threshold_op = ? AND filled_time >= ?
ORDER BY filled_time ASC`

	cutoff := time.Now().Add(-maxWindow).Unix()
	rows, err := w.db.Query(query, w.spec.Symbol, w.spec.MarketType, w.spec.Threshold, w.spec.ThresholdOp, cutoff)
	if err != nil {
		w.log.Warn("restore filled orders failed", logger.ErrorField(err))
		return
	}
	defer rows.Close()

	orders := make([]tracker.FilledOrder, 0, 512)
	for rows.Next() {
		var order tracker.FilledOrder
		if err := rows.Scan(&order.Side, &order.Price, &order.Quantity, &order.FirstSeen, &order.FilledTime, &order.DurationSec); err != nil {
			w.log.Warn("restore filled order scan failed", logger.ErrorField(err))
			return
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		w.log.Warn("restore filled order rows failed", logger.ErrorField(err))
		return
	}

	if len(orders) > 0 {
		w.window.add(orders)
		w.log.Info("restored recent filled orders", zap.Int("count", len(orders)))
	}
}

func buildBinanceURL(spec CoinSpec) string {
	symbol := strings.ToLower(strings.TrimSpace(spec.Symbol))
	depth := spec.Depth
	if depth == 0 {
		depth = 20
	}

	marketType := strings.ToLower(spec.MarketType)
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

	return fmt.Sprintf("%s/%s", base, stream)
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func (w *worker) updateStats(result tracker.ProcessResult, now time.Time) {
	var buyQty, buyValue float64
	for _, order := range result.ActiveBids {
		buyQty += order.Quantity
		buyValue += order.Quantity * order.Price
	}

	var sellQty, sellValue float64
	for _, order := range result.ActiveAsks {
		sellQty += order.Quantity
		sellValue += order.Quantity * order.Price
	}

	w.statsMu.Lock()
	w.stats = StatsSnapshot{
		BuyWallQty:    buyQty,
		SellWallQty:   sellQty,
		BuyWallCount:  len(result.ActiveBids),
		SellWallCount: len(result.ActiveAsks),
		BuyValue:      buyValue,
		SellValue:     sellValue,
		UpdatedAt:     now.Unix(),
	}
	w.statsMu.Unlock()
}

func (w *worker) statsSnapshot() StatsSnapshot {
	w.statsMu.RLock()
	defer w.statsMu.RUnlock()

	return w.stats
}
