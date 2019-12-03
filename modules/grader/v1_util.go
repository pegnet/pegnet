package grader

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/pegnet/pegnet/modules/opr"
)

// V1Payout is the amount of Pegtoshi given to the OPR with the specified index
func V1Payout(index int) int64 {
	if index >= 10 || index < 0 {
		return 0
	}
	switch index {
	case 0:
		return 800 * 1e8
	case 1:
		return 600 * 1e8
	default:
		return 450 * 1e8
	}
}

// ValidateV1 validates the provided data using the specified parameters
func ValidateV1(entryhash []byte, extids [][]byte, height int32, winners []string, content []byte) (*GradingOPR, error) {
	if len(entryhash) != 32 {
		return nil, NewValidateError("invalid entry hash length")
	}

	if len(extids) != 3 {
		return nil, NewValidateError("invalid extid count")
	}

	if len(extids[2]) != 1 || extids[2][0] != 1 {
		return nil, NewValidateError("invalid version")
	}

	if len(extids[1]) != 8 {
		return nil, NewValidateError("self reported difficulty must be 8 bytes")
	}

	o, err := opr.ParseV1Content(content)
	if err != nil {
		return nil, NewDecodeError(err.Error())
	}

	if o.Dbht != height {
		return nil, NewValidateError("invalid height")
	}

	// verify assets
	for _, code := range opr.V1Assets {
		if v, ok := o.Assets[code]; !ok {
			return nil, NewValidateError("asset list is not correct")
		} else if code != "PNT" && v == 0 {
			return nil, NewValidateError("all values other than PNT must be nonzero")
		}
	}

	if !verifyWinnerFormat(o.WinPreviousOPR, 10) {
		return nil, NewValidateError("invalid list of previous winners")
	}

	if !verifyWinners(o.WinPreviousOPR, winners) {
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

// calculate the vector of average prices
func averageV1(oprs []*GradingOPR) []float64 {
	avg := make([]float64, len(oprs[0].OPR.GetOrderedAssetsFloat()))

	// Sum up all the prices
	for _, o := range oprs {
		for i, asset := range o.OPR.GetOrderedAssetsFloat() {
			if asset.Value >= 0 { // Make sure no OPR has negative values for
				avg[i] += asset.Value // assets.  Simply treat all values as positive.
			} else {
				avg[i] -= asset.Value
			}
		}
	}
	// Then divide the prices by the number of OraclePriceRecord records.  Two steps is actually faster
	// than doing everything in one loop (one divide for every asset rather than one divide
	// for every asset * number of OraclePriceRecords)  There is also a little bit of a precision advantage
	// with the two loops (fewer divisions usually does help with precision) but that isn't likely to be
	// interesting here.
	total := float64(len(oprs))
	for i := range avg {
		avg[i] = avg[i] / total
	}

	return avg
}

// v1 grading algorithm, sum of the quadratic differences to the mean
func gradeV1(avg []float64, opr *GradingOPR) float64 {
	assets := opr.OPR.GetOrderedAssetsFloat()
	opr.Grade = 0
	for i, asset := range assets {
		if avg[i] > 0 {
			d := (asset.Value - avg[i]) / avg[i] // compute the difference from the average
			opr.Grade += d * d * d * d           // the grade is the sum of the square of the square of the differences
		}
	}
	return opr.Grade
}
