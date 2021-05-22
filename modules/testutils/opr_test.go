package testutils_test

import (
	"math/rand"
	"testing"

	lxr "github.com/pegnet/LXRHash"

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
	t.Run("V3", func(t *testing.T) {
		testRandomOPR(t, 3)
	})
	t.Run("V4", func(t *testing.T) {
		testRandomOPR(t, 4)
	})
	t.Run("V5", func(t *testing.T) {
		testRandomOPR(t, 5)
	})
	t.Run("Bad Version", func(t *testing.T) {
		a, b, c := RandomOPR(0)
		if a != nil || b != nil || c != nil {
			t.Error("expected all nils")
		}
	})

}

func testRandomOPR(t *testing.T, version uint8) {
	// Just test the random field setting has the right winners and parses
	testRandom := func(f func() (entryhash []byte, extids [][]byte, content []byte), expWinners bool) {
		for i := 0; i < 10; i++ {
			_, _, content := f()

			o, err := opr.Parse(content)
			if err != nil {
				t.Error(err)
			}

			for _, win := range o.GetPreviousWinners() {
				if len(win) != 16 && expWinners {
					t.Error("expected a winner")
				}

				if len(win) != 0 && !expWinners {
					t.Error("expected a winner")
				}
			}
		}
	}

	// Test winners are set
	testRandom(func() (entryhash []byte, extids [][]byte, content []byte) {
		dbht := rand.Int31()
		return RandomOPRWithRandomWinners(version, dbht)
	}, true)

	// Test winners are []string{"", "", ..., ""}
	testRandom(func() (entryhash []byte, extids [][]byte, content []byte) {
		dbht := rand.Int31()
		return RandomOPRWithHeight(version, dbht)
	}, false)

	testRandom(func() (entryhash []byte, extids [][]byte, content []byte) {
		return RandomOPR(version)
	}, false)
}

func TestFlipVersion(t *testing.T) {
	if FlipVersion(1) != 2 {
		t.Error("version 1 not flipped")
	}
	if FlipVersion(2) != 1 {
		t.Error("version 2 not flipped")
	}
}

func TestWinnerAmt(t *testing.T) {
	if WinnerAmt(1) != 10 {
		t.Error("version 1 not flipped")
	}
	if WinnerAmt(2) != 25 {
		t.Error("version 2 not flipped")
	}
}
