package graderStake

import (
	"crypto/sha256"
	"fmt"
	"github.com/pegnet/pegnet/modules/factoidaddress"
	"github.com/pegnet/pegnet/modules/opr"
	"github.com/pegnet/pegnet/modules/spr"
	"math"
	"sort"
)

// S1Payout is the amount of Pegtoshi given to the SPR with the specified index
func S1Payout(index int) int64 {
	if index >= 25 || index < 0 {
		return 0
	}
	return 180 * 1e8
}

// ValidateS1 validates the provided data using the specified parameters
func ValidateS1(entryhash []byte, extids [][]byte, height int32, content []byte) (*GradingSPR, error) {
	if len(entryhash) != 32 {
		return nil, NewValidateError("invalid entry hash length")
	}

	if len(extids) != 3 {
		return nil, NewValidateError("invalid extid count")
	}

	if len(extids[0]) != 1 || extids[0][0] != 5 {
		return nil, NewValidateError("invalid version")
	}

	// ParseS1Content parses the V2 proto format
	// S1 is just the proto format with some more assets.
	o2, err := spr.ParseS1Content(content)
	if err != nil {
		return nil, NewDecodeError(err.Error())
	}
	o := &spr.S1Content{V2Content: *o2}

	if o.Height != height {
		return nil, NewValidateError("invalid height")
	}

	// verify assets
	if len(o.Assets) != len(opr.V5Assets) {
		return nil, NewValidateError("invalid assets")
	}
	for _, val := range o.Assets {
		if val == 0 {
			return nil, NewValidateError("assets must be greater than 0")
		}
	}

	if err := factoidaddress.Valid(o.Address); err != nil {
		return nil, NewValidateError(fmt.Sprintf("factoidaddress is invalid : %s", err.Error()))
	}

	gspr := new(GradingSPR)
	gspr.EntryHash = entryhash
	gspr.CoinbaseAddress = o.Address
	sha := sha256.Sum256(content)
	gspr.SPRHash = sha[:]

	gspr.SPR = o
	return gspr, nil
}

// S1 grading works similar to V1 but the grade is banded
// meaning a record within `band` percentage is considered to be equal
func gradeS1(avg []float64, spr *GradingSPR, band float64) float64 {
	assets := spr.SPR.GetOrderedAssetsFloat()
	spr.Grade = 0
	for i, asset := range assets {
		if avg[i] > 0 {
			d := math.Abs((asset.Value - avg[i]) / avg[i]) // compute the difference from the average
			if d <= band {
				d = 0
			} else {
				d -= band
			}
			spr.Grade += d * d * d * d // the grade is the sum of the square of the square of the differences
		}
	}
	return spr.Grade
}

func TrimmedMeanFloat(data []float64, p int) float64 {
	sort.Slice(data, func(i, j int) bool {
		return data[i] < data[j]
	})

	length := len(data)
	if length <= 3 {
		return data[length/2]
	}

	sum := 0.0
	for i := p; i < length-p; i++ {
		sum = sum + data[i]
	}
	return sum / float64(length-2*p)
}

// calculate the vector of average prices
func averageS1(sprs []*GradingSPR) []float64 {
	data := make([][]float64, len(sprs[0].SPR.GetOrderedAssetsFloat()))
	avg := make([]float64, len(sprs[0].SPR.GetOrderedAssetsFloat()))

	// Sum up all the prices
	for _, o := range sprs {
		for i, asset := range o.SPR.GetOrderedAssetsFloat() {
			data[i] = append(data[i], asset.Value)
		}
	}
	for i := range data {
		sum := 0.0
		for j := range data[i] {
			sum += data[i][j]
		}
		noisyRate := 0.1
		length := int(float64(len(data[i])) * noisyRate)
		avg[i] = TrimmedMeanFloat(data[i], length+1)
	}
	return avg
}
