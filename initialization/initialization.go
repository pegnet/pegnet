package initialization

// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import (
	"encoding/hex"
	"errors"
	"time"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/common"
	log "github.com/sirupsen/logrus"
)

// CreateProtocolChains will attempt to create the set of network chains from the variables set in the config file
func CreateProtocolChains(protocolName string, networkName string, ecAddress string) {
	ec, err := factom.FetchECAddress(ecAddress)
	if err != nil {
		log.WithFields(log.Fields{
			"error":      err.Error(),
			"ec_address": ecAddress,
		}).Fatal("Failed to fetch EC Address")
	}

	chainNames := map[string][][]byte{
		"ProtocolChain":          {[]byte(protocolName), []byte(networkName)},
		"TransactionChain":       {[]byte(protocolName), []byte(networkName), []byte("Transactions")},
		"MinerChain":             {[]byte(protocolName), []byte(networkName), []byte("Miners")},
		"OraclePriceRecordChain": {[]byte(protocolName), []byte(networkName), []byte("Oracle Price Records")},
	}
	for tag, chainName := range chainNames {
		chainID, txID, err := createChain(ec, chainName)
		if err != nil {
			log.WithFields(log.Fields{
				"name":     tag,
				"chain_id": chainID,
				"error":    err.Error(),
			}).Fatal("Failed to create chain")
		} else if txID == "" {
			log.WithFields(log.Fields{
				"name":     tag,
				"chain_id": chainID,
			}).Warn("Chain already exits")
		} else {
			log.WithFields(log.Fields{
				"name":     tag,
				"chain_id": chainID,
				"tx_id":    txID,
			}).Info("Created chain")
		}
	}
}

func createChain(ecAddress *factom.ECAddress, chainName [][]byte) (chainID string, txID string, err error) {
	if len(chainName) == 0 {
		return "", "", errors.New("chain name must be at least length 1")
	}

	chainIDBytes := common.ComputeChainIDFromFields(chainName)
	chainID = hex.EncodeToString(chainIDBytes)
	if factom.ChainExists(chainID) {
		return chainID, "", nil
	}

	entry := factom.Entry{ChainID: chainID, ExtIDs: chainName, Content: []byte{}}
	newChain := factom.NewChain(&entry)
	var commitErr, revealErr error
	for i := 0; i < 1000; i++ {
		if i == 0 || commitErr != nil {
			_, commitErr = factom.CommitChain(newChain, ecAddress)
		}
		if i == 0 || revealErr == nil {
			txID, revealErr = factom.RevealChain(newChain)
		}

		if commitErr == nil && revealErr == nil {
			break
		} else {
			log.WithFields(log.Fields{
				"iteration":    i,
				"chain_id":     chainID,
				"commit_error": commitErr,
				"reveal_error": revealErr,
			}).Debug("Failed to create chain. Retrying in 5 seconds")
			time.Sleep(5 * time.Second)
		}
	}
	if commitErr != nil {
		return chainID, "", commitErr
	}
	if revealErr != nil {
		return chainID, "", revealErr
	}
	return chainID, txID, nil
}
