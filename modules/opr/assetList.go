package opr

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/pegnet/pegnet/common"
)

// Token is a combination of currency Code and Value
type Token struct {
	Code  string
	Value float64
}

// OraclePriceRecordAssetList is used such that the marshaling of the assets
// is in the same order, and we still can use map access in the code
type OraclePriceRecordAssetList struct {
	AssetList map[string]uint64
	Version   uint8
}

func NewOraclePriceRecordAssetList(version uint8) *OraclePriceRecordAssetList {
	o := new(OraclePriceRecordAssetList)
	o.Version = version
	o.AssetList = make(map[string]uint64)

	return o
}

func (o OraclePriceRecordAssetList) Contains(list []string) bool {
	for _, asset := range list {
		if _, ok := o.AssetList[asset]; !ok {
			return false
		}
	}
	return true
}

func (o OraclePriceRecordAssetList) ContainsExactly(list []string) bool {
	if len(o.AssetList) != len(list) {
		return false // Different number of assets
	}

	return o.Contains(list) // Check every asset in list is in ours
}

func (o OraclePriceRecordAssetList) SetValueFromUint64(asset string, value uint64) {
	o.AssetList[asset] = value
}

// SetValue converts the float to uint64
// Deprecated: Should not be using floats, stick to the uint64
func (o OraclePriceRecordAssetList) SetValue(asset string, value float64) {
	// This only keeps 8 decimal places of precision
	o.AssetList[asset] = uint64(math.Round(value * 1e8))
}

// Uint64Value returns 0 for empty assets
func (o OraclePriceRecordAssetList) Uint64Value(asset string) uint64 {
	return o.AssetList[asset]
}

// Value converts the value to a float
// Deprecated: Should not be using floats, stick to the uint64
func (o OraclePriceRecordAssetList) Value(asset string) float64 {
	return float64(o.AssetList[asset]) / 1e8
}

// List returns the list of assets in the global order
func (o OraclePriceRecordAssetList) List() []Token {
	assets := common.AssetsV1
	if o.Version == 2 {
		assets = common.AssetsV2
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
		// v1 uses PNT, switch to PEG
		if o.Version == 1 && k == "PNT" {
			o.SetValue("PEG", v)
		} else {
			o.SetValue(k, v)
		}
	}

	return nil
}

// from https://github.com/iancoleman/orderedmap/blob/master/orderedmap.go#L310
func (o OraclePriceRecordAssetList) MarshalJSON() ([]byte, error) {
	var assets []string
	switch o.Version {
	case 1:
		// We need to read the PNT code instead of peg so the hashes match.
		// When we digest this, we should immediately switch it to PEG
		assets = common.AssetsV1WithPNT
	case 2:
		assets = common.AssetsV2
	default:
		return nil, fmt.Errorf("version unset for json marshalling")
	}

	s := "{"

	for _, k := range assets {
		key := k
		if o.Version == 1 && k == "PEG" {
			key = "PNT" // v1 is PNT
		}

		// add key
		kEscaped := strings.Replace(key, `"`, `\"`, -1)
		s = s + `"` + kEscaped + `":`
		// add value as float.
		vBytes, err := json.Marshal(o.Value(key))
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
