package opr_test

import (
	"testing"

	. "github.com/pegnet/pegnet/opr"
)

func TestEntryBlockSync_AddNewHeadMarker(t *testing.T) {
	e := NewEntryBlockSync("test")
	if e.NextEBlock() != nil {
		t.Errorf("exp a nil")
	}

	e.AddNewHeadMarker(EntryBlockMarker{KeyMr: "test1"})
	if b := e.NextEBlock(); b == nil {
		t.Errorf("block is nil")
	} else {
		if b.KeyMr != "test1" {
			t.Errorf("not expected block")
		}
		e.BlockParsed(*b)
	}

	if e.NextEBlock() != nil {
		t.Errorf("exp a nil")
	}

}
