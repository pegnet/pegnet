package graderStake

// baseGradedBlock is an spr set that has been graded
type baseGradedBlock struct {
	sprs   []*GradingSPR
	cutoff int
	height int32
	count  int

	shorthashes []string
}

func (b *baseGradedBlock) cloneSPRS(sprs []*GradingSPR) {
	b.sprs = nil
	for _, o := range sprs {
		b.sprs = append(b.sprs, o.Clone())
	}
	b.count = len(sprs)
}

func (b *baseGradedBlock) Count() int {
	return b.count
}

// AmountToGrade returns the number of SPRs the grading algorithm attempted to use in the process.
func (b *baseGradedBlock) AmountGraded() int {
	return len(b.sprs)
}

func (b *baseGradedBlock) createShortHashes(count int) {
	shortHashes := make([]string, count)
	if len(b.sprs) >= count {
		for i := 0; i < count; i++ {
			shortHashes[i] = b.sprs[i].Shorthash()
		}
	}
	b.shorthashes = shortHashes
}

// Graded returns the SPRs that made it into the cutoff
func (b *baseGradedBlock) Graded() []*GradingSPR {
	return b.sprs
}

// filter out duplicate GradingSPRs. an SPR is a duplicate when both
// nonce and sprhash are the same
func (b *baseGradedBlock) filterDuplicates() {
	filtered := make([]*GradingSPR, 0)

	added := make(map[string]bool)
	for _, v := range b.sprs {
		id := v.CoinbaseAddress
		if !added[id] {
			filtered = append(filtered, v)
			added[id] = true
		}
	}

	b.sprs = filtered
}

func (b *baseGradedBlock) Cutoff() int { return b.cutoff }
