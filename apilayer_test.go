package oprecord

import (
	"fmt"
	"testing"
	"os"
	"io/ioutil"
	"github.com/zpatrick/go-config"
)

func TestAPICall(t *testing.T) {
	configFileStuff := "" +
		"[Miner]\n" +
		"   Protocol=PegNet\n"+
		"   Network=TestNet\n"+
		"   ECAddress=EC3TsJHUs8bzbbVnratBafub6toRYdgzgbR7kWwCW4tqbmyySRmg\n" +
		"   IdentityChain=45b713101889a561df4028b3197459d2fca9783745d996ae090f672c8387914d\n" +
		"   IdentityChainFields=prototype,miner"
	dir := os.TempDir()+"/"
	iniFile := config.NewINIFile(filepath+"config.ini")
	c :=config.NewConfig([]config.Provider{iniFile})
	c := loadConfig(dir)
	data, err := CallAPILayer()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(data))
}
