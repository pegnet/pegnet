package polling_test

import (
	"net/http"
	"testing"

	"github.com/pegnet/pegnet/common"
	. "github.com/pegnet/pegnet/polling"
	"github.com/zpatrick/go-config"
)

// TestKitcoPeggedAssets tests all the metals assets are found on kitco
func TestKitcoPeggedAssets(t *testing.T) {
	c := config.NewConfig([]config.Provider{common.NewUnitTestConfigProvider()})
	peg := make(PegAssets)

	http.DefaultClient = &http.Client{}

	KitcoInterface(c, peg)
	for _, asset := range common.CommodityAssets {
		_, ok := peg[asset]
		if !ok {
			t.Errorf("Missing %s", asset)
		}
	}
}

// The fixed is huge and a web scrape
