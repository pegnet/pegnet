package mining

import (
	"math/rand"
	"testing"

	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr"
)

func TestEntryWriter(t *testing.T) {
	c := common.NewUnitTestConfig()
	k := 3
	w := NewEntryWriter(c, k)
	miners := []chan<- *opr.NonceRanking{}
	for i := 0; i < 10; i++ {
		miners = append(miners, w.AddMiner())
	}

	go func() {
		for _, m := range miners {
			m <- randNonceRanking(k, 10)
		}
	}()

	w.CollectAndWrite(true)

}

func randNonceRanking(keep, size int) *opr.NonceRanking {
	n := opr.NewNonceRanking(keep)
	for i := 0; i < size; i++ {
		n.AddNonce([]byte{}, rand.Uint64())
	}
	return n
}
