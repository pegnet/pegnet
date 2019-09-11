package opr

import "math"

// FloatToUint64 converts a float to uint with a precision of 8 decimals
func FloatToUint64(f float64) uint64 {
	return uint64(math.Round(f * 1e8))
}

// Uint64ToFloat converts a uint to a float and divides it by 1e8
func Uint64ToFloat(u uint64) float64 {
	return float64(u) / 1e8
}
