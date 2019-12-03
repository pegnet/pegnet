package grader_test

import (
	"testing"

	. "github.com/pegnet/pegnet/modules/grader"
	"github.com/pegnet/pegnet/modules/testutils"
)

func TestValidateV3(t *testing.T) {
	version, height := uint8(3), int32(0)

	winners := testutils.RandomWinners(version)
	ehash, extids, content := testutils.RandomOPRWithFields(version, height, winners)
	_, err := ValidateV3(ehash, extids, height, winners, content)
	if err != nil {
		t.Errorf("expected nil err, got: %s", err.Error())
	}

	// Test 10 previous winners is not enough
	ehash, extids, content = testutils.RandomOPRWithFields(version, height, winners[:10])
	_, err = ValidateV3(ehash, extids, height, winners[:10], content)
	if err == nil || err.Error() != "must have exactly 25 previous winning shorthashes" {
		t.Errorf("did not get expected error")
	}
}
