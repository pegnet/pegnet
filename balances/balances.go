package balances

import (
	"errors"
	"fmt"
	"sync"

	"github.com/pegnet/pegnet/common"
)

type BalanceTracker struct {
	// Balances holds the non-zero balances for every address for every token in a
	// two dimensional map:
	// assetname => { RCD-hash => balance }
	Balances map[string]map[[32]byte]int64
	sync.Mutex
}

func NewBalanceTracker() *BalanceTracker {
	b := new(BalanceTracker)
	b.Balances = make(map[string]map[[32]byte]int64)

	return b
}

// ConvertAddress takes a human-readable address and extracts the prefix and RCD hash
func ConvertAddress(address string) (prefix string, adr [32]byte, err error) {
	prefix, adr2, err := common.ConvertPegNetAssetToRaw(address)
	if err != nil {
		return
	}
	copy(adr[:], adr2)
	return
}

// AddToBalance adds the given value to the human-readable address
// Note that add to balance takes a signed update, so this can be used to add to or
// subtract from a balance.  An error is returned if the value would drive the balance
// negative.  Or if the string doesn't represent a valid token
func (b *BalanceTracker) AddToBalance(address string, value int64) (err error) {
	prefix, addressBytes, err := ConvertAddress(address)
	if err != nil {
		return errors.New("address not properly formed")
	}

	b.Lock()
	defer b.Unlock()
	if _, ok := b.Balances[prefix]; !ok {
		b.Balances[prefix] = make(map[[32]byte]int64)
	}

	prev := b.Balances[prefix][addressBytes]
	if prev+value < 0 {
		return fmt.Errorf("result would be less than zero %d-%d", prev, -value)
	}
	b.Balances[prefix][addressBytes] = prev + value

	//log.WithFields(log.Fields{
	//	"address":       address,
	//	"prefix":        prefix,
	//	"address_bytes": hex.EncodeToString(addressBytes[:]),
	//	"prev_balance":  prev,
	//	"value":         value,
	//	"new_balance":   prev + value,
	//}).Debug("Add to balance")
	return
}

// GetBalance returns the balance for a PegNet asset.
// If the address is invalid, a -1 is returned
func (b *BalanceTracker) GetBalance(address string) (balance int64) {
	prefix, adr, err := ConvertAddress(address)
	if err != nil {
		return -1
	}
	b.Lock()
	defer b.Unlock()
	balance = b.Balances[prefix][adr]
	return
}

// DiagnosticAssetHumanReadablePEGBalances is used to monitor the PEG distribution. It's mainly a diagnosing/debugging
// function.
func (b *BalanceTracker) DiagnosticAssetHumanReadablePEGBalances(prefix string) map[string]string {
	b.Lock()
	defer b.Unlock()
	r := make(map[string]string)
	total := int64(0)
	for k, v := range b.Balances[prefix] {
		r[common.ConvertRawToFCT(k[:])] = fmt.Sprintf("%d", v/1e8)
		total += v / 1e8
	}

	for k, v := range b.Balances[prefix] {
		r[common.ConvertRawToFCT(k[:])+"%"] = fmt.Sprintf("%.2f%%", float64(v/1e8)/float64(total)*100)
	}

	r["all"] = fmt.Sprintf("%d", total)
	return r
}
