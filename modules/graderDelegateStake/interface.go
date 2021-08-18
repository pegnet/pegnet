package graderDelegateStake

// DelegateBlockGrader allows you to grade a single block. Each version has its own struct, which must be instantiated
// with the height and set of previous winners.
type DelegateBlockGrader interface {
	// Height returns the height the block grader is set to
	Height() int32
	// Version returns the version of the underlying grader
	Version() uint8
	// GetPreviousWinners returns the set of previous winners the grader was initialized with
	GetPreviousWinners() []string

	// AddSPR adds an spr to the set to be graded. The content is decoded using the underlying version's format
	// and validated based on the specified height and set of previous winners.
	//
	// Returns an error if an entry could not be validated.
	AddSPRV4(entryhash []byte, extids [][]byte, content []byte, balanceOfPEG uint64) error

	GetDelegatorsAddress(delegatorData []byte, signature []byte, signer string) ([]string, error)

	// Grade grades the block using the default settings for that version.
	// For more details, see each version's Grade() function.
	// If the result is empty, there are no winners.
	Grade() DelegatedGradedBlock

	// GradeCustom grades the SPRs using that version's algorithm and a custom cutoff for the top X
	GradeCustom(cutoff int) DelegatedGradedBlock

	// Count returns the number of SPRs that have been added
	Count() int

	// Payout returns the amount of Pegtoshi awarded to the SPR at the specified index
	Payout(index int) int64
}

// DelegatedGradedBlock is an immutable set of graded sprs
type DelegatedGradedBlock interface {
	// WinnersShortHashes returns the shorthashes of the winning SPRs.
	// This result can be used to set the next block's previous winners.
	// The amount varies between versions.
	// If there are no winners in this block, the previous block's winners are used.
	WinnersShortHashes() []string

	// Winners returns the winning SPRs
	Winners() []*GradingDelegatedSPR

	// Graded returns the top X SPRs that made it to the next stage of grading
	Graded() []*GradingDelegatedSPR

	// Version returns the underlying grader's version
	Version() uint8

	// Cutoff returns the amount SPRs that made it to the second stage of grading
	Cutoff() int

	// Count returns the total count of SPRs that were used in this GradedBlock
	Count() int

	// WinnerAmount returns the version specific amount of winners.
	WinnerAmount() int
}
