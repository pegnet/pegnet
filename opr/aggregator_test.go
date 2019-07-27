// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package opr_test

import (
	"math"
	"math/rand"
	"testing"

	. "github.com/pegnet/pegnet/opr"
)

func TestMergeRankings(t *testing.T) {
	t.Run("Merge vector", func(t *testing.T) {
		total := 100
		rs := make([]*NonceRanking, total)
		for i := range rs {
			rs[i] = NewNonceRanking(1)
			rs[i].AddNonce([]byte{}, uint64(i+1))
		}

		m := MergeNonceRankings(10, rs...)
		if !verifyOrder(m.GetNonces()) {
			t.Errorf("not ordered")
		}

		l := m.GetNonces()
		for i := 0; i < len(l); i++ {
			if l[i].Difficulty != uint64(total-i) {
				t.Errorf("best diff did not win, exp %d, found %d", uint64(total-i), l[i].Difficulty)
			}
		}
	})
}

func TestAggregatorOrder(t *testing.T) {
	t.Run("random adds", func(t *testing.T) {
		list := NewNonceRanking(100)
		vals := make([]uint64, 5000)
		for i := range vals {
			vals[i] = rand.Uint64()
		}

		for i := 0; i < len(vals); i++ {
			list.AddNonce([]byte{}, vals[i])
		}

		l := list.GetNonces()
		for i := 1; i < len(l); i++ {
			if l[i-1].Difficulty < l[i].Difficulty {
				t.Errorf("Not ordered")
			}
		}
	})

	t.Run("vectored", func(t *testing.T) {
		list := NewNonceRanking(10)
		total := 100
		for i := total; i >= 0; i-- {
			list.AddNonce([]byte{}, uint64(i))
		}

		l := list.GetNonces()
		for i := 0; i < len(l); i++ {
			if l[i].Difficulty != uint64(total-i) {
				t.Errorf("Not ordered")
			}
		}
	})
}

func verifyOrder(l []*UniqueOPRData) bool {
	for i := 1; i < len(l); i++ {
		if l[i-1].Difficulty < l[i].Difficulty {
			return false
		}
	}
	return true
}

/* The performance impact to mining on each iteration of using a top X list vs
 * Notes these numbers change run to run, but it's so minor compared to the hashing function
 * BenchmarkNonceAdding/Only_keeping_best_(control)-8         	20000000	       26.7 ns/op
 * BenchmarkNonceAdding/Always_Adding_Nonce_(worst_case)-8    	20000000	        57.9 ns/op
 * BenchmarkNonceAdding/Never_Adding_Nonce_(best_case)-8      	200000000	         6.34 ns/op
 * BenchmarkNonceAdding/Randomly_Adding_Nonce_(avg_case?ish)-8         	50000000	        32.0 ns/op
 */

func BenchmarkNonceAdding(b *testing.B) {
	b.Run("Only keeping best (control)", benchmarkControl)
	b.Run("Always Adding Nonce (worst case)", benchmarkAlwaysAddingNonce)
	b.Run("Never Adding Nonce (best case)", benchmarkNeverAddingNonce)
	b.Run("Randomly Adding Nonce (avg case?ish)", benchmarkRandomlyAddingNonce)
}

// benchmarkAlwaysAddingNonce will be the worst case where each nonce has a higher difficulty
func benchmarkControl(b *testing.B) {
	diffs := make([]uint64, b.N)
	for i := range diffs {
		diffs[i] = rand.Uint64()
	}

	best := uint64(0)

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		if diffs[i] > best {
			best = diffs[i]
		}
	}
}

// benchmarkAlwaysAddingNonce will be the worst case where each nonce has a higher difficulty
func benchmarkAlwaysAddingNonce(b *testing.B) {
	list := NewNonceRanking(10)
	diffs := make([]uint64, b.N)
	for i := range diffs {
		diffs[i] = uint64(i)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		list.AddNonce([]byte{}, diffs[i])
	}
}

// benchmarkAlwaysAddingNonce will be the best case where we do no-ops
func benchmarkNeverAddingNonce(b *testing.B) {
	list := NewNonceRanking(1)
	list.AddNonce([]byte{}, math.MaxUint64)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		list.AddNonce([]byte{}, uint64(i))
	}
}

// benchmarkAlwaysAddingNonce will be the avg case..ish
func benchmarkRandomlyAddingNonce(b *testing.B) {
	list := NewNonceRanking(10)
	diffs := make([]uint64, b.N)
	for i := range diffs {
		diffs[i] = rand.Uint64()
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		list.AddNonce([]byte{}, diffs[i])
	}
}
