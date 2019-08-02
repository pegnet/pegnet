package opr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pegnet/pegnet/common"
)

// OracleFloat has a custom marshal
type OracleFloat float64

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
//
//   X TO         to_usd               1
//  ------    *  --------- = USD * -------- = FROM
//     1            1               from_usd
//
func (o OraclePriceRecordAssetList) ExchangeTo(from string, to string, want int64) (int64, error) {
	fromRate, toRate, err := o.ExchangeRates(from, to)
	if err != nil {
		return 0, err
	}

	// Truncate vs round, so txs will drop the partials, as rounding up would create tokens.
	// Office space code goes here
	return common.TruncateFloat(float64(want) * toRate / fromRate), err
}

// Exchange tells us how much we need to spend given the amount we have is fixed.
//  X FROM -> ?? TO
//
//  X FROM       from_usd             1
//  ------    *  --------- = USD * -------- = TO
//     1            1               to_usd
//
func (o OraclePriceRecordAssetList) ExchangeFrom(from string, have int64, to string) (int64, error) {
	fromRate, toRate, err := o.ExchangeRates(from, to)
	if err != nil {
		return 0, err
	}

	// Truncate vs round, so txs will drop the partials, as rounding up would create tokens.
	// Office space code goes here
	return common.TruncateFloat(float64(have) * fromRate / toRate), err
}

// ExchangeRate finds the exchange rates for FROM and TO in usd as the base pair.
func (o OraclePriceRecordAssetList) ExchangeRates(from, to string) (fromRate float64, toRate float64, err error) {
	var ok bool
	// First we need to ensure we have the pricing for each side of the exchange
	fromRate, ok = o[from]
	if !ok {
		return 0, 0, fmt.Errorf("did not find a rate for %s", from)
	}

	toRate, ok = o[to]
	if !ok {
		return 0, 0, fmt.Errorf("did not find a rate for %s", to)
	}

	if toRate == 0 || fromRate == 0 {
		return 0, 0, fmt.Errorf("one of the rates found is 0")
	}

	// fromRate / toRate
	return fromRate, toRate, nil
}

// ExchangeRate finds the exchange rate going from `FROM` to `TO`.
//	To do the exchange rate, USD is the base pair and used as the intermediary.
//	So to go from FCT -> BTC, the math goes:
//		FCT -> USD -> BTC
func (o OraclePriceRecordAssetList) ExchangeRate(from, to string) (float64, error) {
	fromRate, toRate, err := o.ExchangeRates(from, to)
	if err != nil {
		return 0, err
	}
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
