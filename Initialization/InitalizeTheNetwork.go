package main

import (
	"encoding/hex"
	"fmt"
	"os/user"
	"time"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/support"
	config "github.com/zpatrick/go-config"
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
	protocol, err := Config.String("Miner.Protocol")
	if err != nil {
		panic(err)
	}
	network, err := Config.String("Miner.Network")

	sECAdr, err := Config.String("Miner.ECAddress")
	ecAdr, err := factom.FetchECAddress(sECAdr)
	if err != nil {
		fmt.Println("Failed to initialize EC Address")
		fmt.Println(err.Error())
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
	for i, n := range chainName {
		if i > 0 {
			fmt.Print(" --- ")
		}
		fmt.Print(string(n))
	}
	fmt.Println()
	chainID := support.ComputeChainIDFromFields(chainName)         // Compute the chainID
	chainExists := factom.ChainExists(hex.EncodeToString(chainID)) // Check if it exists
	if chainExists {                                               // If the chain exists, we are done.
		fmt.Println("Chain Exists!")
		return
	}

	entry := factom.Entry{ChainID: hex.EncodeToString(chainID), ExtIDs: chainName, Content: []byte{}}
	newChain := factom.NewChain(&entry)
	var err1, err2 error
	for i := 0; i < 1000; i++ {
		if i == 0 {
			fmt.Println("Creating the Chain")
		} else {
			fmt.Println("Something went wrong.  Waiting 5 seconds to retry (", i*5, ")")
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
		fmt.Println("Failed: ", err1.Error())
		return "", err1
	}
	if err2 != nil {
		fmt.Println("Failed: ", err2.Error())
		return "", err2
	}
	fmt.Println("Success!")
	return txid, nil
}
