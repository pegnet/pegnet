package grader

import (
	"sort"
)

// V1GradedBlock is an opr set that has been graded. The set should be read only through it's interface
// implementation.
type V1GradedBlock struct {
	baseGradedBlock
}

var _ GradedBlock = (*V1GradedBlock)(nil)

// Version returns the underlying grader's version
func (g *V1GradedBlock) Version() uint8 {
	return 1
}

// Winners returns the winning OPRs
func (g *V1GradedBlock) Winners() []*GradingOPR {
	if len(g.oprs) < 10 {
		return nil
	}

	return g.oprs[:10]
}

// WinnersShortHashes returns the shorthashes of the winning OPRs.
func (g *V1GradedBlock) WinnersShortHashes() []string {
	return g.shorthashes
}

// WinnerAmount is the number of OPRs that receive a payout
func (g *V1GradedBlock) WinnerAmount() int {
	return 10
}

func (g *V1GradedBlock) grade() {
	if len(g.oprs) < 10 {
		return
	}

	if g.cutoff > len(g.oprs) {
		g.cutoff = len(g.oprs)
	}

	for i := g.cutoff; i >= 10; i-- {
		avg := averageV1(g.oprs[:i])
		for j := 0; j < i; j++ {
			gradeV1(avg, g.oprs[j])
		}
		// Because this process can scramble the sorted fields, we have to resort with each pass.
		sort.SliceStable(g.oprs[:i], func(i, j int) bool { return g.oprs[i].SelfReportedDifficulty > g.oprs[j].SelfReportedDifficulty })
		sort.SliceStable(g.oprs[:i], func(i, j int) bool { return g.oprs[i].Grade < g.oprs[j].Grade })
	}

	for i := range g.oprs {
		g.oprs[i].position = i
		g.oprs[i].payout = V1Payout(i)
	}
}
