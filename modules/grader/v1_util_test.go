package grader

import "testing"

func TestV1Payout(t *testing.T) {
	tests := []struct {
		name string
		args int
		want int64
	}{
		{"negative index", -1, 0},
		{"first place", 0, 800e8},
		{"second place", 1, 600e8},
		{"third place", 2, 450e8},
		{"fourth place", 3, 450e8},
		{"fifth place", 4, 450e8},
		{"sixth place", 5, 450e8},
		{"seventh place", 6, 450e8},
		{"eighth place", 7, 450e8},
		{"ninth place", 8, 450e8},
		{"tenth place", 9, 450e8},
		{"eleventh place", 10, 0},
		{"twelfth place", 11, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := V1Payout(tt.args); got != tt.want {
				t.Errorf("V1Payout() = %v, want %v", got, tt.want)
			}

			// Also test the function on the grader
			g, _ := NewGrader(1, 0, nil)
			if got := g.Payout(tt.args); got != tt.want {
				t.Errorf("V1Payout() = %v, want %v", got, tt.want)
			}
		})
	}
}
