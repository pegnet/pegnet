package opr

// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import (
	"errors"
	"strings"

	"github.com/FactomProject/factom"
	"github.com/cenkalti/backoff"
	"github.com/pegnet/pegnet/common"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

// OneMiner starts a single mining goroutine that listens to the monitor.
// Starts on minute 1 and writes data at minute 9
// Terminates after writing data.
func OneMiner(verbose bool, config *config.Config, monitor *common.FactomdMonitor, grader *Grader, miner int) {
	alert := monitor.GetAlert()
	gAlert := grader.GetAlert()

	mining := false
	var opr *OraclePriceRecord
	var err error
	for {
		fds := <-alert
		log.WithFields(log.Fields{
			"miner":  miner,
			"height": fds.Dbht,
			"minute": fds.Minute,
		}).Debug("Miner received alert")
		switch fds.Minute {
		case 1:
			if !mining {
				mining = true
				opr, err = NewOpr(miner, fds.Dbht, config, gAlert)
				if err != nil {
					log.WithError(err).Fatal("Error creating an OPR.  Likely a config file issue")
				}
				log.WithFields(log.Fields{
					"did": strings.Join(opr.FactomDigitalID, "-"),
				}).Info("New OPR miner")
				go opr.Mine(verbose)
			}
		case 9:
			if mining {
				opr.StopMining <- 0

				common.Do(func() {
					data, ok := opr.Entry.MarshalBinary()
					if ok != nil {
						log.Fatal("Failed to marshal OPR Entry")
					}
					recordFields := opr.LogFieldsShort()
					recordFields["miner"] = miner
					recordFields["entry_size"] = len(data)
					log.WithFields(recordFields).Info("Created OPR Entry")
				})

				mining = false

				writeMiningRecord(opr)
			}
		}
	}
}

func writeMiningRecord(opr *OraclePriceRecord) {
	operation := func() error {
		var err1, err2 error
		_, err1 = factom.CommitEntry(opr.Entry, opr.EC)
		_, err2 = factom.RevealEntry(opr.Entry)
		if err1 == nil && err2 == nil {
			return nil
		}
		return errors.New("unable to commit Entry to factom")
	}

	err := backoff.Retry(operation, common.PegExponentialBackOff())
	if err != nil {
		// TODO: Handle error in retry
		log.WithError(err).Error("Failed to write mining record")
		return
	}
}
