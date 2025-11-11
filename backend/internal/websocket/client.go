package websocket

import (
	"context"
	"strings"

	"github.com/gorilla/websocket"

	"ordermonitor/internal/tracker"
	"ordermonitor/pkg/logger"
)

func (m *Manager) HandleClient(ctx context.Context, req SessionRequest, conn *websocket.Conn) {
	if strings.TrimSpace(req.Symbol) == "" {
		m.log.Warn("websocket missing symbol")
		conn.WriteJSON(map[string]any{"error": "symbol is required"}) //nolint:errcheck
		conn.Close()
		return
	}

	key := req.Key()

	// 获取或创建 record（用于缓存历史数据）
	value, loaded := m.sessions.LoadOrStore(key, &sessionRecord{req: req})
	record := value.(*sessionRecord)

	// 创建新 session（每个 WebSocket 连接独立）
	sess := newSession(key, req, m, conn)

	// 从 record 继承历史数据
	record.mu.RLock()
	if record.historyLoaded {
		sess.historyLoaded = true
		sess.history = make([]tracker.FilledOrder, len(record.history))
		copy(sess.history, record.history)
		m.log.Info("inherited history from record", logger.Int("count", len(sess.history)))
	}
	record.mu.RUnlock()

	m.log.Info("ws client connected", logger.String("key", key), logger.String("symbol", req.Symbol), logger.Bool("reconnect", loaded))

	// 运行 session（阻塞直到连接断开）
	sess.run(ctx)

	// 保存历史数据到 record
	record.mu.Lock()
	record.history = sess.history
	record.historyLoaded = sess.historyLoaded
	record.mu.Unlock()

	sess.close()

	m.log.Info("ws client disconnected", logger.String("key", key))
}
