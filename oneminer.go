package oprecord

import (
	"github.com/zpatrick/go-config"
	"time"
	"github.com/pegnet/OracleRecord/support"
	"fmt"
	"github.com/FactomProject/btcutil/base58"
	"encoding/hex"
	"encoding/json"
	"github.com/FactomProject/factom"
)

func OneMiner(config *config.Config, monitor *support.FactomdMonitor, miner int) {
	alert := monitor.GetAlert()
	mining := false
	var opr *OraclePriceRecord
	var err error
	for {
		fds := <-alert
		fmt.Println(fds.Dbht," ",fds.Minute)
		switch fds.Minute {
		case 1:
			if !mining {
				mining = true
				for i := 0; i < 100; i++ {
					opr, err = NewOpr(miner, fds.Dbht, config)
					opr.GetOPRecord(config)
					if err == nil {
						break
					}
					time.Sleep(time.Second)
				}
				go opr.Mine(int64(miner)+int64(fds.Dbht)+fds.Minute, true)
			}
		case 9:
			if mining {
				opr.StopMining <- 0
				fmt.Println(opr.String())
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
	bOPR,err := json.Marshal(opr)
	if err != nil {
		panic(err)
	}

	entryExtIDs := [][]byte{[]byte(opr.BestNonce)}
	assetEntry := factom.Entry{ChainID: chainid, ExtIDs: entryExtIDs, Content: bOPR}

	_, err = factom.CommitEntry(&assetEntry,opr.EC)
	if err != nil {
		panic(err)
	}

	_, err = factom.RevealEntry(&assetEntry)
	if err != nil {
		panic(err)
	}
}