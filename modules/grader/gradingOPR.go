package grader

import (
	"encoding/hex"

	"github.com/pegnet/pegnet/modules/opr"
)

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

func (o *GradingOPR) Clone() *GradingOPR {
	clone := new(GradingOPR)
	clone.EntryHash = append(o.EntryHash[:0:0], o.EntryHash...)
	clone.Nonce = append(o.Nonce[:0:0], o.Nonce...)
	clone.SelfReportedDifficulty = o.SelfReportedDifficulty
	clone.Grade = o.Grade
	clone.OPRHash = append(o.OPRHash[:0:0], o.OPRHash...)
	clone.OPR = o.OPR.Clone()
	return clone
}

func (o *GradingOPR) Shorthash() string {
	return hex.EncodeToString(o.EntryHash[:8])
}
