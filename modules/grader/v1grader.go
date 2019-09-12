package grader

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

// WinnerAmount is the number of OPRs that receive a payout
func (v1 *V1BlockGrader) WinnerAmount() int {
	return 10
}

// AddOPR verifies and adds a V1 OPR.
func (v1 *V1BlockGrader) AddOPR(entryhash []byte, extids [][]byte, content []byte) error {

	gopr, err := ValidateV1(entryhash, extids, v1.height, v1.prevWinners, content)
	if err != nil {
		return err
	}

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
	if len(block.oprs) < 10 {
		block.shorthashes = v1.prevWinners
	} else {
		block.createShortHashes(10)
	}
	return block
}

// Payout returns the amount of Pegtoshi awarded to the OPR at the specified index
func (v1 *V1BlockGrader) Payout(index int) int64 {
	return V1Payout(index)
}
