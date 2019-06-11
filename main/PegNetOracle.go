package main

import (
	"fmt"

	"encoding/json"
	"github.com/FactomProject/factom"
	opr2 "github.com/pegnet/OracleRecord/opr"
	"github.com/zpatrick/go-config"
	"os/user"
)

func main() {

	factom.SetFactomdServer("localhost:8088")
	factom.SetWalletServer("localhost:8089")

	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	userPath := u.HomeDir
	configfile := fmt.Sprintf("%s/.%s/miner%03d/config.ini", userPath, "pegnet", 1)
	iniFile := config.NewINIFile(configfile)
	Config := config.NewConfig([]config.Provider{iniFile})
	_, err = Config.String("Miner.Protocol")
	if err != nil {
		configfile = fmt.Sprintf("%s/.%s/defaultconfig.ini", userPath, "pegnet")
		iniFile := config.NewINIFile(configfile)
		Config = config.NewConfig([]config.Provider{iniFile})
		_, err = Config.String("Miner.Protocol")
		if err != nil {
			panic("Failed to open the config file for this miner, and couldn't load the default file either")
		}
	}

	opr, err := opr2.NewOpr(1, 1000, Config)
	if err != nil {
		panic(err)
	}

	for i := 0; true; i++ {
		opr.GetOPRecord(Config)
		js, err := json.Marshal(opr)
		if err != nil {
			panic(err)
		}
		opr.OPRHash = opr2.LX.Hash(js)

		// Just pick a nonce and compute a difficulty
		opr.Difficulty = opr.ComputeDifficulty()

		fmt.Println(i)
		for i := 0; i < 300; i++ {
			opr.Mine(3453262626, true) // Mine for 5 seconds
		}

		fmt.Println(opr.String())

	}
}

/*   Not used right now.  structures are there if you want to use it

 */
