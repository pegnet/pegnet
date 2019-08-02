package polling

import "testing"

func TestVectoredRound(t *testing.T) {
	// 17132.703700
	type Vector struct {
		V   float64
		Exp float64
	}

	testVec := func(t *testing.T, v Vector) {
		if r := RoundRate(v.V); r != v.Exp {
			t.Errorf("Exp %f, found %f", v.Exp, r)
		}
	}

	vectors := []Vector{
		{V: 17132.703700, Exp: 17132.703700},
		{V: 17132.703600, Exp: 17132.703600},
		{V: 17132.703800, Exp: 17132.703800},
		{V: 10014.2259, Exp: 10014.2259},
		{V: 216.1119, Exp: 216.1119},
		{V: 96.4437, Exp: 96.4437},
		{V: 0.0422, Exp: 0.0422},
		{V: 26.6517, Exp: 26.6517},
		{V: 10199.9959, Exp: 10199.9959},
		{V: 215.4847, Exp: 215.4847},
	}

	for _, v := range vectors {
		testVec(t, v)
	}
}
