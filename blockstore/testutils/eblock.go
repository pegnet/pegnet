package testutils

import (
	"bytes"
	"encoding/binary"
	"testing"

	. "github.com/pegnet/pegnet/blockstore/eblockstore"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

// CheckChain verifies the eblock chain is intact and correct
func CheckChain(t *testing.T, store *EblockStore) {
	// Fetch the head
	head, err := store.FetchEblockHead()
	if err != nil {
		t.Error(err)
	}

	next := head
	i := head.Sequence - 1

	for {
		if next.Sequence == 0 {
			next, err := next.Previous()
			if err != nil && err != errors.ErrNotFound {
				t.Error(err)
			}
			if next != nil {
				t.Error("should be nil")
			}
			break
		}

		next, err = next.Previous()
		if err != nil {
			t.Error(err)
		}

		CheckEblock(next, i, t)
		i--
	}

}

// CheckEblockF is just a helper function to accept a function that returns an eblock.
// Then it checks the results
func CheckEblockF(f func() (*EBlock, error), i int32, t *testing.T) {
	eblock, err := f()
	if err != nil {
		t.Error(err)
		return
	}
	CheckEblock(eblock, i, t)
}

// CheckEblock verifies the eblock found matches the eblock that would be made from
// the int i
func CheckEblock(eblock *EBlock, i int32, t *testing.T) {
	if eblock.Sequence != i {
		t.Errorf("seq exp %d, found %d", i, eblock.Sequence)
	}
	if eblock.Height != i {
		t.Errorf("seq exp %d, found %d", i, eblock.Sequence)
	}
	if bytes.Compare(eblock.KeyMr, IntToHash(int(i))) != 0 {
		t.Errorf("keymr exp %d, found %x", i, eblock.KeyMr[:4])
	}
	if i > 0 && bytes.Compare(eblock.PreviousKeyMr, IntToHash(int(i-1))) != 0 {
		t.Errorf("prevkeymr exp %d, found %x", i, eblock.PreviousKeyMr[:4])
	}
}



// MakeEblock generates an eblock for a given int. This means you can easily chain
// eblocks and know MakeEblock(i-1) is the previous to MakeEblock(i)
func MakeEblock(i int) (keymr []byte, prevkeymr []byte, seq int32, dbht int32) {
	keymr = IntToHash(i)
	if i != 0 {
		prevkeymr = IntToHash(i - 1)
	} else {
		prevkeymr = make([]byte, 32)
	}

	seq = int32(i)
	dbht = int32(i)

	return
}

func IntToHash(i int) []byte {
	hash := make([]byte, 32)
	h := make([]byte, 4)
	binary.BigEndian.PutUint32(h, uint32(i))
	copy(hash[:4], h)
	return hash
}
