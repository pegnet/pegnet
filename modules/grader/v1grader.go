package grader

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/pegnet/pegnet/modules/opr"
)

var _ BlockGrader = (*V1BlockGrader)(nil)

// V1BlockGrader implements the first OPR that PegNet launched with.
// Entries are encoded in JSON with 10 winners each block.
// The list of assets can be found in `opr.V1Assets`
type V1BlockGrader struct {
	baseGrader
}

// Version 1
func (v1 *V1BlockGrader) Version() uint8 {
	return 1
}
func (v1 *V1BlockGrader) WinnerAmount() int {
	return 10
}

// AddOPR verifies and adds a V1 OPR.
func (v1 *V1BlockGrader) AddOPR(entryhash []byte, extids [][]byte, content []byte) error {
	if len(entryhash) != 32 {
		return fmt.Errorf("invalid entry hash length")
	}

	if len(extids) != 3 {
		return fmt.Errorf("invalid extid length. expected 3 got %d", len(extids))
	}

	if len(extids[1]) != 8 {
		return fmt.Errorf("self reported difficulty must be 8 bytes")
	}

	if len(extids[2]) != 1 || extids[2][0] != 1 {
		return fmt.Errorf("invalid version")
	}

	var dec *opr.JSONOPR
	err := json.Unmarshal(content, dec)
	if err != nil {
		return err
	}

	if dec.Dbht != v1.height {
		return fmt.Errorf("invalid height")
	}

	// verify assets
	for _, code := range opr.V1Assets {
		if v, ok := dec.Assets[code]; !ok {
			return fmt.Errorf("asset list is not correct")
		} else if code != "PNT" && v == 0 {
			return fmt.Errorf("all values other than PNT must be nonzero")
		}
	}

	if len(dec.WinPreviousOPR) != 10 {
		return fmt.Errorf("must have exactly 10 previous winning shorthashes")
	}

	if len(dec.WinPreviousOPR) != len(v1.prevWinners) {
		return fmt.Errorf("incorrect amount of previous winners")
	}
	for i := range dec.WinPreviousOPR {
		if dec.WinPreviousOPR[i] != v1.prevWinners[i] {
			return fmt.Errorf("incorrect set of previous winners")
		}
	}

	gopr := new(GradingOPR)
	gopr.EntryHash = entryhash
	gopr.Nonce = extids[0]
	gopr.SelfReportedDifficulty = binary.BigEndian.Uint64(extids[1])

	sha := sha256.Sum256(content)
	gopr.OPRHash = sha[:]

	gopr.OPR = dec

	v1.oprs = append(v1.oprs, gopr)
	return nil
}

// Grade the OPRs. The V1 algorithm works the following way:
// 	1. Take the top 50 entries with the best proof of work
// 	2. Calculate the average of each of the 32 assets
// 	3. Calculate the distance for each OPR, where the distance is the sum of the quadratic distances
// 	to the average of each asset
// 	4. Throw out the OPR with the highest distance
// 	5. Repeat 3-4 until there are only 10 OPRs left
func (v1 *V1BlockGrader) Grade() GradedBlock {
	return v1.GradeCustom(50)
}

// GradeCustom grades the block using a custom cutoff for the top X
func (v1 *V1BlockGrader) GradeCustom(cutoff int) GradedBlock {
	block := new(V1GradedBlock)
	block.cutoff = cutoff
	block.height = v1.height
	block.cloneOPRS(v1.oprs)
	block.filterDuplicates()
	block.sortByDifficulty(cutoff)
	block.grade()
	return block
}
