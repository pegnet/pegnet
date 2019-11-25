package conversions

import (
	"fmt"
	"math/big"

	"github.com/pegnet/pegnet/modules/transactionid"
)

const (
	PerBlock uint64 = 5000 * 1e8
)

// ConversionSupplySet indicates the total amount of PEG allowed to be converted
// per block. This amount is currently set to 5,000 PEG, matching the miner
// amount per block.
// All amounts for interacting with this struct should be in PEG.
type ConversionSupplySet struct {
	Bank uint64
	// Key = txid, Value = PegAmount
	ConversionRequests map[string]uint64
	totalRequested     *big.Int
}

func NewConversionSupply(bank uint64) *ConversionSupplySet {
	c := new(ConversionSupplySet)
	c.Bank = bank
	c.ConversionRequests = make(map[string]uint64)
	c.totalRequested = new(big.Int)

	return c
}

func (s *ConversionSupplySet) AddConversion(txid string, pegAmt uint64) error {
	if _, ok := s.ConversionRequests[txid]; ok {
		return fmt.Errorf("txid already exits in the this set")
	}
	s.ConversionRequests[txid] = pegAmt
	s.totalRequested.Add(s.totalRequested, new(big.Int).SetUint64(pegAmt))
	return nil
}

// Payouts returns the amount of PEG to allow each Tx to convert.
func (s *ConversionSupplySet) Payouts() map[string]uint64 {
	payouts := make(map[string]uint64)
	if len(s.ConversionRequests) == 0 {
		return payouts
	}

	if s.totalRequested.IsUint64() && s.totalRequested.Uint64() < s.Bank {
		for txid, c := range s.ConversionRequests {
			payouts[txid] = c
		}
		// No work necessary
		return payouts
	}

	var totalPaid uint64
	for txid, c := range s.ConversionRequests {
		payouts[txid] = PayoutBig(c, s.Bank, s.totalRequested)
		totalPaid += payouts[txid]
	}

	dust := s.Bank - totalPaid
	// Dust goes to the highest request, and ties go to highest entryhash
	// Let's find the highest
	var most uint64
	var top []string

	for txid, amt := range s.ConversionRequests {
		if amt > most {
			top = []string{txid}
		} else if amt == most {
			top = append(top, txid)
		}
	}

	if len(top) == 1 {
		payouts[top[0]] += dust
	} else {
		// More than 1 with the same top amount, highest entryhash wins
		// Sort sorts them with the lowest entryhash first
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
