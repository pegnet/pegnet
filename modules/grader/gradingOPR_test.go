package grader_test

import (
	"reflect"
	"testing"

	. "github.com/pegnet/pegnet/modules/grader"
	"github.com/pegnet/pegnet/modules/testutils"
)

func init() {
	InitLX()
	testutils.SetTestLXR(LX)
}

func TestGradingOPR_Clone(t *testing.T) {
	t.Run("V1", func(t *testing.T) {
		testGradingOPR_Clone(t, 1)
	})
	t.Run("V2", func(t *testing.T) {
		testGradingOPR_Clone(t, 2)
	})
	t.Run("V3", func(t *testing.T) {
		testGradingOPR_Clone(t, 3)
	})
	t.Run("V4", func(t *testing.T) {
		testGradingOPR_Clone(t, 4)
	})
	t.Run("V5", func(t *testing.T) {
		testGradingOPR_Clone(t, 5)
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
