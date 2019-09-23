package polling_test

import (
	"testing"

	"github.com/pegnet/pegnet/polling"
)

func TestFixedPEGPeggedAssets(t *testing.T) {
	FixedDataSourceTest(t, "FixedPEG", []byte{})
}

func TestActualFixedPEGPeggedAssets(t *testing.T) {
	ActualDataSourceTest(t, "FixedPEG")
}

// Ensure price is fixed at 0
func TestNewFixedPEGDataSource(t *testing.T) {
	d, _ := polling.NewDataSource("FixedPEG", nil)
	v, err := d.FetchPegPrice("PEG")
	if err != nil {
		t.Error(err)
	}
	if v.Value != 0 {
		t.Error("Exp a 0 for the value of PEG")
	}
}
