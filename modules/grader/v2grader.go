package grader

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"

	"github.com/pegnet/pegnet/modules/opr"
)

var _ IBlockGrader = (*V1BlockGrader)(nil)

// V2Band is the size of the band employed in the grading algorithm, specified as percentage
const V2Band = float64(0.01) // 1%

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
// 	3. Calculate the grade for each OPR, where the grade is the sum of the quadratic distances
// 	to the average of each asset. If an asset is within `band`% of the average, that asset's
//	grade is 0.
// 	4. Throw out the OPR with the highest grade
// 	5. Repeat 3-4 until there are only 25 OPRs left
//	6. Repeat 3 but this time don't apply the band and don't throw out OPRs, just reorder them
//	until you are left with one
func (v2 *V2BlockGrader) Grade() IGradedBlock {
	v2.filterDuplicates()
	v2.sortByDifficulty(50)
	v2.grade()

	return nil
}

func (v2 *V2BlockGrader) grade() IGradedBlock {
	// TODO: Currently 50 is the default and only option for the number of graded oprs,
	// 		but we should allow a caller to specify a higher number to grade
	set := v2.sortByDifficulty(50)

	if len(set) < 25 {
		return NewV2GradedBlock([]*GradingOPR{}, 50, v2.Height())
	}

	// If the set contains more than 50 sorted by POW, we only calculate the distance for the top 50
	max := 50
	if len(set) < 50 {
		max = len(set)
	}

	for i := max; i >= 1; i-- {
		avg := averageV1(set[:i]) // same average as v1
		band := 0.0
		if i >= 25 {
			band = V2Band
		}
		for j := 0; j < i; j++ {
			gradeV2(avg, set[j], band)
		}
		// Because this process can scramble the sorted fields, we have to resort with each pass.
		sort.SliceStable(set[:i], func(i, j int) bool { return set[i].SelfReportedDifficulty > set[j].SelfReportedDifficulty })
		sort.SliceStable(set[:i], func(i, j int) bool { return set[i].Grade < set[j].Grade })
	}

	return NewV2GradedBlock(set, 50, v2.Height())
}
