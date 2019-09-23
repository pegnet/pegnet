package factoidaddress

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

var FactoidAddressPrefix = []byte{0x5f, 0xb1}

// Valid returns if the address is a valid factoid address
func Valid(addr string) error {
	data := Base58Decode(addr)
	if len(data) == 0 {
		return fmt.Errorf("address must be in base58")
	}

	// prefix + rcd + checksum
	if len(data) != 2+32+4 {
		return fmt.Errorf("address is of wrong length")
	}

	prefix := data[:2]
	if bytes.Compare(prefix, FactoidAddressPrefix) != 0 {
		return fmt.Errorf("address has wrong prefix")
	}

	// Checksum on the address
	checksum := data[len(data)-4:]
	// The expected checksum (we know the length is ok from the above check
	expected, err := Checksum(data[:34])
	if err != nil {
		return err // This error will never actually happen
	}

	if bytes.Compare(expected, checksum) != 0 {
		return fmt.Errorf("checksum is not correct")
	}

	return nil
}

// Checksum returns the 4 byte checksum trailing a factoid address. The input should be
// the 2 byte prefix + the rcd (34 bytes in total).
func Checksum(data []byte) ([]byte, error) {
	if len(data) != 34 {
		return nil, fmt.Errorf("expected 34 bytes, only found %d", len(data))
	}
	sha := sha256.Sum256(data)
	shad := sha256.Sum256(sha[:])

	return shad[:4], nil
}

// Encode takes a given rcd, and returns the factoid address human readable string
func Encode(rcd []byte) (string, error) {
	// Prepend prefix
	addr := append(FactoidAddressPrefix, rcd...)

	// Checksum
	checksum, err := Checksum(addr)
	if err != nil {
		return "", err
	}

	return Base58Encode(append(addr, checksum...)), nil
}

func Random() string {
	rcd := make([]byte, 32)
	_, _ = rand.Read(rcd)
	addr, _ := Encode(rcd)
	return addr
}
