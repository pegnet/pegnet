package common_test

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"testing"

	. "github.com/pegnet/pegnet/common"
)

func TestAmounts(t *testing.T) {
	vectors := []struct {
		V int64
		S string
	}{ // TODO: Add vectors
		{1e8, "1"},
		{2e8, "2"},
		{2e8 + 2e7, "2.2"},
		{1, "0.00000001"},
		{0, "0"},
		{12345678, "0.12345678"},
	}

	for _, v := range vectors {
		// Test the expect
		vS := AmountToString(v.V)
		if vS != v.S {
			t.Errorf("[1] Exp %s, got %s", v.S, vS)
		}

		vV, err := StringToAmount(v.S)
		if err != nil {
			t.Errorf("[2] Exp %d, got %d", v.V, vV)
		}

		// Test the results
		if vS2 := AmountToString(vV); vS2 != vS {
			t.Errorf("[3] Exp %s, got %s", vS2, vS)
		}

		if vV2, _ := StringToAmount(vS); vV2 != vV {
			t.Errorf("[3] Exp %d, got %d", vV2, vV)
		}
	}
}

func TestAmountJsonMarshal(t *testing.T) {
	type TestStruct struct {
		Amt Amount
	}

	for i := uint64(0); i < 100000; i++ {
		ts := &TestStruct{Amount(rand.Int63())}
		d, err := json.Marshal(ts)
		if err != nil {
			t.Error(err)
		}

		t2 := new(TestStruct)
		err = json.Unmarshal(d, t2)
		if err != nil {
			t.Error(err)
		}

		if ts.Amt != t2.Amt {
			fmt.Println(string(d))
			t.Error("json failed")
			t.FailNow()
		}
	}

}

func TestFromFloat(t *testing.T) {
	for i := 0; i < 1000; i++ {
		// test the string
		f := rand.Float64()
		v := FloatToAmount(f)

		// truncate f, so it does not round
		f = math.Trunc(f*1e8) / 1e8

		fS := fmt.Sprintf("%.8f", f)
		fS = strings.TrimRight(fS, "0")
		vS := AmountToString(v)

		if fS != vS {
			t.Errorf("Exp %s, got %s", fS, vS)
		}
	}
}
