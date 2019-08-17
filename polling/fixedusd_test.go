package polling_test

import (
	"testing"
)

func TestFixedUSDPeggedAssets(t *testing.T) {
	FixedDataSourceTest(t, "FixedUSD", []byte{})
}

func TestActualFixedUSDPeggedAssets(t *testing.T) {
	ActualDataSourceTest(t, "FixedUSD")
}
