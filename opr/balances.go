package opr

import (
	"errors"
	"fmt"
	"github.com/pegnet/pegnet/common"
)

var Balances map[string]map[[32]byte]int64

func init() {
	Balances = make(map[string]map[[32]byte]int64)
	for _, asset := range common.AssetNames {
		Balances["p"+asset] = make(map[[32]byte]int64)
		Balances["t"+asset] = make(map[[32]byte]int64)
	}
}

func ConvertAddress(address string) (prefix string, adr [32]byte, err error) {
	prefix, adr2, err := common.ConvertPegAddrToRaw(address)
	if err != nil {
		return
	}
	copy(adr[:], adr2)
	return
}

// AddToBalance()
// Note that add to balance takes a signed update, so this can be used to add to or
// subtract from a balance.  An error is returned if the value would drive the balance
// negative.  Or if the string doesn't represent a valid token
func AddToBalance(address string, value int64) (err error) {
	prefix, adr, err := ConvertAddress(address)
	if err != nil {
		return errors.New("address not properly formed")
	}
	prev := Balances[prefix][adr]
	if prev+value < 0 {
		return fmt.Errorf("result would be less than zero %d-%d", prev, -value)
	}
	Balances[prefix][adr] = prev + value
	return
}

// GetBalance()
// Returns the balance for a PegNet asset.  If the address is invalid, a -1 is returned
func GetBalance(address string) (balance int64) {
	prefix, adr, err := ConvertAddress(address)
	if err != nil {
		return -1
	}
	balance = Balances[prefix][adr]
	return
}
