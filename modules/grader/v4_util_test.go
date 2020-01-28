package grader_test

import (
	"testing"

	. "github.com/pegnet/pegnet/modules/grader"
	"github.com/pegnet/pegnet/modules/testutils"
)

func TestValidateV4(t *testing.T) {
	version, height := uint8(4), int32(0)

	winners := testutils.RandomWinners(version)
	ehash, extids, content := testutils.RandomOPRWithFields(version, height, winners)
	_, err := ValidateV4(ehash, extids, height, winners, content)
	if err != nil {
		t.Errorf("expected nil err, got: %s", err.Error())
	}

	// Test 10 previous winners is not enough
	ehash, extids, content = testutils.RandomOPRWithFields(version, height, winners[:10])
	_, err = ValidateV4(ehash, extids, height, winners[:10], content)
	if err == nil || err.Error() != "incorrect amount of previous winners" {
		t.Errorf("did not get expected error")
	}
}
