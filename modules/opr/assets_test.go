package opr_test

import (
	"testing"

	"github.com/pegnet/pegnet/modules/opr"
)

// Just ensuring no one accidentally changes a list

func TestV1AssetList(t *testing.T) {
	if len(opr.V1Assets) != 32 {
		t.Errorf("V1Asset list was changed!")
	}
}

func TestV2AssetList(t *testing.T) {
	if len(opr.V2Assets) != 30 {
		t.Errorf("V1Asset list was changed!")
	}
}
