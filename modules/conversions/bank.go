package conversions

// The bank is an important aspect of conversions, as it affects PEG conversions.
// These functions serve as the basis of the bank math.

// ArbitrageNeeded returns an integer value indicating if arbitrage is needed
// to close the reference token's exchange rate to the on chain rate.
// The function is named this as arbitrage is the action needed to be taken to
// change the current status of the rates.
//
// Params:
//  refRate refers the the reference token's rate on external exchanges. So in
//  example of pUSD, the refRate of pUSD could be $0.90. Arbitrage is needed
//  to move the refRate towards the chainRate
//
//  chainRate is the rate of USD posted on chain by the miners. For pUSD, this
//  is always 1.
//
// Returns
//  (a 1% tolerance is allowed, so a pUSD of 0.99 will return a 0)
//	 -1 if the refRate is below the pegged price
//    0 if at the pegged price
//    1 if above the pegged price
func ArbitrageNeeded(chainRate, refRate uint64) int {
	tolerance := chainRate / 100
	minBound := chainRate - tolerance
	maxBound := chainRate + tolerance

	if refRate < minBound {
		return -1
	} else if refRate > maxBound {
		return 1
	}

	return 0
}
