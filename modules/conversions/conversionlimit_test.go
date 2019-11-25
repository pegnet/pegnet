package conversions_test

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	. "github.com/pegnet/pegnet/modules/conversions"
	"github.com/pegnet/pegnet/modules/transactionid"
)

func TestPayout(t *testing.T) {
	type Vector struct {
		Requested      uint64
		TotalRequested uint64
		Bank           uint64
		Payout         uint64
	}

	vecs := []Vector{
		{0, 0, 0, 0},
		{2500, 5000, 5000, 2500},
		{2500, 50000, 5000, 250},
		{2500, 500000, 5000, 25},
		{2500, 5000000, 5000, 2},
		{2500, 50000000, 5000, 0},
	}

	for i, v := range vecs {
		if pay := Payout(v.Requested, v.Bank, v.TotalRequested); pay != v.Payout {
			t.Errorf("Vector %d has payout of %d, exp %d", i, pay, v.Payout)
		}
	}

}

func TestNewConversionSupply(t *testing.T) {
	testBank := func(bank uint64) {
		s := NewConversionSupply(bank)
		totalReq := new(big.Int)

		for i := 0; i < rand.Intn(100); i++ {
			eHash := fmt.Sprintf("%032d", i)
			amt := rand.Uint64()
			totalReq.Add(totalReq, new(big.Int).SetUint64(amt))
			err := s.AddConversion(transactionid.FormatTxID(0, eHash), amt)
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
		}

		pays := s.Payouts()
		var totalPay uint64
		for _, a := range pays {
			totalPay += a
		}
		if totalPay > bank {
			t.Errorf("total paid %d, that is over the bank of %d", totalPay, bank)
		}

		if totalReq.IsUint64() && totalReq.Uint64() < bank {
			if totalPay != totalReq.Uint64() {
				t.Errorf("[under bank] exp %d total pay, found %d", totalReq.Uint64(), totalPay)
			}
		} else {
			if totalPay != bank {
				t.Errorf("[over bank] exp %d total pay, found %d", bank, totalPay)
			}
		}

	}

	t.Run("ensure payouts never exceed limit (5K)", func(t *testing.T) {
		for i := 0; i < 10000; i++ {
			testBank(PerBlock)
		}
	})

	t.Run("ensure payouts never exceed limit (Random)", func(t *testing.T) {
		for i := 0; i < 10000; i++ {
			// 10 Mil limit
			testBank(rand.Uint64() % (1e7 * 1e8))
		}
	})

	t.Run("ensure payouts never exceed limit (low bank)", func(t *testing.T) {
		for i := 0; i < 10000; i++ {
			// Low limits
			testBank(rand.Uint64() % PerBlock)
		}
	})

}
