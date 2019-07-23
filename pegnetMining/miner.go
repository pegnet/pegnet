// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package pegnetMining

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/FactomProject/factom"
	"github.com/cenkalti/backoff"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

const MaxMiners = 50

// InitMiners creates the miners requested not-started
func InitMiners(numMiners int, config *config.Config, monitor *common.Monitor, grader *opr.Grader) []*PegnetMiner {
	opr.InitLX() // We intend to use the LX hash
	if numMiners > MaxMiners {
		log.WithFields(log.Fields{
			"attempted": numMiners,
			"limit":     MaxMiners,
		}).Warn("Too many miners specified, defaulting to limit")
		numMiners = MaxMiners
	}
	log.WithFields(log.Fields{
		"miner_count": numMiners,
	}).Info("Initializing miners")

	miners := make([]*PegnetMiner, numMiners)
	for i := range miners { // Init the miners
		miners[i] = NewPegnetMiner(config, monitor, grader)
	}
	return miners
}

var MinerID int64

func GetNextMinerID() int {
	return int(atomic.AddInt64(&MinerID, 1))
}

type PegnetMiner struct {
	// ID is the miner number, starting with "1". Every miner launched gets the next
	// sequential number.
	ID     int            `json:"id"`
	Config *config.Config `json:"-"` //  The config of the miner using the record

	// Factom blockchain related alerts
	FactomMonitor common.IMonitor
	OPRGrader     opr.IGrader

	// miningContext can be passed around, if the context is canceled, everyone who owns this can
	// be notified. If the context is cancelled, that means mining has stopped
	miningContext context.Context
	cancelMining  context.CancelFunc
}

func NewPegnetMiner(config *config.Config, monitor common.IMonitor, grader opr.IGrader) *PegnetMiner {
	m := new(PegnetMiner)
	m.FactomMonitor = monitor
	m.OPRGrader = grader
	m.ID = GetNextMinerID()
	m.Config = config

	return m
}

// StopMining will halt all mining on this miner. It can be restarted
func (p *PegnetMiner) StopMining() bool {
	if p.cancelMining == nil {
		return false // Not mining
	}
	// All contexts listening to this cancel will cancel when they
	// hit.
	p.cancelMining()
	// Set the cancel to nil so we know we are not running.
	// We already called cancel, so it's safe to lose it's reference
	p.cancelMining = nil
	return true
}

// IDString is purely for inner process tracking
func (p *PegnetMiner) IDString() string {
	return fmt.Sprintf("Miner-%d", p.ID)
}

// LaunchMiningThread starts a single mining goroutine that listens to the monitor.
// Starts on minute 1 and writes data at minute 9
func (p *PegnetMiner) LaunchMiningThread(verbose bool) {
	mineLog := log.WithFields(log.Fields{"miner": p.ID})

	// TODO: Also tell Factom Monitor we are done listening
	alert := p.FactomMonitor.NewListener()
	gAlert := p.OPRGrader.GetAlert(p.IDString())
	// Tell OPR grader we are no longer listening
	defer p.OPRGrader.StopAlert(p.IDString())
	p.miningContext, p.cancelMining = context.WithCancel(context.Background())

	numMiners, _ := p.Config.Int("Miner.NumberOfMiners")
	mining := false
	var oprO *opr.OraclePriceRecord
	var err error
MiningLoop:
	for {
		var fds common.MonitorEvent
		select {
		case fds = <-alert:
		case <-p.miningContext.Done(): // If cancelled
			if oprO != nil {
				close(oprO.StopMining)
			}
			return // Stop the mining
		}

		mineLog.WithFields(log.Fields{
			"height": fds.Dbht,
			"minute": fds.Minute,
		}).Debug("Miner received alert")
		switch fds.Minute {
		case 1:
			if !mining {
				mining = true
				oprO, err = opr.NewOpr(p.miningContext, p.ID, fds.Dbht, p.Config, gAlert)
				if err == context.Canceled {
					continue MiningLoop // OPR cancelled
				}
				if err != nil {
					log.WithError(err).Fatal("Error creating an OPR.  Likely a config file issue")
				}
				go oprO.Mine(verbose)
			}
		case 9:
			if mining {
				close(oprO.StopMining)
				mining = false
				writeMiningRecord(oprO)
				if verbose {
					did := strings.Join(oprO.FactomDigitalID[:len(oprO.FactomDigitalID)-1], "-")
					mineLog.WithFields(log.Fields{
						"hashrate":    common.Stats.GetHashRate(),
						"difficulty":  common.FormatDiff(common.Stats.Difficulty, 10),
						"id":          did,
						"miner_count": numMiners,
						"height":      fds.Dbht,
					}).Info("OPR Block Mined")
					common.Stats.Clear()
				}
			}
		}

	}
}

func writeMiningRecord(opr *opr.OraclePriceRecord) {
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
