package common_test

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/pegnet/pegnet/common"
)

func TestBasicAssetList(t *testing.T) {
	for _, asset := range AllAssets {
		if !AssetListContains(AllAssets, asset) {
			t.Errorf("%s is missing from the asset list?", asset)
		}

		if !AssetListContainsCaseInsensitive(AllAssets, strings.ToLower(asset)) {
			t.Errorf("%s is missing from the asset list?", asset)
		}
	}
}

func TestSetSubtract(t *testing.T) {
	for _, asset := range AssetsV2 {
		if asset == "XPD" || asset == "XPT" {
			t.Errorf("contains %s when it should not", asset)
		}
	}

	var set []string
	for i := 0; i < 100; i++ {
		set = append(set, fmt.Sprintf("%d", i))
	}

	for i := 0; i < 50; i += 2 {
		first := fmt.Sprintf("%d", i)
		last := fmt.Sprintf("%d", 100-i)
		midish := fmt.Sprintf("%d", (100-i)/2)
		newset := SubtractFromSet(set,
			first,  // Subtract from index off front
			last,   // Subtract from index off end
			midish, // Subtract from middle-ish of first and last
		)

		if AssetListContains(newset, first) || AssetListContains(newset, last) || AssetListContains(newset, midish) {
			t.Errorf("set subtract failed")
		}
	}
}
