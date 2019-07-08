// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package main

import (
	"flag"
	"fmt"
	"os/user"
	"strings"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

const MaxMiners = 50

var (
	// Global Flags
	LogLevel        string // Logrus global log level
	FactomdLocation string
	WalletdLocation string
	ECAddressString string
)

func init() {
	flag.StringVar(&LogLevel, "log", "info", "Change the logging level. Can choose from 'debug', 'info', 'warn', 'error', or 'fatal'")
	flag.StringVar(&FactomdLocation, "s", "localhost:8088", "IPAddr:port# of factomd API to use to access blockchain (default localhost:8088)")
	flag.StringVar(&WalletdLocation, "w", "localhost:8089", "IPAddr:port# of factom-walletd API to use to create transactions (default localhost:8089)")
	flag.StringVar(&ECAddressString, "ec", "", "EC Address to use in place of the one specified in PegNet config file")
}

// Run a set of miners, as a network debugging aid
func main() {
	u, err := user.Current()
	if err != nil {
		log.WithError(err).Fatal("Failed to read current user's name")
	}
	userPath := u.HomeDir
	configFile := fmt.Sprintf("%s/.%s/defaultconfig.ini", userPath, "pegnet")
	iniFile := config.NewINIFile(configFile)
	Config := config.NewConfig([]config.Provider{iniFile})
	_, err = Config.String("Miner.Protocol")
	if err != nil {
		log.WithError(err).Fatal("Failed to open config file or load the default file")
	}
	numMiners, err := Config.Int("Miner.NumberOfMiners")
	if err != nil {
		log.WithError(err).Fatal("Failed to read number of miners")
	}

	// If miners flag is set use that value otherwise default to the config setting
	flag.IntVar(&numMiners, "m", numMiners, "Number of miners to run")
	flag.Parse()

	factom.SetFactomdServer(FactomdLocation)
	factom.SetWalletServer(WalletdLocation)

	switch strings.ToLower(LogLevel) {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	}

	monitor := common.GetMonitor()
	grader := new(opr.Grader)
	go grader.Run(Config, monitor)

	// Start mining
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
		go opr.OneMiner(false, Config, monitor, grader, i)
	}
	opr.OneMiner(true, Config, monitor, grader, numMiners)
}
