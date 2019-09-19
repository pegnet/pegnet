package blockstore

import "github.com/pegnet/pegnet/modules/grader"

// OPRBlockStore writes opr blocks to disk to speedup syncing
type OPRBlockStore interface {
	// CurrentSyncedEblock indicates the last synced eblock
	CurrentSyncedEblock() (keymr []byte, height int32, sequence int)

	// PreviousWinners returns the previous winners found for a given height
	PreviousWinners() []string

	// MarkInvalidEblock indicates the eblock does not have any winners, and therefore not a valid
	// opr block
	MarkInvalidEblock(eblockKeyMr []byte, seq int, dbht int32) error

	// WriteOPRBlockHead writes the next opr block in the sequence. If the added eblock is not the correct eblock
	// in the sequence, the write should fail.
	WriteOPRBlockHead(eblockKeyMr []byte, seq int, dbht int32, gradedBlock grader.GradedBlock) error

	Close() error
}
