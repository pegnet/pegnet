package opr

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/pegnet/pegnet/common"
)

// OraclePriceRecordAssetList is used such that the marshaling of the assets
// is in the same order, and we still can use map access in the code
type OraclePriceRecordAssetList map[string]uint64

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

func (o OraclePriceRecordAssetList) SetValueFromUint64(asset string, value uint64) {
	o[asset] = value
}

// SetValue converts the float to uint64
// Deprecated: Should not be using floats, stick to the uint64
func (o OraclePriceRecordAssetList) SetValue(asset string, value float64) {
	// This only keeps 8 decimal places of precision
	o[asset] = uint64(math.Round(value * 1e8))
}

// Uint64Value returns 0 for empty assets
func (o OraclePriceRecordAssetList) Uint64Value(asset string) uint64 {
	return o[asset]
}

// Value converts the value to a float
// Deprecated: Should not be using floats, stick to the uint64
func (o OraclePriceRecordAssetList) Value(asset string) float64 {
	return float64(o[asset]) / 1e8
}

// List returns the list of assets in the global order
func (o OraclePriceRecordAssetList) List(version uint8) []Token {
	assets := common.AssetsV1
	if version == 2 || version == 3 {
		assets = common.AssetsV2
	} else if version == 4 {
		assets = common.AssetsV4
	} else if version == 5 {
		assets = common.AssetsV5
	}
	tokens := make([]Token, len(assets))
	for i, asset := range assets {
		tokens[i].Code = asset
		tokens[i].Value = o.Value(asset)
	}
	return tokens
}

func (o OraclePriceRecordAssetList) UnmarshalJSON(data []byte) error {
	floatMap := make(map[string]float64)
	err := json.Unmarshal(data, &floatMap)
	if err != nil {
		return err
	}

	for k, v := range floatMap {
		o.SetValue(k, v)
	}
	return nil
}

// from https://github.com/iancoleman/orderedmap/blob/master/orderedmap.go#L310
func (o OraclePriceRecordAssetList) MarshalJSON() ([]byte, error) {
	if _, ok := o["version"]; !ok {
		return nil, fmt.Errorf("marshaling json must be called through a safe function")
	}
	version := int(o["version"])
	delete(o, "version")

	if version == 0 {
		return nil, fmt.Errorf("version unset for json marshalling")
	}

	var assets []string
	switch version {
	case 1:
		// We need to read the PNT code instead of peg so the hashes match.
		// When we digest this, we should immediately switch it to PEG
		assets = common.AssetsV1WithPNT
	case 2, 3:
		assets = common.AssetsV2
	case 4:
		assets = common.AssetsV4
	case 5:
		assets = common.AssetsV5
	}

	s := "{"

	for _, k := range assets {
		// add key
		kEscaped := strings.Replace(k, `"`, `\"`, -1)
		s = s + `"` + kEscaped + `":`
		// add value as float.
		vBytes, err := json.Marshal(o.Value(k))
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
