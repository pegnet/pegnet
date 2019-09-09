package grader

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/pegnet/pegnet/modules/opr"
)

var _ Block = (*V1Block)(nil)

// V1Block implements the first OPR that PegNet launched with.
// Entries are encoded in JSON with 10 winners each block.
// The list of assets can be found in `opr.V1Assets`
type V1Block struct {
	baseGrader
}

// Version 1
func (v1 *V1Block) Version() uint8 {
	return 1
}

// WinnerAmount is the number of OPRs that receive a payout
func (v1 *V1Block) WinnerAmount() int {
	return 10
}

// AddOPR verifies and adds a V1 OPR.
func (v1 *V1Block) AddOPR(entryhash []byte, extids [][]byte, content []byte) error {
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

	v1.graded = false

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
func (v1 *V1Block) Grade() []*GradingOPR {
	if v1.graded {
		return v1.winners
	}

	v1.filterDuplicates()
	v1.sortByDifficulty(50)
	v1.grade()

	return nil
}

func (v1 *V1Block) grade() {
	v1.graded = true
	if len(v1.oprs) < 10 {
		v1.winners = nil
		return
	}

	for i := len(v1.oprs); i >= 10; i-- {
		avg := averageV1(v1.oprs[:i])
		for j := 0; j < i; j++ {
			gradeV1(avg, v1.oprs[j])
		}
		// Because this process can scramble the sorted fields, we have to resort with each pass.
		sort.SliceStable(v1.oprs[:i], func(i, j int) bool { return v1.oprs[i].SelfReportedDifficulty > v1.oprs[j].SelfReportedDifficulty })
		sort.SliceStable(v1.oprs[:i], func(i, j int) bool { return v1.oprs[i].Grade < v1.oprs[j].Grade })
	}

	v1.winners = v1.oprs[:10]
}

// Graded returns the 50 records that contain a grade.
// Can be used to determine the minimum difficulty to get into the top 50.
// If the set is not already graded, it will be graded.
func (v1 *V1Block) Graded() []*GradingOPR {
	if !v1.graded {
		v1.Grade()
	}

	if len(v1.oprs) >= 50 {
		return v1.oprs[:50]
	}
	return v1.oprs
}

// WinnersShortHashes always returns a slice with 10 elements that are either
// all empty (no winners) or contain the first 8 bytes of each winning opr's
// entryhash encoded as a hexadecimal string of length 16
func (v1 *V1Block) WinnersShortHashes() []string {
	winners := v1.Grade()

	short := make([]string, 10)
	if len(winners) == 10 {
		for i := range short {
			short = append(short, fmt.Sprintf("%x", winners[i].EntryHash[:8]))
		}
	}

	return short
}
