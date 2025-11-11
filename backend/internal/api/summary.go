package api

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"ordermonitor/pkg/logger"
)

type coinSummarySpec struct {
	ID          string
	Symbol      string
	DisplayName string
	Threshold   float64
	ThresholdOp string
	MarketType  string
}

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

var summaryCoins = []coinSummarySpec{
	{ID: "btc", Symbol: "btcusdt", DisplayName: "BTC", Threshold: 3, ThresholdOp: "gt", MarketType: "usdtm"},
	{ID: "eth", Symbol: "ethusdt", DisplayName: "ETH", Threshold: 300, ThresholdOp: "gt", MarketType: "usdtm"},
	{ID: "sol", Symbol: "solusdt", DisplayName: "SOL", Threshold: 800, ThresholdOp: "gt", MarketType: "usdtm"},
	{ID: "wld", Symbol: "wldusdt", DisplayName: "WLD", Threshold: 10000, ThresholdOp: "gt", MarketType: "usdtm"},
	{ID: "doge", Symbol: "dogeusdt", DisplayName: "DOGE", Threshold: 200000, ThresholdOp: "gt", MarketType: "usdtm"},
	{ID: "fil", Symbol: "filusdt", DisplayName: "FIL", Threshold: 10000, ThresholdOp: "gt", MarketType: "usdtm"},
	{ID: "bnb", Symbol: "bnbusdt", DisplayName: "BNB", Threshold: 50, ThresholdOp: "gt", MarketType: "usdtm"},
}

func (h *Handler) Get15MinuteSummary(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}

	ctx := c.Request.Context()

	entries := make([]summaryEntry, 0, len(summaryCoins))
	for _, spec := range summaryCoins {
		entry, err := h.fetchCoinSummary(ctx, spec)
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

func (h *Handler) fetchCoinSummary(ctx context.Context, spec coinSummarySpec) (summaryEntry, error) {
	const query = `SELECT side, COUNT(*) AS cnt, COALESCE(SUM(quantity), 0) AS qty
FROM orders_filled
WHERE symbol = ? AND market_type = ? AND threshold = ? AND threshold_op = ? AND filled_time >= NOW() - INTERVAL 15 MINUTE
GROUP BY side`

	entry := summaryEntry{
		CoinID:      spec.ID,
		Symbol:      spec.Symbol,
		DisplayName: spec.DisplayName,
	}

	rows, err := h.db.QueryContext(ctx, query, spec.Symbol, spec.MarketType, spec.Threshold, spec.ThresholdOp)
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
