package grading_test

import (
	"encoding/hex"
	"fmt"
	"testing"

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

			_, err := NewGradingBlock(0, v, pw)
			if experr && err == nil {
				t.Error(e)
			}
			if !experr && err != nil {
				t.Error(e)
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
