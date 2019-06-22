package database

type Bucket int

const(
	INVALID Bucket = iota
	OPR
	WINNERS
	BALANCES
)


type Database interface {
	Open(pathname string ) error
	Put(bucket Bucket, key []byte, Value []byte) error
	Get(bucket Bucket, key []byte,) ([]byte, error)
	Delete(bucket Bucket, key []byte) error
}