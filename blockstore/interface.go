package blockstore

import "github.com/pegnet/pegnet/modules/grader"

// OPRBlockStore writes opr blocks to disk to speedup syncing
type OPRBlockStore interface {
	// CurrentSyncedEblock indicates the last synced eblock.
	// This is not the latest eblock with a winning set
	CurrentSyncedEblock() (keymr []byte, height int32, sequence int)

	// Returns the minimum set of information for the latest synced opr block to form the next one.
	CurrentSyncedOPRBlock() (keymr []byte, height int32, sequence int) // TODO: Change this up

	// PreviousWinners returns the current set of winners for the latest synced eblock
	// that has a valid winning set
	CurrentWinners() ([]string, error)

	// MarkInvalidEblock indicates the eblock does not have any winners, and therefore not a valid
	// opr block
	MarkInvalidEblock(eblockKeyMr []byte, previousKeyMr []byte, seq int32, dbht int32) error

	// WriteOPRBlockHead writes the next opr block in the sequence. If the added eblock is not the correct eblock
	// in the sequence, the write should fail.
	WriteOPRBlockHead(eblockKeyMr []byte, seq int, dbht int32, gradedBlock grader.GradedBlock) error

	Close() error
}
