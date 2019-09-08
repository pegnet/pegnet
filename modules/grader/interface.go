package grader

// Block is the grading unit that accepts a set of OPRs. A Grading block must be created with
// a height and a version. This will determine the grading and encoding versions it will use.
// When the Grading block is `ungraded` some functions will be unusable. A grading block becomes `graded` when the
// `Grade()` function is called. Adding a new OPR will un-grade the set.
type Block interface {
	// Information needed to setup a grading block. The height and version are determined at construction of
	// the grading block by the caller.
	Height() int32  // Block height in factomd
	Version() uint8 // Indicates the OPR version and grading to be used

	// ---------------------------
	// Functions used for grading

	// AddOPR adds an opr to the set to be graded. If the set is already graded,
	// the set will become ungraded. Some functions require the grading module to be graded to work.
	// Only basic proper entry formed validation occurs at this stage. Most validation of the opr occurs in
	// the grading routine. Validation at this stage is focused on parsing.
	//
	// The params are the components of a Factom Entry as their byte slices
	//
	// Returns
	//		bool	Indicates if the opr was added to the set. An invalid/improperly formed opr will return (false, nil)
	//		error	Indicates an error in the function call. This does not indicate a bad opr, but some other reason.
	AddOPR(entryhash []byte, extids [][]byte, content []byte) error

	/*	// SetPreviousWinners enables checking of the previous winners in the validation function of the grading routine.
		// If the previous winners is unset, then an empty set is accepted. SetPreviousWinners will set the graded block to
		// 'un-graded'.
		//
		// Returns
		//		error	If the length of the previous winners does not comply with the length rules of the version,
		//				then an error is returned and the set is rejected.
		SetPreviousWinners(previousWinners []string) error*/

	// Grade performs the grading operation on the contained set in the module. If the grading is successful, the
	// grading block will be set to `graded`, and the opr set accessors for the resulting grading set will be enabled.
	//
	// The graded slice is the slice of OPRs in their sorted order by their graded rank. Meaning `graded[0]` is
	// the wining opr. And graded[:amt] is the top `amt` (e.g to get paid).
	//
	// If the grading results in a empty block, such as not enough oprs, the graded slice will be of length 0,
	// and `err` will be nil. The maximum length of the slice `graded` will be the number of oprs determined by
	// pure POW. In v1 and v2, this is 50. No more than 50 will ever be returned by the accessors.
	//
	// All OPRs will be fully validated until 50 are obtained sorted by their self reported difficulty order.
	// The grading module will only allow access the top 50 OPRs. A caller does not have access to oprs outside
	// this range.
	//
	// If err != nil, there was an error in grading, and the grading process was not finished.
	//
	// Calling GradedSet more than once will not change the result as long as the set remains graded. Adding a new OPR
	// or setting a new PreviousWinners will unlock the set, and then calling `Grade()` will regrade the oprs with the
	// new state. As long as the set is locked, all future calls will do nothing.
	Grade() []*GradingOPR

	// WinnersShortHashes returns the proper number of winners for the given graded set in the format accepted by the
	// `SetPreviousWinners` function. If the set is not graded, an error is returned.
	//
	// Returns
	//		[]string	The winners of the current set (can be an empty set)
	//		error		If the set is not graded, the winners cannot be asked for
	WinnersShortHashes() []string

	// Graded returns the full set of OPRs that were graded, meaning their POW got them into the top 50.
	Graded() []*GradingOPR

	// ---------------------------
	// Functions used for determining the grading module state

	// Count will return the total number of OPRs properly added to this grading block. If the `AddOPR` returns
	// true, that opr will be included in this count.
	Count() int
	WinnerAmount() int

	// GetPreviousWinners returns the set of previous winners set by SetPreviousWinners
	GetPreviousWinners() []string
}
