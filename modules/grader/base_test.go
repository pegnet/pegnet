package grader_test

import (
	"crypto/rand"
	"fmt"
	"testing"

	. "github.com/pegnet/pegnet/modules/grader"
)

func TestNewGrader(t *testing.T) {
	var winners []string
	for i := 0; i < 26; i++ {
		buf := make([]byte, 8)
		rand.Read(buf)
		winners = append(winners, fmt.Sprintf("%x", buf))
	}

	type args struct {
		version         uint8
		height          int32
		previousWinners []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"unsupported version", args{version: 0, height: 0, previousWinners: nil}, true},
		{"unsupported height", args{version: 0, height: -500, previousWinners: nil}, true},
		{"v1 correct", args{version: 1, height: 1, previousWinners: winners[:10]}, false},
		{"v1 empty winners", args{version: 1, height: 1, previousWinners: nil}, false},
		{"v1 too few winners", args{version: 1, height: 1, previousWinners: winners[:5]}, true},
		{"v1 too many winners", args{version: 1, height: 1, previousWinners: winners[:15]}, true},
		{"v1 non hex winner", args{version: 1, height: 1, previousWinners: append(winners[:9:9], "not a hex string")}, true},
		{"v1 hex too short winner", args{version: 1, height: 1, previousWinners: append(winners[:9:9], "ffffffffffffff")}, true},        // 14
		{"v1 badly formatted hex winner", args{version: 1, height: 1, previousWinners: append(winners[:9:9], "fffffffffffffff")}, true}, // 15
		{"v1 hex too long winner", args{version: 1, height: 1, previousWinners: append(winners[:9:9], "ffffffffffffffffff")}, true},     // 18
		{"v2 correct", args{version: 2, height: 1, previousWinners: winners[:10]}, false},
		{"v2 correct 25", args{version: 2, height: 1, previousWinners: winners[:25]}, false},
		{"v2 empty winners", args{version: 2, height: 1, previousWinners: nil}, false},
		{"v2 too few winners", args{version: 2, height: 1, previousWinners: winners[:5]}, true},
		{"v2 between 10 and 25 winners", args{version: 2, height: 1, previousWinners: winners[:15]}, true},
		{"v2 too many winners", args{version: 2, height: 1, previousWinners: winners[:26]}, true},
		{"v2 non hex winner 10", args{version: 2, height: 1, previousWinners: append(winners[:9:9], "not a hex string")}, true},
		{"v2 non hex winner 25", args{version: 2, height: 1, previousWinners: append(winners[:24:24], "not a hex string")}, true},
		{"v2 hex too short winner 10", args{version: 2, height: 1, previousWinners: append(winners[:9:9], "ffffffffffffff")}, true},          // 14
		{"v2 hex too short winner 25", args{version: 2, height: 1, previousWinners: append(winners[:24:24], "ffffffffffffff")}, true},        // 14
		{"v2 badly formatted hex winner 10", args{version: 2, height: 1, previousWinners: append(winners[:9:9], "fffffffffffffff")}, true},   // 15
		{"v2 badly formatted hex winner 25", args{version: 2, height: 1, previousWinners: append(winners[:24:24], "fffffffffffffff")}, true}, // 15
		{"v2 hex too long winner 10", args{version: 2, height: 1, previousWinners: append(winners[:9:9], "ffffffffffffffffff")}, true},       // 18
		{"v2 hex too long winner 25", args{version: 2, height: 1, previousWinners: append(winners[:24:24], "ffffffffffffffffff")}, true},     // 18

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGrader(tt.args.version, tt.args.height, tt.args.previousWinners)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGrader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.Version() != tt.args.version {
				t.Errorf("NewGrader() version mismatch got = %d, want = %d", got.Version(), tt.args.version)
			}
			if got.Height() != tt.args.height {
				t.Errorf("NewGrader() height mismatch got = %d, want = %d", got.Height(), tt.args.height)
			}

			prev := tt.args.previousWinners
			if len(prev) == 0 {
				if tt.args.version == 1 {
					prev = make([]string, 10)
				} else if tt.args.version == 2 {
					prev = make([]string, 25)
				}
			}
			if !compare(got.GetPreviousWinners(), prev) {
				t.Errorf("NewGrader() previous winners mismatch. got = %s, want = %s", got.GetPreviousWinners(), prev)
			}
		})
	}
}

func compare(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
