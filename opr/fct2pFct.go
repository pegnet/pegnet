// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package opr

import (
	"github.com/FactomProject/factom"
	"github.com/golangci/golangci-lint/pkg/config"
)

git mervar FctDbht int64

func UpdateBurns(config *config.Config) {
	GetEntryBlocks(config)			// make sure we are up to date on Entry Blocks
	if len(OPRBlocks) == 0 {
		return // There is nothing to do if there is no OPR chain with valid OPR blocks
	}
	if FctDbht == 0 {
		FctDbht = OPRBlocks[0].Dbht
	}
	for i:= FctDbht+1;;i++ {
		fc,_,err := factom.GetFBlockByHeight(i)
		if err != nil || fc == nil {
			break
		}
		for _, tx := range fc.Transactions {
			switch {
			case len(tx.Inputs) != 1 ,
			len(tx.Outputs)>0,
			len(tx.ECOutputs) != 1,
			tx.ECOutputs[0].Amount != 0,
			tx.ECOutputs[0].Address !=

			}
		}
	}
}