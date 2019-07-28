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
	for i := FctDbht; ; i++ {
		fc, _, _ := factom.GetFBlockByHeight(i)
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
			case len(ftx["inputs"].([]interface{})) != 1, // This is ugly as I code around some issues in the
				len(ftx["outputs"].([]interface{})) != 0, // factom go library
				len(ftx["outecs"].([]interface{})) != 1,
				ftx["outecs"].([]interface{})[0].(map[string]interface{})["useraddress"] != common.BurnAddresses[net]:
			default:
				fct := ftx["inputs"].([]interface{})[0].(map[string]interface{})["useraddress"].(string)
				amt := ftx["inputs"].([]interface{})[0].(map[string]interface{})["amount"].(float64)

				pFct, err := common.ConvertFCTtoPegNetAsset(network, "FCT", fct)

				if err != nil {
					continue
				}
				if net == "main" {
					_ = AddToBalance(pFct, int64(amt))

				} else if net == "test" {
					_ = AddToBalance(pFct, int64(amt)*1000)
				}
				//log.Printf("Updated address %s balance == %f\n", pFct, float64(GetBalance(pFct))/100000000)
			}
		}
		FctDbht = i
	}
}
