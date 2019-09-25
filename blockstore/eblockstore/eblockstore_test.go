package eblockstore_test

import (
	"testing"

	"github.com/pegnet/pegnet/blockstore/testutils"

	. "github.com/pegnet/pegnet/blockstore/eblockstore"
	"github.com/pegnet/pegnet/database"
)

func TestEblockStore_WriteOPRBlockHead(t *testing.T) {
	t.Run("test sequential", func(t *testing.T) {
		m := database.NewMapDb()
		//m := new(database.Ldb)
		//_ = m.Open("tmp.db")

		s := New(m, make([]byte, 32))
		amt := 100

		for i := 0; i < amt; i++ {
			err := s.WriteEBlockHead(testutils.MakeEblock(i))
			if err != nil {
				t.Error(err)
			}
		}

		// Test retrieval
		for i := int32(0); i < int32(amt); i++ {
			testutils.CheckEblockF(func() (*EBlock, error) {
				return s.FetchEblockByHeight(i)
			}, i, t)

			testutils.CheckEblockF(func() (*EBlock, error) {
				return s.FetchEblockBySequence(i)
			}, i, t)

			testutils.CheckEblockF(func() (*EBlock, error) {
				return s.FetchEblockByKeyMr(testutils.IntToHash(int(i)))
			}, i, t)
		}

		testutils.CheckChain(t, s)
	})
}
