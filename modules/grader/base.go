package grader

import (
	"fmt"
)

// baseGrader provides common functionality that is deemed useful in all versions
type baseGrader struct {
	oprs   []*GradingOPR
	height int32

	prevWinners []string

	count int
}

// NewGrader instantiates a IBlockGrader Grader for a specific version.
// Once set, the Height and list of previous winners can't be changed.
func NewGrader(version uint8, height int32, previousWinners []string) (BlockGrader, error) {
	if height < 0 {
		return nil, fmt.Errorf("Height must be > 0")
	}
	switch version {
	case 1:
		if len(previousWinners) == 0 {
			previousWinners = make([]string, 10)
		} else if !verifyWinnerFormat(previousWinners, 10) {
			return nil, fmt.Errorf("invalid previous winners")
		}
		v1 := new(V1BlockGrader)
		v1.height = height
		v1.prevWinners = previousWinners
		return v1, nil
	case 2:
		if len(previousWinners) == 0 {
			previousWinners = make([]string, 25)
		} else if !verifyWinnerFormat(previousWinners, 10) && !verifyWinnerFormat(previousWinners, 25) {
			return nil, fmt.Errorf("invalid previous winners")
		}
		v2 := new(V2BlockGrader)
		v2.height = height
		v2.prevWinners = previousWinners
		return v2, nil
	default:
		// most likely developer error or outdated package
		return nil, fmt.Errorf("unsupported version")
	}
}

// Count will return the total number of OPRs stored in the block.
// If the set has been graded, this number may be less than the amount of OPRs added
// due to duplicate filter and self reported difficulty checks
func (bg *baseGrader) Count() int {
	return len(bg.oprs)
}

// GetPreviousWinners returns the set of previous winners
func (bg *baseGrader) GetPreviousWinners() []string {
	return bg.prevWinners
}

// Height returns the Height the block grader is set to
func (bg *baseGrader) Height() int32 {
	return bg.height
}
