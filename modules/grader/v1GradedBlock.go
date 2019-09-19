package grader

import (
	"sort"
)

// V1GradedBlock is an opr set that has been graded. The set should be read only through it's interface
// implementation.
type V1GradedBlock struct {
	BaseGradedBlock
}

var _ GradedBlock = (*V1GradedBlock)(nil)

// Version returns the underlying grader's version
func (g *V1GradedBlock) Version() uint8 {
	return 1
}

// Winners returns the winning OPRs
func (g *V1GradedBlock) Winners() []*GradingOPR {
	if len(g.OPRs) < 10 {
		return nil
	}

	return g.OPRs[:10]
}

// WinnersShortHashes returns the ShortHashes of the winning OPRs.
func (g *V1GradedBlock) WinnersShortHashes() []string {
	return g.ShortHashes
}

// WinnerAmount is the number of OPRs that receive a RewardPayout
func (g *V1GradedBlock) WinnerAmount() int {
	return 10
}

func (g *V1GradedBlock) grade() {
	if len(g.OPRs) < 10 {
		return
	}

	if g.CutOff > len(g.OPRs) {
		g.CutOff = len(g.OPRs)
	}

	for i := g.CutOff; i >= 10; i-- {
		avg := averageV1(g.OPRs[:i])
		for j := 0; j < i; j++ {
			gradeV1(avg, g.OPRs[j])
		}
		// Because this process can scramble the sorted fields, we have to resort with each pass.
		sort.SliceStable(g.OPRs[:i], func(i, j int) bool { return g.OPRs[i].SelfReportedDifficulty > g.OPRs[j].SelfReportedDifficulty })
		sort.SliceStable(g.OPRs[:i], func(i, j int) bool { return g.OPRs[i].Grade < g.OPRs[j].Grade })
	}

	for i := range g.OPRs {
		g.OPRs[i].GradePosition = i
		g.OPRs[i].RewardPayout = V1Payout(i)
	}
}
