package mining_test

import (
	"testing"

	. "github.com/pegnet/pegnet/mining"
)

func TestStats(t *testing.T) {
	stats := NewGlobalStatTracker()

	items := []int{10, 5, 2, 100, 20, 19}
	for _, it := range items {
		stats.InsertStats(&GroupMinerStats{BlockHeight: it})
	}

	gs := stats.FetchAllStats()
	if !verifyOrder(gs) {
		t.Error("Not ordered right")
	}

	for _, it := range items {
		if stats.FetchStats(it) == nil {
			t.Error("Item not found when it should")
		}
	}

	if stats.FetchStats(18) != nil {
		t.Error("Item found when it should not")
	}

	if stats.FetchStats(1) != nil {
		t.Error("Item found when it should not")
	}

}

func verifyOrder(gs []*StatisticBucket) bool {
	for i := 1; i < len(gs); i++ {
		if gs[i].BlockHeight > gs[i-1].BlockHeight {
			return false
		}
	}
	return true
}
