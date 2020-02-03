package grader

import (
	"os"
	"testing"
)

func TestV3BlockGrader_Grade(t *testing.T) {
	test := os.Getenv("LXRBITSIZE")
	if len(test) > 0 && test != "30" {
		// can't do 30-bit tests in travis
		return
	}

	if !GradeTestBlock(LoadTestBlock(222270), 3) {
		t.Errorf("Failed to validate v3 genesis block")
	}

	if !GradeTestBlock(LoadTestBlock(222271), 3) {
		t.Errorf("Failed to validate first real v3 test block")
	}

	if !GradeTestBlock(LoadTestBlock(229145), 3) {
		t.Errorf("Failed to validate v3 block 229145")
	}
}
