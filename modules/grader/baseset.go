package grader

import "fmt"

// baseGradedBlock is an opr set that has been graded. The set should be read only through it's interface
// implementation.
type baseGradedBlock struct {
	// GradedOprs should be sorted in their graded order
	GradedOprs []*GradingOPR

	// GradedUpTo indicates the max number of OPRs graded. By default, only the top 50
	// oprs are graded
	GradedUpTo int

	// WinnerAmount is the amount to expect in the winner set
	WinnerAmount int

	// Height indicates what height this graded set is for on the blockchain
	Height int32
}

func newBaseGradedBlock(graded []*GradingOPR, gradedTo int, height int32, winnerCount int) *baseGradedBlock {
	b := new(baseGradedBlock)
	b.GradedOprs = graded
	b.GradedUpTo = gradedTo
	b.Height = height
	b.WinnerAmount = winnerCount

	return b
}

// AmountToGrade returns the number of OPRs the grading algorithm attempted to use in the process.
func (s *baseGradedBlock) AmountToGrade() int {
	return s.GradedUpTo
}

// WinnersShortHashes returns the shorthashes of the winning OPRs.
// This result can be used to set the next block's previous winners.
// The amount varies between versions.
// If there are no winners, all strings will be empty.
func (s *baseGradedBlock) WinnersShortHashes() []string {
	wins := s.Winners()
	shortHashes := make([]string, s.WinnerAmount, s.WinnerAmount)
	for i := range wins {
		// TODO: Should the ShortHash() function be apart of the OPR?
		//		`opr.ShortHash()` ?
		shortHashes[i] = fmt.Sprintf("%x", wins[i].EntryHash[:8])
	}
	return shortHashes
}

// Winners returns the set of oprs rewarded with PEG
func (s *baseGradedBlock) Winners() []*GradingOPR {
	if len(s.GradedOprs) < s.WinnerAmount {
		return []*GradingOPR{}
	}
	return s.GradedOprs[:s.WinnerAmount]
}

// Graded returns the
func (s *baseGradedBlock) Graded() []*GradingOPR {
	return s.GradedOprs
}
