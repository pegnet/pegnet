// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package main

import (
	"encoding/hex"
	"fmt"
	"github.com/pegnet/pegnet/common"
	"os/user"
	"time"

	"github.com/FactomProject/factom"
	"github.com/zpatrick/go-config"
)

func main() {
	factom.SetFactomdServer("localhost:8088")
	factom.SetWalletServer("localhost:8089")

	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	userPath := u.HomeDir
	configfile := fmt.Sprintf("%s/.%s/defaultconfig.ini", userPath, "pegnet")
	iniFile := config.NewINIFile(configfile)
	Config := config.NewConfig([]config.Provider{iniFile})

	// Init Logs
	common.InitLogs(Config)

	protocol, err := Config.String("Miner.Protocol")
	if err != nil {
		panic(err)
	}
	network, err := Config.String("Miner.Network")

	sECAdr, err := Config.String("Miner.ECAddress")
	ecAdr, err := factom.FetchECAddress(sECAdr)
	if err != nil {
		common.Logf("OPR", "Failed to initialize EC Address %v ", err)
		return
	}
	_ = ecAdr

	CreateChain(ecAdr, [][]byte{[]byte(protocol), []byte(network)})
	CreateChain(ecAdr, [][]byte{[]byte(protocol), []byte(network), []byte("Transactions")})
	CreateChain(ecAdr, [][]byte{[]byte(protocol), []byte(network), []byte("Miners")})
	CreateChain(ecAdr, [][]byte{[]byte(protocol), []byte(network), []byte("Oracle Price Records")})

}

func CreateChain(ec_adr *factom.ECAddress, chainName [][]byte) (txid string, err error) {
	fmt.Print("Looking to create the chain: ")
	cn := string(chainName[0])
	for i, n := range chainName[1:] {
		cn = cn + " --- " + string(n)
	}
	common.Logf("CreateChain", "Chain Name %s", cn)

	chainID := common.ComputeChainIDFromFields(chainName)          // Compute the chainID
	chainExists := factom.ChainExists(hex.EncodeToString(chainID)) // Check if it exists
	if chainExists {                                               // If the chain exists, we are done.
		common.Logf("CreateChain", "Chain Exists!")
		return
	}

	entry := factom.Entry{ChainID: hex.EncodeToString(chainID), ExtIDs: chainName, Content: []byte{}}
	newChain := factom.NewChain(&entry)
	var err1, err2 error
	for i := 0; i < 1000; i++ {
		if i == 0 {
			common.Logf("CreateChain", "Creating the Chain")
		} else {
			common.Logf("CreateChain", "Something went wrong.  Waiting 5 seconds to retry (%d)", i*5)
			time.Sleep(5 * time.Second)
		}
		if i == 0 || err1 != nil {
			_, err1 = factom.CommitChain(newChain, ec_adr)
		}
		if i == 0 || err2 == nil {
			txid, err2 = factom.RevealChain(newChain)
		}
		if err1 == nil && err2 == nil { // Success?  Go to reveal!
			break
		}
	}
	if err1 != nil {
		common.Logf("error", "CreateChain Failed: %v", err1)
		return "", err1
	}
	if err2 != nil {
		common.Logf("error", "CreateChain Failed: %v", err2)
		return "", err2
	}
	return txid, nil
}
