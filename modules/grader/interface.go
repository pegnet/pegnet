package grader

// BlockGrader allows you to grade a single block. Each version has its own struct, which must be instantiated
// with the height and set of previous winners.
type BlockGrader interface {
	// Height returns the height the block grader is set to
	Height() int32
	// Version returns the version of the underlying grader
	Version() uint8
	// GetPreviousWinners returns the set of previous winners the grader was initialized with
	GetPreviousWinners() []string

	// AddOPR adds an opr to the set to be graded. The content is decoded using the underlying version's format
	// and validated based on the specified height and set of previous winners.
	//
	// Returns an error if an entry could not be validated.
	AddOPR(entryhash []byte, extids [][]byte, content []byte) error

	// Grade grades the block using the default settings for that version.
	// For more details, see each version's Grade() function.
	// If the result is empty, there are no winners.
	Grade() GradedBlock

	// GradeCustom grades the OPRs using that version's algorithm and a custom cutoff for the top X
	GradeCustom(cutoff int) GradedBlock

	// Count returns the number of OPRs that have been added
	Count() int

	// Payout returns the amount of Pegtoshi awarded to the OPR at the specified index
	Payout(index int) int64
}

// GradedBlock is an immutable set of graded oprs
type GradedBlock interface {
	// WinnersShortHashes returns the shorthashes of the winning OPRs.
	// This result can be used to set the next block's previous winners.
	// The amount varies between versions.
	// If there are no winners in this block, the previous block's winners are used.
	WinnersShortHashes() []string

	// Winners returns the winning OPRs
	Winners() []*GradingOPR

	// Graded returns the top X OPRs that made it to the next stage of grading
	Graded() []*GradingOPR

	// Version returns the underlying grader's version
	Version() uint8

	// Cutoff returns the amount OPRs that made it to the second stage of grading
	Cutoff() int

	// Count returns the total count of OPRs that were used in this GradedBlock
	Count() int

	// WinnerAmount returns the version specific amount of winners.
	WinnerAmount() int
}
