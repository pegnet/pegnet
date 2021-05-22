// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package opr

import "sort"

// UniqueOPRData is the minimum set of information we need
// to change for our OPR submissions
type UniqueOPRData struct {
	Nonce      []byte
	Difficulty uint64
}

// NonceRanking is a sorted list of nonces and their difficulties.
//	It allows us to only keep the top X oprs.
type NonceRanking struct {
	// Keep determines the depth of the list.
	// If you wish to keep the top 10 OPRs, set the keep to 10.
	Keep int

	List []*UniqueOPRData

	// Ability to set a minimum difficulty threshold
	// TODO: Set the minimum difficulty based on the previous block.
	//		 That way we don't submit oprs to the network if we know they will lose.
	MinimumDifficulty uint64

	WorstDiff  uint64 // Keep track of the worst difficulty to make the lookup cheap
	WorstIndex int

	taken int // The number of slots in the list taken. If we have space, accept it
}

func NewNonceRanking(keep int) *NonceRanking {
	r := new(NonceRanking)
	r.Keep = keep
	r.List = make([]*UniqueOPRData, keep)

	return r
}

// MergeNonceRankings will merge a set of rankings lists into 1
func MergeNonceRankings(keep int, rankings ...*NonceRanking) *NonceRanking {
	list := make([]*UniqueOPRData, 0)
	for _, l := range rankings {
		if l == nil {
			// Do not append nil's
			// We get a nil from a cancelled miner
			continue
		}
		list = append(list, l.GetNonces()...)
	}
	list = SortNonceRanks(list)

	n := NewNonceRanking(keep)
	for _, v := range list { // Adding nonces will only keep the top `keep`
		n.AddNonce(v.Nonce, v.Difficulty)
	}
	return n
}

// GetNonces returns the sorted nonce list
func (r *NonceRanking) GetNonces() []*UniqueOPRData {
	r.List = SortNonceRanks(r.List[:r.taken])
	return r.List
}

// AddNonce will only add the nonce information if it is better than the current worst
// or we have extra room that is not taken in the list.
func (r *NonceRanking) AddNonce(nonce []byte, difficulty uint64) bool {
	if difficulty < r.MinimumDifficulty {
		return false // Below min, we don't care
	}

	// If we have room, add it to the next avail slot
	if r.taken < r.Keep {
		newNonce := append([]byte{}, nonce...)
		r.List[r.taken] = &UniqueOPRData{newNonce, difficulty} // Add to empty slot
		// If we have our first, or new worst. Set the worst
		if r.taken == 0 || difficulty < r.WorstDiff {
			r.WorstDiff = difficulty
			r.WorstIndex = r.taken
		}
		r.taken++
		return true
	}

	// If we are full, we check against the worst
	if difficulty < r.WorstDiff {
		return false // Our worst in the list is better
	}

	newNonce := append([]byte{}, nonce...)
	// Replace the worst
	r.List[r.WorstIndex] = &UniqueOPRData{newNonce, difficulty}
	r.WorstDiff = difficulty

	// Update the worst
	for i, v := range r.List {
		if v.Difficulty < r.WorstDiff {
			r.WorstDiff = v.Difficulty
			r.WorstIndex = i
		}
	}
	return true
}

func SortNonceRanks(list []*UniqueOPRData) []*UniqueOPRData {
	sort.Slice(list, func(i, j int) bool {
		return list[i].Difficulty > list[j].Difficulty
	})
	return list
}
