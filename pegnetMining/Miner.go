// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package main

import (
	"fmt"
	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr"
	"github.com/zpatrick/go-config"
	"os"
	"os/user"
)

// Run a set of miners, as a network debugging aid
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
	_, err = Config.String("Miner.Protocol")
	if err != nil {
		panic("Failed to open the config file for this miner, and couldn't load the default file either")
	}

	monitor := new(common.FactomdMonitor)
	monitor.Start()
	grader := new(opr.Grader)
	go grader.Run(Config, monitor)

	numMiners, err := Config.Int("Miner.NumberOfMiners")
	if err != nil {
		panic(err)
	}
	if numMiners > 50 {
		fmt.Fprintln(os.Stderr, "Miner Limit is 50.  Config file specified too many Miners: ", numMiners, ".  Using 50")
		numMiners = 50
	}

	fmt.Println("Mining with ", numMiners, " miner(s).")

	for i := 1; i < numMiners; i++ {
		go opr.OneMiner(false, Config, monitor, grader, i)
	}
	opr.OneMiner(true, Config, monitor, grader, 16)
}
