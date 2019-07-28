// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package opr

import (
	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/common"
	"github.com/zpatrick/go-config"
)

var FctDbht int64

func UpdateBurns(c *config.Config) {

	network, err := c.String("Miner.Network")
	if err != nil {
		panic("cannot find the network designation for updating burn txs")
	}
	net := "test"

	if network == "MainNet" {
		net = "main"
	} else if network != "TestNet" {
		panic("unknown network found when updating burn txs")
	}
	_ = net

	if len(OPRBlocks) == 0 {
		return // There is nothing to do if there is no OPR chain with valid OPR blocks
	}
	if FctDbht == 0 {
		FctDbht = OPRBlocks[0].Dbht
	}
	for i := 1; ; i++ {
		fc, _, _ := factom.GetFBlockByHeight(int64(i))
		if fc == nil {
			break
		}
		for _, txid := range fc.Transactions {
			tx, _ := factom.GetTransaction(txid.TxID)
			ftx := tx.FactoidTransaction.(map[string]interface{})
			if ftx == nil {
				continue
			}

			switch {
			case len(ftx["inputs"].([]interface{})) != 1,
				len(ftx["outputs"].([]interface{})) != 0,
				len(ftx["outecs"].([]interface{})) != 1,
				ftx["outecs"].([]interface{})[0].(map[string]interface{})["useraddress"] != common.BurnAddresses[net]:
			default:
				fct := ftx["inputs"].([]interface{})[0].(map[string]interface{})["useraddress"].(string)
				amt := ftx["inputs"].([]interface{})[0].(map[string]interface{})["amount"].(float64)
				_ = fct
				_ = amt
				var pFct string

				pFct, err := common.ConvertFCTtoPegNetAsset(network, fct, "FCT")

				if err != nil {
					continue
				}
				if net == "main" {
					_ = AddToBalance(pFct, int64(amt))

				} else if net == "test" {
					_ = AddToBalance(pFct, int64(amt)*1000)
				}
			}
			//     tx.FactoidTransaction["inputs"] !=nil,
			//
			//len(tx.Inputs) != 1,
			//	//	len(tx.Outputs) != 0,
			//	len(tx.ECOutputs) != 1,
			//	//	tx.ECOutputs[0].Amount != 0,
			//	tx.ECOutputs[0].Address != common.BurnAddresses[net+"RCD"]:
			//default:
			//	address := tx.Inputs[0].Address
			//	pFCT, err := common.ConvertFCTtoPNT(network, address)
			//	if err != nil {
			//		panic("FCT address conversion to pFCT error should not happen")
			//	}
			//	err = AddToBalance(pFCT, int64(tx.Inputs[0].Amount))
			//	if err != nil {
			//		panic("pFCT balance update errors should not happen")
			//	}
			//}

			//BlockHeight    uint32          `json:"blockheight,omitempty"`
			//FeesPaid       uint64          `json:"feespaid,omitempty"`
			//FeesRequired   uint64          `json:"feesrequired,omitempty"`
			//IsSigned       bool            `json:"signed"`
			//Name           string          `json:"name,omitempty"`
			//Timestamp      time.Time       `json:"timestamp"`
			//TotalECOutputs uint64          `json:"totalecoutputs"`
			//TotalInputs    uint64          `json:"totalinputs"`
			//TotalOutputs   uint64          `json:"totaloutputs"`
			//Inputs         []*TransAddress `json:"inputs"`
			//Outputs        []*TransAddress `json:"outputs"`
			//ECOutputs      []*TransAddress `json:"ecoutputs"`
			//TxID           string          `json:"txid,omitempty"`

		}
		FctDbht = int64(i)
	}
}
