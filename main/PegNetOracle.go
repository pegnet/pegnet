package main

import (
	"fmt"

	"github.com/pegnet/OracleRecord"
	"os/user"
	"github.com/zpatrick/go-config"
)

func main() {
	var opr oprecord.OraclePriceRecord

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

	opr.GetOPRecord(Config)
	fmt.Println(opr)
}

/*   Not used right now.  structures are there if you want to use it

 */
