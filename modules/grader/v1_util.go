package grader

func averageV1(oprs []*GradingOPR) []float64 {
	avg := make([]float64, len(oprs[0].OPR.GetOrderedAssets()))

	// Sum up all the prices
	for _, o := range oprs {
		for i, asset := range o.OPR.GetOrderedAssets() {
			if asset.Value >= 0 { // Make sure no OPR has negative values for
				avg[i] += asset.Value // assets.  Simply treat all values as positive.
			} else {
				avg[i] -= asset.Value
			}
		}
	}
	// Then divide the prices by the number of OraclePriceRecord records.  Two steps is actually faster
	// than doing everything in one loop (one divide for every asset rather than one divide
	// for every asset * number of OraclePriceRecords)  There is also a little bit of a precision advantage
	// with the two loops (fewer divisions usually does help with precision) but that isn't likely to be
	// interesting here.
	total := float64(len(avg))
	for i := range avg {
		avg[i] = avg[i] / total
	}

	return avg
}

func gradeV1(avg []float64, opr *GradingOPR) float64 {
	assets := opr.OPR.GetOrderedAssets()
	opr.Grade = 0
	for i, asset := range assets {
		if avg[i] > 0 {
			d := (asset.Value - avg[i]) / avg[i] // compute the difference from the average
			opr.Grade += d * d * d * d           // the grade is the sum of the square of the square of the differences
		}
	}
	return opr.Grade
}
