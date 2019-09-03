package opr

import (
	"encoding/json"
	"fmt"
	"math/big"
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

// ExchangeTo tells us how much we need to spend given the amount we want is fixed. All inputs must be in their
// lowest divisible unit as whole numbers.
//	?? FROM -> X TO
//
//   X TO         to_usd               1
//  ------    *  --------- = USD * -------- = FROM
//     1            1               from_usd
func (o OraclePriceRecordAssetList) ExchangeTo(from string, to string, want int64) (int64, error) {
	return o.Exchange(to, want, from)
}

// ExchangeFrom tells us how much we need to spend given the amount we have is fixed. All inputs must be in their
// lowest divisible unit as whole numbers.
//  X FROM -> ?? TO
//
//  X FROM       from_usd             1
//  ------    *  --------- = USD * -------- = TO
//     1            1               to_usd
//
func (o OraclePriceRecordAssetList) ExchangeFrom(from string, have int64, to string) (int64, error) {
	return o.Exchange(from, have, to)
}

// Exchange will use big ints to avoid overflows. All inputs must be in their
// lowest divisible unit as whole numbers.
// TODO: Will we ever overflow a int64?
func (o OraclePriceRecordAssetList) Exchange(input string, amount int64, output string) (int64, error) {
	fromRate, toRate, err := o.ExchangeRates(input, output)
	if err != nil {
		return 0, err
	}

	// Convert the rates to integers. Because these rates are in USD, we will switch all our inputs to
	// 1e-8 fixed point. The `want` should already be in this format. This should be the most amount of
	// accuracy a miner reports. Anything beyond the 8th decimal point, we cannot account for.
	fr := big.NewInt(int64(fromRate * 1e8))
	tr := big.NewInt(int64(toRate * 1e8))
	amt := big.NewInt(amount)

	// Now we can run the conversion
	// ALWAYS multiply first. If you do not adhere to the order of operations shown
	// explicitly below, your answer will be incorrect. When doing a conversion,
	// always multiply before you divide.
	//  (amt * fromrate) / torate
	num := big.NewInt(0).Mul(amt, fr)
	num = num.Div(num, tr)
	return num.Int64(), nil
}

// ExchangeRates finds the exchange rates for FROM and TO in usd as the base pair.
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
// Don't use this for any math.
// TODO: Remove this from core? It's just for printing purposes and informing users.
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
func (o OraclePriceRecordAssetList) List(version uint8) []Token {
	assets := common.AssetsV1
	if version == 2 {
		assets = common.AssetsV2
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
		assets = common.AssetsV1
	case 2:
		assets = common.AssetsV2
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
