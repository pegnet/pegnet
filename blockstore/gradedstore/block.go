package gradedstore

import (
	"crypto/rand"
	"fmt"

	"github.com/pegnet/pegnet/modules/grader"
)

// OprBlock is the graded block that we can save/retrieve from disk
type OprBlock struct {
	// GradedOPRs
	GradedOPRs grader.GradedBlock

	// Eblock Information
	EblockKeyMr    string
	EblockSequence int32
	Dbht           int32

	// TotalNumberRecords is the total number of oprs that parsed for the block.
	// We tend to only keep the top 50, it's nice to know how many existed before we cut it
	TotalNumberRecords int
	EmptyOPRBlock      bool // An empty opr block is an eblock that had no winners
}

// RandomOPRBlock is very useful for unit testing
func RandomOPRBlock(version uint8, dbht int32) (*OprBlock, error) {
	g, err := grader.NewGrader(version, dbht, nil)
	if err != nil {
		return nil, err
	}
	for i := 0; i < 50; i++ {
		if err := g.AddOPR(RandomOPRWithFields(version, dbht)); err != nil {
			return nil, err
		}
	}

	keymr := make([]byte, 32)
	_, _ = rand.Read(keymr)

	block := g.Grade()
	return &OprBlock{
		GradedOPRs:         block,
		EblockKeyMr:        fmt.Sprintf("%x", keymr),
		EblockSequence:     dbht,
		Dbht:               dbht,
		TotalNumberRecords: 50,
		EmptyOPRBlock:      false,
	}, nil
}
