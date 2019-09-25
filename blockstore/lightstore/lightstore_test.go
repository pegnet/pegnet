package lightstore_test

import (
	"testing"

	"github.com/pegnet/pegnet/blockstore/eblockstore"

	. "github.com/pegnet/pegnet/blockstore/lightstore"
	"github.com/pegnet/pegnet/blockstore/testutils"
	"github.com/pegnet/pegnet/database"
)

// TestLightStore_MarkInvalidEblock mainly tests eblock ordering since they all are invalid
func TestLightStore_MarkInvalidEblock(t *testing.T) {
	t.Run("test sequential", func(t *testing.T) {
		m := database.NewMapDb()
		//m := new(database.Ldb)
		//_ = m.Open("tmp.db")

		s := New(m, make([]byte, 32))
		amt := 100

		for i := 0; i < amt; i++ {
			err := s.MarkInvalidEblock(testutils.MakeEblock(i))
			if err != nil {
				t.Error(err)
			}
		}

		// Test retrieval
		for i := int32(0); i < int32(amt); i++ {
			testutils.CheckEblockF(func() (*eblockstore.EBlock, error) {
				min, err := s.FetchMinimumSetByHeight(i)
				if err != nil {
					return nil, err
				}

				if !min.InvalidBlock {
					t.Error("Should be invalid")
				}

				return s.EblockStore.FetchEblockByKeyMr(min.EblockKeyMr)
			}, i, t)

			testutils.CheckEblockF(func() (*eblockstore.EBlock, error) {
				min, err := s.FetchMinimumSetByKeyMr(testutils.IntToHash(int(i)))
				if err != nil {
					return nil, err
				}

				if !min.InvalidBlock {
					t.Error("Should be invalid")
				}

				return s.EblockStore.FetchEblockByKeyMr(min.EblockKeyMr)
			}, i, t)
		}

		// Validate the eblock chain is in tact
		testutils.CheckChain(t, s.EblockStore)
	})
}

func TestLightStore_WriteOPRBlockHead(t *testing.T) {

}
