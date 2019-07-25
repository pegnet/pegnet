package common

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/FactomProject/btcutil/base58"
)

// Address holds the raw RCD which lets us generate any desired address
type Address struct {
	RCD    []byte
	Prefix string
}

// ParseAddress takes any string and converts it to an Address if valid.
// The resulting Address will remember which asset it came from.
func ParseAddress(input string) (*Address, error) {
	if len(input) < 42 || len(input) > 56 {
		return nil, fmt.Errorf("input must be between 42 and 56 characters")
	}

	a := new(Address)

	split := strings.Split(input, "_")
	if len(split) == 2 {
		err := a.SetPrefix(split[0])
		if err != nil {
			return nil, err
		}
		split[0] = split[1]
	}

	if err := a.loadBase58(split[0]); err != nil {
		return nil, err
	}
	return a, nil
}

// SetPrefix verifies the asset exists and then assigns it to the Address.
// example: "pPNT" for the pegnet token or an empty string for FA addresses
func (a *Address) SetPrefix(prefix string) error {
	if len(prefix) == 0 {
		a.Prefix = ""
		return nil
	}
	if len(prefix) < 4 {
		return fmt.Errorf("invalid prefix length. must be t or p followed by the prefix name")
	}

	if !CheckPrefix(prefix) {
		return fmt.Errorf("%s is not a valid prefix", prefix)
	}
	a.Prefix = prefix
	return nil
}

// loadBase58 is a wrapper function to decode a base58 string and load that as raw
func (a *Address) loadBase58(b58 string) error {
	raw := base58.Decode(b58)

	if len(raw) == 0 {
		return fmt.Errorf("could not decode base58")
	}

	return a.loadRaw(raw)
}

// loadRaw takes raw bytes in the form of "RCD1|checksum"
func (a *Address) loadRaw(raw []byte) error {
	if len(raw) < 4 {
		return fmt.Errorf("raw data must contain at least a checksum")
	}
	a.RCD = raw[:len(raw)-4]
	raw = raw[len(raw)-4:]

	if len(a.RCD) == 34 && bytes.Compare(fcPubPrefix, a.RCD[0:2]) == 0 {
		a.RCD = a.RCD[2:]
	}

	cs := a.Checksum(a.Prefix)
	if bytes.Compare(raw, cs) != 0 {
		return fmt.Errorf("invalid checksum")
	}

	return nil
}

// Checksum calculates the 4 byte checksum for the specified asset
func (a *Address) Checksum(asset string) []byte {
	var base []byte
	if len(asset) > 0 {
		base = append([]byte(asset), '_')
	} else {
		base = fcPubPrefix
	}

	hash := sha256.Sum256(append(append(base, a.RCD...)))
	hash = sha256.Sum256(hash[:])
	return hash[:4]
}

// String returns the human readable address for the rcd and set asset
func (a *Address) String() string {
	return a.ToAsset(a.Prefix)
}

// ToAsset converts the address to any given asset.
// An empty string returns the FA address.
func (a *Address) ToAsset(asset string) string {
	if len(asset) == 0 {
		return a.FactomAddress()
	}

	checksum := a.Checksum(asset)
	data := append(a.RCD, checksum...)
	return asset + "_" + base58.Encode(data)
}

// FactomAddress turns the factom address representation
func (a *Address) FactomAddress() string {
	data := append(fcPubPrefix, a.RCD...)
	data = append(data, a.Checksum("")...)
	return base58.Encode(data)
}

// IsSame returns true if both the RCD and the prefix match
func (a *Address) IsSame(b *Address) bool {
	return a.Prefix == b.Prefix && bytes.Compare(a.RCD, b.RCD) == 0
}

// IsSameBase returns true if the base RCD matches
func (a *Address) IsSameBase(b *Address) bool {
	return bytes.Compare(a.RCD, b.RCD) == 0
}
