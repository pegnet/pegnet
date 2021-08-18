package graderDelegateStake

import (
	"fmt"
)

// baseDelegatedGrader provides common functionality that is deemed useful in all versions
type baseDelegatedGrader struct {
	sprs   []*GradingDelegatedSPR
	height int32

	prevWinners []string

	count int
}

// NewDelegatedGrader instantiates a IBlockGrader Grader for a specific version.
func NewDelegatedGrader(version uint8, height int32) (DelegateBlockGrader, error) {
	if height < 0 {
		return nil, fmt.Errorf("height must be > 0")
	}
	switch version {
	case 8:
		s4 := new(S4BlockGrader)
		s4.height = height
		return s4, nil
	default:
		// most likely developer error or outdated package
		return nil, fmt.Errorf("unsupported version")
	}
}

// Count will return the total number of SPRs stored in the block.
// If the set has been graded, this number may be less than the amount of SPRs added
// due to duplicate filter and self reported difficulty checks
func (bg *baseDelegatedGrader) Count() int {
	return len(bg.sprs)
}

// GetPreviousWinners returns the set of previous winners
func (bg *baseDelegatedGrader) GetPreviousWinners() []string {
	return bg.prevWinners
}

// Height returns the height the block grader is set to
func (bg *baseDelegatedGrader) Height() int32 {
	return bg.height
}
