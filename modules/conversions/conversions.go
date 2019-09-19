package conversions

import (
	"fmt"
	"math/big"
)

// Convert takes an input amount and returns an output amount that can be created
// from it, given the two rates `fromRate` and `toRate` denominated in 1e-8 USD.
// All parameters must be in their lowest divisible unit as whole numbers.
//  X fromType -> ?? toType
//
//  X fromType     fromRate USD      1 toType
//  ----------  *  ------------  *  ----------  =  ?? toType
//     1            1 fromType      toRate USD
func Convert(amount int64, fromRate, toRate uint64) (int64, error) {
	if fromRate == 0 || toRate == 0 {
		return 0, fmt.Errorf("invalid rate: 0")
	}

	// Convert the rates to integers. Because these rates are in USD, we will switch all our inputs to
	// 1e-8 fixed point. The `want` should already be in this format. This should be the most amount of
	// accuracy a miner reports. Anything beyond the 8th decimal point, we cannot account for.
	//
	// Uses big ints to avoid overflows. TODO: Will we ever overflow a int64?
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
