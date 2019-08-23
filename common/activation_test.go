package common_test

import (
	"testing"

	"github.com/pegnet/pegnet/common"
)

// Just checking around the hardcoded activation
func TestNetworkActive(t *testing.T) {
	for i := int64(0); i < 100; i++ {
		if common.NetworkActive(common.MainNetwork, i) {
			t.Errorf("Mainnet is not active yet")
		}
	}

	for i := int64(0); i < 100; i++ {
		if common.NetworkActive(common.MainNetwork, 206421-i) {
			t.Errorf("Mainnet is not active yet")
		}
	}

	for i := int64(0); i < 100; i++ {
		if !common.NetworkActive(common.MainNetwork, 206422+i) {
			t.Errorf("Mainnet is active")
		}
	}
}
