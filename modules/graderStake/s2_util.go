package graderStake

import (
	"crypto/sha256"
	"fmt"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/pegnet/pegnet/modules/factoidaddress"
	"github.com/pegnet/pegnet/modules/opr"
	"github.com/pegnet/pegnet/modules/spr"
)

// ValidateS2 validates the provided data using the specified parameters
func ValidateS2(entryhash []byte, extids [][]byte, height int32, content []byte) (*GradingSPR, error) {
	if len(entryhash) != 32 {
		return nil, NewValidateError("invalid entry hash length")
	}

	if len(extids) != 3 {
		return nil, NewValidateError("invalid extid count")
	}

	if len(extids[0]) != 1 || extids[0][0] != 6 {
		return nil, NewValidateError("invalid version")
	}

	// ParseS1Content parses the V2 proto format
	// S1 is just the proto format with some more assets.
	o2, err := spr.ParseS1Content(content)
	if err != nil {
		return nil, NewDecodeError(err.Error())
	}
	o := &spr.S1Content{V2Content: *o2}

	// Verify Signature
	if len(extids[2]) != 96 {
		return nil, NewValidateError("invalid signature length")
	}
	pubKey := extids[2][:32]
	signData := extids[2][32:]
	err2 := primitives.VerifySignature(content, pubKey, signData)
	if err2 != nil {
		fmt.Printf("%v \n", err2)
		return nil, NewValidateError("invalid signature")
	}

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
