package mining_test

import (
	"fmt"
	"sort"
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

func TestThe(t *testing.T) {
	var _ = t
	data := []int{27, 15, 8, 9, 12, 4, 17, 19, 21, 23, 25}
	sort.Ints(data)
	fmt.Println(data)
	x := 9
	notpresent := false
	i := sort.Search(len(data), func(i int) bool { return data[i] >= x })
	if i >= len(data) || data[i] != x {
		// x is not present in data,
		// but i is the index where it would be inserted.
		notpresent = true
	}
	fmt.Println(x, notpresent)
}

func verifyOrder(gs []*GroupMinerStats) bool {
	for i := 1; i < len(gs); i++ {
		if gs[i].BlockHeight > gs[i-1].BlockHeight {
			return false
		}
	}
	return true
}
