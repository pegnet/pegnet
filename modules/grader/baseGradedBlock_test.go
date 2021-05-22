package grader_test

import (
	"reflect"
	"testing"

	"github.com/pegnet/pegnet/modules/opr"

	. "github.com/pegnet/pegnet/modules/grader"
	"github.com/pegnet/pegnet/modules/testutils"
)

func init() {
	InitLX()
	testutils.SetTestLXR(LX)
}

// TestBaseGradedBlock_Invalid tests various invalid sets and checks the resulting block
// has nil winners and is obviously a no win block
func TestBaseGradedBlock_Invalid(t *testing.T) {
	t.Run("V1", func(t *testing.T) {
		testBaseGradedBlock_Invalid(t, 1)
	})
	t.Run("V2", func(t *testing.T) {
		testBaseGradedBlock_Invalid(t, 2)
	})
	t.Run("V3", func(t *testing.T) {
		testBaseGradedBlock_Invalid(t, 3)
	})
	t.Run("V4", func(t *testing.T) {
		testBaseGradedBlock_Invalid(t, 4)
	})
	t.Run("V5", func(t *testing.T) {
		testBaseGradedBlock_Invalid(t, 5)
	})
}

func testBaseGradedBlock_Invalid(t *testing.T, version uint8) {
	checkEmpty := func(g BlockGrader, reason string) {
		block := g.Grade()
		if block.Winners() != nil {
			t.Error(reason)
		}

		if block.Version() != version {
			t.Error("wrong version")
		}
	}

	dbht := int32(1)

	// Test the empty set
	t.Run("test the empty set", func(t *testing.T) {
		g, _ := NewGrader(version, dbht, nil)
		checkEmpty(g, "no oprs")
	})

	t.Run("test not enough winners", func(t *testing.T) {
		// Test an incomplete set
		winners := testutils.RandomWinners(version)
		g, _ := NewGrader(version, dbht, winners)
		for i := 0; i < testutils.WinnerAmt(version)-1; i++ {
			err := g.AddOPR(testutils.RandomOPRWithFields(version, dbht, winners))
			if err != nil {
				t.Error(err)
			}
		}

		if g.Count() != testutils.WinnerAmt(version)-1 {
			t.Errorf("exp %d count, found %d", testutils.WinnerAmt(version)-1, g.Count())
		}

		checkEmpty(g, "not enough oprs to have winners")
	})

	t.Run("test winner short hashes resorts to the previous", func(t *testing.T) {
		// Test an incomplete set
		winners := testutils.RandomWinners(version)
		g, _ := NewGrader(version, dbht, winners)
		block := g.Grade()
		if len(block.WinnersShortHashes()) != len(winners) {
			t.Error("short hashes did not grab the previous")
		} else {
			for i := range winners {
				if block.WinnersShortHashes()[i] != winners[i] {
					t.Error("wrong winner")
				}
			}
		}
	})

	t.Run("invalid oprs", func(t *testing.T) {
		winners := testutils.RandomWinners(version)
		g, _ := NewGrader(version, dbht, winners)
		// First check the random is valid
		if err := g.AddOPR(testutils.RandomOPRWithFieldsAndModify(version, dbht, winners, nil)); err != nil {
			t.Errorf("error should be nil: %s", err.Error())
		}

		// Test bad FA address (bad checksum)
		if err := g.AddOPR(testutils.RandomOPRWithFieldsAndModify(version, dbht, winners, func(o interface{}) {
			switch o.(type) {
			case *opr.V1Content:
				// V1 allows bad FA addresses.
			case *opr.V2Content:
				obj := o.(*opr.V2Content)
				obj.Address = "FA2FK18Hdr2SBzUXqtfEAbGaNJUdr7VQBNLgRK7JKnR8wLQzYwUa"
			default:
				panic(reflect.TypeOf(o))
			}
		})); err == nil && version != 1 {
			t.Errorf("[%d] expected an error for bad fa address", version)
		}

		// Test bad Identity
		if err := g.AddOPR(testutils.RandomOPRWithFieldsAndModify(version, dbht, winners, func(o interface{}) {
			switch o.(type) {
			case *opr.V1Content:
				// V1 allows bad identities.
			case *opr.V2Content:
				obj := o.(*opr.V2Content)
				obj.ID = "random-hyphen"
			default:
				panic(reflect.TypeOf(o))
			}
		})); err == nil && version != 1 {
			t.Errorf("[%d] expected an error for bad identity", version)
		}

		// Test 0 PEG value
		err := g.AddOPR(testutils.RandomOPRWithFieldsAndModify(version, dbht, winners, func(o interface{}) {
			switch o.(type) {
			case *opr.V1Content:
				// V1 allows bad identities.
			case *opr.V2Content:
				obj := o.(*opr.V2Content)
				obj.Assets[0] = 0
			default:
				panic(reflect.TypeOf(o))
			}
		}))
		switch version {
		case 1, 2:
			if err != nil {
				t.Errorf("[%d] expected no error for 0 peg value: %s", version, err.Error())
			}
		case 3, 4, 5:
			if err == nil || err.Error() != NewValidateError("assets must be greater than 0").Error() {
				t.Errorf("[%d] expected error for 0 peg value", version)
			}
		}

	})

}

