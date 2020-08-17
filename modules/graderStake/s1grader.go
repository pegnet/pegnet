package graderStake

var _ BlockGrader = (*S1BlockGrader)(nil)

// S1BlockGrader implements the S1 grading algorithm.
// Entries are encoded in Protobuf with 25 winners each block.
// Valid assets can be found in ´opr.V5Assets´
type S1BlockGrader struct {
	baseGrader
}

// Version 5
func (s1 *S1BlockGrader) Version() uint8 {
	return 5
}

// WinnerAmount is the number of SPRs that receive a payout
func (s1 *S1BlockGrader) WinnerAmount() int {
	return 25
}

// AddSPR verifies and adds a S1 SPR.
func (s1 *S1BlockGrader) AddSPR(entryhash []byte, extids [][]byte, content []byte) error {
	gspr, err := ValidateS1(entryhash, extids, s1.height, content)
	if err != nil {
		return err
	}
	s1.sprs = append(s1.sprs, gspr)
	return nil
}

// Grade the SPRs. The S1 algorithm works the following way:
//	1. Take the top 50 entries with the best proof of work
//	2. Remove top and low's 1% band from each of the 32 assets
//	3. Calculate the average of each of the 32 assets
//	4. Calculate the distance of each SPR to the average, where distance is the sum of quadratic differences
//	to the average of each asset. If an asset is within `band`% of the average, that asset's
//	distance is 0.
//	5. Throw out the SPR with the highest distance
//	6. Repeat 3-4 until there are only 25 SPRs left
//	7. Repeat 3 but this time don't apply the band and don't throw out SPRs, just reorder them
//	until you are left with one
func (s1 *S1BlockGrader) Grade() GradedBlock {
	return s1.GradeCustom(50)
}

// GradeCustom grades the block using a custom cutoff for the top X
func (s1 *S1BlockGrader) GradeCustom(cutoff int) GradedBlock {
	block := new(S1GradedBlock)
	block.cutoff = cutoff
	block.height = s1.height
	block.cloneSPRS(s1.sprs)
	block.filterDuplicates()
	block.grade()
	return block
}

// Payout returns the amount of Pegtoshi awarded to the SPR at the specified index
func (s1 *S1BlockGrader) Payout(index int) int64 {
	return S1Payout(index)
}
