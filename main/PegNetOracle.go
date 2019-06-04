package main

import (
	"fmt"

	"github.com/pegnet/OracleRecord"
	"os/user"
	"github.com/zpatrick/go-config"
	"encoding/json"
	"github.com/FactomProject/factom"
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

	opr,err := oprecord.NewOpr(1,1000,Config)
	if err != nil {
		panic(err)
	}

	for i:=0;true;i++ {
		opr.GetOPRecord(Config)
		js,err := json.Marshal(opr)
		if err != nil {
			panic(err)
		}
		opr.OPRHash = oprecord.LX.Hash(js)

		// Just pick a nonce and compute a difficulty
		opr.BestNonce = oprecord.LX.Hash([]byte(fmt.Sprintf("junk %d",i)))
		opr.Difficulty = opr.ComputeDifficulty(opr.OPRHash,opr.BestNonce)


		fmt.Println(i)
		for i := 0; i < 300; i++ {
			opr.Mine(true, .1) // Mine for 5 seconds
		}

		fmt.Println(opr.String())

	}
}

/*   Not used right now.  structures are there if you want to use it

 */
