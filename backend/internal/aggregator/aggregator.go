package aggregator

import (
	"time"
)

type Window string

const (
	Window15m Window = "15m"
	Window30m Window = "30m"
	Window1h  Window = "1h"
	Window4h  Window = "4h"
)

type Snapshot struct {
	Window  Window  `json:"window"`
	BuyQty  float64 `json:"buyQty"`
	SellQty float64 `json:"sellQty"`
	BuyCnt  int     `json:"buyCnt"`
	SellCnt int     `json:"sellCnt"`
}

type WindowSpec struct {
	ID       Window
	Duration time.Duration
}

var DefaultWindows = []WindowSpec{
	{ID: Window15m, Duration: 15 * time.Minute},
	{ID: Window30m, Duration: 30 * time.Minute},
	{ID: Window1h, Duration: time.Hour},
	{ID: Window4h, Duration: 4 * time.Hour},
}

type FilledOrder interface {
	GetSide() string
	GetQuantity() float64
	GetFilledTime() int64
}

func Aggregate(orders []FilledOrder, specs []WindowSpec, now time.Time) []Snapshot {
	if specs == nil {
		specs = DefaultWindows
	}

	snapshots := make([]Snapshot, 0, len(specs))

	for _, spec := range specs {
		from := now.Add(-spec.Duration).Unix()
		var buyQty, sellQty float64
		var buyCnt, sellCnt int

		for _, order := range orders {
			if order.GetFilledTime() < from {
				continue
			}
			switch order.GetSide() {
			case "buy":
				buyCnt++
				buyQty += order.GetQuantity()
			case "sell":
				sellCnt++
				sellQty += order.GetQuantity()
			}
		}

		snapshots = append(snapshots, Snapshot{
			Window:  spec.ID,
			BuyQty:  buyQty,
			SellQty: sellQty,
			BuyCnt:  buyCnt,
			SellCnt: sellCnt,
		})
	}

	return snapshots
}
