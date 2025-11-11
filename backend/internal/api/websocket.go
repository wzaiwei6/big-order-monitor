package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	ws "ordermonitor/internal/websocket"
	"ordermonitor/pkg/logger"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Handler) WebSocket(c *gin.Context) {
	req := ws.SessionRequest{
		Symbol:       c.Query("symbol"),
		MarketType:   c.Query("marketType"),
		ThresholdOp:  c.Query("thresholdOp"),
		Gateway:      c.Query("gateway"),
		GatewayProxy: c.Query("gatewayProxy"),
	}

	if depth, ok := parseInt(c.Query("depth")); ok {
		req.Depth = depth
	}
	if threshold, ok := parseFloat(c.Query("threshold")); ok {
		req.Threshold = threshold
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Error("websocket upgrade failed", logger.ErrorField(err))
		return
	}

	h.wsManager.HandleClient(c.Request.Context(), req, conn)
}

func parseInt(val string) (int, bool) {
	if val == "" {
		return 0, false
	}
	var out int
	_, err := fmt.Sscanf(val, "%d", &out)
	return out, err == nil
}

func parseFloat(val string) (float64, bool) {
	if val == "" {
		return 0, false
	}
	var out float64
	_, err := fmt.Sscanf(val, "%f", &out)
	return out, err == nil
}

