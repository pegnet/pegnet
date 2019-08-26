package opr

import (
	"encoding/json"
	"strings"

	"github.com/pegnet/pegnet/common"
)

// OraclePriceRecordAssetList is used such that the marshaling of the assets
// is in the same order, and we still can use map access in the code
type OraclePriceRecordAssetList map[string]float64

func (o OraclePriceRecordAssetList) Contains(list []string) bool {
	for _, asset := range list {
		if _, ok := o[asset]; !ok {
			return false
		}
	}
	return true
}

func (o OraclePriceRecordAssetList) ContainsExactly(list []string) bool {
	if len(o) != len(list) {
		return false // Different number of assets
	}

	return o.Contains(list) // Check every asset in list is in ours
}

// List returns the list of assets in the global order
func (o OraclePriceRecordAssetList) List(version uint8) []Token {
	assets := common.VersionOneAssets
	if version == 2 {
		assets = common.VersionTwoAssets
	}
	tokens := make([]Token, len(assets))
	for i, asset := range assets {
		tokens[i].Code = asset
		tokens[i].Value = o[asset]
	}
	return tokens
}

// from https://github.com/iancoleman/orderedmap/blob/master/orderedmap.go#L310
func (o OraclePriceRecordAssetList) MarshalJSON() ([]byte, error) {
	var assets []string
	switch OPRVersion {
	case 1:
		assets = common.VersionOneAssets
	case 2:
		assets = common.VersionTwoAssets
	}

	s := "{"

	for _, k := range assets {
		// add key
		kEscaped := strings.Replace(k, `"`, `\"`, -1)
		s = s + `"` + kEscaped + `":`
		// add value
		v := o[k]
		vBytes, err := json.Marshal(v)
		if err != nil {
			return []byte{}, err
		}
		s = s + string(vBytes) + ","
	}
	if len(assets) > 0 {
		s = s[0 : len(s)-1]
	}
	s = s + "}"
	return []byte(s), nil
}
