package polling_test

import (
	"testing"

	"github.com/pegnet/pegnet/polling"
)

func TestFixedUSDPeggedAssets(t *testing.T) {
	FixedDataSourceTest(t, "FixedUSD", []byte{})
}

func TestActualFixedUSDPeggedAssets(t *testing.T) {
	ActualDataSourceTest(t, "FixedUSD")
}

// Ensure price is fixed at 1
func TestNewFixedUSDDataSource(t *testing.T) {
	d, _ := polling.NewDataSource("FixedUSD", nil)
	v, err := d.FetchPegPrice("USD")
	if err != nil {
		t.Error(err)
	}
	if v.Value != 1 {
		t.Error("Exp a 1 for the value of USD")
	}
}
