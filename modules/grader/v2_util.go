package grader

import (
	"crypto/sha256"
	"encoding/binary"
	"math"

	"github.com/pegnet/pegnet/modules/opr"
)

func ValidateV2(entryhash []byte, extids [][]byte, height int32, winners []string, content []byte) (*GradingOPR, error) {
	if len(entryhash) != 32 {
		return nil, NewValidateError("invalid entry hash length")
	}

	if len(extids) != 3 {
		return nil, NewValidateError("invalid extid count")
	}

	if len(extids[1]) != 8 {
		return nil, NewValidateError("self reported difficulty must be 8 bytes")
	}

	if len(extids[2]) != 1 || extids[2][0] != 2 {
		return nil, NewValidateError("invalid version")
	}

	var dec *opr.V2Content
	err := dec.Unmarshal(content)
	if err != nil {
		// All errors are parse errors. We silence them here
		return nil, NewDecodeError(err.Error())
	}

	if dec.Height != height {
		return nil, NewValidateError("invalid height")
	}

	// verify assets
	if len(dec.Assets) != len(opr.V2Assets) {
		return nil, NewValidateError("invalid assets")
	}
	for i, val := range dec.Assets {
		if i > 0 && val == 0 {
			return nil, NewValidateError("assets must be greater than 0")
		}
	}

	if len(dec.Winners) != 10 && len(dec.Winners) != 25 {
		return nil, NewValidateError("must have exactly 10 or 25 previous winning shorthashes")
	}

	if !verifyWinnerFormat(dec.GetPreviousWinners(), 10) && !verifyWinnerFormat(dec.GetPreviousWinners(), 25) {
		return nil, NewValidateError("incorrect amount of previous winners")
	}

	if !verifyWinners(dec.GetPreviousWinners(), winners) {
		return nil, NewValidateError("incorrect set of previous winners")
	}

	gopr := new(GradingOPR)
	gopr.EntryHash = entryhash
	gopr.Nonce = extids[0]
	gopr.SelfReportedDifficulty = binary.BigEndian.Uint64(extids[1])

	sha := sha256.Sum256(content)
	gopr.OPRHash = sha[:]

	gopr.OPR = dec

	return gopr, nil
}

// V2 grading works similar to V1 but the grade is banded
// meaning a record within `band` percentage is considered to be
// equal
func gradeV2(avg []float64, opr *GradingOPR, band float64) float64 {
	assets := opr.OPR.GetOrderedAssets()
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
