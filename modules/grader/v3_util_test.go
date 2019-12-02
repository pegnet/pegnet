package grader

import (
	"fmt"
	"testing"
)

func TestV3Payout(t *testing.T) {
	for i := -10; i < 100; i++ {
		t.Run(fmt.Sprintf("V3 Payout %d", i), func(t *testing.T) {
			got := V3Payout(i)
			if i >= 0 && i < 25 && got != 200e8 {
				t.Errorf("V3Payout() = %v, want %v", got, 200e8)
			}
			if (i < 0 || i >= 25) && got != 0 {
				t.Errorf("V3Payout() = %v, want %v", got, 0)
			}

			// Also test the function on the grader
			g, _ := NewGrader(3, 0, nil)
			if graderGot := g.Payout(i); got != graderGot {
				t.Errorf("Grdader.V3Payout() = %v, V3Payout() = %v", graderGot, got)
			}
		})
	}
}
