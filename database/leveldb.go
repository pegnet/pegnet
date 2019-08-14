// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package database

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var _ = leveldb.OpenFile

type Ldb struct {
	Pathname string
	DB       *leveldb.DB
	lock     sync.Mutex
}

// BuildKey()
// appends the bucket as a number to the front of the key, only using 7 bits
// and keeping the high bit set.  Much like a varint.  This does not work well
// for negative numbers or really big numbers.
//
// This usually this adds two bytes, the value and a zero.  The bucket is always
// separated from the key by one zero. We set the high order bit
func BuildKey(bucket Bucket, key []byte) (bkey []byte) {
	for {
		bkey = append(key, byte(bucket)|0x80)
		if bucket == 0 {
			return bkey
		}
		bucket = bucket >> 7
	}
}

// Open()
// Open the database.  The path to the database is int he Ldb struct.
func (db *Ldb) Open(pathname string) (err error) {
	db.lock.Lock() // make database access concurrent safe
	defer db.lock.Unlock()

	db.Pathname = pathname
	if db.DB == nil { // Don't try and open the Database if it is already open
		db.DB, err = leveldb.OpenFile(db.Pathname, nil) // call sets err, no special processing, just return
	}
	return
}

// Close()
// Close the database if it has not yet been closed.  We know this by
// checking that the DB field isn't nil.
func (db *Ldb) Close() (err error) {
	db.lock.Lock() // make database access concurrent safe
	defer db.lock.Unlock()

	if db.DB != nil {
		err = db.DB.Close()
	}
	db.DB = nil
	return nil
}

func (db *Ldb) Put(bucket Bucket, key []byte, Value []byte) error {
	bkey := BuildKey(bucket, key)

	db.lock.Lock() // make database access concurrent safe
	defer db.lock.Unlock()

	err := db.DB.Put(bkey, Value, nil)
	if err != nil {
		return err
	}

	return nil
}

func (db *Ldb) Get(bucket Bucket, key []byte) ([]byte, error) {
	bkey := BuildKey(bucket, key)

	db.lock.Lock() // make database access concurrent safe
	defer db.lock.Unlock()

	value, err := db.DB.Get(bkey, nil)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (db *Ldb) Delete(bucket Bucket, key []byte) error {
	bkey := BuildKey(bucket, key)

	db.lock.Lock() // make database access concurrent safe
	defer db.lock.Unlock()

	err := db.DB.Delete(bkey, nil)
	if err != nil {
		return err
	}

	return nil
}

// Iterate()
// Creates an iterator to iterate over the elements in a bucket.
// A particular iterator cannot be used in multiple processes, but
// multiple processes can have their own iterators without issue.
//
// Iterators must be released.
func (db *Ldb) Iterate(bucket Bucket) (iter Iterator) {
	bkey := BuildKey(bucket, []byte{})
	iter = db.DB.NewIterator(util.BytesPrefix(bkey), nil)
	return iter
}
