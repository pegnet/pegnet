package graderStake

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"
)

func CompareTwoLists(t *testing.T, match bool, ref, chg *baseGradedBlock, errorMsg string) {
	cnt := 0
	for i, spr := range ref.sprs {
		if bytes.Equal(chg.sprs[i].EntryHash, spr.EntryHash) {
			cnt++
		}
	}
	if match && cnt != len(ref.sprs) {
		t.Error(errorMsg)
	}
	if !match && cnt > len(ref.sprs)/2 {
		t.Error(errorMsg)
	}

	fmt.Println("================")
	return

}

func Test_shuffle(t *testing.T) {
	// First create a list of sprs in a baseGradedBlock
	// These will be out of order.
	b := new(baseGradedBlock)
	for i := 0; i < 50; i++ {
		e := sha256.Sum256([]byte{byte(i)})
		spr := new(GradingSPR)
		spr.EntryHash = e[:]
		b.sprs = append(b.sprs, spr)
	}

	// hold the original list, then shuffle.  These two list should be quite different.
	hld := new(baseGradedBlock)
	hld2 := new(baseGradedBlock)
	hld.sprs = append(hld.sprs, b.sprs...)
	b.shuffleSPRs()
	CompareTwoLists(t, false, hld, b, "shuffle of initial values shouldn't match")

	// Shuffle the sprs, then capture that result, shuffle again.
	// Should not matter, should be exactly the same
	b.shuffleSPRs()
	hld.sprs = append(hld.sprs[:0], b.sprs...)
	b.shuffleSPRs()
	CompareTwoLists(t, true, hld, b, "double shuffle should have no impact")

	// Different length lists should not result in the same lists
	ne := sha256.Sum256([]byte{51})
	nspr := new(GradingSPR)
	nspr.EntryHash = ne[:]
	b.sprs = append(b.sprs, nspr)
	b.shuffleSPRs()
	CompareTwoLists(t, false, hld, b, "different length lists should be very different")

	// Shuffle the sprs, then capture that result, shuffle again.
	// Still should not matter, should be exactly the same
	hld2.sprs = append(hld2.sprs[:0], b.sprs...)
	b.shuffleSPRs()
	CompareTwoLists(t, true, hld2, b, "double shuffle still should have no impact")
	b.sprs = append(b.sprs[:0], hld.sprs...) // reset b

	// Changing even a bit of a hash's first 24 bits should mix the lists
	b.sprs[49].EntryHash[0] ^= 1
	b.shuffleSPRs()
	CompareTwoLists(t, false, hld, b, "bit changes to the first 24 bits of an entry hash should give different lists")

	// Shuffle again, and that should not matter, should be exactly the same
	hld2.sprs = append(hld2.sprs[:0], b.sprs...)
	b.shuffleSPRs()
	CompareTwoLists(t, true, hld2, b, "double shuffle still should have no impact")
	b.sprs = append(b.sprs[:0], hld.sprs...) // reset b

}
