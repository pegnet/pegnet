package opr_test

import (
	"math/rand"
	"testing"

	. "github.com/pegnet/pegnet/modules/opr"
)

// TestUint64ToFloatAndBack ensures our price quotes have symmetry in their
// conversions. This especially concerns v1
func TestUint64ToFloatAndBack(t *testing.T) {
	// Uint64 -> Float -> Uint64 should be symmetric for v2
	for i := 0; i < 10000; i++ {
		// We set a 100K USD max since this only conerns price quotes.
		// Once the number of significant digits is too high, then this symmetry might fail
		u := rand.Uint64() % 100000 * 1e8 // 100K max
		f := Uint64ToFloat(u)
		if y := FloatToUint64(f); y != u {
			t.Errorf("uint64 -> float64 -> uint64 failed. exp %d, got %d", u, y)
		}
	}

	// Float64 (truncated to 4) -> Uint64 -> Float64 should be symmetric
	for i := 0; i < 10000; i++ {
		// V1 has price quotes in floats truncated to 4 decimal places
		f := float64(int64(rand.Float64()*1e4)) / 1e4
		u := FloatToUint64(f)
		if y := Uint64ToFloat(u); y != f {
			t.Errorf("float64 -> uint64 -> float64 failed. exp %.8f, got %.8f", f, y)
		}
	}
}
