package grader

import (
	"encoding/binary"
	"sort"

	"github.com/pegnet/pegnet/modules/lxr30"
)

// BaseGradedBlock is an opr set that has been graded
type BaseGradedBlock struct {
	OPRs     []*GradingOPR
	CutOff   int
	Height   int32
	OPRCount int

	ShortHashes []string
}

func (b *BaseGradedBlock) cloneOPRS(oprs []*GradingOPR) {
	b.OPRs = nil
	for _, o := range oprs {
		b.OPRs = append(b.OPRs, o.Clone())
	}
	b.OPRCount = len(oprs)
}

func (b *BaseGradedBlock) Count() int {
	return b.OPRCount
}

// AmountToGrade returns the number of OPRs the grading algorithm attempted to use in the process.
func (b *BaseGradedBlock) AmountGraded() int {
	return len(b.OPRs)
}

func (b *BaseGradedBlock) createShortHashes(count int) {
	shortHashes := make([]string, count)
	if len(b.OPRs) >= count {
		for i := 0; i < count; i++ {
			shortHashes[i] = b.OPRs[i].Shorthash()
		}
	}
	b.ShortHashes = shortHashes
}

// Graded returns the OPRs that made it into the CutOff
func (b *BaseGradedBlock) Graded() []*GradingOPR {
	return b.OPRs
}

// sortByDifficulty uses an efficient algorithm based on self-reported difficulty
// to avoid having to LXRhash the entire set.
// calculates at most `limit + misreported difficulties` hashes
func (b *BaseGradedBlock) sortByDifficulty(limit int) {
	sort.SliceStable(b.OPRs, func(i, j int) bool {
		return b.OPRs[i].SelfReportedDifficulty > b.OPRs[j].SelfReportedDifficulty
	})

	lx := lxr30.Init()

	topX := make([]*GradingOPR, 0)
	for _, o := range b.OPRs {
		hash := lx.Hash(append(o.OPRHash, o.Nonce...))
		diff := binary.BigEndian.Uint64(hash)

		if diff != o.SelfReportedDifficulty {
			continue
		}

		topX = append(topX, o)

		if len(topX) >= limit {
			break
		}
	}

	b.OPRs = topX
}

// filter out duplicate gradingOPRs. an OPR is a duplicate when both
// nonce and oprhash are the same
func (b *BaseGradedBlock) filterDuplicates() {
	filtered := make([]*GradingOPR, 0)

	added := make(map[string]bool)
	for _, v := range b.OPRs {
		id := string(append(v.Nonce, v.OPRHash...))
		if !added[id] {
			filtered = append(filtered, v)
			added[id] = true
		}
	}

	b.OPRs = filtered
}

func (b *BaseGradedBlock) Cutoff() int { return b.CutOff }
