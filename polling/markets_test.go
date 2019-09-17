package polling_test

import (
	"testing"
	"time"

	"github.com/pegnet/pegnet/polling"
)

func TestIsMarketOpen_Forex(t *testing.T) {
	type IsOpenVec struct {
		Reference time.Time
		Exp       bool
	}

	vecs := []IsOpenVec{
		// Monday
		{Reference: silenceParse("16 Sep 19 15:04 UTC"), Exp: true},
		{Reference: silenceParse("16 Sep 19 00:04 UTC"), Exp: true},
		{Reference: silenceParse("16 Sep 19 23:04 UTC"), Exp: true},
		{Reference: silenceParse("16 Sep 19 12:04 UTC"), Exp: true},

		// Tuesday
		{Reference: silenceParse("17 Sep 19 15:04 UTC"), Exp: true},
		{Reference: silenceParse("17 Sep 19 00:04 UTC"), Exp: true},
		{Reference: silenceParse("17 Sep 19 23:04 UTC"), Exp: true},
		{Reference: silenceParse("17 Sep 19 12:04 UTC"), Exp: true},

		// Wednesday
		{Reference: silenceParse("18 Sep 19 15:04 UTC"), Exp: true},
		{Reference: silenceParse("18 Sep 19 00:04 UTC"), Exp: true},
		{Reference: silenceParse("18 Sep 19 23:04 UTC"), Exp: true},
		{Reference: silenceParse("18 Sep 19 12:04 UTC"), Exp: true},

		// Thursday
		{Reference: silenceParse("19 Sep 19 15:04 UTC"), Exp: true},
		{Reference: silenceParse("19 Sep 19 00:04 UTC"), Exp: true},
		{Reference: silenceParse("19 Sep 19 23:04 UTC"), Exp: true},
		{Reference: silenceParse("19 Sep 19 12:04 UTC"), Exp: true},

		// Friday
		{Reference: silenceParse("20 Sep 19 15:04 UTC"), Exp: true},
		{Reference: silenceParse("20 Sep 19 00:04 UTC"), Exp: true},
		{Reference: silenceParse("20 Sep 19 12:04 UTC"), Exp: true},
		{Reference: silenceParse("20 Sep 19 20:59 UTC"), Exp: true},
		{Reference: silenceParse("20 Sep 19 21:00 UTC"), Exp: false},
		{Reference: silenceParse("20 Sep 19 23:04 UTC"), Exp: false},
		{Reference: silenceParse("20 Sep 19 21:04 UTC"), Exp: false},

		// Saturday
		{Reference: silenceParse("21 Sep 19 15:04 UTC"), Exp: false},
		{Reference: silenceParse("21 Sep 19 00:04 UTC"), Exp: false},
		{Reference: silenceParse("21 Sep 19 12:04 UTC"), Exp: false},
		{Reference: silenceParse("21 Sep 19 23:04 UTC"), Exp: false},
		{Reference: silenceParse("21 Sep 19 21:04 UTC"), Exp: false},

		// Sunday
		{Reference: silenceParse("22 Sep 19 15:04 UTC"), Exp: false},
		{Reference: silenceParse("22 Sep 19 00:04 UTC"), Exp: false},
		{Reference: silenceParse("22 Sep 19 12:04 UTC"), Exp: false},
		{Reference: silenceParse("22 Sep 19 20:59 UTC"), Exp: false},
		{Reference: silenceParse("22 Sep 19 23:04 UTC"), Exp: true},
		{Reference: silenceParse("22 Sep 19 21:04 UTC"), Exp: true},
	}

	for _, vec := range vecs {
		open := polling.IsMarketOpen("EUR", vec.Reference)
		if open != vec.Exp {
			t.Errorf("%s: exp %t,found %t", vec.Reference, vec.Exp, open)
		}
	}
}

