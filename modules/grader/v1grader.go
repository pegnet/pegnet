package grader

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/pegnet/pegnet/modules/opr"
)

var _ IBlockGrader = (*V1BlockGrader)(nil)

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

// WinnerAmount is the number of OPRs that receive a payout
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
// 	3. Calculate the grade for each OPR, where the grade is the sum of the quadratic distances
// 	to the average of each asset
// 	4. Throw out the OPR with the highest grade
// 	5. Repeat 3-4 until there are only 10 OPRs left
func (v1 *V1BlockGrader) Grade() IGradedBlock {
	v1.filterDuplicates()
	return v1.grade()
}

func (v1 *V1BlockGrader) grade() IGradedBlock {
	// TODO: Currently 50 is the default and only option for the number of graded oprs,
	// 		but we should allow a caller to specify a higher number to grade
	set := v1.sortByDifficulty(50)

	if len(set) < 10 {
		return NewV1GradedBlock([]*GradingOPR{}, 50, v1.Height())
	}

	// If the set contains more than 50 sorted by POW, we only calculate the distance for the top 50
	max := 50
	if len(set) < 50 {
		max = len(set)
	}

	for i := max; i >= 10; i-- {
		avg := averageV1(set[:i])
		for j := 0; j < i; j++ {
			gradeV1(avg, set[j])
		}
		// Because this process can scramble the sorted fields, we have to resort with each pass.
		sort.SliceStable(set[:i], func(i, j int) bool { return set[i].SelfReportedDifficulty > v1.oprs[j].SelfReportedDifficulty })
		sort.SliceStable(set[:i], func(i, j int) bool { return set[i].Grade < set[j].Grade })
	}

	return NewV1GradedBlock(set, 50, v1.Height())
}
