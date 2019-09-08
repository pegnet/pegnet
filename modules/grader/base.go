package grader

import (
	"encoding/binary"
	"sort"

	"github.com/pegnet/pegnet/modules/lxr30"
)

type baseGrader struct {
	oprs    []*GradingOPR
	winners []*GradingOPR
	graded  bool

	height int32

	prevWinners []string
}

func NewGrader(version int, height int32) Block {
	switch version {
	case 1:
		v1 := new(V1Block)
		v1.height = height
		return v1
	case 2:
		v2 := new(V2Block)
		v2.height = height
		return v2
	default:
		panic("invalid grader version")
	}
}

// Count will return the total number of OPRs properly added to this grading block. If the `AddOPR` returns
// true, that opr will be included in this count.
func (bg *baseGrader) Count() int {
	return len(bg.oprs)
}

// GetPreviousWinners returns the set of previous winners set by SetPreviousWinners
func (bg *baseGrader) GetPreviousWinners() []string {
	return bg.prevWinners
}

func (bg *baseGrader) Height() int32 {
	return bg.height
}

func (bg *baseGrader) filterDuplicates() {
	filtered := make([]*GradingOPR, 0)

	added := make(map[string]bool)
	for _, v := range bg.oprs {
		id := string(append(v.Nonce, v.OPRHash...))
		if !added[id] {
			filtered = append(filtered, v)
			added[id] = true
		}
	}

	bg.oprs = filtered
}

func (bg *baseGrader) sortByDifficulty(limit int) {
	sort.SliceStable(bg.oprs, func(i, j int) bool {
		return bg.oprs[i].SelfReportedDifficulty > bg.oprs[i].SelfReportedDifficulty
	})

	lx := lxr30.Init()

	topX := make([]*GradingOPR, 0)
	for _, o := range bg.oprs {
		hash := lx.Hash(append(o.OPRHash, o.Nonce...))
		diff := binary.BigEndian.Uint64(hash)

		if diff != o.SelfReportedDifficulty {
			continue
		}

		topX = append(topX, o)
	}

	bg.oprs = topX
}
