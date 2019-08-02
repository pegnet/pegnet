package opr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pegnet/pegnet/common"
)

// OraclePriceRecordAssetList is used such that the marshaling of the assets
// is in the same order, and we still can use map access in the code
// 	Key: Asset
//	Value: Exchange rate to USD
type OraclePriceRecordAssetList map[string]float64

func (o OraclePriceRecordAssetList) Contains(list []string) bool {
	for _, asset := range list {
		if _, ok := o[asset]; !ok {
			return false
		}
	}

	return true
}

// Exchange tells us how much we need to spend given the amount we want is fixed.
//	?? FROM -> X TO
//	TODO: Ensure float calculations are ok.
func (o OraclePriceRecordAssetList) ExchangeTo(from string, to string, want int64) (int64, error) {
	rate, err := o.ExchangeRate(from, to)
	if err != nil {
		return 0, err
	}
	if rate == 0 {
		return 0, fmt.Errorf("exchrate is 0")
	}

	return Int64RoundedCast(float64(want) / rate), err
}

// Exchange tells us how much we need to spend given the amount we have is fixed.
//	X FROM -> ?? TO
//	TODO: Ensure float calculations are ok.
func (o OraclePriceRecordAssetList) ExchangeFrom(from string, have int64, to string) (int64, error) {
	rate, err := o.ExchangeRate(from, to)
	// The have is in 'sats'.
	return Int64RoundedCast(float64(have) * rate), err
}

// Int64RoundedCast will cast the amt to int64 and round rather than truncate
func Int64RoundedCast(amt float64) int64 {
	round := (int64(amt*10) % 10) / 5
	return int64(amt) + round
}

// ExchangeRate finds the exchange rate going from `FROM` to `TO`.
//	To do the exchange rate, USD is the base pair and used as the intermediary.
//	So to go from FCT -> BTC, the math goes:
//		FCT -> USD -> BTC
//	TODO: Ensure float calculations are ok.
func (o OraclePriceRecordAssetList) ExchangeRate(from, to string) (float64, error) {
	// First we need to ensure we have the pricing for each side of the exchange
	fromRate, ok := o[from]
	if !ok {
		return 0, fmt.Errorf("did not find a rate for %s", from)
	}

	toRate, ok := o[to]
	if !ok {
		return 0, fmt.Errorf("did not find a rate for %s", to)
	}

	// TODO: Should I round this?
	return fromRate / toRate, nil
}

func (o OraclePriceRecordAssetList) ContainsExactly(list []string) bool {
	if len(o) != len(list) {
		return false // Different number of assets
	}

	return o.Contains(list) // Check every asset in list is in ours
}

// List returns the list of assets in the global order
func (o OraclePriceRecordAssetList) List() []Token {
	tokens := make([]Token, len(common.AllAssets))
	for i, asset := range common.AllAssets {
		tokens[i].code = asset
		tokens[i].value = o[asset]
	}
	return tokens
}

// from https://github.com/iancoleman/orderedmap/blob/master/orderedmap.go#L310
func (o OraclePriceRecordAssetList) MarshalJSON() ([]byte, error) {
	s := "{"
	for _, k := range common.AllAssets {
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
	if len(common.AllAssets) > 0 {
		s = s[0 : len(s)-1]
	}
	s = s + "}"
	return []byte(s), nil
}
