package summary

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"

	orderagg "ordermonitor/internal/aggregator"
	"ordermonitor/internal/config"
)

type Manager struct {
	cfg config.Config
	log *zap.Logger
	db  *sql.DB

	mu       sync.Mutex
	workers  map[string]*worker
	windows  map[string]*windowAggregator
	refCount int
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewManager(cfg config.Config, log *zap.Logger, db *sql.DB) *Manager {
	return &Manager{
		cfg:     cfg,
		log:     log,
		db:      db,
		workers: make(map[string]*worker),
		windows: make(map[string]*windowAggregator),
	}
}

func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db == nil {
		return errors.New("database not configured")
	}

	if m.refCount == 0 {
		m.ctx, m.cancel = context.WithCancel(context.Background())
		m.workers = make(map[string]*worker)
		m.windows = make(map[string]*windowAggregator)
		for _, spec := range DefaultSpecs {
			agg := newWindowAggregator()
			m.windows[spec.ID] = agg
			w := newWorker(m.cfg, m.log, m.db, spec, agg)
			m.workers[spec.ID] = w
			m.wg.Add(1)
			go func(worker *worker) {
				defer m.wg.Done()
				worker.run(m.ctx)
			}(w)
		}
		m.log.Info("summary workers started")
	}

	m.refCount++
	m.log.Info("summary start", zap.Int("active_clients", m.refCount))
	return nil
}

func (m *Manager) Stop() {
	m.mu.Lock()
	if m.refCount == 0 {
		m.mu.Unlock()
		return
	}

	m.refCount--
	remaining := m.refCount
	cancel := m.cancel
	if remaining == 0 {
		m.cancel = nil
		m.ctx = nil
	}
	m.mu.Unlock()

	m.log.Info("summary stop", zap.Int("active_clients", remaining))

	if remaining == 0 {
		if cancel != nil {
			cancel()
		}
		m.wg.Wait()
		m.mu.Lock()
		m.workers = make(map[string]*worker)
		m.windows = make(map[string]*windowAggregator)
		m.mu.Unlock()
		m.log.Info("summary workers stopped")
	}
}

func (m *Manager) ActiveClients() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.refCount
}

func (m *Manager) Snapshot(window time.Duration) ([]SummarySnapshot, bool) {
	m.mu.Lock()
	if len(m.windows) == 0 {
		m.mu.Unlock()
		return nil, false
	}

	windows := make(map[string]*windowAggregator, len(m.windows))
	for key, agg := range m.windows {
		windows[key] = agg
	}
	m.mu.Unlock()

	now := time.Now()
	snapshots := make([]SummarySnapshot, 0, len(DefaultSpecs))
	for _, spec := range DefaultSpecs {
		agg, ok := windows[spec.ID]
		if !ok {
			snapshots = append(snapshots, SummarySnapshot{
				CoinID:      spec.ID,
				Symbol:      spec.Symbol,
				DisplayName: spec.DisplayName,
			})
			continue
		}
		buyQty, sellQty, buyCnt, sellCnt := agg.snapshot(now, window)
		snapshots = append(snapshots, SummarySnapshot{
			CoinID:      spec.ID,
			Symbol:      spec.Symbol,
			DisplayName: spec.DisplayName,
			BuyQty:      buyQty,
			SellQty:     sellQty,
			BuyCnt:      buyCnt,
			SellCnt:     sellCnt,
		})
	}

	return snapshots, true
}

func (m *Manager) CoinSnapshot(coinID string) (CoinRuntimeSnapshot, bool) {
	spec, ok := LookupSpec(coinID)
	if !ok {
		return CoinRuntimeSnapshot{}, false
	}

	m.mu.Lock()
	worker := m.workers[spec.ID]
	window := m.windows[spec.ID]
	m.mu.Unlock()

	snapshot := CoinRuntimeSnapshot{
		CoinID:      spec.ID,
		Symbol:      spec.Symbol,
		DisplayName: spec.DisplayName,
		Windows:     make([]orderagg.Snapshot, 0, len(orderagg.DefaultWindows)),
	}

	if worker != nil {
		snapshot.Stats = worker.statsSnapshot()
	}

	if window != nil {
		snapshot.Windows = window.snapshots(time.Now(), orderagg.DefaultWindows)
	}

	return snapshot, true
}

func LookupSpec(coinID string) (CoinSpec, bool) {
	for _, spec := range DefaultSpecs {
		if spec.ID == coinID {
			return spec, true
		}
	}

	return CoinSpec{}, false
}
