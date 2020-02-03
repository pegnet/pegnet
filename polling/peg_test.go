package polling_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/pegnet/pegnet/common"
	. "github.com/pegnet/pegnet/polling"
)

func TestPegAssets_Clone(t *testing.T) {
	for i := 0; i < 1000; i++ {
		pa := make(PegAssets)
		for _, asset := range common.AllAssets {
			if rand.Int()%2 == 0 {
				ts := time.Now().Add(time.Duration(rand.Int63n(int64(time.Hour))))
				pa[asset] = PegItem{
					Value:    TruncateTo8(rand.Float64()),
					WhenUnix: ts.Unix(),
					When:     ts,
				}
			}
		}
		c := pa.Clone(0)
		if len(pa) != len(c) {
			t.Errorf("clone failed")
		}
	}
}
