package api

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	orderagg "ordermonitor/internal/aggregator"
	"ordermonitor/internal/summary"
	"ordermonitor/pkg/logger"
)

type summaryResponse struct {
	GeneratedAt int64          `json:"generatedAt"`
	Window      string         `json:"window"`
	Coins       []summaryEntry `json:"coins"`
}

type summaryEntry struct {
	CoinID      string  `json:"coinId"`
	Symbol      string  `json:"symbol"`
	DisplayName string  `json:"displayName"`
	BuyQty      float64 `json:"buyQty"`
	SellQty     float64 `json:"sellQty"`
	BuyCnt      int     `json:"buyCnt"`
	SellCnt     int     `json:"sellCnt"`
}

type monitorOrderEntry struct {
	Side        string  `json:"side"`
	Price       float64 `json:"price"`
	Quantity    float64 `json:"quantity"`
	FirstSeen   int64   `json:"firstSeen"`
	FilledTime  int64   `json:"filledTime"`
	DurationSec int64   `json:"durationSec"`
}

type coinMonitorResponse struct {
	GeneratedAt int64                 `json:"generatedAt"`
	CoinID      string                `json:"coinId"`
	Symbol      string                `json:"symbol"`
	DisplayName string                `json:"displayName"`
	Stats       summary.StatsSnapshot `json:"stats"`
	Windows     []orderagg.Snapshot   `json:"windows"`
	Orders      []monitorOrderEntry   `json:"orders"`
}

func (h *Handler) Get15MinuteSummary(c *gin.Context) {
	if snapshots, ok := h.summaryManager.Snapshot(15 * time.Minute); ok {
		entries := make([]summaryEntry, 0, len(snapshots))
		for _, snap := range snapshots {
			entries = append(entries, summaryEntry{
				CoinID:      snap.CoinID,
				Symbol:      snap.Symbol,
				DisplayName: snap.DisplayName,
				BuyQty:      snap.BuyQty,
				SellQty:     snap.SellQty,
				BuyCnt:      snap.BuyCnt,
				SellCnt:     snap.SellCnt,
			})
		}

		c.JSON(http.StatusOK, summaryResponse{
			GeneratedAt: time.Now().Unix(),
			Window:      "15m",
			Coins:       entries,
		})
		return
	}

	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}

	entries := make([]summaryEntry, 0, len(summary.DefaultSpecs))
	for _, spec := range summary.DefaultSpecs {
		entry, err := h.fetchCoinSummary(c.Request.Context(), spec)
		if err != nil {
			h.log.Warn("fetch coin summary failed", logger.String("coin", spec.ID), logger.ErrorField(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch summary"})
			return
		}
		entries = append(entries, entry)
	}

	c.JSON(http.StatusOK, summaryResponse{
		GeneratedAt: time.Now().Unix(),
		Window:      "15m",
		Coins:       entries,
	})
}

func (h *Handler) fetchCoinSummary(ctx context.Context, spec summary.CoinSpec) (summaryEntry, error) {
	const query = `SELECT side, COUNT(*) AS cnt, COALESCE(SUM(quantity), 0) AS qty
FROM orders_filled
WHERE symbol = ? AND market_type = ? AND threshold = ? AND threshold_op = ? AND filled_time >= ?
GROUP BY side`

	entry := summaryEntry{
		CoinID:      spec.ID,
		Symbol:      spec.Symbol,
		DisplayName: spec.DisplayName,
	}

	cutoff := time.Now().Add(-15 * time.Minute).Unix()
	rows, err := h.db.QueryContext(ctx, query, spec.Symbol, spec.MarketType, spec.Threshold, spec.ThresholdOp, cutoff)
	if err != nil {
		return entry, err
	}
	defer rows.Close()

	for rows.Next() {
		var side string
		var cnt sql.NullInt64
		var qty sql.NullFloat64
		if err := rows.Scan(&side, &cnt, &qty); err != nil {
			return entry, err
		}

		count := 0
		if cnt.Valid {
			count = int(cnt.Int64)
		}
		quantity := 0.0
		if qty.Valid {
			quantity = qty.Float64
		}

		switch side {
		case "buy":
			entry.BuyCnt = count
			entry.BuyQty = quantity
		case "sell":
			entry.SellCnt = count
			entry.SellQty = quantity
		}
	}

	if err := rows.Err(); err != nil {
		return entry, err
	}

	return entry, nil
}

func (h *Handler) GetCoinMonitor(c *gin.Context) {
	coinID := strings.ToLower(strings.TrimSpace(c.Param("coinId")))
	spec, ok := summary.LookupSpec(coinID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "coin not found"})
		return
	}

	snapshot, ok := h.summaryManager.CoinSnapshot(coinID)
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "collector not ready"})
		return
	}

	orders, err := h.fetchRecentOrders(c.Request.Context(), spec, 100)
	if err != nil {
		h.log.Warn("fetch recent orders failed", logger.String("coin", coinID), logger.ErrorField(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, coinMonitorResponse{
		GeneratedAt: time.Now().Unix(),
		CoinID:      snapshot.CoinID,
		Symbol:      snapshot.Symbol,
		DisplayName: snapshot.DisplayName,
		Stats:       snapshot.Stats,
		Windows:     snapshot.Windows,
		Orders:      orders,
	})
}

func (h *Handler) fetchRecentOrders(ctx context.Context, spec summary.CoinSpec, limit int) ([]monitorOrderEntry, error) {
	if h.db == nil {
		return nil, nil
	}

	const query = `SELECT side, price, quantity, first_seen, filled_time, duration_seconds
FROM orders_filled
WHERE symbol = ? AND market_type = ? AND threshold = ? AND threshold_op = ? AND filled_time >= ?
ORDER BY filled_time DESC
LIMIT ?`

	cutoff := time.Now().Add(-time.Duration(h.cfg.Monitor.DataRetentionHours) * time.Hour).Unix()
	rows, err := h.db.QueryContext(ctx, query, spec.Symbol, spec.MarketType, spec.Threshold, spec.ThresholdOp, cutoff, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]monitorOrderEntry, 0, limit)
	for rows.Next() {
		var entry monitorOrderEntry
		if err := rows.Scan(&entry.Side, &entry.Price, &entry.Quantity, &entry.FirstSeen, &entry.FilledTime, &entry.DurationSec); err != nil {
			return nil, err
		}
		orders = append(orders, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
