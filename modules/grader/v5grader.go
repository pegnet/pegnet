package grader

var _ BlockGrader = (*V3BlockGrader)(nil)

// V5BlockGrader implements the V2 grading algorithm but requires PEG to have a price.
type V5BlockGrader struct {
	// intentionally v2 and not v3.
	// v2 implements the correct grading for v3 and v5. The only
	// difference is validation rules and in v5, some additional currencies.
	V2BlockGrader
}

// Version 5
func (v5 *V5BlockGrader) Version() uint8 {
	return 5
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
func (v5 *V5BlockGrader) Grade() GradedBlock {
	return v5.GradeCustom(50)
}

// GradeCustom grades the block using a custom cutoff for the top X
func (v5 *V5BlockGrader) GradeCustom(cutoff int) GradedBlock {
	// Use the 2 Grading
	block := v5.V2BlockGrader.GradeCustom(cutoff)
	return &V5GradedBlock{block}
}
