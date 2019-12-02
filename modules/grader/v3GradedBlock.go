package grader

import "sort"

// V3GradedBlock is an opr set that has been graded. The set should be read only through it's interface
// implementation.
type V3GradedBlock struct {
	baseGradedBlock
}

// V3Band is the size of the band employed in the grading algorithm, specified as percentage
const V3Band = float64(0.01) // 1%

// Version returns the underlying grader's version
func (g *V3GradedBlock) Version() uint8 {
	return 3
}

// WinnerAmount returns the version specific amount of winners.
func (g *V3GradedBlock) WinnerAmount() int {
	return 25
}

// Winners returns the winning OPRs
func (g *V3GradedBlock) Winners() []*GradingOPR {
	if len(g.oprs) < 25 {
		return nil
	}

	return g.oprs[:25]
}

func (g *V3GradedBlock) grade() {
	if len(g.oprs) < 25 {
		return
	}

	if g.cutoff > len(g.oprs) {
		g.cutoff = len(g.oprs)
	}

	for i := g.cutoff; i >= 1; i-- {
		avg := averageV1(g.oprs[:i]) // same average as v1
		band := 0.0
		if i >= 25 {
			band = V2Band
		}
		for j := 0; j < i; j++ {
			gradeV2(avg, g.oprs[j], band) // same grade as v2
		}
		// Because this process can scramble the sorted fields, we have to resort with each pass.
		sort.SliceStable(g.oprs[:i], func(i, j int) bool { return g.oprs[i].SelfReportedDifficulty > g.oprs[j].SelfReportedDifficulty })
		sort.SliceStable(g.oprs[:i], func(i, j int) bool { return g.oprs[i].Grade < g.oprs[j].Grade })
	}

	for i := range g.oprs {
		g.oprs[i].position = i
		g.oprs[i].payout = V3Payout(i)
	}
}

// WinnersShortHashes returns the shorthashes of the winning OPRs.
func (g *V3GradedBlock) WinnersShortHashes() []string {
	return g.shorthashes
}
