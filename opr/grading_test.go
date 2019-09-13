// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/common"
)

// "dupe opr". hijacks OPRChainID to store the full name to use while testing
// the ID is the first character of the name
func dopr(oprhash, nonce, index string) *OraclePriceRecord {
	//split := strings.Split("name", "")
	o := NewOraclePriceRecord()
	o.OPRHash = []byte(oprhash)
	o.Nonce = []byte(nonce)
	o.OPRChainID = index // NOT A REAL VALUE, only used to track indices for unit test
	return o
}

func dupeCheck(got []*OraclePriceRecord, want []string) error {
	if len(got) != len(want) {
		return fmt.Errorf("results are not the same length, got = %d, want = %d", len(got), len(want))
	}

	for i, o := range got {
		name := string(o.OPRHash) + string(o.Nonce) + o.OPRChainID
		if name != want[i] {
			return fmt.Errorf("pos %d. got = %s, want = %s", i, name, want[i])
		}
	}

	return nil
}

func TestApplyBand(t *testing.T) {
	for i := 0; i < 200; i++ {
		f := rand.Float64()
		if i&1 == 0 {
			f = f * -1
		}
		if d := ApplyBand(f, 0); d != math.Abs(f) {
			t.Errorf("exp %.4f, got %.4f", f, d)
		}

		if d := ApplyBand(f, 0.1); d != math.Abs(f) {
			if math.Abs(f) <= 0.1 {
				if d != 0 {
					t.Errorf("1] from %.4f, exp %.4f, got %.4f", f, float64(0), d)
				}
			} else {
				if d != math.Abs(f)-0.1 {
					t.Errorf("2] from %.4f, exp %.4f, got %.4f, %t, %t", f, f-0.1, d, f <= 0.1, d == 0)
				}
			}
		}
	}
}

func TestRemoveDuplicateSubmissions(t *testing.T) {
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
		{"one input", args{dopr("a", "1", "0")}, []string{"a10"}},
		{"two normal inputs", args{dopr("a", "1", "0"), dopr("b", "1", "0")}, []string{"a10", "b10"}},
		{"1 dupe", args{dopr("a", "1", "0"), dopr("a", "1", "1")}, []string{"a10"}},
		{"1 dupe, 1 normal", args{dopr("a", "1", "0"), dopr("a", "1", "1"), dopr("b", "1", "0")}, []string{"a10", "b10"}},
		{"double dupes", args{dopr("a", "1", "0"), dopr("a", "1", "1"), dopr("a", "2", "0"), dopr("a", "2", "1")}, []string{"a10", "a20"}},
		{"3 dupes", args{dopr("a", "1", "0"), dopr("a", "1", "1"), dopr("a", "1", "2"), dopr("a", "2", "0")}, []string{"a10", "a20"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveDuplicateSubmissions(tt.args)
			if err := dupeCheck(got, tt.want); err != nil {
				t.Errorf("RemoveDuplicateSubmissions() = %v", err)
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
		opr.Nonce = []byte{byte(i)}
		opr.Difficulty = opr.ComputeDifficulty(opr.Nonce)
		binary.BigEndian.PutUint64(buf, opr.Difficulty)
		opr.SelfReportedDifficulty = buf
		//opr.Entry.Content = []byte(fmt.Sprintf("Entry %05d Content for this entry", i))
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
		opr.Assets.SetValue(k, entry.data)
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

	// Only check the top 10 graded, as anything over the 10 won't necessarily be in graded order.
	//		Each grading pass changes the grades relative to the new avg. So the grades 'jiggle' as we
	//		close in on 10.
	if !sort.SliceIsSorted(winners[:10], func(i, j int) bool {
		// i is before j when:
		// grade is smaller (better)
		//  or difficulty higher
		return winners[i].Grade < winners[j].Grade || (winners[i].Grade == winners[j].Grade && winners[i].Difficulty > winners[j].Difficulty)
	}) {
		return fmt.Errorf("the graded results are not sorted")
	}

	if !sort.SliceIsSorted(sorted, func(i, j int) bool {
		// i is before j when:
		// grade is smaller (better)
		//  or difficulty higher
		return sorted[i].Difficulty > sorted[j].Difficulty
	}) {
		return fmt.Errorf("the difficulty results are not sorted")
	}

	if len(winners) < 10 {
		return fmt.Errorf("there are fewer than 10 winners")
	}

	// TODO: Why was this a test? The lists are sorted differently.
	//for i := range winners {
	//	if winners[i] != sorted[i] {
	//		return fmt.Errorf("winners and sorted are not the same at index %d", i)
	//	}
	//}

	dupe := make(map[string]bool)
	for i, e := range winners {
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

// for BENCHMARKING grading, it's not important what the data is, we just need data
// we need ASSETS, OPRHASH, DIFFICULTY, and SELFREPORTEDDIFFICULTY
func makeBenchmarkOPR() *OraclePriceRecord {
	o := new(OraclePriceRecord)
	o.Assets = make(OraclePriceRecordAssetList)
	for _, a := range common.AllAssets {
		o.Assets.SetValue(a, rand.Float64()*50)
	}
	o.Nonce = make([]byte, 8) // random nonce
	rand.Read(o.Nonce)
	json, _ := json.Marshal(o)
	sha := sha256.Sum256(json)
	o.OPRHash = sha[:]
	o.EntryHash = json // FOR BENCHMARK ONLY
	difficulty := ComputeDifficulty(o.OPRHash, o.Nonce)
	o.SelfReportedDifficulty = make([]byte, 8)
	binary.BigEndian.PutUint64(o.SelfReportedDifficulty, difficulty)

	return o
}

func BenchmarkGradeBlock(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	InitLX()

	var oprs []*OraclePriceRecord
	for i := 0; i < 10000; i++ {
		oprs = append(oprs, makeBenchmarkOPR())
	}

	b.Run("ten", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GradeBlock(oprs[:10])
		}
	})
	b.Run("fifty", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GradeBlock(oprs[:50])
		}
	})
	b.Run("200", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GradeBlock(oprs[:200])
		}
	})
	b.Run("500", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GradeBlock(oprs[:500])
		}
	})
	b.Run("1000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GradeBlock(oprs[:1000])
		}
	})
	b.Run("5000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GradeBlock(oprs[:5000])
		}
	})
	b.Run("10000", func(b *testing.B) { // 10k OPRs = ~17 tps on factom
		for i := 0; i < b.N; i++ {
			GradeBlock(oprs[:10000])
		}
	})

}

