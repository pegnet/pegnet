package conversions

import (
	"fmt"
	"math/big"
)

// OraclePriceRecordAssetList is used such that the marshaling of the assets
// is in the same order, and we still can use map access in the code
// 	Key: Asset
//	Value: Exchange rate to USD * 1e8
type OraclePriceRecordAssetList map[string]uint64

// MinimumInputNeeded tells us how many `inputType` tokens are needed to create the requested amount of
// 'outputType` tokens.
// All inputs must be in their lowest divisible unit as whole numbers.
//	?? inputType -> X outputType
//
//  X outputType     outputRate USD     1 outputType
//  ------------  *  --------------  *  -------------  =  ?? inputType
//        1           1 outputType      inputRate USD
func (o OraclePriceRecordAssetList) MinimumInputNeeded(inputType string, outputType string, requestedOutputAmount int64) (int64, error) {
	return o.exchange(outputType, requestedOutputAmount, inputType)
}

// MaximumOutputPossible tells us how many `outputType` tokens can be created from a given number of `inputType` tokens.
// All inputs must be in their lowest divisible unit as whole numbers.
//  X inputType -> ?? outputType
//
//  X inputType     inputRate USD     1 outputType
//  -----------  *  -------------  * --------------  =  ?? outputType
//     1             1 inputType     outputRate USD
func (o OraclePriceRecordAssetList) MaximumOutputPossible(inputType string, availableInputAmount int64, outputType string) (int64, error) {
	return o.exchange(inputType, availableInputAmount, outputType)
}

// exchange will use big ints to avoid overflows. All inputs must be in their
// lowest divisible unit as whole numbers.
// TODO: Will we ever overflow a int64?
func (o OraclePriceRecordAssetList) exchange(inputType string, amount int64, outputType string) (int64, error) {
	fromRate, toRate, err := o.exchangeRates(inputType, outputType)
	if err != nil {
		return 0, err
	}

	// Convert the rates to integers. Because these rates are in USD, we will switch all our inputs to
	// 1e-8 fixed point. The `want` should already be in this format. This should be the most amount of
	// accuracy a miner reports. Anything beyond the 8th decimal point, we cannot account for.
	fr := big.NewInt(int64(fromRate))
	tr := big.NewInt(int64(toRate))
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

// exchangeRates finds the exchange rates for inputType and outputType denominated in USD.
func (o OraclePriceRecordAssetList) exchangeRates(inputType, outputType string) (fromRate uint64, toRate uint64, err error) {
	var ok bool
	fromRate, ok = o[inputType]
	if !ok {
		return 0, 0, fmt.Errorf("no rate found for %s", inputType)
	}

	toRate, ok = o[outputType]
	if !ok {
		return 0, 0, fmt.Errorf("no rate found for %s", outputType)
	}

	if toRate == 0 || fromRate == 0 {
		return 0, 0, fmt.Errorf("one of the rates found is 0")
	}
	return fromRate, toRate, nil
}
