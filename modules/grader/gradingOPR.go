package grader

import "github.com/pegnet/pegnet/modules/opr"

type GradingOPR struct {
	EntryHash              []byte
	Nonce                  []byte
	Grade                  float64
	SelfReportedDifficulty uint64
	validSRD               bool
	OPRHash                []byte
	OPR                    opr.OPR
}
