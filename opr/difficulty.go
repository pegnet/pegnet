package opr

import (
	"math"
	"math/big"
)

const (
	MiningPeriod = 480 // in seconds
)

// CalculateMinimumDifficultyFromOPRs that we should submit for a block.
//	Params:
//		oprs 		Sorted by difficulty, must be > 0 oprs
//		cutoff 		Is 1 based, not 0. So cutoff 50 is at index 49.
func CalculateMinimumDifficultyFromOPRs(oprs []*OraclePriceRecord, cutoff int) uint64 {
	var min *OraclePriceRecord
	var spot = 0
	// grab the least difficult in the top 50
	if len(oprs) >= 50 {
		// Oprs is indexed at 0, where the math (spot) is indexed at 1.
		min = oprs[49]
		spot = 50
	} else {
		min = oprs[len(oprs)-1]
		spot = len(oprs)
	}

	// minDiff is our number to create a tolerance around
	minDiff := min.Difficulty

	return CalculateMinimumDifficulty(spot, minDiff, cutoff)
}

// CalculateMinimumDifficulty
//		spot		The index of the difficulty param in the list. Sorted by difficulty
//		difficulty	The difficulty at index 'spot'
//		cutoff		The targeted index difficulty estimate
func CalculateMinimumDifficulty(spot int, difficulty uint64, cutoff int) uint64 {
	// Calculate the effective hash rate of the network in hashes/s
	hashrate := EffectiveHashRate(difficulty, spot)

	// Given that hashrate, aim to be above the cutoff
	floor := ExpectedMinimumDifficulty(hashrate, cutoff)
	return floor
}

// The effective hashrate of the network given the difficulty of the 50th opr
// sorted by difficulty.
// Using https://github.com/WhoSoup/pegnet/wiki/Mining-Probabilities
func EffectiveHashRate(min uint64, spot int) float64 {
	minF := big.NewFloat(float64(min))

	// 2^64
	space := big.NewFloat(math.MaxUint64)

	// Assume min is the 50th spot
	minSpot := big.NewFloat(float64(spot))

	num := new(big.Float).Mul(minSpot, space)
	den := new(big.Float).Sub(space, minF)

	ehr := new(big.Float).Quo(num, den)
	ehr = ehr.Quo(ehr, big.NewFloat(MiningPeriod))
	f, _ := ehr.Float64()
	return f
}

// ExpectedMinimumDifficulty will report what minimum difficulty we would expect given a hashrate for
// a given position.
// Using https://github.com/WhoSoup/pegnet/wiki/Mining-Probabilities#expected-minimum-difficulty
func ExpectedMinimumDifficulty(hashrate float64, spot int) uint64 {
	// 2^64
	space := big.NewFloat(math.MaxUint64)
	ehrF := big.NewFloat(hashrate * MiningPeriod)
	spotF := new(big.Float).Sub(ehrF, big.NewFloat(float64(spot)))
	num := new(big.Float).Mul(space, spotF)

	den := ehrF

	expMin := new(big.Float).Quo(num, den)
	f, _ := expMin.Float64()
	return uint64(f)
}