func BenchmarkOPRHash(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	LX.Init(0xfafaececfafaecec, 30, 256, 5)

	var oprs []*OraclePriceRecord
	for i := 0; i < 10000; i++ {
		oprs = append(oprs, makeBenchmarkOPR())
	}
	b.Run("opr hash for 10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, o := range oprs[:10] {
				LX.Hash(o.EntryHash) // contains json
			}
		}
	})
	b.Run("opr hash for 50", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, o := range oprs[:50] {
				LX.Hash(o.EntryHash) // contains json
			}
		}
	})
	b.Run("opr hash for 200", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, o := range oprs[:200] {
				LX.Hash(o.EntryHash) // contains json
			}
		}
	})

	b.Run("opr hash for 500", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, o := range oprs[:500] {
				LX.Hash(o.EntryHash) // contains json
			}
		}
	})
	b.Run("opr hash for 1000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, o := range oprs[:1000] {
				LX.Hash(o.EntryHash) // contains json
			}
		}
	})
	b.Run("opr hash for 10000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, o := range oprs[:10000] {
				LX.Hash(o.EntryHash) // contains json
			}
		}
	})
}

type winner struct {
	opr     *OraclePriceRecord
	winners []*OraclePriceRecord
}

func genWinnerOPR(prev [10]string, winners []string) winner {
	opr := new(OraclePriceRecord)
	opr.WinPreviousOPR = prev[:]

	if winners == nil {
		return winner{opr, nil}
	}
	win := make([]*OraclePriceRecord, len(winners))
	for i := range win {
		win[i] = new(OraclePriceRecord)
		h, err := hex.DecodeString(winners[i])
		if err != nil {
			panic("developer error by putting invalid hex into list of winners: " + err.Error())
		}
		win[i].EntryHash = h
	}
	return winner{opr, win}
}

func TestVerifyWinners(t *testing.T) {
	var empty [10]string

	base := [10]string{
		"0000000000000000",
		"1111111111111111",
		"2222222222222222",
		"3333333333333333",
		"4444444444444444",
		"5555555555555555",
		"6666666666666666",
		"7777777777777777",
		"8888888888888888",
		"9999999999999999",
	}

	onewrong := base
	onewrong[0] = "ffffffffffffffff"

	oneempty := base
	oneempty[3] = ""

	wrongorder := base
	wrongorder[3], wrongorder[8] = wrongorder[8], wrongorder[3]

	oneshort := base
	oneshort[9] = "999999999999999"
	onelong := base
	onelong[9] = "99999999999999999"

	tests := []struct {
		name string
		args winner
		want bool
	}{
		{"empty, empty", genWinnerOPR(empty, nil), true},
		{"not empty, empty", genWinnerOPR(base, nil), false},
		{"matching", genWinnerOPR(base, base[:]), true},
		{"one wrong", genWinnerOPR(onewrong, base[:]), false},
		{"one empty", genWinnerOPR(oneempty, base[:]), false},
		{"wrong order", genWinnerOPR(wrongorder, base[:]), false},
		{"one short", genWinnerOPR(oneshort, base[:]), false},
		{"one long", genWinnerOPR(onelong, base[:]), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VerifyWinners(tt.args.opr, tt.args.winners); got != tt.want {
				t.Errorf("VerifyWinners() = %v, want %v", got, tt.want)
			}
		})
	}
}

// GradeBlock is put here to maintain the old method signature
// It is only used for v1 grading
//
// Old Description:
// 	GradeBlock takes all OPRs in a block, sorts them according to Difficulty, and grades the top 50.
// 	The top ten graded entries are considered the winners. Returns the top 50 sorted by grade, then the original list
// 	sorted by difficulty.
func GradeBlock(list []*OraclePriceRecord) (graded []*OraclePriceRecord, sorted []*OraclePriceRecord) {
	sort.SliceStable(list, func(i, j int) bool {
		return binary.BigEndian.Uint64(list[i].SelfReportedDifficulty) > binary.BigEndian.Uint64(list[j].SelfReportedDifficulty)
	})

	common.SetTestingVersion(1)
	graded = GradeMinimum(list, common.UnitTestNetwork, 0)
	return graded, list
}
