// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package opr

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/FactomProject/factom"
)

// "dupe opr". hijacks OPRChainID to store the full name to use while testing
// the ID is the first character of the name
func dopr(name string, difficulty uint64) *OraclePriceRecord {
	//split := strings.Split("name", "")
	o := new(OraclePriceRecord)
	o.FactomDigitalID = []string{string(name[0])}
	o.Difficulty = difficulty
	o.OPRChainID = name
	return o
}

func dupeCheck(got []*OraclePriceRecord, want []string) error {
	if len(got) != len(want) {
		return fmt.Errorf("results are not the same length, got = %d, want = %d", len(got), len(want))
	}

	for i, o := range got {
		if o.OPRChainID != want[i] {
			return fmt.Errorf("wrong entry at position %d. got = %s, want = %s", i, o.OPRChainID, want[i])
		}
	}

	return nil
}

func TestRemoveDuplicateMiningIDs(t *testing.T) {
	// dopr() uses the FIRST CHARACTER as id, and the full name as identifier
	// eg "a1" and "a2" are duplicate entries
	type args []*OraclePriceRecord
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"no input", nil, []string{}},
		{"empty input", args{}, []string{}},
		{"one input", args{dopr("a1", 1)}, []string{"a1"}},
		{"two normal inputs", args{dopr("a1", 1), dopr("b1", 1)}, []string{"a1", "b1"}},
		{"dupe, equal copy", args{dopr("a1", 1), dopr("a2", 1)}, []string{"a1"}},
		{"dupe, higher copy", args{dopr("a1", 1), dopr("a2", 2)}, []string{"a2"}},
		{"dupe, lower copy", args{dopr("a1", 2), dopr("a2", 1)}, []string{"a1"}},
		{"many dupes", args{dopr("a1", 2), dopr("a2", 1), dopr("a3", 5), dopr("a4", 0)}, []string{"a3"}},
		{"mixed 1", args{dopr("b1", 50), dopr("a1", 2), dopr("a2", 1)}, []string{"b1", "a1"}},
		{"middle test, keep last", args{dopr("a1", 1), dopr("b1", 50), dopr("a2", 2)}, []string{"b1", "a2"}},
		{"middle test, keep first", args{dopr("a1", 2), dopr("b1", 50), dopr("a2", 1)}, []string{"a1", "b1"}},
		{"two dupes", args{dopr("a1", 2), dopr("b1", 50), dopr("a2", 1), dopr("b2", 100)}, []string{"a1", "b2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveDuplicateMiningIDs(tt.args)
			if err := dupeCheck(got, tt.want); err != nil {
				t.Errorf("RemoveDuplicateMiningIDs() = %v", err)
			}
		})
	}
}

func BenchmarkSprintVsHash(b *testing.B) {
	words := strings.Split("Lorem ipsum dolor sit amet consectetur adipiscing elit Nullam aliquam lacinia ipsum eu blandit Nullam varius ut libero ut vulputate Maecenas id purus quis ligula molestie eleifend Duis maximus neque vitae tempor blandit Praesent commodo orci quis magna imperdiet pulvinar Morbi eget eleifend lectus Nunc eget ligula eu velit faucibus suscipit in non tellus Maecenas nec dictum neque Sed a diam et nisi tincidunt ullamcorper sit amet sit amet sapien Suspendisse ut sollicitudin justo Phasellus tincidunt mauris a dapibus ultrices elit est tincidunt libero nec dictum dui elit id magna Praesent ac magna commodo molestie neque ac imperdiet ipsum Cras sapien felis iaculis porttitor consectetur vel blandit malesuada metus", " ")

	store := make(map[int][][]string)
	gen := func(n int) [][]string {
		if existing, ok := store[n]; ok {
			return existing
		}
		random := make([][]string, n)
		for i := range random {
			random[i] = make([]string, 0)

			count := 1 + rand.Intn(4) // between 1 and 5 entries for a typical id
			for c := 0; c < count; c++ {
				random[i] = append(random[i], words[rand.Intn(len(words))])
			}
		}
		store[n] = random
		return random
	}

	b.Run("Sprint ID", func(b *testing.B) {
		b.StopTimer()
		random := gen(b.N)
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			_ = fmt.Sprintf("%d+%s", len(random[i]), strings.Join(random[i], "-"))
		}
	})
	b.Run("Sha256 ID", func(b *testing.B) {
		b.StopTimer()
		random := gen(b.N)
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			_ = factom.ChainIDFromStrings(random[i])
		}
	})
}
