package chainsync

import (
	"encoding/gob"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"

	"github.com/pegnet/pegnet/modules/grader"
	"github.com/pegnet/pegnet/modules/opr"

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
	// Register all the types
	gob.Register(OprBlock{})
	gob.Register(&grader.V1GradedBlock{})
	gob.Register(&grader.V2GradedBlock{})
	gob.Register(&opr.V1Content{})
	gob.Register(&opr.V2Content{})

	o.DB = db

	return o
}
func (d *OPRBlockStore) Close() error {
	return d.DB.Close()
}

// WriteInvalidOPRBlock is for writing a dbht to disk that has an invalid oprblock. This could
// be because the oprblock does not have enough entries, or another error
func (d *OPRBlockStore) WriteInvalidOPRBlock(dbht int32) error {
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
	data, err := database.Encode(opr)
	if err != nil {
		return err
	}

	// The multiple indexes. First we write the OPR block to the first index by height. This is where
	// the raw data will live
	err = d.DB.Put(database.BUCKET_OPR_HEIGHT, database.HeightToBytes(opr.Dbht), data)
	if err != nil {
		return err
	}

	// TODO: Add more indexing if you need more

	return nil
}

// WriteOPRBlockHead will write the oprblock and also update the current head
func (d *OPRBlockStore) WriteOPRBlockHead(opr *OprBlock) error {
	if err := d.WriteOPRBlock(opr); err != nil {
		return err
	}

	data, err := database.Encode(opr)
	if err != nil {
		return err
	}

	// Grab the head so we can write the previous height to our previous bucket
	head, err := d.FetchOPRBlockHead()
	if err != nil && err != leveldb.ErrNotFound {
		return err
	}

	if err != leveldb.ErrNotFound && head != nil {
		// There is head to write, write it's height at our previous height bucket
		err = d.DB.Put(database.BUCKET_PREVIOUS_OPR_HEIGHT, database.HeightToBytes(opr.Dbht), database.HeightToBytes(head.Dbht))
		if err != nil {
			return err
		}
	}

	// Write the head
	if err := d.DB.Put(database.BUCKET_CURRENT_HEAD, database.RECORD_OPR_CHAIN_HEAD, data); err != nil {
		return err
	}
	var _ = data

	return nil
}

func (d *OPRBlockStore) FetchPreviousOPRHeight(height int32) (int32, error) {
	data, err := d.DB.Get(database.BUCKET_PREVIOUS_OPR_HEIGHT, database.HeightToBytes(height))
	if err != nil {
		return -1, err
	}

	p := database.BytesToHeight(data)
	if p == -1 {
		return -1, fmt.Errorf("height unable to be decoded from db")
	}
	return p, nil
}

func (d *OPRBlockStore) FetchOPRBlockHead() (*OprBlock, error) {
	obj := new(OprBlock)

	data, err := d.DB.Get(database.BUCKET_CURRENT_HEAD, database.RECORD_OPR_CHAIN_HEAD)
	if err != nil {
		return nil, err
	}

	err = database.Decode(obj, data)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (d *OPRBlockStore) FetchOPRBlock(height int32) (*OprBlock, error) {
	obj := new(OprBlock)

	data, err := d.DB.Get(database.BUCKET_OPR_HEIGHT, database.HeightToBytes(height))
	if err != nil {
		return nil, err
	}

	err = database.Decode(obj, data)
	if err != nil {
		return nil, err
	}

	return obj, nil
}
