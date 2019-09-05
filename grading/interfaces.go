/*
Package grading implements the basic grading module unit for pegnet opr grading.
*/
package grading

// IGradingBlock is the grading unit that accepts a set of OPRs. A Grading block must be created with
// a height and a network type. This will determine the grading and encoding versions it will use.
type IGradingBlock interface {
	// Information needed to setup a grading block. The height and network are determined at construction of
	// the grading block by the caller. The version will be computed from those values.
	Height() int64 // Block height in factomd
	Network() string
	Version() uint8 // Indicates the OPR version and grading to be used

	// Functions used for grading

	// AddOPR adds an opr to the set to be graded. If the set is already graded,
	// an error will be returned. Only basic proper entry formed validation occurs at this stage.
	// Most validation of the opr occurs in the grading routine. Validation at this stage is focused on parsing.
	//
	// The params are the components of a Factom Entry as their byte slices
	//
	// Returns
	//		bool	Indicates if the opr was added to the set. An invalid/improperly formed opr will return (false, nil)
	//		error	Indicates an error in the function call. This does not indicate a bad opr, but some other reason
	//				such as the set already being graded. If the grading module is graded, the set is locked.
	AddOPR(entryhash []byte, extids [][]byte, content []byte) (added bool, err error)

	// TotalOPRs will return the total number of OPRs properly added to this grading block. If the `AddOPR` returns
	// true, that opr will be included in this count.
	TotalOPRs() int

	// SetPreviousWinners enables checking of the previous winners in the validation function of the grading routine.
	// If the previous winners is unset, then an empty set is accepted.
	//
	// Returns
	//		error	If the length of the previous winners does not comply with the length rules of the version,
	//				then an error is returned and the set is rejected. An error is also returned if the previousWinners
	//				was already set.
	SetPreviousWinners(previousWinners []string) error

	// GetPreviousWinners returns the set of previous winners set by SetPreviousWinners
	GetPreviousWinners() []string

	// GradedSet performs the grading operation on the contained set in the module. If the grading is successful, the
	// returned slice of OPRs is in sorted order by their graded rank. Meaning `graded[0]` is the wining opr. And
	// graded[:amt] is the top `amt` (e.g to get paid). If the grading results in a empty block, such as not enough oprs,
	// the resulting slice will be of length 0, and `err` will be nil. The maximum length of the slice `graded` will be
	// the number of oprs determined by pure POW. In v1 and v2, this is 50. No more than 50 will ever be returned.
	// All OPRs will be fully validated until 50 are obtained sorted by their self reported difficulty order.
	// The grading module will only keep the top 50 OPRs, while the rest are discarded. If the caller wants to know if
	// a particular OPR is valid/proper, that cannot be done through the grading module.
	//
	// If err != nil, there was an error in grading, and the grading process was not finished.
	//
	// Calling GradedSet more than once will not change the result, as once the set is graded, the set is locked
	// and all future calls will return the same set.
	GradedSet() (graded []OPR, err error)

	// Winners returns the proper number of winners for the given graded set in the format accepted by the
	// `SetPreviousWinners` function. If the set is not graded, an error is returned.
	//
	// Returns
	//		[]string	The winners of the current set (can be an empty set)
	//		error		If the set is not graded, the winners cannot be asked for
	Winners() ([]string, error)

	// Functions used for determining the grading module state

	// The set can only be graded once. Once the set is graded, all future calls to `GradedSet` are idempotent
	// This function let's the caller know if the set is already graded, meaning all future calls will run in O(1).
	Graded() bool
}

// TODO: Construct another interface for the parse/validate of a single OPR for a caller to use if they wish to debug
// 		a specific entry

type OPR struct {
	// TBD
}
