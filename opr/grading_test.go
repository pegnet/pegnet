// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package opr

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"testing"
)

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

func genOPR(entry gradeEntry) *OraclePriceRecord {
	opr := new(OraclePriceRecord)
	opr.Difficulty = entry.difficulty
	opr.FactomDigitalID = []string{entry.id}
	opr.PNT = entry.data
	opr.USD = entry.data
	opr.EUR = entry.data
	opr.JPY = entry.data
	opr.GBP = entry.data
	opr.CAD = entry.data
	opr.CHF = entry.data
	opr.INR = entry.data
	opr.SGD = entry.data
	opr.CNY = entry.data
	opr.HKD = entry.data
	opr.XAU = entry.data
	opr.XAG = entry.data
	opr.XPD = entry.data
	opr.XPT = entry.data
	opr.XBT = entry.data
	opr.ETH = entry.data
	opr.LTC = entry.data
	opr.XBC = entry.data
	opr.FCT = entry.data
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

// entry with set id
func e2(difficulty uint64, data float64, id string) gradeEntry {
	return gradeEntry{id: id, difficulty: difficulty, data: data}
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
		gt.args = append(gt.args, genOPR(e))
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
		id := strings.Join(e.FactomDigitalID, "-")
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
		id := strings.Join(e.FactomDigitalID, "-")
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

func prettyPrint(a []*OraclePriceRecord) {
	for _, e := range a {
		fmt.Printf("[id=%s, grade=%f, diff=%d]", strings.Join(e.FactomDigitalID, "-"), e.Grade, e.Difficulty)
	}
	fmt.Println()
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
			e1(1, 1.00),
			e1(3, 1.00),
			e1(4, 1.00),
			e1(5, 1.00),
			e1(6, 1.00),
			e1(7, 1.00),
			e1(8, 1.00),
			e1(9, 1.00),
			e1(10, 1.00), //                         stable order
		}, []string{"10", "9", "8", "7", "6", "5", "4", "3", "1", "2"}),
		genTest("stable order 3 items", []gradeEntry{
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
			e2(1, 1.00, "3"),
			e2(1, 1.00, "9"),
			e2(1, 1.00, "4"),
			e2(1, 1.00, "6"),
			e2(1, 1.00, "2"),
			e2(1, 1.00, "5"),
			e2(1, 1.00, "7"),
			e2(1, 1.00, "10"),
			e2(1, 1.00, "8"),
		}, []string{"1", "3", "9", "4", "6", "2", "5", "7", "10", "8"}),
		genTest("same difficulty, diff results (10)", []gradeEntry{
			e1(1, 5.00), // avg = 3.3
			e1(1, 4.00),
			e1(1, 3.00),
			e1(1, 2.00),
			e1(1, 1.00),
			e1(1, 3.00),
			e1(1, 4.00),
			e1(1, 5.00),
			e1(1, 6.00),
			e1(1, 7.00),
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
			e1(18446744073709551615, 1.00),
			e1(18446744073709551615, 1.00),
			e1(18446744073709551615, 1.00),
			e1(18446744073709551615, 1.00),
			e1(18446744073709551615, 1.00),
			e1(18446744073709551615, 1.00),
			e1(18446744073709551615, 1.00),
			e1(18446744073709551615, 1.00),
			e1(18446744073709551615, 1.00),
			e1(18446744073709551615, 1.00),
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
