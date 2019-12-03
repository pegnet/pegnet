package transactionid_test

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/pegnet/pegnet/modules/transactionid"
)

func TestFormatTxID(t *testing.T) {
	assert := assert.New(t)

	// Ensure formats are valid
	t.Run("default", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			data := make([]byte, 32)
			rand.Read(data)
			idx, hash := rand.Intn(10000), hex.EncodeToString(data)
			txid := FormatTxID(idx, hash)

			fIdx, fHash, err := SplitTxID(txid)
			assert.NoError(err)
			assert.Equal(idx, fIdx)
			assert.Equal(hash, fHash)
		}
	})

	// Ensure formats are valid with different pads
	t.Run("random pad", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			data := make([]byte, 32)
			rand.Read(data)
			idx, hash := rand.Intn(10000), hex.EncodeToString(data)
			txid := FormatTxIDWithPad(rand.Intn(40), idx, hash)

			fIdx, fHash, err := SplitTxID(txid)
			assert.NoError(err)
			assert.Equal(idx, fIdx)
			assert.Equal(hash, fHash)
		}
	})
}

func TestVerifyTransactionHash(t *testing.T) {
	assert := assert.New(t)
	type TestVec struct {
		TxID string
		// If Error is set, an error is expected
		Error     string
		EntryHash string // If this is set, the entryhash and txindex are checked
		TxIndex   int
		Pad       int
	}
	vects := []TestVec{
		{TxID: "0-", Error: "tx has no entryhash"},
		{TxID: "0-aa", Error: "entryhash too short"},
		{TxID: "0-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a", Error: "entryhash too short"},
		{TxID: "0-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25aaaaa", Error: "entryhash too long"},
		{TxID: "0-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a5", Error: "entryhash odd character length"},
		{TxID: "-2-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", Error: "negative, and too many splits"},
		{TxID: "a2-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", Error: "txindex not a number"},
		{TxID: "179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", Error: "hash too short"},
		{TxID: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57aa", Error: "hash too long"},
		{TxID: "0-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a5X", Error: "hash must be a valid hex string"},

		// Valids
		{TxID: "0-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57",
			EntryHash: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", TxIndex: 0, Pad: 1},
		{TxID: "0010-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57",
			EntryHash: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", TxIndex: 10, Pad: 4},
		{TxID: "012-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57",
			EntryHash: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", TxIndex: 12, Pad: 3},
		{TxID: "9-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57",
			EntryHash: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", TxIndex: 9, Pad: 1},
		{TxID: "00000-17c05acb2fec5add1bfadc4c5d7fbd532a1a3fdad0b7b8dee97a544b4ab77396",
			EntryHash: "17c05acb2fec5add1bfadc4c5d7fbd532a1a3fdad0b7b8dee97a544b4ab77396", TxIndex: 0, Pad: 5},
		{TxID: "999999-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57"},

		// Test some under valued pads
		{TxID: "999999-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57",
			EntryHash: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", TxIndex: 999999, Pad: 0},
		{TxID: "12345-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57",
			EntryHash: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", TxIndex: 12345, Pad: 1},
		{TxID: "12-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57",
			EntryHash: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", TxIndex: 12, Pad: 1},

		// Test Batch Hashes
		{TxID: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57",
			EntryHash: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", TxIndex: -1},
	}

	for i := range vects {
		vec := vects[i]
		index, entryhash, err := VerifyTransactionHash(vec.TxID)
		if err != nil && vec.Error == "" {
			// Error not expected!
			t.Errorf("%d: should be good, found err: %s", i, err.Error())
		} else if err == nil && vec.Error != "" {
			t.Errorf("%d: found no error, expected one. %s", i, vec.Error)
		} else if err == nil && vec.Error == "" {
			if vec.EntryHash != "" && index != vec.TxIndex {
				t.Errorf("exp index of %d, found %d", vec.TxIndex, index)
			}
			if vec.EntryHash != "" && vec.EntryHash != entryhash {
				t.Errorf("exp ehash of %s, found %s", vec.EntryHash, entryhash)
			}

			// -1 tx idx means don't check the reformatting
			if vec.EntryHash != "" && vec.TxIndex != -1 {
				exp := FormatTxIDWithPad(vec.Pad, vec.TxIndex, vec.EntryHash)
				assert.Equal(exp, vec.TxID)
			}
		} else {
		}
	}
}

func TestSortTxIDS(t *testing.T) {
	check := func(list []string, expList []string) {
		for i := range list {
			if list[i] != expList[i] {
				t.Errorf("Sort failed")
			}
		}
	}

	t.Run("vector", func(t *testing.T) {
		list := []string{
			"10-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a00",
			"0-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a00",
			"999999-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a00",
			"999998-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a00",
			"000000-F000000000000000000000000000000000000000000000000000000000000000",
			"000000-0000000000000000000000000000000000000000000000000000000000000000",
			"000000-0000000000000000000000000000000000000000000000000000000000000004",
			"000000-0000000000000000000000000000000000000000000000000000000000000003",
		}

		expList := []string{
			"000000-0000000000000000000000000000000000000000000000000000000000000000",
			"000000-0000000000000000000000000000000000000000000000000000000000000003",
			"000000-0000000000000000000000000000000000000000000000000000000000000004",
			"0-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a00",
			"10-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a00",
			"999998-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a00",
			"999999-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a00",
			"000000-F000000000000000000000000000000000000000000000000000000000000000",
		}

		for i := 0; i < 100; i++ {
			rand.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
			list = SortTxIDS(list)
			check(list, expList)
		}

	})

	tID := func(i int) string {
		return fmt.Sprintf("0-%064d", i)
	}

	t.Run("set of 3", func(t *testing.T) {
		exp := []string{tID(0), tID(1), tID(2)}
		check(SortTxIDS([]string{tID(0), tID(2), tID(1)}), exp)
		check(SortTxIDS([]string{tID(1), tID(0), tID(2)}), exp)
		check(SortTxIDS([]string{tID(1), tID(2), tID(0)}), exp)
		check(SortTxIDS([]string{tID(2), tID(0), tID(1)}), exp)
		check(SortTxIDS([]string{tID(2), tID(1), tID(0)}), exp)
	})
}
