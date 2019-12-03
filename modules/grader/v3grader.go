package grader

var _ BlockGrader = (*V3BlockGrader)(nil)

// V3BlockGrader implements the V2 grading algorithm but requires PEG to have a price.
type V3BlockGrader struct {
	V2BlockGrader
}

// Version 3
func (v3 *V3BlockGrader) Version() uint8 {
	return 3
}

// AddOPR verifies and adds a V3 OPR.
func (v3 *V3BlockGrader) AddOPR(entryhash []byte, extids [][]byte, content []byte) error {
	gopr, err := ValidateV3(entryhash, extids, v3.height, v3.prevWinners, content)
	if err != nil {
		return err
	}

	v3.oprs = append(v3.oprs, gopr)
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
func (v3 *V3BlockGrader) Grade() GradedBlock {
	return v3.GradeCustom(50)
}

// GradeCustom grades the block using a custom cutoff for the top X
func (v3 *V3BlockGrader) GradeCustom(cutoff int) GradedBlock {
	// Use the 2 Grading
	block := v3.V2BlockGrader.GradeCustom(cutoff)
	return &V3GradedBlock{block}
}
