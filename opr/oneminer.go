package opr

import (
	"github.com/zpatrick/go-config"
	"github.com/pegnet/OracleRecord/support"
	"fmt"
	"github.com/FactomProject/btcutil/base58"
	"encoding/hex"
	"encoding/json"
	"github.com/FactomProject/factom"
)

func OneMiner(verbose bool, config *config.Config, monitor *support.FactomdMonitor, miner int) {
	alert := monitor.GetAlert()
	mining := false
	var opr *OraclePriceRecord
	for {
		fds := <-alert
		if verbose {
			fmt.Println(fds.Dbht, " ", fds.Minute)
		}
		switch fds.Minute {
		case 1:
			if !mining {
				mining = true
					opr, _ = NewOpr(miner, fds.Dbht, config)
					opr.GetOPRecord(config)
				go opr.Mine(int64(miner)+int64(fds.Dbht)+fds.Minute, verbose)
			}
		case 9:
			if mining {
				opr.StopMining <- 0
				if verbose {
					fmt.Println(opr.String())
				}
				mining = false

				writeMiningRecord(opr)
			}
		}
	}
}

func writeMiningRecord(opr *OraclePriceRecord) {
	// Encode the OPR ChainID to a hex string
	chainid := hex.EncodeToString(base58.Decode(opr.OPRChainID))
	// Get the binary representation of the opr
	bOPR, err := json.Marshal(opr)
	if err != nil {
		panic(err)
	}

	fmt.Println("Difficulty ", opr.ComputeDifficulty(LX.Hash(bOPR), opr.BestNonce))

	entryExtIDs := [][]byte{[]byte(opr.BestNonce)}
	assetEntry := factom.Entry{ChainID: chainid, ExtIDs: entryExtIDs, Content: bOPR}

	var err1, err2 error
	for i := 0; i < 100; i++ {
		if i == 0 || err1 != nil {
			_, err1 = factom.CommitEntry(&assetEntry, opr.EC)
		}
		if i == 0 || err2 != nil {
			_, err2 = factom.RevealEntry(&assetEntry)
		}
		if err1 == nil && err2 == nil {
			break
		}
	}
}
