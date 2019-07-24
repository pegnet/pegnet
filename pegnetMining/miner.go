// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package pegnetMining

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/FactomProject/factom"
	"github.com/cenkalti/backoff"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

const MaxMiners = 50

// Singleton counter
var MinerID int64

// GetNextMinerID just grabs the next minerid available
func GetNextMinerID() int {
	return int(atomic.AddInt64(&MinerID, 1))
}

// InitMiners creates the miners requested, not-started
func InitMiners(config *config.Config, monitor *common.Monitor, grader *opr.Grader) []*PegnetMiner {
	numMiners, _ := config.Int("Miner.NumberOfMiners")
	top, _ := config.Int("Miner.RecordsPerBlock")

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
		"records_per": top,
	}).Info("Initializing miners")

	keep, err := config.Int("Miner.RecordsPerBlock")
	if err != nil {
		panic(err)
	}

	// All miners share the same writer. As the writer will
	// select the best `keep` records
	writer := NewEntryWriter(config, keep)
	err = writer.populateECAddress()
	if err != nil {
		panic(err)
	}

	miners := make([]*PegnetMiner, numMiners)
	for i := range miners { // Init the miners
		miners[i] = NewPegnetMiner(config, monitor, grader, writer)
	}
	return miners
}

type PegnetMiner struct {
	// ID is the miner number, starting with "1". Every miner launched gets the next
	// sequential number.
	ID     int            `json:"id"`
	Config *config.Config `json:"-"` //  The config of the miner using the record

	// Factom blockchain related alerts
	FactomMonitor common.IMonitor
	OPRGrader     opr.IGrader

	FactomEntryWriter *EntryWriter

	// miningContext can be passed around, if the context is canceled, everyone who owns this can
	// be notified. If the context is cancelled, that means mining has stopped
	contextMutex  sync.Mutex // Go race complaining about context access
	miningContext context.Context
	cancelMining  context.CancelFunc
}

func NewPegnetMiner(config *config.Config, monitor common.IMonitor, grader opr.IGrader, writer *EntryWriter) *PegnetMiner {
	m := new(PegnetMiner)
	m.FactomMonitor = monitor
	m.OPRGrader = grader
	m.FactomEntryWriter = writer
	m.ID = GetNextMinerID()
	m.Config = config

	return m
}

// StopMining will halt all mining on this miner. It can be restarted
func (p *PegnetMiner) StopMining() bool {
	p.contextMutex.Lock()
	defer p.contextMutex.Unlock()
	if p.cancelMining == nil {
		return false // Not mining
	}
	// All contexts listening to this cancel will cancel when they
	// hit.
	p.cancelMining()
	// Set the cancel to nil so we know we are not running.
	// We already called cancel, so it's safe to lose it's reference
	p.cancelMining = nil
	p.FactomEntryWriter.Cancel()
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

	p.contextMutex.Lock()
	p.miningContext, p.cancelMining = context.WithCancel(context.Background())
	p.contextMutex.Unlock()

	var writeChannel chan<- *opr.NonceRanking

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
				p.FactomEntryWriter = p.FactomEntryWriter.NextBlockWriter()
				writeChannel = p.FactomEntryWriter.AddMiner()
				oprO, err = opr.NewOpr(p.miningContext, p.ID, fds.Dbht, p.Config, gAlert)
				if err == context.Canceled {
					continue MiningLoop // OPR cancelled
				}
				if err != nil {
					log.WithError(err).Fatal("Error creating an OPR.  Likely a config file issue")
				}
				p.FactomEntryWriter.SetOPR(oprO) // Makes a copy of this opr for our final entry
				go oprO.Mine(verbose)
			}
		case 9:
			if mining {
				close(oprO.StopMining)
				mining = false
				// Time to write
				p.FactomEntryWriter.CollectAndWrite(false)
				writeChannel <- oprO.NonceAggregate
			}
		}

	}
}
