package websocket

import (
	"fmt"
	"strings"
)

type SessionRequest struct {
	Symbol       string  `json:"symbol" binding:"required"`
	MarketType   string  `json:"marketType" binding:"required"`
	Depth        int     `json:"depth" binding:"required"`
	Threshold    float64 `json:"threshold" binding:"required"`
	ThresholdOp  string  `json:"thresholdOp" binding:"required"`
	Gateway      string  `json:"gateway"`
	GatewayProxy string  `json:"gatewayProxy"`
}

func (r SessionRequest) Key() string {
	return fmt.Sprintf("%s:%s:%s:%.2f", strings.ToLower(r.Symbol), strings.ToLower(r.MarketType), r.ThresholdOp, r.Threshold)
}
