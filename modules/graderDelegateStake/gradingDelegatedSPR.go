package graderDelegateStake

import (
	"encoding/hex"

	"github.com/pegnet/pegnet/modules/spr"
)

// GradingDelegatedSPR holds the temporary variables used during the grading process
type GradingDelegatedSPR struct {
	// Factom Entry variables
	EntryHash       []byte
	CoinbaseAddress string

	// Grading Variables
	Grade    float64
	SPRHash  []byte
	position int
	payout   int64

	// Decoded SPR
	SPR spr.SPR

	balanceOfPEG uint64
}

// Clone the GradingDelegatedSPR
func (o *GradingDelegatedSPR) Clone() *GradingDelegatedSPR {
	clone := new(GradingDelegatedSPR)
	clone.EntryHash = append(o.EntryHash[:0:0], o.EntryHash...)
	clone.Grade = o.Grade
	clone.CoinbaseAddress = o.CoinbaseAddress
	clone.SPRHash = append(o.SPRHash[:0:0], o.SPRHash...)
	clone.SPR = o.SPR.Clone()
	clone.payout = o.payout
	clone.position = o.position
	clone.balanceOfPEG = o.balanceOfPEG
	return clone
}

// Shorthash is the hex-encoded first 8 bytes of the entry hash
func (o *GradingDelegatedSPR) Shorthash() string {
	return hex.EncodeToString(o.EntryHash[:8])
}

// Payout is the amount of Pegtoshi this SPR would be rewarded with.
// Only valid for GradingDelegatedSPRs coming from a GradedBlock
func (o *GradingDelegatedSPR) Payout() int64 {
	return o.payout
}

// Position is the index of the SPR in the Graded set. Position 0 is the winner.
// Only valid for GradingDelegatedSPRs coming from a GradedBlock
func (o *GradingDelegatedSPR) Position() int {
	return o.position
}
