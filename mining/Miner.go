package main

import (
	"github.com/pegnet/OracleRecord/support"
	"github.com/FactomProject/factom"
	"os/user"
	"fmt"
	"github.com/zpatrick/go-config"
	"github.com/pegnet/OracleRecord/opr"
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
	go opr.OneMiner(true,Config,monitor,1)
	go opr.OneMiner(false,Config,monitor,2)
	go opr.OneMiner(false,Config,monitor,3)
	go opr.OneMiner(false,Config,monitor,4)
	go opr.OneMiner(false,Config,monitor,5)
	go opr.OneMiner(false,Config,monitor,6)
	go opr.OneMiner(false,Config,monitor,7)
	go opr.OneMiner(false,Config,monitor,8)
	go opr.OneMiner(false,Config,monitor,9)
	go opr.OneMiner(false,Config,monitor,10)
	go opr.OneMiner(false,Config,monitor,11)
	go opr.OneMiner(false,Config,monitor,12)
	go opr.OneMiner(false,Config,monitor,13)
	go opr.OneMiner(false,Config,monitor,14)
	opr.OneMiner(false, Config,monitor,15)
}
