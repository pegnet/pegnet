// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package database

type Bucket int

const (
	INVALID Bucket = iota
	OPR
	WINNERS
	BALANCES
)

type Database interface {
	Open(pathname string) error
	Put(bucket Bucket, key []byte, Value []byte) error
	Get(bucket Bucket, key []byte) ([]byte, error)
	Delete(bucket Bucket, key []byte) error
}
