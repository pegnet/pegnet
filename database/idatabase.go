// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package database

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
)

type Bucket int

const (
	INVALID Bucket = iota // Don't use zero, that's a good invalid value

	// The bucket indexed by height that has the raw oprblock data
	//	Key -> Height
	//	Value -> Graded opr list
	BUCKET_OPR_HEIGHT

	// These are unused
	BUCKET_OPR        // Mostly Valid (has the prior winners, has proper structure
	BUCKET_ALL_EB     // OPR chain Entry Blocks indexed by Directory Block Height
	BUCKET_VALID_EB   // OPR chain Entry Blocks that actually qualify to pay out mining fees and set asset prices
	BUCKET_VALID_OPRS // OPR Lists of valid OPRS, indexed by Directory Block Height, ordered as graded
	BUCKET_BALANCES   // PEG payout balances
)

type Iterator interface {
	First() bool
	Last() bool
	Next() bool
	Key() []byte
	Value() []byte
	Release()
}

type IDatabase interface {
	Open(pathname string) error
	Put(bucket Bucket, key []byte, Value []byte) error
	Get(bucket Bucket, key []byte) ([]byte, error)
	Delete(bucket Bucket, key []byte) error
	Close() error

	Iterate(bucket Bucket) Iterator
}

// Decode is a gob decode into the target object
func Decode(o interface{}, data []byte) error {
	var buf bytes.Buffer
	buf.Write(data)
	dec := gob.NewDecoder(&buf)
	err := dec.Decode(o)
	if err != nil {
		return err
	}
	return nil
}

// Encode is a gob encode
func Encode(o interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(o)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func HeightToBytes(v int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(v))
	return buf
}
