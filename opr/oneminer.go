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
func OneMiner(verbose bool, config *config.Config, monitor *common.Monitor, grader *Grader, miner int) {
	alert := monitor.NewListener()
	gAlert := grader.GetAlert()

	numMiners, _ := config.Int("Miner.NumberOfMiners")
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
				go opr.Mine(verbose)
			}
		case 9:
			if mining {
				opr.StopMining <- 0
				mining = false
				writeMiningRecord(opr)
				if verbose {
					did := strings.Join(opr.FactomDigitalID[:len(opr.FactomDigitalID)-1], "-")
					log.WithFields(log.Fields{
						"hashrate": common.Stats.GetHashRate(),
						"difficulty": common.FormatDiff(common.Stats.Difficulty, 10),
						"id": did,
						"miner_count": numMiners,
						"height": fds.Dbht,
						}).Info("OPR Block Mined")
					common.Stats.Clear()
				}
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
		return errors.New("Unable to commit entry to factom")
	}

	err := backoff.Retry(operation, common.PegExponentialBackOff())
	if err != nil {
		// TODO: Handle error in retry
		log.WithError(err).Error("Failed to write mining record")
		return
	}
}
