package balances

import (
	"encoding/json"
	"fmt"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/common"
	"github.com/zpatrick/go-config"
)

type BurnTracking struct {
	FctDbht  int64
	Balances *BalanceTracker
}

func NewBurnTracking(balanceTracker *BalanceTracker) *BurnTracking {
	b := new(BurnTracking)
	b.Balances = balanceTracker

	return b
}

func (b *BurnTracking) UpdateBurns(c *config.Config, startBlock int64) error {
	network, err := common.LoadConfigNetwork(c)
	if err != nil {
		panic("cannot find the network designation for updating burn txs")
	}

	if b.FctDbht == 0 {
		b.FctDbht = startBlock
	}

	heights, err := factom.GetHeights()
	if err != nil {
		return err
	}

	for i := b.FctDbht + 1; i < heights.DirectoryBlockHeight; i++ {
		deltas := make(map[string]int64)

		fc, _, err := factom.GetFBlockByHeight(i)
		if err != nil {
			return err
		}
		if fc == nil {
			return fmt.Errorf("fblock is nil")
		}

		for _, txid := range fc.Transactions {
			txInterface, err := factom.GetTransaction(txid.TxID)
			if err != nil {
				return err
			}

			txData, err := json.Marshal(txInterface.FactoidTransaction)
			if err != nil {
				return err
			}

			tx := new(FactoidTransaction)
			err = json.Unmarshal(txData, tx)
			if err != nil {
				return err
			}

			// Is this a burn?
			if len(tx.Outecs) == 1 && tx.Outecs[0].Useraddress == common.BurnAddresses[network] && tx.Outecs[0].Amount == 0 {
				// The output is a burn. Let's check some other properties
				if len(tx.Outputs) > 0 || len(tx.Inputs) > 1 {
					continue // must only have 1 output, and 1 input, being the burn
				}

				burnAmt := tx.Inputs[0].Amount
				pFct, err := common.ConvertFCTtoPegNetAsset(network, "FCT", tx.Inputs[0].Useraddress)
				if err != nil {
					return err
				}
				if network == common.MainNetwork {
					deltas[pFct] += int64(burnAmt)

				} else if network == common.TestNetwork {
					deltas[pFct] += int64(burnAmt) * 1000
				}

			}
		}

		// Process them as a block
		for pFct, delta := range deltas {
			_ = b.Balances.AddToBalance(pFct, delta)
		}
		b.FctDbht = i
	}
	return nil
}

type FactoidTransaction struct {
	Millitimestamp int64               `json:"millitimestamp"`
	Inputs         []TransactionOutput `json:"inputs"`
	Outputs        []TransactionOutput `json:"outputs"`
	Outecs         []TransactionOutput `json:"outecs"`
	Rcds           []string            `json:"rcds"`
	Sigblocks      []struct {
		Signatures []string `json:"signatures"`
	} `json:"sigblocks"`
	Blockheight int `json:"blockheight"`
}

type TransactionOutput struct {
	Amount      int64  `json:"amount"`
	Address     string `json:"address"`
	Useraddress string `json:"useraddress"`
}
