package database_test

import (
	"bytes"
	"math/rand"
	"reflect"
	"testing"

	"github.com/pegnet/pegnet/opr"

	. "github.com/pegnet/pegnet/database"
)

func TestDatabaseOverlay_Coding(t *testing.T) {

	for i := 0; i < 100; i++ {
		orig := RandomOPRBlock()
		data, err := Encode(orig)
		if err != nil {
			t.Error(err)
		}

		newO := new(opr.OPRBlockDatabaseObject)
		err = Decode(newO, data)
		if err != nil {
			t.Error(err)
		}

		// Check same going back in
		data2, err := Encode(newO)
		if err != nil {
			t.Error(err)
		}

		if bytes.Compare(data, data2) != 0 {
			t.Error("encode is not the same after a decode")
		}

		if !reflect.DeepEqual(orig, newO) {
			t.Error("new is not the same as the orig")
		}

	}

}

func RandomOPRBlock() *opr.OPRBlockDatabaseObject {
	o := new(opr.OPRBlockDatabaseObject)

	o.DblockHeight = rand.Int63n(1e6)
	o.GradedOprs = make([]*opr.OraclePriceRecord, rand.Intn(100)+1)
	for i := range o.GradedOprs {
		tmp := new(opr.OraclePriceRecord)
		tmp.OPRHash = make([]byte, 32)
		rand.Read(tmp.OPRHash)
		// TODO: Add more fields to this
		o.GradedOprs[i] = tmp
	}

	return o
}
