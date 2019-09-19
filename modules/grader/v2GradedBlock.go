package grader

import "sort"

// V2GradedBlock is an opr set that has been graded. The set should be read only through it's interface
// implementation.
type V2GradedBlock struct {
	BaseGradedBlock
}

// V2Band is the size of the band employed in the grading algorithm, specified as percentage
const V2Band = float64(0.01) // 1%

// Version returns the underlying grader's version
func (g *V2GradedBlock) Version() uint8 {
	return 2
}

// WinnerAmount returns the version specific amount of winners.
func (g *V2GradedBlock) WinnerAmount() int {
	return 25
}

// Winners returns the winning OPRs
func (g *V2GradedBlock) Winners() []*GradingOPR {
	if len(g.OPRs) < 25 {
		return nil
	}

	return g.OPRs[:25]
}

func (g *V2GradedBlock) grade() {
	if len(g.OPRs) < 25 {
		return
	}

	if g.CutOff > len(g.OPRs) {
		g.CutOff = len(g.OPRs)
	}

	for i := g.CutOff; i >= 1; i-- {
		avg := averageV1(g.OPRs[:i]) // same average as v1
		band := 0.0
		if i >= 25 {
			band = V2Band
		}
		for j := 0; j < i; j++ {
			gradeV2(avg, g.OPRs[j], band)
		}
		// Because this process can scramble the sorted fields, we have to resort with each pass.
		sort.SliceStable(g.OPRs[:i], func(i, j int) bool { return g.OPRs[i].SelfReportedDifficulty > g.OPRs[j].SelfReportedDifficulty })
		sort.SliceStable(g.OPRs[:i], func(i, j int) bool { return g.OPRs[i].Grade < g.OPRs[j].Grade })
	}

	for i := range g.OPRs {
		g.OPRs[i].GradePosition = i
		g.OPRs[i].RewardPayout = V2Payout(i)
	}
}

// WinnersShortHashes returns the ShortHashes of the winning OPRs.
func (g *V2GradedBlock) WinnersShortHashes() []string {
	return g.ShortHashes
}
