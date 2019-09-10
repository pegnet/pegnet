package opr

import "math"

func FloatToUint64(f float64) uint64 {
	return uint64(math.Round(f * 1e8))
}
func Uint64ToFloat(u uint64) float64 {
	return float64(u) / 1e8
}
