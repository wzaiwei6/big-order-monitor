package websocket

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ordermonitor/internal/config"
	"ordermonitor/internal/tracker"
	"ordermonitor/pkg/logger"
)

type sessionRecord struct {
	req           SessionRequest
	history       []tracker.FilledOrder // 缓存的历史数据
	historyLoaded bool                  // 是否已加载历史
	mu            sync.RWMutex          // 保护 history 和 historyLoaded
}

type Manager struct {
	cfg       config.Config
	log       *zap.Logger
	redis     *redis.Client
	db        *sql.DB
	sessions  sync.Map // key -> *sessionRecord
	shutdown  chan struct{}
	shutdownW sync.Once
}

func NewManager(cfg config.Config, log *zap.Logger, redis *redis.Client, db *sql.DB) *Manager {
	return &Manager{
		cfg:      cfg,
		log:      log,
		redis:    redis,
		db:       db,
		shutdown: make(chan struct{}),
	}
}

func (m *Manager) StartSession(ctx context.Context, req SessionRequest) error {
	if strings.TrimSpace(req.Symbol) == "" {
		return errors.New("symbol is required")
	}

	key := req.Key()
	record := &sessionRecord{req: req}
	m.sessions.Store(key, record)
	m.log.Info("session registered", logger.String("key", key), logger.String("symbol", req.Symbol))
	return nil
}

func (m *Manager) StopSession(key string) {
	if _, ok := m.sessions.Load(key); ok {
		m.sessions.Delete(key)
		m.log.Info("session record deleted", logger.String("key", key))
	}
}

func (m *Manager) Shutdown() {
	m.shutdownW.Do(func() {
		close(m.shutdown)
		// 清理所有 session records
		m.sessions.Range(func(key, _ any) bool {
			m.sessions.Delete(key)
			m.log.Info("session record cleaned", logger.String("key", key.(string)))
			return true
		})
	})
}

func (m *Manager) ActiveSessions() []SessionRequest {
	res := make([]SessionRequest, 0)
	m.sessions.Range(func(_, value any) bool {
		if record, ok := value.(*sessionRecord); ok {
			res = append(res, record.req)
		}
		return true
	})
	return res
}

func (m *Manager) HeartbeatTicker(ctx context.Context) <-chan time.Time {
	return time.Tick(30 * time.Second) //nolint:staticcheck
}
