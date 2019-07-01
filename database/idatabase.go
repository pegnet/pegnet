// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package database

type Bucket int

const (
	INVALID           Bucket = iota // Don't use zero, that's a good invalid value
	BUCKET_OPR                      // Mostly Valid (has the prior winners, has proper structure
	BUCKET_ALL_EB                   // OPR chain Entry Blocks indexed by Directory Block Height
	BUCKET_VALID_EB                 // OPR chain Entry Blocks that actually qualify to pay out mining fees and set asset prices
	BUCKET_VALID_OPRS               // OPR Lists of valid OPRS, indexed by Directory Block Height, ordered as graded
	BUCKET_BALANCES                 // PNT payout balances
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
	Iterate(bucket Bucket) Iterator
}
