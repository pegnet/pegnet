package grader

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"regexp"
	"sort"

	"github.com/pegnet/pegnet/modules/factoidaddress"
	"github.com/pegnet/pegnet/modules/opr"
)

// V5Payout is the amount of Pegtoshi given to the OPR with the specified index
func V5Payout(index int) int64 {
	if index >= 25 || index < 0 {
		return 0
	}
	return 360 * 1e8
}

// ValidateV5 validates the provided data using the specified parameters
func ValidateV5(entryhash []byte, extids [][]byte, height int32, winners []string, content []byte) (*GradingOPR, error) {
	if len(entryhash) != 32 {
		return nil, NewValidateError("invalid entry hash length")
	}

	if len(extids) != 3 {
		return nil, NewValidateError("invalid extid count")
	}

	if len(extids[1]) != 8 {
		return nil, NewValidateError("self reported difficulty must be 8 bytes")
	}

	if len(extids[2]) != 1 || extids[2][0] != 5 {
		return nil, NewValidateError("invalid version")
	}

	// ParseV2Content parses the V2 proto format
	// V5 is just the proto format with some more assets.
	o2, err := opr.ParseV2Content(content)
	if err != nil {
		return nil, NewDecodeError(err.Error())
	}
	o := &opr.V5Content{V2Content: *o2}

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

	if valid, _ := regexp.MatchString("^[a-zA-Z0-9,]+$", o.ID); !valid {
		return nil, NewValidateError("only alphanumeric characters and commas are allowed in the identity")
	}

	if !verifyWinnerFormat(o.GetPreviousWinners(), 25) {
		return nil, NewValidateError("incorrect amount of previous winners")
	}

	if !verifyWinners(o.GetPreviousWinners(), winners) {
		return nil, NewValidateError("incorrect set of previous winners")
	}

	gopr := new(GradingOPR)
	gopr.EntryHash = entryhash
	gopr.Nonce = extids[0]
	gopr.SelfReportedDifficulty = binary.BigEndian.Uint64(extids[1])

	sha := sha256.Sum256(content)
	gopr.OPRHash = sha[:]

	gopr.OPR = o

	return gopr, nil
}

// V5 grading works similar to V1 but the grade is banded
// meaning a record within `band` percentage is considered to be equal
func gradeV5(avg []float64, opr *GradingOPR, band float64) float64 {
	assets := opr.OPR.GetOrderedAssetsFloat()
	opr.Grade = 0
	for i, asset := range assets {
		if avg[i] > 0 {
			d := math.Abs((asset.Value - avg[i]) / avg[i]) // compute the difference from the average
			if d <= band {
				d = 0
			} else {
				d -= band
			}
			opr.Grade += d * d * d * d // the grade is the sum of the square of the square of the differences
		}
	}
	return opr.Grade
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
func averageV5(oprs []*GradingOPR) []float64 {
	data := make([][]float64, len(oprs[0].OPR.GetOrderedAssetsFloat()))
	avg := make([]float64, len(oprs[0].OPR.GetOrderedAssetsFloat()))

	// Sum up all the prices
	for _, o := range oprs {
		for i, asset := range o.OPR.GetOrderedAssetsFloat() {
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
