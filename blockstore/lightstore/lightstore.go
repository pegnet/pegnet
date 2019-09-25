package lightstore

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb/errors"

	"github.com/pegnet/pegnet/blockstore/eblockstore"
	"github.com/pegnet/pegnet/database"
	"github.com/pegnet/pegnet/modules/grader"
)

// Lightstore Buckets
const (
	// Just marks where the eblock buckets start
	_ = iota + database.LightsStoreBucketStart
	BucketLightStoreHead
	BucketLightKeyMrIndexed
	BucketLightHeightIndexed

	BucketEBlockKeyMrIndexed
	BucketEBlockHeightIndexed
	BucketEBlockSequenceIndexed
	BucketEblockHead
)

// Some reserved keys
var (
	KeyLightStoreHead = []byte("lightstorehead")
)

// LightStore only stores the absolute bare minimum needed to support mining and constructing an opr.
type LightStore struct {
	DB database.IDatabase

	// EblockStore will contain the eblock chain that we are tracking
	EblockStore *eblockstore.EblockStore

	// We keep the chainid for the eblock chain
	chainID []byte
}

func New(db database.IDatabase, chain []byte) *LightStore {
	es := new(LightStore)
	es.DB = db
	es.chainID = chain
	es.EblockStore = eblockstore.New(db, chain)

	return es
}

type MinimumStoreSet struct {
	// The previous winning eblock might be more than 1 eblock behind. So this value can skip over invalid eblocks
	PreviousWinningEblockKeyMr []byte

	// Data about this eblock
	EblockKeyMr          []byte
	EBlock               *eblockstore.EBlock
	LastGradedIndex      int
	LastGradedDifficulty uint64 // We will mask off the first bit because sqlite can only store 63 bit uints

	// Data about this set
	WinnerShortHashes []string
	AssetPrices       map[string]uint64

	// Extra data nice to have for some indexing. Users tend to use heights
	Height       int32
	InvalidBlock bool
}

func (l *LightStore) WriteOPRBlockHead(eblockKeyMr, previousKeyMr []byte, seq int32, dbht int32, gradedBlock grader.GradedBlock) error {
	// each time we write a block, we want to update our eblock store for traversing later
	if err := l.EblockStore.WriteEBlockHead(eblockKeyMr, previousKeyMr, seq, dbht); err != nil {
		return err
	}

	// Now we need to write the minimum information that we will need for mining.
	//  - WinnerShortHashes
	//	- Winning opr prices (we might need to use old prices for some reason)
	//  - Last graded opr difficulty + index (incase we want to use dynamic difficulty targeting with older data)
	//	TODO: Store data to preserve the performance api?

	if len(gradedBlock.Graded()) < gradedBlock.Cutoff() {
		return fmt.Errorf("found cutoff of %d, but only %d graded oprs", gradedBlock.Cutoff(), len(gradedBlock.Graded()))
	}

	// Fetch the current head
	head, err := l.FetchLightStoreHead()
	if seq == 0 && err == errors.ErrNotFound {
		// This is an expected error, init the blank fields
		head = emptyHead()
	} else if err != nil {
		// This is an error
		return err
	} else if head == nil {
		// If our head is nil, the head was not found
		return fmt.Errorf("no lightstore head found, and seq is not 0")
	}

	min := &MinimumStoreSet{
		// This is to find the previous winners for the current block
		PreviousWinningEblockKeyMr: head.EblockKeyMr,

		// Set the new values
		EblockKeyMr:          eblockKeyMr,
		LastGradedIndex:      gradedBlock.Cutoff(),
		LastGradedDifficulty: gradedBlock.Graded()[gradedBlock.Cutoff()-1].SelfReportedDifficulty,
		WinnerShortHashes:    gradedBlock.WinnersShortHashes(),

		// Init the map, still need to populate it
		AssetPrices: make(map[string]uint64),
	}

	// Set the prices for this block
	quotes := gradedBlock.Winners()[0].OPR.GetOrderedAssetsUint()
	for _, q := range quotes {
		min.AssetPrices[q.Name] = q.Value
	}

	data, err := database.Encode(min)
	if err != nil {
		return err
	}

	// We want to update our LightStore synced head
	if err := l.DB.Put(BucketLightStoreHead, l.key(KeyLightStoreHead), data); err != nil {
		return err
	}

	// Store the standard indexes
	if err := l.storeStandard(min, data); err != nil {
		return err
	}

	return nil
}

// MarkInvalidEblock is when we encounter an invalid eblock for opr winners. We still want to write
// the eblock to maintain the eblock chain. It will also allow a user to query for an eblock, to realize it
// was invalid
func (l *LightStore) MarkInvalidEblock(eblockKeyMr []byte, previousKeyMr []byte, seq int32, dbht int32) error {
	if err := l.EblockStore.WriteEBlockHead(eblockKeyMr, previousKeyMr, seq, dbht); err != nil {
		return err
	}

	head, err := l.FetchLightStoreHead()
	if err != nil && err != errors.ErrNotFound {
		return err
	}

	if head == nil {
		head = emptyHead()
	}

	min := &MinimumStoreSet{
		// This is to find the previous winners for the current block
		PreviousWinningEblockKeyMr: head.EblockKeyMr,

		// Set the new values
		EblockKeyMr:  eblockKeyMr,
		Height:       dbht,
		InvalidBlock: true,
	}

	data, err := database.Encode(min)
	if err != nil {
		return err
	}

	if err := l.storeStandard(min, data); err != nil {
		return err
	}

	return nil
}

// storeStandard stores the standard indexing for a minimum set. We pass the min set and the data so
// we can access the fields, and not have to remarshal the data
func (l *LightStore) storeStandard(min *MinimumStoreSet, data []byte) error {
	err := l.DB.Put(BucketLightKeyMrIndexed, l.key(min.EblockKeyMr), data)
	if err != nil {
		return err
	}

	err = l.DB.Put(BucketLightHeightIndexed, l.key(database.HeightToBytes(min.Height)), data)
	if err != nil {
		return err
	}
	return nil
}

func (l *LightStore) FetchLightStoreHead() (*MinimumStoreSet, error) {
	return l.minimumSet(BucketLightStoreHead, l.key(KeyLightStoreHead))
}

func (l *LightStore) FetchMinimumSetByKeyMr(keymr []byte) (*MinimumStoreSet, error) {
	return l.minimumSet(BucketLightKeyMrIndexed, l.key(keymr))
}

func (l *LightStore) FetchMinimumSetByHeight(height int32) (*MinimumStoreSet, error) {
	return l.minimumSet(BucketLightHeightIndexed, l.key(database.HeightToBytes(height)))
}

func (l *LightStore) minimumSet(bucket database.Bucket, key []byte) (*MinimumStoreSet, error) {
	data, err := l.DB.Get(bucket, key)
	if err != nil {
		return nil, err
	}

	var set MinimumStoreSet
	err = database.Decode(&set, data)
	if err != nil {
		return nil, err
	}

	return &set, nil
}

func emptyHead() *MinimumStoreSet {
	return &MinimumStoreSet{
		EblockKeyMr:       make([]byte, 32),
		WinnerShortHashes: nil,
		AssetPrices:       nil,
	}
}

func (l *LightStore) key(key []byte) []byte {
	return append(l.chainID, key...)
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
