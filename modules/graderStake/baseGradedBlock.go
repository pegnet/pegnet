package graderStake

import (
	"bytes"
	"sort"

	log "github.com/sirupsen/logrus"
	"github.com/FactomProject/factom"
)

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

// randomize the SPRs order.  This provides the reward to a random qualified submission
func (b *baseGradedBlock) shuffleSPRs() {
	// First of all, let's make sure all instances of pegnetd use the same order for the SPRs
	// We sort them by their entry hashes
	sort.Slice(b.sprs, func(i, j int) bool {
		return bytes.Compare(b.sprs[i].EntryHash, b.sprs[j].EntryHash) < 0
	})
	// Add the directory block KMR to the mix to prevent mining to improve odds on the random
	// selection of staking rewards.
	DBlock, _, err := factom.GetDBlockByHeight(int64(b.height))
	if err != nil || DBlock == nil {
		log.WithFields(log.Fields{
			"dbheight": b.height,
		}).Info("Could not get the DBlock at this height to shuffle the SPRs")
		return
	}
	// Now we want to create a seed from the first 24 bits of all the accepted SPRs.
	// We will combine them in a way that all the bits in each 24 bits of an Entryhash contribute
	// the the randomness of our selection.
	seed := int64(DBlock.KeyMR[0]) ^ int64(DBlock.KeyMR[1])<<8 ^ int64(DBlock.KeyMR[2])<<16
	for _, spr := range b.sprs {
		// Mix up the bits in seed by replacing it with its bits scrambled.
		seed = seed>>1 ^ seed<<5
		// Now take the top 24 bits of SPR's entryhash and xor it in.
		seed ^= int64(spr.EntryHash[0]) ^ int64(spr.EntryHash[1])<<8 ^ int64(spr.EntryHash[2])<<16
		// Each time through the loop, bits of each 24 bits of each entryhash are scrambled again and again.
		// This is a pretty solid sort of pseudorandom number generation. see https://en.wikipedia.org/wiki/Xorshift
	}

	// Now we randomize b.sprs by using our random number generator and telling our sort that
	// spr[i] < spr[j] not by comparing them, but looking at the result of a stream of random numbers
	sort.Slice(b.sprs, func(i, j int) bool {
		seed = seed>>7 ^ seed<<3
		return seed&0xFF > 0x7F
	})
}

func (b *baseGradedBlock) Cutoff() int { return b.cutoff }
