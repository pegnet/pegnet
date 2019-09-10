package grader

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/pegnet/pegnet/modules/opr"
)

var _ BlockGrader = (*V2BlockGrader)(nil)

// V2BlockGrader implements the V2 grading algorithm.
// Entries are encoded in Protobuf with 25 winners each block.
// Valid assets can be found in ´opr.V2Assets´
type V2BlockGrader struct {
	baseGrader
}

// Version 2
func (v2 *V2BlockGrader) Version() uint8 {
	return 2
}

// WinnerAmount is the number of OPRs that receive a payout
func (v2 *V2BlockGrader) WinnerAmount() int {
	return 25
}

// AddOPR verifies and adds a V2 OPR.
func (v2 *V2BlockGrader) AddOPR(entryhash []byte, extids [][]byte, content []byte) error {
	if len(entryhash) != 32 {
		return fmt.Errorf("invalid entry hash length")
	}

	if len(extids) != 3 {
		return fmt.Errorf("invalid extid length. expected 3 got %d", len(extids))
	}

	if len(extids[1]) != 8 {
		return fmt.Errorf("self reported difficulty must be 8 bytes")
	}

	if len(extids[2]) != 1 || extids[2][0] != 2 {
		return fmt.Errorf("invalid version")
	}

	var dec *opr.ProtoOPR
	err := dec.Unmarshal(content)
	if err != nil {
		// All errors are parse errors. We silence them here
		return err
	}

	if dec.Height != v2.height {
		return fmt.Errorf("invalid height")
	}

	// verify assets
	if len(dec.Assets) != len(opr.V2Assets) {
		return fmt.Errorf("invalid assets")
	}
	for i, val := range dec.Assets {
		if i > 0 && val == 0 {
			return fmt.Errorf("assets must be greater than 0")
		}
	}

	if len(dec.Winners) != 10 && len(dec.Winners) != 25 {
		return fmt.Errorf("must have exactly 10 or 25 previous winning shorthashes")
	}

	if len(dec.Winners) != len(v2.prevWinners) {
		return fmt.Errorf("incorrect amount of previous winners")
	}
	for i, w := range dec.GetPreviousWinners() {
		if w != v2.prevWinners[i] {
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

	v2.oprs = append(v2.oprs, gopr)
	return nil
}

// Grade the OPRs. The V2 algorithm works the following way:
// 	1. Take the top 50 entries with the best proof of work
// 	2. Calculate the average of each of the 32 assets
// 	3. Calculate the distance of each OPR to the average, where distance is the sum of quadratic differences
// 	to the average of each asset. If an asset is within `band`% of the average, that asset's
//	distance is 0.
// 	4. Throw out the OPR with the highest distance
// 	5. Repeat 3-4 until there are only 25 OPRs left
//	6. Repeat 3 but this time don't apply the band and don't throw out OPRs, just reorder them
//	until you are left with one
func (v2 *V2BlockGrader) Grade() GradedBlock {
	return v2.GradeCustom(50)
}

// GradeCustom grades the block using a custom cutoff for the top X
func (v2 *V2BlockGrader) GradeCustom(cutoff int) GradedBlock {

	block := new(V2GradedBlock)
	block.cutoff = cutoff
	block.height = v2.height
	block.cloneOPRS(v2.oprs)
	block.filterDuplicates()
	block.sortByDifficulty(cutoff)
	block.grade()

	return block
}
