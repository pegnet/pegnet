package factoidaddress

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

var FactoidAddressPrefix = [2]byte{0x5f, 0xb1}

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
	if bytes.Compare(prefix, FactoidAddressPrefix[:]) != 0 {
		return fmt.Errorf("address has wrong prefix")
	}

	checksum := data[len(data)-4:]
	sha := sha256.Sum256(data[:34])
	shad := sha256.Sum256(sha[:])

	if bytes.Compare(shad[:4], checksum) != 0 {
		return fmt.Errorf("checksum is not correct")
	}

	return nil
}

func Random() string {
	addr := make([]byte, 32)
	_, _ = rand.Read(addr)
	// Prepend prefix
	addr = append(FactoidAddressPrefix[:], addr...)

	// Shad
	sha := sha256.Sum256(append(addr))
	shad := sha256.Sum256(sha[:])

	return Base58Encode(append(addr, shad[:4]...))
}
