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

	for i := 0; i < 100; i++ {
		h := rand.Int63()
		// Test an invalid
		err = o.WriteInvalidOPRBlock(h)
		if err != nil {
			t.Error(err)
		}

		block, err := o.FetchOPRBlock(h)
		if err != nil {
			t.Error(err)
		}
		if !block.EmptyOPRBlock {
			t.Error("Should be empty")
		}
	}
}
