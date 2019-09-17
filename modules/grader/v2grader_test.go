package grader

import (
	"os"
	"testing"
)

func TestV2BlockGrader_Grade(t *testing.T) {

	test := os.Getenv("LXRBITSIZE")
	if len(test) > 0 && test != "30" {
		// can't do 30-bit tests in travis
		return
	}

	if !GradeTestBlock(LoadTestBlock(210330), 2) {
		t.Errorf("Failed to validate v2 genesis block")
	}

	if !GradeTestBlock(LoadTestBlock(210419), 2) {
		t.Errorf("Failed to validate block 210419")
	}
}
