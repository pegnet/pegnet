package graderDelegateStake

import (
	"crypto/sha256"
	"fmt"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/pegnet/pegnet/modules/factoidaddress"
	"github.com/pegnet/pegnet/modules/graderStake"
	"github.com/pegnet/pegnet/modules/opr"
	"github.com/pegnet/pegnet/modules/spr"
	"math"
)

// S4Payout is the amount of Pegtoshi given to the SPR with the specified index
func S4Payout(index int) int64 {
	if index >= 25 || index < 0 {
		return 0
	}
	return 180 * 1e8
}

func removeDuplicateValues(inputSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range inputSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func getDelegatorsAddress(delegatorData []byte, signature []byte, signer string) ([]string, error) {
	if len(signature) != 96 {
		return nil, graderStake.NewValidateError("Invalid signature length")
	}
	dPubKey := signature[:32]
	dSignData := signature[32:]

	err3 := primitives.VerifySignature(delegatorData, dPubKey[:], dSignData[:])
	if err3 != nil {
		return nil, graderStake.NewValidateError("Invalid signature")
	}

	var listOfDelegatorsAddress []string
	for bI := 0; bI < len(delegatorData); bI += 148 {
		delegator := delegatorData[bI : bI+148]
		addressOfDelegator := delegator[:52]
		signDataOfDelegator := delegator[52:116]
		pubKeyOfDelegator := delegator[116:]

		err2 := primitives.VerifySignature([]byte(signer), pubKeyOfDelegator[:], signDataOfDelegator[:])
		if err2 != nil {
			continue
		}
		listOfDelegatorsAddress = append(listOfDelegatorsAddress, string(addressOfDelegator[:]))
	}
	removeDuplicateValuesSlice := removeDuplicateValues(listOfDelegatorsAddress)
	return removeDuplicateValuesSlice, nil
}

// ValidateS4 validates the provided data using the specified parameters
func ValidateS4(entryhash []byte, extids [][]byte, height int32, content []byte, balanceOfPEG uint64) (*GradingDelegatedSPR, error) {
	if len(entryhash) != 32 {
		return nil, graderStake.NewValidateError("invalid entry hash length")
	}

	if len(extids) != 5 {
		return nil, graderStake.NewValidateError("invalid extid count")
	}

	if len(extids[0]) != 1 || extids[0][0] != 8 {
		return nil, graderStake.NewValidateError("invalid version")
	}

	// ParseS1Content parses the V2 proto format
	// S1 is just the proto format with some more assets.
	o2, err := spr.ParseS1Content(content)
	if err != nil {
		return nil, graderStake.NewDecodeError(err.Error())
	}
	o := &spr.S1Content{V2Content: *o2}

	/**
	 *  Verify Signature of Oracle Price Data
	 */
	if len(extids[2]) != 96 {
		return nil, graderStake.NewValidateError("invalid signature length")
	}
	pubKey := extids[2][:32]
	signData := extids[2][32:]
	err2 := primitives.VerifySignature(content, pubKey, signData)
	if err2 != nil {
		fmt.Printf("%v \n", err2)
		return nil, graderStake.NewValidateError("invalid signature")
	}

	if o.Height != height {
		return nil, graderStake.NewValidateError("invalid height")
	}

	// verify assets
	if len(o.Assets) != len(opr.V5Assets) {
		return nil, graderStake.NewValidateError("invalid assets")
	}
	for _, val := range o.Assets {
		if val == 0 {
			return nil, graderStake.NewValidateError("assets must be greater than 0")
		}
	}

	if err := factoidaddress.Valid(o.Address); err != nil {
		return nil, graderStake.NewValidateError(fmt.Sprintf("factoidaddress is invalid : %s", err.Error()))
	}

	/**
	 *  Set GradingDelegatedSPR object
	 */
	gspr := new(GradingDelegatedSPR)
	gspr.EntryHash = entryhash
	gspr.CoinbaseAddress = o.Address
	sha := sha256.Sum256(content)
	gspr.SPRHash = sha[:]
	gspr.SPR = o
	gspr.balanceOfPEG = balanceOfPEG

	return gspr, nil
}

// S4 grading works similar to V1 but the grade is banded
// meaning a record within `band` percentage is considered to be equal
func gradeS4(avg []float64, spr *GradingDelegatedSPR, band float64) float64 {
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

// calculate the vector of average prices
func averageS4(sprs []*GradingDelegatedSPR) []float64 {
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
		avg[i] = graderStake.TrimmedMeanFloat(data[i], length+1)
	}
	return avg
}
