package polling

import (
	"testing"

	"github.com/pegnet/pegnet/common"
	"github.com/zpatrick/go-config"
)

// TestCoinCapPeggedAssets tests all the crypto assets are found on coinmarket cap
func TestCoinCapPeggedAssets(t *testing.T) {
	c := config.NewConfig([]config.Provider{common.NewDefaultConfigProvider()})
	peg := make(PegAssets)
	CoinCapInterface(c, peg)
	for _, asset := range common.CryptoAssets {
		_, ok := peg[asset]
		if !ok {
			t.Errorf("Missing %s", asset)
		}
	}
}
