package common_test

import (
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
