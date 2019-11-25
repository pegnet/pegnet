package transactionid

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	DefaultPad = 1
)

// VerifyTransactionHash checks if a given hash or txid is valid.
// There are 2 types of transaction hashes:
//    - batches, 64 hex characters indicates a batch of transactions.
//	             All burns and coinbases are considered batches of length 1.
//    - txid, [TxIndex]-[BatchHash] indicates a single transaction in a batch.
//
// All hashes in pure hash format (64 hex characters) will return an index
// of -1, meaning the hash indicates a batch of transactions.
//
// All hashes in txid format, [TxIndex]-[BatchHash] will return an index
// number >= 0.
func VerifyTransactionHash(hash string) (index int, batchHash string, err error) {
	// First identify if the hash is a batch hash
	if len(hash) == 64 {
		_, err := hex.DecodeString(hash)
		if err != nil { // Not valid hex
			return -1, "", err
		}
		return -1, hash, nil
	}

	return SplitTxID(hash)
}

// SplitTxID splits a TxID into it's parts.
// TxID format : [TxIndex]-[BatchHash]
//				 1-c99dedea0e4e0c40118fe7e4d515b23cc0489269c8cef187b4f15a4ccbd880be
func SplitTxID(txid string) (index int, batchHash string, err error) {
	arr := strings.Split(txid, "-")
	if len(arr) != 2 {
		return -1, "", fmt.Errorf("txid does not match txid format, format: [TxIndex]-[EntryHash]")
	}

	txIndex, err := strconv.ParseInt(arr[0], 10, 32)
	if err != nil {
		return -1, "", fmt.Errorf("index must be a valid integer")
	}

	if len(arr[1]) != 64 {
		return -1, "", fmt.Errorf("hash must be 32 bytes (64 hex characters)")
	}

	// Verify the hash is valid hex
	// There might be a more efficient check, such as a regex string.
	_, err = hex.DecodeString(arr[1])
	if err != nil {
		return -1, "", fmt.Errorf("hash must be a valid hex string")
	}

	return int(txIndex), arr[1], nil
}

// FormatTxID constructs a txid from an entryhash and its index
func FormatTxID(index int, hash string) string {
	return FormatTxIDWithPad(DefaultPad, index, hash)
}

// FormatTxIDWithPad constructs a txid from an entryhash and its index.
// It will pad the index such that it is of at least 'pad' characters in lenght.
// pad = 2 -> 01-entryhash
// pad = 3 -> 001-entryhash
func FormatTxIDWithPad(pad, index int, hash string) string {
	// format is the "%0Nd-%s
	format := fmt.Sprintf("%%0%dd-%%s", pad)
	return fmt.Sprintf(format, index, hash)
}

// SortTxIDS will sort txids by entryhash, then by index
// TODO: This is probably not very efficient since we call "SplitTxID" each time
//		We should really cache that for larger sets
func SortTxIDS(txids []string) []string {
	sort.SliceStable(txids, func(i, j int) bool {
		idxI, hI, _ := SplitTxID(txids[i])
		idxJ, hJ, _ := SplitTxID(txids[j])
		if hI != hJ {
			return hI < hJ
		}
		return idxI < idxJ
	})
	return txids
}
