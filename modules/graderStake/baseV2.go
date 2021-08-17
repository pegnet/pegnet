package graderStake

import "fmt"

// baseGraderV2 provides common functionality that is deemed useful in all versions
type baseGraderV2 struct {
	sprs   []*GradingSPRV2
	height int32

	prevWinners []string

	count int
}

// NewGrader instantiates a IBlockGrader Grader for a specific version.
func NewGraderV2(version uint8, height int32) (BlockGrader, error) {
	if height < 0 {
		return nil, fmt.Errorf("height must be > 0")
	}
	switch version {
	case 5:
		s1 := new(S1BlockGrader)
		s1.height = height
		return s1, nil
	case 6:
		s2 := new(S2BlockGrader)
		s2.height = height
		return s2, nil
	case 7:
		s3 := new(S3BlockGrader)
		s3.height = height
		return s3, nil
	default:
		// most likely developer error or outdated package
		return nil, fmt.Errorf("unsupported version")
	}
}

// Count will return the total number of SPRs stored in the block.
// If the set has been graded, this number may be less than the amount of SPRs added
// due to duplicate filter and self reported difficulty checks
func (bg *baseGraderV2) Count() int {
	return len(bg.sprs)
}

// GetPreviousWinners returns the set of previous winners
func (bg *baseGraderV2) GetPreviousWinners() []string {
	return bg.prevWinners
}

// Height returns the height the block grader is set to
func (bg *baseGraderV2) Height() int32 {
	return bg.height
}
