// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package opr

import (
	"fmt"

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

	if len(OPRBlocks) == 0 {
		return // There is nothing to do if there is no OPR chain with valid OPR blocks
	}
	if FctDbht == 0 {
		FctDbht = OPRBlocks[0].Dbht
	}
	for i := FctDbht + 1; ; i++ {
		fc, _, err := factom.GetFBlockByHeight(i)
		if err != nil || fc == nil {
			break
		}
		for _, tx := range fc.Transactions {
			fmt.Println(tx.String())
			switch {
			case len(tx.Inputs) != 1,
				len(tx.Outputs) != 0,
				len(tx.ECOutputs) != 1,
				tx.ECOutputs[0].Amount != 0,
				tx.ECOutputs[0].Address != common.BurnAddresses[net+"RCD"]:
			default:
				address := tx.Inputs[0].Address
				pFCT, err := common.ConvertFCTtoPNT(network, address)
				if err != nil {
					panic("FCT address conversion to pFCT error should not happen")
				}
				err = AddToBalance(pFCT, int64(tx.Inputs[0].Amount))
				if err != nil {
					panic("pFCT balance update errors should not happen")
				}
			}
		}
		FctDbht = i
	}
}
