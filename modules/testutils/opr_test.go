package testutils_test

import (
	"math/rand"
	"testing"

	"github.com/pegnet/LXRHash"
	"github.com/pegnet/pegnet/modules/opr"
	. "github.com/pegnet/pegnet/modules/testutils"
)

// Test the random oprs actually parse correctly
func TestRandomOPR(t *testing.T) {
	SetTestLXR(lxr.Init(lxr.Seed, 8, lxr.HashSize, lxr.Passes))

	t.Run("V1", func(t *testing.T) {
		testRandomOPR(t, 1)
	})
	t.Run("V2", func(t *testing.T) {
		testRandomOPR(t, 2)
	})
	t.Run("Bad Version", func(t *testing.T) {
		a, b, c := RandomOPR(0)
		if a != nil || b != nil || c != nil {
			t.Error("expected all nils")
		}
	})

}

func testRandomOPR(t *testing.T, version uint8) {
	for i := 0; i < 10; i++ {
		dbht := rand.Int31()
		_, _, content := RandomOPRWithFields(version, dbht)
		o, err := opr.Parse(content)
		if err != nil {
			t.Error(err)
		}

		PopulateRandomWinners(o)
		for _, win := range o.GetPreviousWinners() {
			if len(win) != 16 {
				t.Error("expected a winner")
			}
		}
	}
}