func TestIsMarketOpen_Commodity(t *testing.T) {
	type IsOpenVec struct {
		Reference time.Time
		Exp       bool
	}

	vecs := []IsOpenVec{
		// Monday
		{Reference: silenceParse("16 Sep 19 15:04 UTC"), Exp: true},
		{Reference: silenceParse("16 Sep 19 00:04 UTC"), Exp: true},
		{Reference: silenceParse("16 Sep 19 23:04 UTC"), Exp: true},
		{Reference: silenceParse("16 Sep 19 12:04 UTC"), Exp: true},

		// Tuesday
		{Reference: silenceParse("17 Sep 19 15:04 UTC"), Exp: true},
		{Reference: silenceParse("17 Sep 19 00:04 UTC"), Exp: true},
		{Reference: silenceParse("17 Sep 19 23:04 UTC"), Exp: true},
		{Reference: silenceParse("17 Sep 19 12:04 UTC"), Exp: true},

		// Wednesday
		{Reference: silenceParse("18 Sep 19 15:04 UTC"), Exp: true},
		{Reference: silenceParse("18 Sep 19 00:04 UTC"), Exp: true},
		{Reference: silenceParse("18 Sep 19 23:04 UTC"), Exp: true},
		{Reference: silenceParse("18 Sep 19 12:04 UTC"), Exp: true},

		// Thursday
		{Reference: silenceParse("19 Sep 19 15:04 UTC"), Exp: true},
		{Reference: silenceParse("19 Sep 19 00:04 UTC"), Exp: true},
		{Reference: silenceParse("19 Sep 19 23:04 UTC"), Exp: true},
		{Reference: silenceParse("19 Sep 19 12:04 UTC"), Exp: true},

		// Friday
		{Reference: silenceParse("20 Sep 19 15:04 UTC"), Exp: true},
		{Reference: silenceParse("20 Sep 19 00:04 UTC"), Exp: true},
		{Reference: silenceParse("20 Sep 19 12:04 UTC"), Exp: true},
		{Reference: silenceParse("20 Sep 19 20:59 UTC"), Exp: true},
		{Reference: silenceParse("20 Sep 19 21:00 UTC"), Exp: false},
		{Reference: silenceParse("20 Sep 19 23:04 UTC"), Exp: false},
		{Reference: silenceParse("20 Sep 19 21:04 UTC"), Exp: false},

		// Saturday
		{Reference: silenceParse("21 Sep 19 15:04 UTC"), Exp: false},
		{Reference: silenceParse("21 Sep 19 00:04 UTC"), Exp: false},
		{Reference: silenceParse("21 Sep 19 12:04 UTC"), Exp: false},
		{Reference: silenceParse("21 Sep 19 23:04 UTC"), Exp: false},
		{Reference: silenceParse("21 Sep 19 21:04 UTC"), Exp: false},

		// Sunday
		{Reference: silenceParse("22 Sep 19 15:04 UTC"), Exp: false},
		{Reference: silenceParse("22 Sep 19 00:04 UTC"), Exp: false},
		{Reference: silenceParse("22 Sep 19 12:04 UTC"), Exp: false},
		{Reference: silenceParse("22 Sep 19 21:59 UTC"), Exp: false},
		{Reference: silenceParse("22 Sep 19 23:04 UTC"), Exp: true},
		{Reference: silenceParse("22 Sep 19 22:04 UTC"), Exp: true},
	}

	for _, vec := range vecs {
		open := polling.IsMarketOpen("XAG", vec.Reference)
		if open != vec.Exp {
			t.Errorf("%s: exp %t,found %t", vec.Reference, vec.Exp, open)
		}
	}
}

func TestIsMarketOpen_Crypto(t *testing.T) {
	type IsOpenVec struct {
		Reference time.Time
		Exp       bool
	}

	vecs := []IsOpenVec{
		// Monday
		{Reference: silenceParse("16 Sep 19 15:04 UTC"), Exp: true},
		{Reference: silenceParse("16 Sep 19 00:04 UTC"), Exp: true},
		{Reference: silenceParse("16 Sep 19 23:04 UTC"), Exp: true},
		{Reference: silenceParse("16 Sep 19 12:04 UTC"), Exp: true},

		// Tuesday
		{Reference: silenceParse("17 Sep 19 15:04 UTC"), Exp: true},
		{Reference: silenceParse("17 Sep 19 00:04 UTC"), Exp: true},
		{Reference: silenceParse("17 Sep 19 23:04 UTC"), Exp: true},
		{Reference: silenceParse("17 Sep 19 12:04 UTC"), Exp: true},

		// Wednesday
		{Reference: silenceParse("18 Sep 19 15:04 UTC"), Exp: true},
		{Reference: silenceParse("18 Sep 19 00:04 UTC"), Exp: true},
		{Reference: silenceParse("18 Sep 19 23:04 UTC"), Exp: true},
		{Reference: silenceParse("18 Sep 19 12:04 UTC"), Exp: true},

		// Thursday
		{Reference: silenceParse("19 Sep 19 15:04 UTC"), Exp: true},
		{Reference: silenceParse("19 Sep 19 00:04 UTC"), Exp: true},
		{Reference: silenceParse("19 Sep 19 23:04 UTC"), Exp: true},
		{Reference: silenceParse("19 Sep 19 12:04 UTC"), Exp: true},

		// Friday
		{Reference: silenceParse("20 Sep 19 15:04 UTC"), Exp: true},
		{Reference: silenceParse("20 Sep 19 00:04 UTC"), Exp: true},
		{Reference: silenceParse("20 Sep 19 12:04 UTC"), Exp: true},
		{Reference: silenceParse("20 Sep 19 20:59 UTC"), Exp: true},
		{Reference: silenceParse("20 Sep 19 21:00 UTC"), Exp: true},
		{Reference: silenceParse("20 Sep 19 23:04 UTC"), Exp: true},
		{Reference: silenceParse("20 Sep 19 21:04 UTC"), Exp: true},

		// Saturday
		{Reference: silenceParse("21 Sep 19 15:04 UTC"), Exp: true},
		{Reference: silenceParse("21 Sep 19 00:04 UTC"), Exp: true},
		{Reference: silenceParse("21 Sep 19 12:04 UTC"), Exp: true},
		{Reference: silenceParse("21 Sep 19 23:04 UTC"), Exp: true},
		{Reference: silenceParse("21 Sep 19 21:04 UTC"), Exp: true},

		// Sunday
		{Reference: silenceParse("22 Sep 19 15:04 UTC"), Exp: true},
		{Reference: silenceParse("22 Sep 19 00:04 UTC"), Exp: true},
		{Reference: silenceParse("22 Sep 19 12:04 UTC"), Exp: true},
		{Reference: silenceParse("22 Sep 19 21:59 UTC"), Exp: true},
		{Reference: silenceParse("22 Sep 19 23:04 UTC"), Exp: true},
		{Reference: silenceParse("22 Sep 19 22:04 UTC"), Exp: true},
	}

	for _, vec := range vecs {
		open := polling.IsMarketOpen("XBT", vec.Reference)
		if open != vec.Exp {
			t.Errorf("%s: exp %t,found %t", vec.Reference, vec.Exp, open)
		}
	}
}

func silenceParse(s string) time.Time {
	// 02 Jan 06 15:04 MST
	t, _ := time.Parse(time.RFC822, s)
	return t
}
