package tracker

import (
	"fmt"
	"math"
	"strconv"
	"sync"
)

type OrderSide string

type ThresholdOp string

const (
	SideBuy  OrderSide = "buy"
	SideSell OrderSide = "sell"

	OpGreater ThresholdOp = "gt"
	OpLess    ThresholdOp = "lt"
)

type ActiveOrder struct {
	Price     float64 `json:"price"`
	Quantity  float64 `json:"quantity"`
	FirstSeen int64   `json:"firstSeen"`
}

type FilledOrder struct {
	Side        OrderSide `json:"side"`
	Price       float64   `json:"price"`
	Quantity    float64   `json:"quantity"`
	FirstSeen   int64     `json:"firstSeen"`
	FilledTime  int64     `json:"filledTime"`
	DurationSec int64     `json:"durationSec"`
}

func (o FilledOrder) GetSide() string {
	return string(o.Side)
}

func (o FilledOrder) GetQuantity() float64 {
	return o.Quantity
}

func (o FilledOrder) GetFilledTime() int64 {
	return o.FilledTime
}

type ProcessResult struct {
	Filled     []FilledOrder
	ActiveBids []ActiveOrder
	ActiveAsks []ActiveOrder
}

type Config struct {
	Threshold   float64
	ThresholdOp ThresholdOp
	MaxTracked  int
}

type Engine struct {
	mu      sync.Mutex
	tracked map[string]*trackedOrder
	config  Config
}

type trackedOrder struct {
	side      OrderSide
	price     float64
	quantity  float64
	firstSeen int64
	lastSeen  int64
}

func NewEngine(cfg Config) *Engine {
	if cfg.MaxTracked <= 0 {
		cfg.MaxTracked = 5000
	}
	if cfg.ThresholdOp == "" {
		cfg.ThresholdOp = OpGreater
	}
	return &Engine{
		tracked: make(map[string]*trackedOrder),
		config:  cfg,
	}
}

func (e *Engine) UpdateConfig(cfg Config) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if cfg.MaxTracked > 0 {
		e.config.MaxTracked = cfg.MaxTracked
	}
	if cfg.ThresholdOp != "" {
		e.config.ThresholdOp = cfg.ThresholdOp
	}
	if !math.IsNaN(cfg.Threshold) && cfg.Threshold >= 0 {
		e.config.Threshold = cfg.Threshold
	}
	// 如果阈值改变，重置已有跟踪避免脏数据
	e.tracked = make(map[string]*trackedOrder)
}

func (e *Engine) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.tracked = make(map[string]*trackedOrder)
}

func (e *Engine) Process(bids [][]string, asks [][]string, now int64) (ProcessResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	result := ProcessResult{}
	current := make(map[string]struct{})

	compare := e.getComparator()

	// 处理买单
	for _, pair := range bids {
		price, qty, ok := parsePair(pair)
		if !ok {
			continue
		}
		key := fmt.Sprintf("bid_%s", pair[0])
		if compare(qty, e.config.Threshold) {
			e.trackOrder(key, SideBuy, price, qty, now, current)
			result.ActiveBids = append(result.ActiveBids, ActiveOrder{Price: price, Quantity: qty, FirstSeen: e.tracked[key].firstSeen})
		}
	}

	// 处理卖单
	for _, pair := range asks {
		price, qty, ok := parsePair(pair)
		if !ok {
			continue
		}
		key := fmt.Sprintf("ask_%s", pair[0])
		if compare(qty, e.config.Threshold) {
			e.trackOrder(key, SideSell, price, qty, now, current)
			result.ActiveAsks = append(result.ActiveAsks, ActiveOrder{Price: price, Quantity: qty, FirstSeen: e.tracked[key].firstSeen})
		}
	}

	// 判断消失的挂单
	for key, tracked := range e.tracked {
		if _, ok := current[key]; !ok {
			duration := now - tracked.firstSeen
			if duration < 0 {
				duration = 0
			}
			result.Filled = append(result.Filled, FilledOrder{
				Side:        tracked.side,
				Price:       tracked.price,
				Quantity:    tracked.quantity,
				FirstSeen:   tracked.firstSeen,
				FilledTime:  now,
				DurationSec: duration,
			})
			delete(e.tracked, key)
		}
	}

	return result, nil
}

func (e *Engine) trackOrder(key string, side OrderSide, price, qty float64, now int64, current map[string]struct{}) {
	current[key] = struct{}{}
	if existing, ok := e.tracked[key]; ok {
		existing.quantity = qty
		existing.lastSeen = now
		return
	}
	if len(e.tracked) >= e.config.MaxTracked {
		// 简单回收：找到最早的记录移除
		var oldestKey string
		var oldestLast int64 = math.MaxInt64
		for k, v := range e.tracked {
			if v.lastSeen < oldestLast {
				oldestLast = v.lastSeen
				oldestKey = k
			}
		}
		if oldestKey != "" {
			delete(e.tracked, oldestKey)
		}
	}
	e.tracked[key] = &trackedOrder{
		side:      side,
		price:     price,
		quantity:  qty,
		firstSeen: now,
		lastSeen:  now,
	}
}

func (e *Engine) getComparator() func(qty, threshold float64) bool {
	switch e.config.ThresholdOp {
	case OpLess:
		return func(qty, threshold float64) bool {
			if threshold <= 0 {
				return false
			}
			return qty <= threshold
		}
	default:
		return func(qty, threshold float64) bool {
			if threshold <= 0 {
				return qty > 0
			}
			return qty >= threshold
		}
	}
}

func parsePair(pair []string) (price float64, qty float64, ok bool) {
	if len(pair) < 2 {
		return 0, 0, false
	}
	p, err := strconv.ParseFloat(pair[0], 64)
	if err != nil {
		return 0, 0, false
	}
	q, err := strconv.ParseFloat(pair[1], 64)
	if err != nil {
		return 0, 0, false
	}
	return p, q, true
}
