package grader_test

import (
	"reflect"
	"testing"

	lxr "github.com/pegnet/LXRHash"

	. "github.com/pegnet/pegnet/modules/grader"
	"github.com/pegnet/pegnet/modules/testutils"
)

func TestGradingOPR_Clone(t *testing.T) {
	// TODO: Change the bitsize to something smaller. The grader needs to use the same one
	testutils.SetTestLXR(lxr.Init(lxr.Seed, 30, lxr.HashSize, lxr.Passes))

	t.Run("V1", func(t *testing.T) {
		testGradingOPR_Clone(t, 1)
	})
	t.Run("V2", func(t *testing.T) {
		testGradingOPR_Clone(t, 2)
	})
}

func testGradingOPR_Clone(t *testing.T, version uint8) {
	for i := 0; i < 5; i++ {

		dbht := int32(100)
		prevWinners := testutils.RandomWinners(version)
		g, _ := NewGrader(version, dbht, prevWinners)
		for i := 0; i < 50; i++ {
			err := g.AddOPR(testutils.RandomOPRWithFields(version, dbht, prevWinners))
			if err != nil {
				t.Error(err)
			}
		}

		block := g.Grade()
		for _, g := range block.Graded() {
			if !reflect.DeepEqual(g.Clone(), g) {
				t.Error("clone failed deep equal")
			}
		}
	}

}
