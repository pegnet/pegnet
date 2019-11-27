package conversions

import (
	"fmt"
	"math/big"

	"github.com/pegnet/pegnet/modules/transactionid"
)

const (
	// PEG allocated per block for conversions (not including the bank)
	PerBlock uint64 = 5000 * 1e8
)

// ConversionSupplySet indicates the total amount of PEG allowed to be converted
// into per block. This amount is currently set to 5,000 PEG, matching
// matching the miner amount per block. (not including the bank)
// All amounts for interacting with this struct should be in PEG.
//
// All conversion requests are conversions INTO PEG. So if PEG is $0.50,
// then a conversion request of 10 pUSD -> PEG, would be
// ConversionRequests[txid] = 20*1e8
type ConversionSupplySet struct {
	Bank uint64 // Can be set to any positive number. Set to 0 if negative
	// Key = txid, Value = PegAmount requested in the conversion
	ConversionRequests map[string]uint64
	totalRequested     *big.Int
}

// NewConversionSupply will allocate up to the bank amount of PEG based
// on the proportions requested
func NewConversionSupply(bank uint64) *ConversionSupplySet {
	c := new(ConversionSupplySet)
	c.Bank = bank
	c.ConversionRequests = make(map[string]uint64)
	c.totalRequested = new(big.Int)

	return c
}

// AddConversion will add a PEG conversion request to the set.
// All conversion requests will be pXXX -> PEG
// The `pegAmt` is the amount of PEG the total conversion would yield.
// Because of the supply limit, this conversion request might not have
// 100% yield.
func (s *ConversionSupplySet) AddConversion(txid string, pegAmt uint64) error {
	if _, _, err := transactionid.VerifyTransactionHash(txid); err != nil {
		return err
	}

	if _, ok := s.ConversionRequests[txid]; ok {
		return fmt.Errorf("txid already exists in the this set")
	}
	s.ConversionRequests[txid] = pegAmt
	s.totalRequested.Add(s.totalRequested, new(big.Int).SetUint64(pegAmt))
	return nil
}

// Payouts returns the amount of PEG to allow each Tx to convert into.
// This is the actual PEG yield of each conversion request (pXXX -> PEG)
func (s *ConversionSupplySet) Payouts() map[string]uint64 {
	payouts := make(map[string]uint64)
	if len(s.ConversionRequests) == 0 {
		return payouts // No one to pay. That was easy
	}

	// If the total requested is less than the bank, we can fill the orders
	// with exactly what they want.
	if s.totalRequested.IsUint64() && s.totalRequested.Uint64() < s.Bank {
		for txid, c := range s.ConversionRequests {
			payouts[txid] = c
		}
		// No work necessary
		return payouts
	}

	var totalPaid uint64
	for txid, c := range s.ConversionRequests {
		// PayoutBig pays out proportionally to their requested amount.
		payouts[txid] = PayoutBig(c, s.Bank, s.totalRequested)
		totalPaid += payouts[txid]
	}

	// The function should stop here, but we have some dust. In order
	// to make the inflation "even" and not "4,999.99999997", we account
	// for the dust. So we need to allocate the dust to a lucky winner.
	dust := s.Bank - totalPaid
	// Dust goes to the highest request, and ties go to highest entryhash
	// Let's find the highest
	var most uint64
	var top []string

	for txid, amt := range s.ConversionRequests {
		if amt > most {
			top = []string{txid}
			most = amt
		} else if amt == most {
			// Tied for the highest amount requested
			top = append(top, txid)
		}
	}

	if len(top) == 1 {
		// Only 1 top requester.
		payouts[top[0]] += dust
	} else {
		// More than 1 with the same top amount, highest entryhash wins
		// Sort sorts them with the lowest entryhash first
		// If two conversions are in the same entryhash, the lowest txindex
		// wins.
		top = transactionid.SortTxIDS(top)
		payouts[top[0]] += dust
	}

	return payouts
}

// PayoutBig, denoted p(c), is the payout amount a requested peg amount
// can receive.
func PayoutBig(requested, bank uint64, totalRequested *big.Int) uint64 {
	if requested == 0 || bank == 0 || (totalRequested.IsUint64() && totalRequested.Uint64() == 0) {
		// Requested 0, means 0 payout
		return 0
	}

	in := new(big.Int).SetUint64(requested)
	b := new(big.Int).SetUint64(bank)
	in = in.Mul(in, b)

	res := new(big.Int).Quo(in, totalRequested)
	return res.Uint64()
}

// Payout, denoted p(c), is the payout amount a requested peg amount
// can receive.
func Payout(requested, bank uint64, totalRequested uint64) uint64 {
	t := new(big.Int).SetUint64(totalRequested)
	return PayoutBig(requested, bank, t)
}

// Refund calculates the refund based on the input amount and pegYield.
// The refund for a pXXX -> PEG conversion is in the original asset units.
// It takes the yielded peg and the rates to determine how much of the
// original asset to return.
// Params:
//	inputAmount 	Original pXXX asset amount
//	pegYield		Amount of PEG allocated in the conversion
//	inputRate		pUSD rate of the original asset
//	pegRate			pUSD rate of PEG
func Refund(inputAmount, pegYield int64, inputRate, pegRate uint64) int64 {
	consumedInput, _ := Convert(pegYield, pegRate, inputRate)
	refund := inputAmount - consumedInput
	return refund
}
