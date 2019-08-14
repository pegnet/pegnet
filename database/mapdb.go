// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package database

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb/errors"
)

type MapDb struct {
	data map[string][]byte
	lock sync.Mutex
}

func NewMapDb() *MapDb {
	m := new(MapDb)
	m.data = make(map[string][]byte)

	return m
}

func (db *MapDb) Open(pathname string) (err error) {
	db.data = make(map[string][]byte)
	return
}

func (db *MapDb) Close() (err error) {
	return nil
}

func (db *MapDb) Put(bucket Bucket, key []byte, Value []byte) error {
	bkey := BuildKey(bucket, key)

	db.lock.Lock() // make database access concurrent safe
	defer db.lock.Unlock()
	db.data[string(bkey)] = Value

	return nil
}

func (db *MapDb) Get(bucket Bucket, key []byte) ([]byte, error) {
	bkey := BuildKey(bucket, key)

	db.lock.Lock() // make database access concurrent safe
	defer db.lock.Unlock()

	v, ok := db.data[string(bkey)]
	if !ok {
		return nil, errors.ErrNotFound
	}

	return v, nil
}

func (db *MapDb) Delete(bucket Bucket, key []byte) error {
	bkey := BuildKey(bucket, key)

	db.lock.Lock() // make database access concurrent safe
	defer db.lock.Unlock()

	delete(db.data, string(bkey))
	return nil
}

// TODO: implement this for a map? Are we going to use this?
func (db *MapDb) Iterate(bucket Bucket) (iter Iterator) {
	return nil
}
