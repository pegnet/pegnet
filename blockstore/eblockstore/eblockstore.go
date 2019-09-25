package eblockstore

import (
	"bytes"
	"fmt"

	"github.com/pegnet/pegnet/database"
	"github.com/pegnet/pegnet/modules/grader"
	"github.com/syndtr/goleveldb/leveldb"
)

// EBlock Buckets
const (
	// Just marks where the eblock buckets start
	_ = iota + database.EBlockBucketStart
	BucketEBlockKeyMrIndexed
	BucketEBlockHeightIndexed
	BucketEBlockSequenceIndexed
	BucketEblockHead
)

// Some reserved keys
var (
	KeyEblockHead = []byte("eblockhead")
)

// EblockStore handles eblock indexing in a key/val db
type EblockStore struct {
	DB database.IDatabase

	// We keep the chainid for the eblock chain
	chainID []byte
}

func New(db database.IDatabase, chain []byte) *EblockStore {
	es := new(EblockStore)
	es.DB = db
	es.chainID = chain

	return es
}

type Eblock struct {
	KeyMr         []byte
	PreviousKeyMr []byte
	Height        int32
	Sequence      int32
}

func (s *EblockStore) FetchEblockByHeight(dbht int32) (*Eblock, error) {
	return s.eblock(BucketEBlockHeightIndexed, s.key(database.HeightToBytes(dbht)))
}

func (s *EblockStore) FetchEblockBySequence(seq int32) (*Eblock, error) {
	return s.eblock(BucketEBlockSequenceIndexed, s.key(database.HeightToBytes(seq)))
}

func (s *EblockStore) FetchEblockByKeyMr(keyMr []byte) (*Eblock, error) {
	return s.eblock(BucketEBlockKeyMrIndexed, s.key(keyMr))
}

func (s *EblockStore) FetchEblockHead() (*Eblock, error) {
	return s.eblock(BucketEblockHead, s.key(KeyEblockHead))
}

func (s *EblockStore) eblock(bucket database.Bucket, key []byte) (*Eblock, error) {
	data, err := s.DB.Get(bucket, key)
	if err != nil {
		return nil, err
	}

	var block Eblock
	err = database.Decode(&block, data)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

// WriteOPRBlockHead indicates a new synced eblock
//	Params:
//		eblockKeyMr
//		previousKeyMr
//		seq
//		dbht
//		gradedBlock			Not used, but provided to match the interface
func (s *EblockStore) WriteOPRBlockHead(eblockKeyMr, previousKeyMr []byte, seq int32, dbht int32, gradedBlock grader.GradedBlock) error {
	head, err := s.FetchEblockHead()
	if seq == 0 && err == leveldb.ErrNotFound {
		// This is an expected error
	} else if err != nil {
		// This is an error
		return err
	} else if head == nil {
		// If our head is nil, the head was not found
		return fmt.Errorf("no eblock head found, and seq is not 0")
	} else if bytes.Compare(previousKeyMr, head.KeyMr) != 0 || head.Sequence != seq-1 {
		return fmt.Errorf("attempt to sync the eblocks out of order")
	}

	// First we write indexes for:
	// 	keymr -> eblock
	//	height -> eblock
	// sequence -> eblock

	block := Eblock{
		KeyMr:         eblockKeyMr,
		PreviousKeyMr: previousKeyMr,
		Height:        dbht,
		Sequence:      seq,
	}

	data, err := database.Encode(&block)
	if err != nil {
		return err
	}

	err = s.DB.Put(BucketEBlockKeyMrIndexed, s.key(eblockKeyMr), data)
	if err != nil {
		return err
	}

	err = s.DB.Put(BucketEBlockHeightIndexed, s.key(database.HeightToBytes(dbht)), data)
	if err != nil {
		return err
	}

	err = s.DB.Put(BucketEBlockSequenceIndexed, s.key(database.HeightToBytes(seq)), data)
	if err != nil {
		return err
	}

	// Now write the head index
	err = s.DB.Put(BucketEblockHead, s.key(KeyEblockHead), data)
	if err != nil {
		return err
	}

	return nil
}

func (s *EblockStore) key(key []byte) []byte {
	return append(s.chainID, key...)
}

//// CurrentSyncedEblock indicates the last synced eblock
//CurrentSyncedEblock() (keymr []byte, height int32, sequence int)
//
//// PreviousWinners returns the previous winners found for a given height
//PreviousWinners(height int32) ([]string, error)
//
//// MarkInvalidEblock indicates the eblock does not have any winners, and therefore not a valid
//// opr block
//MarkInvalidEblock(eblockKeyMr []byte, seq int, dbht int32) error
//
//// WriteOPRBlockHead writes the next opr block in the sequence. If the added eblock is not the correct eblock
//// in the sequence, the write should fail.
//WriteOPRBlockHead(eblockKeyMr []byte, seq int, dbht int32, gradedBlock grader.GradedBlock) error
//
//Close() error
