// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package database

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
)

type Bucket int

// Bucket list
const (
	INVALID Bucket = iota // Don't use zero, that's a good invalid value

	// BUCKET_OPR_HEIGHT bucket indexed by height that has the raw oprblock data
	//	Key -> Height
	//	Value -> Graded opr list
	BUCKET_OPR_HEIGHT

	// BUCKET_PREVIOUS_OPR_HEIGHT bucket indexed by height that has the height of the previous opr block
	//	Key -> Height
	//	Value -> height of the prior oprblock
	BUCKET_PREVIOUS_OPR_HEIGHT

	// BUCKET_CURRENT_HEAD bucket stores the current head of the graded opr blocks
	BUCKET_CURRENT_HEAD

	// These are unused
	BUCKET_OPR        // Mostly Valid (has the prior winners, has proper structure
	BUCKET_ALL_EB     // OPR chain Entry Blocks indexed by Directory Block Height
	BUCKET_VALID_EB   // OPR chain Entry Blocks that actually qualify to pay out mining fees and set asset prices
	BUCKET_VALID_OPRS // OPR Lists of valid OPRS, indexed by Directory Block Height, ordered as graded
	BUCKET_BALANCES   // PEG payout balances
)

// Bucket Sets
//	Some buckets are described in their respective packages.
//	We keep them spaced so they have enough space to do what they need to do
const (
	EBlockBucketStart = 100
)

// Records are fixed records in a certain bucket
var (
	// BUCKET_CURRENT_HEAD Records
	RECORD_OPR_CHAIN_HEAD = []byte("OPRChainHead")
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

func HeightToBytes(v int32) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(v))
	return buf
}

func BytesToHeight(d []byte) int32 {
	if len(d) != 8 {
		return -1
	}
	return int32(binary.BigEndian.Uint64(d))
}
