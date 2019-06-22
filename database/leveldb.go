// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package database

import (
	"github.com/syndtr/goleveldb/leveldb"
)

var _ = leveldb.OpenFile

type Ldb struct {
	Pathname string
	DB       *leveldb.DB
}

// BuildKey()
// appends the bucket as a number to the front of the key
// usually this adds two bytes, the value and a zero.  The bucket is always
// separated from the key by one zero.
func BuildKey(bucket Bucket, key []byte) (bkey []byte) {
	for {
		bkey = append(bkey, byte(bucket))
		if bucket == 0 {
			return bkey
		}
		bucket = bucket >> 8
	}
	bkey = append(bkey, key...)
	return
}

// Open()
// Open the database.  The path to the database is int he Ldb struct.
func (db *Ldb) Open() (err error) {
	db.DB, err = leveldb.OpenFile(db.Pathname, nil)
	return nil
}

// Close()
// Close the database if it has not yet been closed.  We know this by
// checking that the DB field isn't nil.
func (db *Ldb) Close() (err error) {
	if db.DB != nil {
		db.DB.Close()
	}
	db.DB = nil
	return nil
}

func (db *Ldb) Put(bucket Bucket, key []byte, Value []byte) error {
	bkey := BuildKey(bucket, key)

	err := db.DB.Put(bkey, Value, nil)
	if err != nil {
		return err
	}

	return nil
}

func (db *Ldb) Get(bucket Bucket, key []byte) ([]byte, error) {
	bkey := BuildKey(bucket, key)

	value, err := db.DB.Get(bkey, nil)
	if err != nil {
		return err
	}

	return value, nil
}

func (db *Ldb) Delete(bucket Bucket, key []byte) error {
	bkey := BuildKey(bucket, key)

	err := db.DB.Delete(bkey, nil)
	if err != nil {
		return err
	}

	return nil
}

//// Get()
//// Get a value from a particular bucket
//// The contents of the returned slice should not be modified.
//func (db *Ldb) Get(bucket Bucket, key []byte,) ([]byte, error){
//
//	data, err := db.Get([]byte("key"), nil)
//	...
//	err = db.Put([]byte("key"), []byte("value"), nil)
//	...
//	err = db.Delete([]byte("key"), nil)
//	return nil,nil
//}
