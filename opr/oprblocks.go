package opr

import (
	"encoding/gob"
	"sort"

	"github.com/pegnet/pegnet/database"
)

type IOPRBlockStore interface {
	WriteInvalidOPRBlock(dbht int64) error
	WriteOPRBlock(opr *OprBlock) error
	FetchOPRBlock(height int64) (*OprBlock, error)
	Close() error
}

// OPRBlockStore is where we store the oprblock list
type OPRBlockStore struct {
	DB database.IDatabase
}

func NewOPRBlockStore(db database.IDatabase) *OPRBlockStore {
	o := new(OPRBlockStore)
	gob.Register(OPRBlockDatabaseObject{})
	o.DB = db

	return o
}
func (d *OPRBlockStore) Close() error {
	return d.DB.Close()
}

// WriteInvalidOPRBlock is for writing a dbht to disk that has an invalid oprblock. This could
// be because the oprblock does not have enough entries, or another error
func (d *OPRBlockStore) WriteInvalidOPRBlock(dbht int64) error {
	// An invalid oprblock is
	oprblock := new(OprBlock)
	oprblock.EmptyOPRBlock = true
	oprblock.Dbht = dbht

	return d.WriteOPRBlock(oprblock)
}

// WriteOPRBlock will write the top 50 graded oprs and write their corresponding indexes
func (d *OPRBlockStore) WriteOPRBlock(opr *OprBlock) error {
	// And opr block has both a graded component and sorted by difficulty component.
	// To save space, we can just keep the graded component, and resort the oprs when we pull them.

	obj := OPRBlockDatabaseObject{
		GradedOprs:         opr.GradedOPRs,
		DblockHeight:       opr.Dbht,
		EmptyOPRBlock:      opr.EmptyOPRBlock,
		TotalNumberRecords: opr.TotalNumberRecords,
	}

	data, err := database.Encode(obj)
	if err != nil {
		return err
	}

	// The multiple indexes. First we write the OPR block to the first index by height. This is where
	// the raw data will live
	err = d.DB.Put(database.BUCKET_OPR_HEIGHT, database.HeightToBytes(obj.DblockHeight), data)
	if err != nil {
		return err
	}

	// TODO: Add more indexing if you need more

	return nil
}

func (d *OPRBlockStore) FetchOPRBlock(height int64) (*OprBlock, error) {
	obj := new(OPRBlockDatabaseObject)

	data, err := d.DB.Get(database.BUCKET_OPR_HEIGHT, database.HeightToBytes(height))
	if err != nil {
		return nil, err
	}

	err = database.Decode(obj, data)
	if err != nil {
		return nil, err
	}

	return obj.ToOPRBlock(), nil
}

type OPRBlockDatabaseObject struct {
	GradedOprs         []*OraclePriceRecord
	DblockHeight       int64
	EmptyOPRBlock      bool
	TotalNumberRecords int
}

func (o *OPRBlockDatabaseObject) ToOPRBlock() *OprBlock {
	oprBlock := new(OprBlock)
	oprBlock.Dbht = o.DblockHeight
	oprBlock.GradedOPRs = o.GradedOprs
	oprBlock.OPRs = make([]*OraclePriceRecord, len(oprBlock.GradedOPRs))
	copy(oprBlock.OPRs, oprBlock.GradedOPRs)
	oprBlock.EmptyOPRBlock = o.EmptyOPRBlock

	sort.SliceStable(oprBlock.OPRs, func(i, j int) bool { return oprBlock.OPRs[i].Difficulty > oprBlock.OPRs[j].Difficulty })

	return oprBlock
}
