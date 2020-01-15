package conversions_test

import (
	"math/rand"
	"testing"

	. "github.com/pegnet/pegnet/modules/conversions"
)

func TestArbitrageNeeded(t *testing.T) {
	type vec struct {
		RefRate   uint64
		ChainRate uint64
		Ret       int
	}

	vecs := []vec{
		// Easy ones
		{RefRate: 1, ChainRate: 1, Ret: 0},
		{RefRate: 1e8, ChainRate: 1e8, Ret: 0},
		{RefRate: 5e10, ChainRate: 5e10, Ret: 0},

		// On the boundary
		{RefRate: 1e8 - 1e6, ChainRate: 1e8, Ret: 0},
		{RefRate: 1e8 - 1e6 + 1, ChainRate: 1e8, Ret: 0},
		{RefRate: 1e8 - 1e6 - 1, ChainRate: 1e8, Ret: -1},

		{RefRate: 1e8 + 1e6, ChainRate: 1e8, Ret: 0},
		{RefRate: 1e8 + 1e6 - 1, ChainRate: 1e8, Ret: 0},
		{RefRate: 1e8 + 1e6 + 1, ChainRate: 1e8, Ret: 1},

		// Far off
		{RefRate: 1e8 - 1e7, ChainRate: 1e8, Ret: -1},
		{RefRate: 1e8 + 1e7, ChainRate: 1e8, Ret: 1},
	}

	t.Run("vectored", func(t *testing.T) {
		for i, v := range vecs {
			r := ArbitrageNeeded(v.ChainRate, v.RefRate)
			if r != v.Ret {
				// The floats are close enough so you know what test it is for.
				// But don't forget the prints are using floats, so might not be
				// perfect.
				t.Errorf("[%d]Unexpected result, expected %d, found %d. Chain: %.8f, Ref: %.8f",
					i, v.Ret, r, float64(v.ChainRate)/1e8, float64(v.RefRate)/1e8)
			}
		}
	})

	t.Run("random", func(t *testing.T) {
		// Random tests
		for i := 0; i < 10000; i++ {
			x := rand.Uint64() % (1e6 * 1e8) // 100K
			y := rand.Uint64() % (1e6 * 1e8) // 100K

			r := ArbitrageNeeded(x, y)
			min := uint64(float64(x) * 0.99)
			max := uint64(float64(x) * 1.01)
			exp := 0

			if y > max {
				exp = 1
			} else if y < min {
				exp = -1
			}
			if r != exp {
				t.Errorf("[%d]Unexpected result, expected %d, found %d. Chain: %.8f, Ref: %.8f",
					i, exp, r, float64(x)/1e8, float64(y)/1e8)
			}

		}
	})
}
