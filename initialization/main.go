// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package main

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os/user"
	"strings"
	"time"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/common"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

var (
	// Global Flags
	LogLevel        string // Logrus global log level
	FactomdFlag     string
	WalletdFlag     string
	ECAddressString string
)

func init() {
	flag.StringVar(&LogLevel, "log", "info", "Change the logging level. Can choose from 'debug', 'info', 'warn', 'error', or 'fatal'")
	flag.StringVar(&FactomdFlag, "s", "", "IPAddr:port# of factomd API to use to access blockchain (default localhost:8088)")
	flag.StringVar(&WalletdFlag, "w", "", "IPAddr:port# of factom-walletd API to use to create transactions (default localhost:8089)")
	flag.StringVar(&ECAddressString, "ec", "", "EC Address to use in place of the one specified in PegNet config file")
}

func main() {
	// Config file parsing
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	userPath := u.HomeDir
	configFile := fmt.Sprintf("%s/%s/defaultconfig.ini", userPath, ".pegnet")
	iniFile := config.NewINIFile(configFile)
	Config := config.NewConfig([]config.Provider{iniFile})
	protocol, err := Config.String("Miner.Protocol")
	if err != nil {
		panic(err)
	}
	network, err := common.LoadConfigNetwork(Config)
	if err != nil {
		panic(err)
	}
	configECAddress, err := Config.String("Miner.ECAddress")
	if err != nil {
		panic(err)
	}

	// CLI overrides
	flag.Parse()

	FactomdLocation, _ := Config.String("Miner.FactomdLocation")
	WalletdLocation, _ := Config.String("Miner.WalletdLocation")

	if len(FactomdFlag) > 0 {
		FactomdLocation = FactomdFlag // flag override
	}
	if FactomdLocation == "" {
		FactomdLocation = "localhost:8088"
	}

	if len(WalletdFlag) > 0 {
		WalletdLocation = WalletdFlag // flag override
	}
	if WalletdLocation == "" {
		WalletdLocation = "localhost:8089"
	}

	factom.SetFactomdServer(FactomdLocation)
	factom.SetWalletServer(WalletdLocation)

	switch strings.ToLower(LogLevel) {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	}

	if ECAddressString == "" {
		ECAddressString = configECAddress
	}
	ecAddress, err := factom.FetchECAddress(ECAddressString)
	if err != nil {
		log.WithFields(log.Fields{
			"error":      err.Error(),
			"ec_address": ECAddressString,
		}).Fatal("Failed to fetch EC Address")
	}

	// Try to create the network chains
	chainNames := map[string][][]byte{
		"ProtocolChain":          {[]byte(protocol), []byte(network)},
		"TransactionChain":       {[]byte(protocol), []byte(network), []byte(common.TransactionChainTag)},
		"MinerChain":             {[]byte(protocol), []byte(network), []byte(common.MinerChainTag)},
		"OraclePriceRecordChain": {[]byte(protocol), []byte(network), []byte(common.OPRChainTag)},
		"StakingPriceRecordChain": {[]byte(protocol), []byte(network), []byte(common.SPRChainTag)},
	}
	for tag, chainName := range chainNames {
		chainID, txID, err := CreateChain(ecAddress, chainName)
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

func CreateChain(ecAddress *factom.ECAddress, chainName [][]byte) (chainID string, txID string, err error) {
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
