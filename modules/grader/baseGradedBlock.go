package grader

import (
	"encoding/binary"
	"sort"
)

// baseGradedBlock is an opr set that has been graded
type baseGradedBlock struct {
	oprs   []*GradingOPR
	cutoff int
	height int32
	count  int

	shorthashes []string
}

func (b *baseGradedBlock) cloneOPRS(oprs []*GradingOPR) {
	b.oprs = nil
	for _, o := range oprs {
		b.oprs = append(b.oprs, o.Clone())
	}
	b.count = len(oprs)
}

func (b *baseGradedBlock) Count() int {
	return b.count
}

// AmountToGrade returns the number of OPRs the grading algorithm attempted to use in the process.
func (b *baseGradedBlock) AmountGraded() int {
	return len(b.oprs)
}

func (b *baseGradedBlock) createShortHashes(count int) {
	shortHashes := make([]string, count)
	if len(b.oprs) >= count {
		for i := 0; i < count; i++ {
			shortHashes[i] = b.oprs[i].Shorthash()
		}
	}
	b.shorthashes = shortHashes
}

// Graded returns the OPRs that made it into the cutoff
func (b *baseGradedBlock) Graded() []*GradingOPR {
	return b.oprs
}

// sortByDifficulty uses an efficient algorithm based on self-reported difficulty
// to avoid having to LXRhash the entire set.
// calculates at most `limit + misreported difficulties` hashes
func (b *baseGradedBlock) sortByDifficulty(limit int) {
	sort.SliceStable(b.oprs, func(i, j int) bool {
		return b.oprs[i].SelfReportedDifficulty > b.oprs[j].SelfReportedDifficulty
	})

	topX := make([]*GradingOPR, 0)
	for _, o := range b.oprs {
		hash := LX.Hash(append(o.OPRHash, o.Nonce...))
		diff := binary.BigEndian.Uint64(hash)

		if diff != o.SelfReportedDifficulty {
			continue
		}

		topX = append(topX, o)

		if len(topX) >= limit {
			break
		}
	}

	b.oprs = topX
}

// filter out duplicate gradingOPRs. an OPR is a duplicate when both
// nonce and oprhash are the same
func (b *baseGradedBlock) filterDuplicates() {
	filtered := make([]*GradingOPR, 0)

	added := make(map[string]bool)
	for _, v := range b.oprs {
		id := string(append(v.Nonce, v.OPRHash...))
		if !added[id] {
			filtered = append(filtered, v)
			added[id] = true
		}
	}

	b.oprs = filtered
}

func (b *baseGradedBlock) Cutoff() int { return b.cutoff }
