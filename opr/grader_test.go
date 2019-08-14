package opr_test

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/pegnet/pegnet/database"
	. "github.com/pegnet/pegnet/opr"
)

func TestOPRBlockStore(t *testing.T) {
	o := NewOPRBlockStore(database.NewMapDb())
	var err error
	for i := 0; i < 100; i++ {
		orig := RandomOPRBlock()

		err = o.WriteOPRBlock(orig.ToOPRBlock())
		if err != nil {
			t.Error(err)
		}

		newO, err := o.FetchOPRBlock(orig.DblockHeight)
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(orig.ToOPRBlock(), newO) {
			t.Error("new is not the same as the orig")
		}

	}

}

func RandomOPRBlock() *OPRBlockDatabaseObject {
	o := new(OPRBlockDatabaseObject)

	o.DblockHeight = rand.Int63n(1e6)
	o.GradedOprs = make([]*OraclePriceRecord, rand.Intn(100)+1)
	for i := range o.GradedOprs {
		tmp := new(OraclePriceRecord)
		tmp.OPRHash = make([]byte, 32)
		rand.Read(tmp.OPRHash)
		// TODO: Add more fields to this
		o.GradedOprs[i] = tmp
	}

	return o
}
