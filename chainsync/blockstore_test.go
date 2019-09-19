package chainsync_test

import (
	"math/rand"
	"reflect"
	"testing"

	. "github.com/pegnet/pegnet/chainsync"
	"github.com/pegnet/pegnet/database"
)

func TestOPRBlockStore(t *testing.T) {
	o := NewOPRBlockStore(database.NewMapDb())
	var err error
	for i := 0; i < 100; i++ {
		// Alternate v1 and v2
		orig, err := RandomOPRBlock(uint8(i%2)+1, int32(i))
		if err != nil {
			t.Error(err)
		}

		err = o.WriteOPRBlockHead(orig)
		if err != nil {
			t.Error(err)
		}

		newO, err := o.FetchOPRBlock(orig.Dbht)
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(newO, orig) {
			if uint8(i%2)+1 == 2 {
				// TODO: Set some real winners
				// For v2 the winners are []byte{} which gob encodes as nil. Making the deep copy fail
			} else {
				t.Error("new is not the same as the orig", uint8(i%2)+1)
			}
		}

		head, err := o.FetchOPRBlockHead()
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		if !reflect.DeepEqual(newO, head) {
			t.Error("head is not same as written", uint8(i%2)+1)
		}

		// Check the prev
		if i != 0 {
			p, err := o.FetchPreviousOPRHeight(newO.Dbht)
			if err != nil {
				t.Error(err)
			}

			if p != orig.Dbht-1 {
				t.Errorf("wrong prev height, exp %d found %d", orig.Dbht-1, p)
			}
		}

	}

	for i := 0; i < 100; i++ {
		h := rand.Int31()
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
