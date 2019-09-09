package grader

import "github.com/pegnet/pegnet/modules/opr"

// GradingOPR holds the temporary variables used during the grading process
type GradingOPR struct {
	// Factom Entry variables
	EntryHash              []byte
	Nonce                  []byte
	SelfReportedDifficulty uint64

	// Grading Variables
	Grade   float64
	OPRHash []byte

	// Decoded OPR
	OPR opr.OPR
}
