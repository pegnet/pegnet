package eblockstore

import (
	"github.com/pegnet/pegnet/database"
	"github.com/pegnet/pegnet/modules/grader"
)

// EblockStore handles eblock indexing in a key/val db
type EblockStore struct {
	DB database.IDatabase
}

// WriteOPRBlockHead indicates a new synced eblock
func (e *EblockStore) WriteOPRBlockHead(eblockKeyMr []byte, seq int, dbht int32, gradedBlock grader.GradedBlock) error {
	// First we write indexes for:
	// 	keymr -> eblock
	//	height -> eblock
	// sequence -> eblock

	return nil
}
