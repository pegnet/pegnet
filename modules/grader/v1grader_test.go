package grader

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

type TestBlock struct {
	Height          int32
	PreviousWinners []string
	Winners         []string
	Entries         []TestEntry
}
type TestEntry struct {
	Hash    string
	ExtIDs  [][]byte
	Content []byte
}

func LoadTestBlock(height int) *TestBlock {

	data, err := ioutil.ReadFile(fmt.Sprintf("testdata/%d.json", height))
	if err != nil {
		panic(err)
	}

	block := new(TestBlock)
	err = json.Unmarshal(data, block)
	if err != nil {
		panic(err)
	}
	return block
}

func GradeTestBlock(tb *TestBlock, version uint8) bool {
	fmt.Println("Grading eblock", tb.Height, "with previous winners", tb.PreviousWinners)
	g, err := NewGrader(version, tb.Height, tb.PreviousWinners)
	if err != nil {
		panic(err)
	}

	for _, entry := range tb.Entries {
		hexhash, err := hex.DecodeString(entry.Hash)
		if err != nil {
			panic(err)
		}
		err = g.AddOPR(hexhash, entry.ExtIDs, entry.Content)
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println(tb.Height, "Block contains", g.Count(), "entries")

	graded := g.Grade()

	winners := graded.WinnersShortHashes()

	if len(winners) != len(tb.Winners) {
		fmt.Println(tb.Height, "unexpected amount of winners")
		return false
	}
	for i := range winners {
		if winners[i] != tb.Winners[i] {
			fmt.Println(tb.Height, "winner mismatch. expected =", tb.Winners, ". got =", graded.WinnersShortHashes())
			return false
		}
	}
	return true
}

func TestV1BlockGrader_Grade(t *testing.T) {
	test := os.Getenv("LXRBITSIZE")
	if len(test) > 0 && test != "30" {
		// can't do 30-bit tests in travis
		return
	}

	if !GradeTestBlock(LoadTestBlock(206422), 1) {
		t.Errorf("Failed to validate genesis block")
	}

	if !GradeTestBlock(LoadTestBlock(209000), 1) {
		t.Errorf("Failed to validate block 209000")
	}
}
