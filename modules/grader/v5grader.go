package grader

var _ BlockGrader = (*V5BlockGrader)(nil)

// V5BlockGrader implements the V5 grading algorithm.
// Entries are encoded in Protobuf with 25 winners each block.
// Valid assets can be found in ´opr.V5Assets´
type V5BlockGrader struct {
	baseGrader
}

// Version 5
func (v5 *V5BlockGrader) Version() uint8 {
	return 5
}

// WinnerAmount is the number of OPRs that receive a payout
func (v5 *V5BlockGrader) WinnerAmount() int {
	return 25
}

// AddOPR verifies and adds a V5 OPR.
func (v5 *V5BlockGrader) AddOPR(entryhash []byte, extids [][]byte, content []byte) error {
	gopr, err := ValidateV5(entryhash, extids, v5.height, v5.prevWinners, content)
	if err != nil {
		return err
	}
	v5.oprs = append(v5.oprs, gopr)
	return nil
}

// Grade the OPRs. The V5 algorithm works the following way:
//	1. Take the top 50 entries with the best proof of work
//	2. Remove top and low's 1% band from each of the 32 assets
//	3. Calculate the average of each of the 32 assets
//	4. Calculate the distance of each OPR to the average, where distance is the sum of quadratic differences
//	to the average of each asset. If an asset is within `band`% of the average, that asset's
//	distance is 0.
//	5. Throw out the OPR with the highest distance
//	6. Repeat 3-4 until there are only 25 OPRs left
//	7. Repeat 3 but this time don't apply the band and don't throw out OPRs, just reorder them
//	until you are left with one
func (v5 *V5BlockGrader) Grade() GradedBlock {
	return v5.GradeCustom(50)
}

// GradeCustom grades the block using a custom cutoff for the top X
func (v5 *V5BlockGrader) GradeCustom(cutoff int) GradedBlock {

	block := new(V5GradedBlock)
	block.cutoff = cutoff
	block.height = v5.height
	block.cloneOPRS(v5.oprs)
	block.filterDuplicates()
	block.sortByDifficulty(cutoff)
	block.grade()
	if len(block.oprs) < 25 {
		block.shorthashes = v5.prevWinners
	} else {
		block.createShortHashes(25)
	}
	return block
}

// Payout returns the amount of Pegtoshi awarded to the OPR at the specified index
func (v5 *V5BlockGrader) Payout(index int) int64 {
	return V5Payout(index)
}