// TestBaseGradedBlock_Valid tests various valid sets and checks the resulting block
// has some winners and a good block. It will also check the cutoff and number of graded
// for sets of varying amounts.
func TestBaseGradedBlock_Valid(t *testing.T) {
	t.Run("V1", func(t *testing.T) {
		testBaseGradedBlock_valid(t, 1)
	})
	t.Run("V2", func(t *testing.T) {
		testBaseGradedBlock_valid(t, 2)
	})
	t.Run("V3", func(t *testing.T) {
		testBaseGradedBlock_valid(t, 3)
	})
	t.Run("V4", func(t *testing.T) {
		testBaseGradedBlock_valid(t, 4)
	})
	t.Run("V5", func(t *testing.T) {
		testBaseGradedBlock_valid(t, 5)
	})
}

func testBaseGradedBlock_valid(t *testing.T, version uint8) {
	prevWinners := testutils.RandomWinners(version)
	dbht := int32(1)

	testAmt := func(amt int) {
		g, _ := NewGrader(version, dbht, prevWinners)
		for i := 0; i < amt; i++ {
			err := g.AddOPR(testutils.RandomOPRWithFields(version, dbht, prevWinners))
			if err != nil {
				t.Error(err)
			}
		}

		if g.Count() != amt {
			t.Errorf("exp %d count, found %d", amt, g.Count())
		}

		block := g.Grade()

		if block.Winners() == nil {
			t.Error("no winners found")
		}

		// This should never fail, but if it does, then the next checks will also fail.
		// Our test util and WinnerAmount on the interface should always match.
		if block.WinnerAmount() != testutils.WinnerAmt(version) {
			t.Errorf("exp winner amt %d, found %d", testutils.WinnerAmt(version), block.WinnerAmount())
		}

		if len(block.Winners()) != testutils.WinnerAmt(version) {
			t.Error("not the right number of winners")
		} else {
			for i, sh := range block.WinnersShortHashes() {
				if sh == prevWinners[i] {
					t.Error("shorthashes showing previous winners instead of new winners")
				}
			}
		}

		if block.Version() != version {
			t.Error("wrong version")
		}

		for i, winner := range block.Winners() {
			if winner.Payout() == 0 {
				t.Error("winners should have payouts")
			}

			if winner.Position() != i {
				t.Error("winner position is incorrect")
			}

			if len(winner.Shorthash()) != 16 {
				t.Error("shorthash is of wrong length")
			}
		}

		cutoff := 50
		if amt < 50 {
			cutoff = amt
		}

		if block.Cutoff() != cutoff {
			t.Errorf("exp cutoff of %d, found %d", cutoff, block.Cutoff())
		}

		if len(block.Graded()) != cutoff {
			t.Errorf("exp graded of %d, found %d", cutoff, len(block.Graded()))
		}

		if amt > 50 {
			// Test a custom graded
			block = g.GradeCustom(amt)
			if block.Cutoff() != amt {
				t.Errorf("exp cutoff of %d, found %d", amt, block.Cutoff())
			}

			if len(block.Graded()) != amt {
				t.Errorf("exp graded of %d, found %d", amt, len(block.Graded()))
			}
		}

	}

	t.Run("test just enough winners", func(t *testing.T) {
		testAmt(testutils.WinnerAmt(version))
	})

	t.Run("test 35 oprs", func(t *testing.T) {
		testAmt(35)
	})

	t.Run("test 50 oprs", func(t *testing.T) {
		testAmt(50)
	})

	t.Run("test 100 oprs", func(t *testing.T) {
		testAmt(100)
	})
}
