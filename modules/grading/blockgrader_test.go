package grading_test

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/pegnet/pegnet/modules/opr"

	"github.com/pegnet/pegnet/common"
	. "github.com/pegnet/pegnet/modules/grading"
)

// Just testing the very basic contructor rules
func TestBlockGrader_Constructor(t *testing.T) {
	t.Run("proper version", func(t *testing.T) {
		// Bad versions
		_, err := NewGradingBlock(0, 0, nil)
		if err == nil {
			t.Errorf("exp error for bad version")
		}
		for i := 3; i < 10; i++ {
			_, err := NewGradingBlock(0, 0, nil)
			if err == nil {
				t.Errorf("exp error for bad version")
			}
		}

		// Good versions
		_, err = NewGradingBlock(0, 1, nil)
		if err != nil {
			t.Errorf("exp no error for good version: %s", err.Error())
		}
		_, err = NewGradingBlock(0, 2, nil)
		if err != nil {
			t.Errorf("exp no error for good version: %s", err.Error())
		}
	})

	t.Run("proper prevWinners", func(t *testing.T) {
		gb := func(v uint8, amt, l int, experr bool, reason string) {
			pw := make([]string, amt, amt)
			for i := range pw {
				pw[i] = string(make([]byte, l, l))
				if l%2 == 0 {
					// Make it a hex string
					pw[i] = hex.EncodeToString(make([]byte, l/2, l/2))
				}
			}

			e := common.DetailErrorCallStack(fmt.Errorf(reason), 2)
			h := rand.Int31()

			b, err := NewGradingBlock(h, v, pw)
			if experr && err == nil {
				t.Error(e)
			}
			if !experr && err != nil {
				t.Error(e)
			}

			if !experr && err == nil {
				// Just verify the values were set
				if b.Height() != h {
					t.Error("height is not set, but should be")
				}

				if b.Version() != v {
					t.Error("version is not set, but should be")
				}

				if !reflect.DeepEqual(pw, b.GetPreviousWinners()) {
					t.Error("prev winners not set right")
				}
			}
		}

		// Bad winner sets
		//	Improper num of winners
		gb(1, 9, 16, true, "bad length winners")
		gb(1, 11, 16, true, "bad length winners")
		gb(1, 1, 16, true, "bad length winners")
		gb(1, 99, 16, true, "bad length winners")

		gb(2, 24, 16, true, "bad length winners")
		gb(2, 26, 16, true, "bad length winners")
		gb(2, 1, 16, true, "bad length winners")
		gb(2, 99, 16, true, "bad length winners")

		// 	Improper length strings
		gb(1, 10, 15, true, "bad length winners")
		gb(1, 10, 17, true, "bad length winners")
		gb(1, 10, 1, true, "bad length winners")

		gb(1, 25, 15, true, "bad length winner string")
		gb(1, 25, 17, true, "bad length winner string")
		gb(1, 25, 1, true, "bad length winner string")

		// Proper winner sets
		gb(1, 10, 16, false, "good winner set failed")
		gb(1, 10, 0, false, "good winner set failed")

		gb(2, 10, 16, false, "good winner set failed")
		gb(2, 25, 16, false, "good winner set failed")
		gb(2, 25, 0, false, "good winner set failed")

		// Special case
		//	This should fail, an empty set for v2 should always be 25
		gb(2, 10, 0, true, "empty set is 25 for v2")
	})
}

// Ensure the AddOPR has the very basic functionality of fliping the graded, and some basic parsing checks
func TestBlockGrader_AddOPR(t *testing.T) {
	b, err := NewGradingBlock(0, 2, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	st := func(entryhash []byte, extids [][]byte, content []byte, expadded bool, reason string) {
		e := common.DetailErrorCallStack(fmt.Errorf(reason), 2)
		prior := b.TotalOPRs()

		added := b.AddOPR(entryhash, extids, content)
		// Error checking
		if added != expadded {
			t.Error(e)
		}

		if expadded {
			if b.TotalOPRs() != prior+1 {
				t.Errorf("exp %d total oprs, found %d", prior+1, b.TotalOPRs())
			}
		}
	}

	template := opr.RandomOPR(2)

	zerohash := make([]byte, 32, 32)
	v2ExtIds := make([][]byte, 3, 3)
	v2ExtIds[1] = make([]byte, 8, 8)
	v2ExtIds[2] = []byte{2}
	content, err := template.SafeMarshal()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	var _ = zerohash

	// First try adding some improper
	st(zerohash, v2ExtIds, []byte{0x00}, false, "bad content")
	st(zerohash, v2ExtIds, append(content, 0x00), false, "bad content")
	st(zerohash, v2ExtIds[:2], content, false, "not enough extids")
	st(zerohash, append(v2ExtIds, []byte{}), content, false, "too many extids")

	// Now add a proper
	st(template.EntryHash, template.ExtIDs(), content, true, "proper opr rejected")

	// -------------
	// Check the grading switch
	// -------------
	gradedShouldBeUnset(t, b) // Should be ungraded

	// Set to graded
	b.Grade() // There will be no winners, as there is not enough oprs
	if !b.IsGraded() {
		t.Error("should be graded")
	}

	st(template.EntryHash, template.ExtIDs(), content, true, "proper opr rejected")
	gradedShouldBeUnset(t, b) // Should be ungraded
}

// TestBlockGrader_GradedBlocksCalls makes sure calls are blocked when the set is not graded
func TestBlockGrader_GradedBlocksCalls(t *testing.T) {
	b, err := NewGradingBlock(0, 2, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if b.IsGraded() {
		t.Error("should not be graded")
	}

	gradedShouldBeUnset(t, b)

}

// gradedShouldBeUnset just verifies the stat is `ungraded`
func gradedShouldBeUnset(t *testing.T, b *BlockGrader) {
	var err error
	if b.IsGraded() {
		t.Error("expected the blockgrader to be ungraded")
	}

	_, err = b.Winners()
	if err == nil {
		t.Error("Winners() : we expected an err since the set is ungraded")
	}

	_, err = b.Graded()
	if err == nil {
		t.Error("Winners() : we expected an err since the set is ungraded")
	}

	_, err = b.WinnersShortHashes()
	if err == nil {
		t.Error("Winners() : we expected an err since the set is ungraded")
	}
}
