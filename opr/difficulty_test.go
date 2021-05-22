package opr_test

import (
	"math"
	"testing"

	. "github.com/pegnet/pegnet/opr"
)

// Just some sample vectors taken from 1 miner on a network
func TestEffectiveHashRate(t *testing.T) {
	vectors := []struct {
		WorstDiff uint64
		BestDiff  uint64
		HashRate  float64
		Height    uint64
	}{
		{WorstDiff: 18446586878042484810, BestDiff: 18446742466885828062, HashRate: 14399.63, Height: 204797},
		{WorstDiff: 18446633144064883604, BestDiff: 18446743743681176119, HashRate: 14337.81, Height: 204796},
		{WorstDiff: 18446587826224330316, BestDiff: 18446740789917189647, HashRate: 14319.85, Height: 204795},

		{WorstDiff: 18446595274788757708, BestDiff: 18446743410821960935, HashRate: 14265.29, Height: 204794},
		{WorstDiff: 18446611696110517928, BestDiff: 18446743091015240431, HashRate: 14235.62, Height: 204793},
		{WorstDiff: 18446572367808766797, BestDiff: 18446739418174104138, HashRate: 14299.25, Height: 204792},

		{WorstDiff: 18446608794088792336, BestDiff: 18446743251485133710, HashRate: 14335.20, Height: 204791},
		{WorstDiff: 18446610101229625045, BestDiff: 18446744035146642516, HashRate: 14332.53, Height: 204790},
		{WorstDiff: 18446625595880705622, BestDiff: 18446739129668856235, HashRate: 14356.89, Height: 204789},
	}

	for _, v := range vectors {
		ehr := EffectiveHashRate(v.WorstDiff, 50)
		diff := math.Abs(ehr - v.HashRate)
		if diff/v.HashRate > 0.25 {
			// 25% tolerance?
			t.Errorf("Hashrate calculated is over 20%% different from the actual. Difference is %.2f%%", diff/v.HashRate*100)
		}
		//fmt.Printf("%d %.4f, %.4f\n", v.Height, ehr, v.HashRate)

		expMin := ExpectedMinimumDifficulty(v.HashRate, 50)
		diff = math.Abs(float64(expMin) - float64(v.WorstDiff))
		if diff/float64(v.WorstDiff) > 0.25 {
			// 25% tolerance?
			t.Errorf("Minimum difficulty calculated is over 20%% different from the actual. Difference is %.2f%%", diff/float64(v.WorstDiff)*100)
		}
		//fm
		//fmt.Printf("%d %x, %x\n", v.Height, expMin, v.WorstDiff)
	}

}
