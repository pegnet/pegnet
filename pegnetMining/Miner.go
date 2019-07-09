package pegnetMining

// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import (
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

const MaxMiners = 50

// Mine runs a set of miners, as a network debugging aid
func Mine(numMiners int, config *config.Config, monitor *common.FactomdMonitor, grader *opr.Grader) {
	if numMiners > MaxMiners {
		log.WithFields(log.Fields{
			"attempted": numMiners,
			"limit":     MaxMiners,
		}).Warn("Too many miners specified, defaulting to limit")
		numMiners = MaxMiners
	}
	log.WithFields(log.Fields{
		"miner_count": numMiners,
	}).Info("Starting to mine")

	for i := 1; i < numMiners; i++ {
		go opr.OneMiner(false, config, monitor, grader, i)
	}
	opr.OneMiner(true, config, monitor, grader, numMiners)
}
