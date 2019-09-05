package opr

import (
	"crypto/sha256"
	"fmt"

	"github.com/pegnet/pegnet/modules/lxr30"
)

// The OPRFactory is for processing factom entries into OPRs.
// All functions are static

// ParseOPR attempts to process the OPR from a given factom entry. Assuming the extid's
// are proper, the version is taken from the entry itself. In the case the decoding fails
// the OPR returned is nil, and an error is returned.
//
// ParseOPR does not validate the OPR beyond just parsing
func ParseOPR(entryhash []byte, extids [][]byte, content []byte) (*OPR, error) {
	opr := new(OPR)

	// Can only have three ExtIDs which must be:
	//	[0] the nonce for the entry
	//	[1] Self reported difficulty
	//  [2] Version number
	if len(extids) != 3 {
		return nil, fmt.Errorf("found %d extids, must have 3", len(extids))
	}

	// Check each extid
	opr.Nonce = extids[0]

	if len(extids[1]) != 8 { // self reported difficulty must be a uint64
		return nil, fmt.Errorf("extid[1] is %d bytes, must be 8 byte", len(extids[1]))
	}
	opr.SelfReportedDifficulty = extids[1]

	// Need the version number
	if len(extids[2]) != 1 {
		return nil, fmt.Errorf("extid[2] is %d bytes, must be 1 byte", len(extids[2]))
	}
	opr.Version = extids[2][0]
	opr.EntryHash = entryhash

	err := opr.SafeUnmarshal(content)
	if err != nil {
		return nil, err
	}

	// Go ahead and compute the OPRHash, if we don't save the data, re-marshaling it is not deterministic for v1
	sha := sha256.Sum256(content)
	opr.OPRHash = sha[:] // Save the OPRHash

	// TODO: Save the content if we want to get the entry back out?

	return opr, nil
}

func ComputeDifficulty(oprhash, nonce []byte) (difficulty uint64) {
	no := append(oprhash, nonce...)
	h := lxr30.Hash(no)

	// The high eight bytes of the hash(hash(entry.Content) + nonce) is the difficulty.
	// Because we don't have a difficulty bar, we can define difficulty as the greatest
	// value, rather than the minimum value.  Our bar is the greatest difficulty found
	// within a 10 minute period.  We compute difficulty as Big Endian.
	for i := uint64(0); i < 8; i++ {
		difficulty = difficulty<<8 + uint64(h[i])
	}
	return difficulty
}
