package summary

import (
	"testing"
	"time"

	orderagg "ordermonitor/internal/aggregator"
	"ordermonitor/internal/tracker"
)

func TestSnapshotDoesNotPruneLongerWindows(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	agg := newWindowAggregator()
	agg.add([]tracker.FilledOrder{
		{Side: tracker.SideBuy, Quantity: 10, FilledTime: now.Add(-10 * time.Minute).Unix()},
		{Side: tracker.SideBuy, Quantity: 20, FilledTime: now.Add(-45 * time.Minute).Unix()},
		{Side: tracker.SideSell, Quantity: 30, FilledTime: now.Add(-2 * time.Hour).Unix()},
	})

	buyQty, sellQty, buyCnt, sellCnt := agg.snapshot(now, 15*time.Minute)
	if buyQty != 10 || sellQty != 0 || buyCnt != 1 || sellCnt != 0 {
		t.Fatalf("15m snapshot = buy %.0f/%d sell %.0f/%d", buyQty, buyCnt, sellQty, sellCnt)
	}

	snapshots := agg.snapshots(now, orderagg.DefaultWindows)
	byWindow := make(map[orderagg.Window]orderagg.Snapshot, len(snapshots))
	for _, snapshot := range snapshots {
		byWindow[snapshot.Window] = snapshot
	}

	if got := byWindow[orderagg.Window30m]; got.BuyQty != 10 || got.BuyCnt != 1 || got.SellQty != 0 || got.SellCnt != 0 {
		t.Fatalf("30m snapshot after 15m read = %+v", got)
	}
	if got := byWindow[orderagg.Window1h]; got.BuyQty != 30 || got.BuyCnt != 2 || got.SellQty != 0 || got.SellCnt != 0 {
		t.Fatalf("1h snapshot after 15m read = %+v", got)
	}
	if got := byWindow[orderagg.Window4h]; got.BuyQty != 30 || got.BuyCnt != 2 || got.SellQty != 30 || got.SellCnt != 1 {
		t.Fatalf("4h snapshot after 15m read = %+v", got)
	}
}
