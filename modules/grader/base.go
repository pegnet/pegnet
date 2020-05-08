package grader

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	lxr "github.com/pegnet/LXRHash"
)

// LX holds an instance of lxrhash
var LX *lxr.LXRHash
var lxInitializer sync.Once

// The init function for LX is expensive. So we should explicitly call the init if we intend
// to use it. Make the init call idempotent
func InitLX() {
	lxInitializer.Do(func() {
		// This code will only be executed ONCE, no matter how often you call it
		if size, err := strconv.Atoi(os.Getenv("LXRBITSIZE")); err == nil && size >= 8 && size <= 30 {
			LX = lxr.Init(lxr.Seed, uint64(size), lxr.HashSize, lxr.Passes)
		} else {
			LX = lxr.Init(lxr.Seed, lxr.MapSizeBits, lxr.HashSize, lxr.Passes)
		}
	})
}

// baseGrader provides common functionality that is deemed useful in all versions
type baseGrader struct {
	oprs   []*GradingOPR
	height int32

	prevWinners []string

	count int
}

// NewGrader instantiates a IBlockGrader Grader for a specific version.
// Once set, the height and list of previous winners can't be changed.
func NewGrader(version uint8, height int32, previousWinners []string) (BlockGrader, error) {
	if height < 0 {
		return nil, fmt.Errorf("height must be > 0")
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
	case 3:
		if len(previousWinners) == 0 {
			previousWinners = make([]string, 25)
		} else if !verifyWinnerFormat(previousWinners, 25) {
			// V2 has 25 winners, we can enforce a 25 winner previous rule
			return nil, fmt.Errorf("invalid previous winners")
		}
		v3 := new(V3BlockGrader)
		v3.height = height
		v3.prevWinners = previousWinners
		return v3, nil
	case 4:
		if len(previousWinners) == 0 {
			previousWinners = make([]string, 25)
		} else if !verifyWinnerFormat(previousWinners, 25) {
			// V2 has 25 winners, we can enforce a 25 winner previous rule
			return nil, fmt.Errorf("invalid previous winners")
		}
		v4 := new(V4BlockGrader)
		v4.height = height
		v4.prevWinners = previousWinners
		return v4, nil
	case 5:
		if len(previousWinners) == 0 {
			previousWinners = make([]string, 25)
		} else if !verifyWinnerFormat(previousWinners, 25) {
			// V2 has 25 winners, we can enforce a 25 winner previous rule
			return nil, fmt.Errorf("invalid previous winners")
		}
		v5 := new(V5BlockGrader)
		v5.height = height
		v5.prevWinners = previousWinners
		return v5, nil
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

// Height returns the height the block grader is set to
func (bg *baseGrader) Height() int32 {
	return bg.height
}
