package grader_test

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/pegnet/pegnet/modules/testutils"

	. "github.com/pegnet/pegnet/modules/grader"
)

func init() {
	InitLX()
	testutils.SetTestLXR(LX)
}

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

		{"v3 incorrect 10 (not allowed 10 prev winner)", args{version: 3, height: 1, previousWinners: winners[:10]}, true},
		{"v3 correct 25", args{version: 3, height: 1, previousWinners: winners[:25]}, false},
		{"v3 empty winners", args{version: 3, height: 1, previousWinners: nil}, false},
		{"v3 too few winners", args{version: 3, height: 1, previousWinners: winners[:5]}, true},
		{"v3 between 10 and 25 winners", args{version: 3, height: 1, previousWinners: winners[:15]}, true},
		{"v3 too many winners", args{version: 3, height: 1, previousWinners: winners[:26]}, true},
		{"v3 non hex winner 10", args{version: 3, height: 1, previousWinners: append(winners[:9:9], "not a hex string")}, true},
		{"v3 non hex winner 25", args{version: 3, height: 1, previousWinners: append(winners[:24:24], "not a hex string")}, true},
		{"v3 hex too short winner 10", args{version: 3, height: 1, previousWinners: append(winners[:9:9], "ffffffffffffff")}, true},
		{"v3 hex too short winner 25", args{version: 3, height: 1, previousWinners: append(winners[:24:24], "ffffffffffffff")}, true},
		{"v3 badly formatted hex winner 10", args{version: 3, height: 1, previousWinners: append(winners[:9:9], "fffffffffffffff")}, true},
		{"v3 badly formatted hex winner 25", args{version: 3, height: 1, previousWinners: append(winners[:24:24], "fffffffffffffff")}, true},
		{"v3 hex too long winner 10", args{version: 3, height: 1, previousWinners: append(winners[:9:9], "ffffffffffffffffff")}, true},
		{"v3 hex too long winner 25", args{version: 3, height: 1, previousWinners: append(winners[:24:24], "ffffffffffffffffff")}, true},

		{"v4 incorrect 10 (not allowed 10 prev winner)", args{version: 4, height: 1, previousWinners: winners[:10]}, true},
		{"v4 correct 25", args{version: 4, height: 1, previousWinners: winners[:25]}, false},
		{"v4 empty winners", args{version: 4, height: 1, previousWinners: nil}, false},
		{"v4 too few winners", args{version: 4, height: 1, previousWinners: winners[:5]}, true},
		{"v4 between 10 and 25 winners", args{version: 4, height: 1, previousWinners: winners[:15]}, true},
		{"v4 too many winners", args{version: 4, height: 1, previousWinners: winners[:26]}, true},
		{"v4 non hex winner 10", args{version: 4, height: 1, previousWinners: append(winners[:9:9], "not a hex string")}, true},
		{"v4 non hex winner 25", args{version: 4, height: 1, previousWinners: append(winners[:24:24], "not a hex string")}, true},
		{"v4 hex too short winner 10", args{version: 4, height: 1, previousWinners: append(winners[:9:9], "ffffffffffffff")}, true},
		{"v4 hex too short winner 25", args{version: 4, height: 1, previousWinners: append(winners[:24:24], "ffffffffffffff")}, true},
		{"v4 badly formatted hex winner 10", args{version: 4, height: 1, previousWinners: append(winners[:9:9], "fffffffffffffff")}, true},
		{"v4 badly formatted hex winner 25", args{version: 4, height: 1, previousWinners: append(winners[:24:24], "fffffffffffffff")}, true},
		{"v4 hex too long winner 10", args{version: 4, height: 1, previousWinners: append(winners[:9:9], "ffffffffffffffffff")}, true},
		{"v4 hex too long winner 25", args{version: 4, height: 1, previousWinners: append(winners[:24:24], "ffffffffffffffffff")}, true},

		{"v5 incorrect 10 (not allowed 10 prev winner)", args{version: 5, height: 1, previousWinners: winners[:10]}, true},
		{"v5 correct 25", args{version: 5, height: 1, previousWinners: winners[:25]}, false},
		{"v5 empty winners", args{version: 5, height: 1, previousWinners: nil}, false},
		{"v5 too few winners", args{version: 5, height: 1, previousWinners: winners[:5]}, true},
		{"v5 between 10 and 25 winners", args{version: 5, height: 1, previousWinners: winners[:15]}, true},
		{"v5 too many winners", args{version: 5, height: 1, previousWinners: winners[:26]}, true},
		{"v5 non hex winner 10", args{version: 5, height: 1, previousWinners: append(winners[:9:9], "not a hex string")}, true},
		{"v5 non hex winner 25", args{version: 5, height: 1, previousWinners: append(winners[:24:24], "not a hex string")}, true},
		{"v5 hex too short winner 10", args{version: 5, height: 1, previousWinners: append(winners[:9:9], "ffffffffffffff")}, true},
		{"v5 hex too short winner 25", args{version: 5, height: 1, previousWinners: append(winners[:24:24], "ffffffffffffff")}, true},
		{"v5 badly formatted hex winner 10", args{version: 5, height: 1, previousWinners: append(winners[:9:9], "fffffffffffffff")}, true},
		{"v5 badly formatted hex winner 25", args{version: 5, height: 1, previousWinners: append(winners[:24:24], "fffffffffffffff")}, true},
		{"v5 hex too long winner 10", args{version: 5, height: 1, previousWinners: append(winners[:9:9], "ffffffffffffffffff")}, true},
		{"v5 hex too long winner 25", args{version: 5, height: 1, previousWinners: append(winners[:24:24], "ffffffffffffffffff")}, true},
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
				} else if tt.args.version > 1 {
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

// TestBlockGrader_AddOPR will check the various fail conditions around a failed
// AddOPR and ensure the oprs are properly validated
func TestBlockGrader_AddOPR(t *testing.T) {
	t.Run("V1 AddOPR", func(t *testing.T) {
		testBlockGrader_AddOPR(t, 1)
	})
	t.Run("V2 AddOPR", func(t *testing.T) {
		testBlockGrader_AddOPR(t, 2)
	})
	t.Run("V3 AddOPR", func(t *testing.T) {
		testBlockGrader_AddOPR(t, 3)
	})
	t.Run("V4 AddOPR", func(t *testing.T) {
		testBlockGrader_AddOPR(t, 4)
	})
	t.Run("V5 AddOPR", func(t *testing.T) {
		testBlockGrader_AddOPR(t, 5)
	})
}

func testBlockGrader_AddOPR(t *testing.T, version uint8) {
	winners := testutils.RandomWinners(version)
	dbht := int32(100)
	g, err := NewGrader(version, dbht, winners)
	if err != nil {
		t.Error(err)
	}

	// Test various adds
	addOpr := func(f func() (entryhash []byte, extids [][]byte, content []byte), valid bool, reason string) {
		err := g.AddOPR(f())
		if valid && err != nil {
			t.Errorf("%s: %s", reason, err.Error())
		} else if !valid && err == nil {
			t.Error(reason)
		}
	}

	addOpr(func() (entryhash []byte, extids [][]byte, content []byte) {
		return testutils.RandomOPR(version)
	}, false, "totally random oprs are not valid")

	addOpr(func() (entryhash []byte, extids [][]byte, content []byte) {
		return testutils.RandomOPR(testutils.FlipVersion(version))
	}, false, "wrong version")

	addOpr(func() (entryhash []byte, extids [][]byte, content []byte) {
		return testutils.RandomOPRWithFields(testutils.FlipVersion(version), dbht, winners)
	}, false, "wrong version with right params")

	addOpr(func() (entryhash []byte, extids [][]byte, content []byte) {
		return testutils.RandomOPRWithHeight(version, dbht)
	}, false, "winners are not correct, blank")

	addOpr(func() (entryhash []byte, extids [][]byte, content []byte) {
		return testutils.RandomOPRWithRandomWinners(version, dbht)
	}, false, "winners are not correct, random")

	// Edge case testing
	addOpr(func() (entryhash []byte, extids [][]byte, content []byte) {
		return nil, nil, nil
	}, false, "all nil data")

	addOpr(func() (entryhash []byte, extids [][]byte, content []byte) {
		return []byte{}, [][]byte{}, []byte{}
	}, false, "all empty data")

	addOpr(func() (entryhash []byte, extids [][]byte, content []byte) {
		_, b, c := testutils.RandomOPRWithFields(version, dbht, winners)
		return []byte{}, b, c
	}, false, "no entryhash")

	addOpr(func() (entryhash []byte, extids [][]byte, content []byte) {
		a, _, c := testutils.RandomOPRWithFields(version, dbht, winners)
		return a, [][]byte{}, c
	}, false, "no extids")

	addOpr(func() (entryhash []byte, extids [][]byte, content []byte) {
		a, b, _ := testutils.RandomOPRWithFields(version, dbht, winners)
		return a, b, []byte{}
	}, false, "no content")

	addOpr(func() (entryhash []byte, extids [][]byte, content []byte) {
		a, b, c := testutils.RandomOPRWithFields(version, dbht, winners)
		b[1] = []byte{} // Self report difficulty
		return a, b, c
	}, false, "Self report difficulty is not 8 bytes")

	addOpr(func() (entryhash []byte, extids [][]byte, content []byte) {
		a, b, c := testutils.RandomOPRWithFields(version, dbht, winners)
		b[2] = []byte{} // Version is blank
		return a, b, c
	}, false, "bad version, too short")

	addOpr(func() (entryhash []byte, extids [][]byte, content []byte) {
		a, b, c := testutils.RandomOPRWithFields(version, dbht, winners)
		b[2] = []byte{version, 0x00} // Version is too long
		return a, b, c
	}, false, "bad version, too long")

	addOpr(func() (entryhash []byte, extids [][]byte, content []byte) {
		a, b, c := testutils.RandomOPRWithFields(version, dbht, winners[1:])
		return a, b, c
	}, false, "winners not enough")

	addOpr(func() (entryhash []byte, extids [][]byte, content []byte) {
		tmp := make([]string, len(winners))
		copy(tmp, winners)
		tmp[0] = winners[0][2:] // 1 byte short
		a, b, c := testutils.RandomOPRWithFields(version, dbht, tmp)
		return a, b, c
	}, false, "first winner not correct length")

	addOpr(func() (entryhash []byte, extids [][]byte, content []byte) {
		tmp := make([]string, len(winners))
		copy(tmp, winners)
		tmp[len(winners)-1] = winners[len(winners)-1][2:] // 1 byte short
		a, b, c := testutils.RandomOPRWithFields(version, dbht, tmp)
		return a, b, c
	}, false, "last winner not correct length")

	//
	// Things that can be added
	//
	addOpr(func() (entryhash []byte, extids [][]byte, content []byte) {
		a, b, c := testutils.RandomOPRWithFields(version, dbht, winners)
		b[0] = []byte{} // Nonce
		return a, b, c
	}, true, "blank nonce is ok")

	addOpr(func() (entryhash []byte, extids [][]byte, content []byte) {
		return testutils.RandomOPRWithFields(version, dbht, winners)
	}, true, "should be added")

	if g.Count() != 2 {
		t.Error("Exp only 1 added")
	}
}
