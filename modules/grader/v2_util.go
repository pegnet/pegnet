package grader

import "math"

func gradeV2(avg []float64, opr *GradingOPR, band float64) float64 {
	assets := opr.OPR.GetOrderedAssets()
	opr.Grade = 0
	for i, asset := range assets {
		if avg[i] > 0 {
			d := math.Abs((asset.Value - avg[i]) / avg[i]) // compute the difference from the average
			if d <= band {
				d = 0
			} else {
				d -= band
			}
			opr.Grade += d * d * d * d // the grade is the sum of the square of the square of the differences
		}
	}
	return opr.Grade
}
