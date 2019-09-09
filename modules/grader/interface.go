package grader

// Block allows you to grade a single block. Each version has its own struct, which must be instantiated
// with the height and set of previous winners.
type Block interface {
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

	// Grade grades the block. For more details, see each version's Grade() function.
	// If the result is empty, there are no winners.
	Grade() []*GradingOPR

	// WinnersShortHashes returns the shorthashes of the winning OPRs.
	// This result can be used to set the next block's previous winners.
	// The amount varies between versions.
	// If there are no winners, all strings will be empty.
	//
	// Grades the block if it has not yet been graded.
	WinnersShortHashes() []string

	// Graded returns the
	//
	// Grades the block if it has not yet been graded.
	Graded() []*GradingOPR

	// Count returns the number of valid OPRs, after duplicates are filtered out and
	// self-reported difficulty is validated
	Count() int

	// WinnerAmount returns the version specific amount of winners.
	// More efficient than len(WinnersShortHashes()).
	WinnerAmount() int
}
