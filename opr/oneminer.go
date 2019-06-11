package opr

import (
	"fmt"
	"github.com/FactomProject/factom"
	"github.com/pegnet/OracleRecord/support"
	"github.com/zpatrick/go-config"
)

func OneMiner(verbose bool, config *config.Config, monitor *support.FactomdMonitor, grader *Grader, miner int) {
	alert := monitor.GetAlert()
	gAlert := grader.GetAlert()

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
				opr, _ = NewOpr(miner, fds.Dbht, config, gAlert)
				go opr.Mine(verbose)
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

	var err1, err2 error
	for i := 0; i < 100; i++ {
		if i == 0 || err1 != nil {
			_, err1 = factom.CommitEntry(opr.Entry, opr.EC)
		}
		if i == 0 || err2 != nil {
			_, err2 = factom.RevealEntry(opr.Entry)
		}
		if err1 == nil && err2 == nil {
			break
		}
	}
}
