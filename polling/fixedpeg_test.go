package polling_test

import (
	"testing"
)

func TestFixedPEGPeggedAssets(t *testing.T) {
	FixedDataSourceTest(t, "FixedPEG", []byte{})
}

func TestActualFixedPEGPeggedAssets(t *testing.T) {
	ActualDataSourceTest(t, "FixedPEG")
}
