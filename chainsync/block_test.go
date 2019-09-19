package chainsync_test

import (
	"encoding/gob"
	"testing"

	. "github.com/pegnet/pegnet/chainsync"
)

func TestOPRBlock(t *testing.T) {
	gob.Register(&OprBlock{})

	// tmp := new(OprBlock)
	// var buf bytes.Buffer
	// gob.NewEncoder(buf)
}
