package main

import (
	"fmt"
	"github.com/FactomProject/factom"
	"github.com/pegnet/OracleRecord/opr"
	"github.com/pegnet/OracleRecord/support"
	"github.com/zpatrick/go-config"
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

	monitor := new(support.FactomdMonitor)
	monitor.Start()
	grader := new(opr.Grader)
	go grader.Run(Config, monitor)

	for i:=1;i<=15;i++ {
		go opr.OneMiner(false, Config, monitor, grader, i)
	}
	opr.OneMiner(true,Config,monitor,grader,16)
}
