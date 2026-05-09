package summary

import (
	"sync"
	"time"

	orderagg "ordermonitor/internal/aggregator"
	"ordermonitor/internal/tracker"
)

type SummarySnapshot struct {
	CoinID      string
	Symbol      string
	DisplayName string
	BuyQty      float64
	SellQty     float64
	BuyCnt      int
	SellCnt     int
}

type windowAggregator struct {
	mu     sync.Mutex
	orders []tracker.FilledOrder
}

func newWindowAggregator() *windowAggregator {
	return &windowAggregator{orders: make([]tracker.FilledOrder, 0, 256)}
}

func (a *windowAggregator) add(orders []tracker.FilledOrder) {
	if len(orders) == 0 {
		return
	}

	a.mu.Lock()
	a.orders = append(a.orders, orders...)
	a.mu.Unlock()
}

func (a *windowAggregator) snapshot(now time.Time, window time.Duration) (buyQty, sellQty float64, buyCnt, sellCnt int) {
	cutoff := now.Add(-window).Unix()
	retentionCutoff := now.Add(-maxWindowDuration(orderagg.DefaultWindows)).Unix()

	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.orders) == 0 {
		return
	}

	filtered := a.orders[:0]
	for _, order := range a.orders {
		if order.FilledTime < retentionCutoff {
			continue
		}

		filtered = append(filtered, order)

		if order.FilledTime < cutoff {
			continue
		}

		switch order.Side {
		case tracker.SideBuy:
			buyCnt++
			buyQty += order.Quantity
		case tracker.SideSell:
			sellCnt++
			sellQty += order.Quantity
		}
	}

	a.orders = filtered

	return
}

func (a *windowAggregator) snapshots(now time.Time, specs []orderagg.WindowSpec) []orderagg.Snapshot {
	if specs == nil {
		specs = orderagg.DefaultWindows
	}

	results := make([]orderagg.Snapshot, len(specs))
	for i, spec := range specs {
		results[i].Window = spec.ID
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.orders) == 0 {
		return results
	}

	filterCutoff := now.Add(-maxWindowDuration(specs)).Unix()
	filtered := a.orders[:0]

	for _, order := range a.orders {
		if order.FilledTime < filterCutoff {
			continue
		}

		filtered = append(filtered, order)

		for idx, spec := range specs {
			if order.FilledTime < now.Add(-spec.Duration).Unix() {
				continue
			}

			switch order.Side {
			case tracker.SideBuy:
				results[idx].BuyCnt++
				results[idx].BuyQty += order.Quantity
			case tracker.SideSell:
				results[idx].SellCnt++
				results[idx].SellQty += order.Quantity
			}
		}
	}

	a.orders = filtered

	return results
}

func maxWindowDuration(specs []orderagg.WindowSpec) time.Duration {
	if len(specs) == 0 {
		return 4 * time.Hour
	}

	maxWindow := specs[0].Duration
	for _, spec := range specs[1:] {
		if spec.Duration > maxWindow {
			maxWindow = spec.Duration
		}
	}

	return maxWindow
}
