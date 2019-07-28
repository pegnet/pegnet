// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/common"
)

// "dupe opr". hijacks OPRChainID to store the full name to use while testing
// the ID is the first character of the name
func dopr(oprhash, nonce string) *OraclePriceRecord {
	//split := strings.Split("name", "")
	o := NewOraclePriceRecord()
	o.OPRHash = []byte(oprhash)
	o.Nonce = []byte(nonce)
	return o
}

func dupeCheck(got []*OraclePriceRecord, want []string) error {
	if len(got) != len(want) {
		return fmt.Errorf("results are not the same length, got = %d, want = %d", len(got), len(want))
	}

	for i, o := range got {
		name := string(o.OPRHash) + string(o.Nonce)
		if name != want[i] {
			return fmt.Errorf("pos %d. got = %s, want = %s", i, name, want[i])
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
		{"one input", args{dopr("a", "1")}, []string{"a1"}},
		{"two normal inputs", args{dopr("a", "1"), dopr("b", "1")}, []string{"a1", "b1"}},
		{"1 dupe", args{dopr("a", "1"), dopr("a", "1")}, []string{"a1"}},
		{"1 dupe, 1 normal", args{dopr("a", "1"), dopr("a", "1"), dopr("b", "1")}, []string{"a1", "b1"}},
		{"double dupes", args{dopr("a", "1"), dopr("a", "1"), dopr("a", "2"), dopr("a", "2")}, []string{"a1", "a2"}},
		{"3 dupes", args{dopr("a", "1"), dopr("a", "1"), dopr("a", "1"), dopr("a", "2")}, []string{"a1", "a2"}},
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

type gradeTest struct {
	name   string
	args   []*OraclePriceRecord // input data
	sorted []string             // expected order of ids
}

var id int // gets reset to 0 by genTest

func uniqID() string {
	id++
	return strconv.Itoa(id)
}

var difficulty []*OraclePriceRecord

func init() {
	InitLX()
	// create difficulties that are in order and indexed in an array so I can assign them and compare
	// them in the tests.
	for i := 0; i < 100; i++ {
		opr := NewOraclePriceRecord()
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, opr.Difficulty)
		opr.Nonce = []byte{byte(i)}
		opr.SelfReportedDifficulty = buf
		//opr.Entry.Content = []byte(fmt.Sprintf("Entry %05d Content for this entry", i))
		opr.Difficulty = opr.ComputeDifficulty(opr.Nonce)
		difficulty = append(difficulty, opr)
	}
	sort.Slice(difficulty, func(i, j int) bool {
		return difficulty[i].Difficulty < difficulty[j].Difficulty
	})
}

func genOPR(entry gradeEntry) *OraclePriceRecord {
	opr := NewOraclePriceRecord()
	opr.FactomDigitalID = entry.id
	for _, k := range common.AllAssets {
		opr.Assets[k] = entry.data
	}

	return opr
}

// this gets turned into *OraclePriceRecord in genTest
type gradeEntry struct {
	id         string
	difficulty uint64
	data       float64
}

// entry with unique id
func e1(difficulty uint64, data float64) gradeEntry {
	return gradeEntry{id: uniqID(), difficulty: difficulty, data: data}
}

// generate the OPR test case from input and outcome
func genTest(name string, entries []gradeEntry, results []string) (gt gradeTest) {
	id = 0 // reset uniq id
	gt.name = name
	if len(entries) < 10 {
		panic("genTest needs at least ten entries")
	}
	gt.args = make([]*OraclePriceRecord, 0)
	for _, e := range entries {

		diff := e.difficulty
		en := genOPR(e)

		en.Nonce = difficulty[diff].Nonce
		en.SelfReportedDifficulty = difficulty[diff].SelfReportedDifficulty
		en.Difficulty = e.difficulty

		gt.args = append(gt.args, en)
	}
	gt.sorted = results
	return
}

func gradeCompare(ids []string, entries, winners, sorted []*OraclePriceRecord) error {
	if len(entries) < len(sorted) { // dropped records
		return fmt.Errorf("there are more results than input")
	}

	exists := make(map[string]bool)
	unique := make(map[string]uint64)
	for _, e := range entries {
		id := e.FactomDigitalID
		exists[fmt.Sprintf("%s-%d", id, e.Difficulty)] = true
		if e.Difficulty >= unique[id] {
			unique[id] = e.Difficulty
		}
	}

	if !sort.SliceIsSorted(sorted, func(i, j int) bool {
		// i is before j when:
		// grade is smaller (better)
		//  or difficulty higher
		return sorted[i].Grade < sorted[j].Grade || (sorted[i].Grade == sorted[j].Grade && sorted[i].Difficulty > sorted[j].Difficulty)
	}) {
		return fmt.Errorf("the results are not sorted")
	}

	if len(winners) < 10 {
		return fmt.Errorf("there are fewer than 10 winners")
	}

	for i := range winners {
		if winners[i] != sorted[i] {
			return fmt.Errorf("winners and sorted are not the same at index %d", i)
		}
	}

	dupe := make(map[string]bool)
	for i, e := range sorted {
		id := e.FactomDigitalID
		if !exists[fmt.Sprintf("%s-%d", id, e.Difficulty)] {
			return fmt.Errorf("unknown record showed up in sorted set: id=%s", id)
		}
		if unique[id] != e.Difficulty {
			return fmt.Errorf("duplicate record with highest difficulty wasn't picked. id=%s, wanted=%d, got=%d", id, unique[id], e.Difficulty)
		}

		if dupe[id] {
			return fmt.Errorf("record id=%s is duplicate", id)
		}
		dupe[id] = true

		if id != ids[i] {
			return fmt.Errorf("Did not get the results we expected. at position %d, want = %s, got = %s", i, ids[i], id)
		}
	}

	return nil
}

func TestGradeBlock(t *testing.T) {
	r1, r2 := GradeBlock(nil)
	if r1 != nil || r2 != nil {
		t.Errorf("nil param produced non nil results: %v, %v", r1, r2)
	}

	tests := []gradeTest{
		genTest("higher difficulty wins", []gradeEntry{
			e1(1, 1.00),
			e1(2, 1.00),
			e1(3, 1.00),
			e1(4, 1.00),
			e1(5, 1.00),
			e1(6, 1.00),
			e1(7, 1.00),
			e1(8, 1.00),
			e1(9, 1.00),
			e1(10, 1.00),
		}, []string{"10", "9", "8", "7", "6", "5", "4", "3", "2", "1"}),
		genTest("one outlier = last", []gradeEntry{
			e1(1, 1.00),
			e1(2, 1.00),
			e1(3, 1.00),
			e1(4, 1.00),
			e1(5, 1.00),
			e1(6, 1.00),
			e1(7, 1.00),
			e1(8, 1.00),
			e1(9, 1.00),
			e1(10, 1.00),
			e1(11, 2.00),
		}, []string{"10", "9", "8", "7", "6", "5", "4", "3", "2", "1", "11"}),
		genTest("one outlier = first", []gradeEntry{
			e1(1, 2.00),
			e1(2, 1.00),
			e1(3, 1.00),
			e1(4, 1.00),
			e1(5, 1.00),
			e1(6, 1.00),
			e1(7, 1.00),
			e1(8, 1.00),
			e1(9, 1.00),
			e1(10, 1.00),
			e1(11, 1.00),
		}, []string{"11", "10", "9", "8", "7", "6", "5", "4", "3", "2", "1"}),
		genTest("one outlier = middle", []gradeEntry{
			e1(1, 1.00),
			e1(2, 1.00),
			e1(3, 1.00),
			e1(4, 1.00),
			e1(5, 0.00),
			e1(6, 1.00),
			e1(7, 1.00),
			e1(8, 1.00),
			e1(9, 1.00),
			e1(10, 1.00),
			e1(11, 1.00),
		}, []string{"11", "10", "9", "8", "7", "6", "4", "3", "2", "1", "5"}),
		genTest("stable order 2 items", []gradeEntry{
			e1(1, 1.00), // same difficulty
			e1(2, 1.00),
			e1(3, 1.00),
			e1(4, 1.00),
			e1(5, 1.00),
			e1(6, 1.00),
			e1(7, 1.00),
			e1(8, 1.00),
			e1(9, 1.00),
			e1(10, 1.00), //                         stable order
		}, []string{"10", "9", "8", "7", "6", "5", "4", "3", "2", "1"}),
		/*		genTest("stable order 3 items", []gradeEntry{
					e1(1, 1.00), // same difficulty
					e1(1, 1.00),
					e1(1, 1.00),
					e1(4, 1.00),
					e1(5, 1.00),
					e1(6, 1.00),
					e1(7, 1.00),
					e1(8, 1.00),
					e1(9, 1.00),
					e1(10, 1.00), //                         stable order
				}, []string{"10", "9", "8", "7", "6", "5", "4", "1", "2", "3"}),
				genTest("reordered input", []gradeEntry{
					e2(1, 1.00, "1"),
					e2(3, 1.00, "3"),
					e2(9, 1.00, "9"),
					e2(4, 1.00, "4"),
					e2(6, 1.00, "6"),
					e2(2, 1.00, "2"),
					e2(5, 1.00, "5"),
					e2(7, 1.00, "7"),
					e2(10, 1.00, "10"),
					e2(8, 1.00, "8"),
				}, []string{"10", "9", "8", "7", "6", "5", "4", "3", "2", "1"}),
				genTest("reordered input (stable)", []gradeEntry{
					e2(1, 1.00, "1"),
					e2(2, 1.00, "3"),
					e2(3, 1.00, "9"),
					e2(4, 1.00, "4"),
					e2(5, 1.00, "6"),
					e2(6, 1.00, "2"),
					e2(7, 1.00, "5"),
					e2(8, 1.00, "7"),
					e2(9, 1.00, "10"),
					e2(10, 1.00, "8"),
				}, []string{"1", "3", "9", "4", "6", "2", "5", "7", "10", "8"}),
				genTest("same difficulty, diff results (10)", []gradeEntry{
					e1(2, 5.00), // avg = 3.3
					e1(7, 4.00),
					e1(1, 3.00),
					e1(3, 2.00),
					e1(6, 1.00),
					e1(8, 3.00),
					e1(4, 4.00),
					e1(9, 5.00),
					e1(5, 6.00),
					e1(10, 7.00),
				}, []string{"2", "7", "1", "3", "6", "8", "4", "9", "5", "10"}),
				genTest("same difficulty, strong outlier (10)", []gradeEntry{
					e1(1, 5.00), // avg = 3.3
					e1(1, 4.00),
					e1(1, 3.00),
					e1(1, 2.00),
					e1(1, 1.00),
					e1(1, 3.00),
					e1(1, 4.00),
					e1(1, 5.00),
					e1(1, 6.00),
					e1(1, 1000000.00),
				}, []string{"9", "1", "8", "2", "7", "3", "6", "4", "5", "10"}),
				genTest("normal, strong outlier (10)", []gradeEntry{
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 10000000.00),
				}, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}),
				genTest("normal, strong outlier 2 (10)", []gradeEntry{
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 2.00), // weak outlier
					e1(1, 10000000.00),
				}, []string{"9", "1", "2", "3", "4", "5", "6", "7", "8", "10"}),
				genTest("low difficulty but strong outlier 2 (10)", []gradeEntry{
					e1(10, 1.00),
					e1(10, 1.00),
					e1(10, 1.00),
					e1(10, 1.00),
					e1(10, 1.00),
					e1(10, 1.00),
					e1(10, 1.00),
					e1(10, 1.00),
					e1(1, 2.00), // weak outlier
					e1(1, 10000000.00),
				}, []string{"9", "1", "2", "3", "4", "5", "6", "7", "8", "10"}),
				genTest("low difficulty but strong outlier 2 (10)", []gradeEntry{
					e1(10, 1.00),
					e1(10, 1.00),
					e1(10, 1.00),
					e1(10, 1.00),
					e1(10, 1.00),
					e1(10, 1.00),
					e1(10, 1.00),
					e1(10, 1.00),
					e1(10, 1.00),
					e1(10, 1.00),
					e1(1, 20.00), // weaker outlier
					e1(1, 1000000.00),
				}, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"}),
				genTest("all equal 10", []gradeEntry{
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
				}, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}),
				genTest("all equal 20", []gradeEntry{
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
				}, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20"}),
				genTest("higher difficulty wins 20", []gradeEntry{
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(2, 1.00),
				}, []string{"20", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19"}),
				genTest("zero difficulty", []gradeEntry{
					e1(0, 1.00),
					e1(0, 1.00),
					e1(0, 1.00),
					e1(0, 1.00),
					e1(0, 1.00),
					e1(0, 1.00),
					e1(0, 1.00),
					e1(0, 1.00),
					e1(0, 1.00),
					e1(0, 1.00),
				}, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}),
				genTest("max difficulty", []gradeEntry{
					e1(19, 1.00),
					e1(19, 1.00),
					e1(19, 1.00),
					e1(19, 1.00),
					e1(19, 1.00),
					e1(19, 1.00),
					e1(19, 1.00),
					e1(19, 1.00),
					e1(19, 1.00),
					e1(19, 1.00),
				}, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}),
				genTest("max float64", []gradeEntry{
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, math.MaxFloat64),
				}, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}),
				genTest("min float64", []gradeEntry{
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, math.SmallestNonzeroFloat64),
				}, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}),
				genTest("negative result", []gradeEntry{
					e1(1, -1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
					e1(1, 1.00),
				}, []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "1"}),

		*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			winners, sorted := GradeBlock(tt.args)
			if err := gradeCompare(tt.sorted, tt.args, winners, sorted); err != nil {
				t.Error(err)
			}
		})
	}
}
