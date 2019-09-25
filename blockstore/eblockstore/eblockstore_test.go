package eblockstore_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/pegnet/pegnet/modules/grader"

	. "github.com/pegnet/pegnet/blockstore/eblockstore"
	"github.com/pegnet/pegnet/database"
)

func TestEblockStore_WriteOPRBlockHead(t *testing.T) {
	t.Run("test sequential", func(t *testing.T) {
		m := database.NewMapDb()
		s := New(m, make([]byte, 32))
		amt := 100

		for i := 0; i < amt; i++ {
			err := s.WriteOPRBlockHead(eblock(i))
			if err != nil {
				t.Error(err)
			}
		}

		// Test retrieval
		for i := int32(0); i < int32(amt); i++ {
			checkEblock(func() (*Eblock, error) {
				return s.FetchEblockByHeight(i)
			}, i, t)

			checkEblock(func() (*Eblock, error) {
				return s.FetchEblockBySequence(i)
			}, i, t)

			checkEblock(func() (*Eblock, error) {
				return s.FetchEblockByKeyMr(intToHash(int(i)))
			}, i, t)

		}
	})
}

func checkEblock(f func() (*Eblock, error), i int32, t *testing.T) {
	eblock, err := f()
	if err != nil {
		t.Error(err)
		return
	}

	if eblock.Sequence != i {
		t.Errorf("seq exp %d, found %d", i, eblock.Sequence)
	}
	if eblock.Height != i {
		t.Errorf("seq exp %d, found %d", i, eblock.Sequence)
	}
	if bytes.Compare(eblock.KeyMr, intToHash(int(i))) != 0 {
		t.Errorf("keymr exp %d, found %x", i, eblock.KeyMr[:4])
	}
	if i > 0 && bytes.Compare(eblock.PreviousKeyMr, intToHash(int(i-1))) != 0 {
		t.Errorf("prevkeymr exp %d, found %x", i, eblock.PreviousKeyMr[:4])
	}
}

func eblock(i int) (keymr []byte, prevkeymr []byte, seq int32, dbht int32, block grader.GradedBlock) {
	keymr = intToHash(i)
	if i != 0 {
		prevkeymr = intToHash(i - 1)
	} else {
		prevkeymr = make([]byte, 32)
	}

	seq = int32(i)
	dbht = int32(i)

	return
}

func intToHash(i int) []byte {
	hash := make([]byte, 32)
	h := make([]byte, 4)
	binary.BigEndian.PutUint32(h, uint32(i))
	copy(hash[:4], h)
	return hash
}
