package graderStake

import "sort"

// S1GradedBlock is an spr set that has been graded. The set should be read only through it's interface
// implementation.
type S1GradedBlock struct {
	baseGradedBlock
}

// S1Band is the size of the band employed in the grading algorithm, specified as percentage
const S1Band = float64(0.01) // 1%

// Version returns the underlying grader's version
func (g *S1GradedBlock) Version() uint8 {
	return 5
}

// WinnerAmount returns the version specific amount of winners.
func (g *S1GradedBlock) WinnerAmount() int {
	return 25
}

// Winners returns the winning SPRs
func (g *S1GradedBlock) Winners() []*GradingSPR {
	if len(g.sprs) < 25 {
		return nil
	}

	return g.sprs[:25]
}

func (g *S1GradedBlock) grade() {
	if len(g.sprs) < 25 {
		return
	}

	if g.cutoff > len(g.sprs) {
		g.cutoff = len(g.sprs)
	}

	for i := g.cutoff; i >= 1; i-- {
		avg := averageS1(g.sprs[:i]) // same average as v1
		band := 0.0
		if i >= 25 {
			band = S1Band
		}
		for j := 0; j < i; j++ {
			gradeS1(avg, g.sprs[j], band)
		}
		// Because this process can scramble the sorted fields, we have to resort with each pass.
		sort.SliceStable(g.sprs[:i], func(i, j int) bool { return g.sprs[i].Grade < g.sprs[j].Grade })
	}

	for i := range g.sprs {
		g.sprs[i].position = i
		g.sprs[i].payout = S1Payout(i)
	}
}

// WinnersShortHashes returns the shorthashes of the winning SPRs.
func (g *S1GradedBlock) WinnersShortHashes() []string {
	return g.shorthashes
}
