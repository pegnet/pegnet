package common

import (
	"fmt"

	"github.com/go-ini/ini"
)

//
type DefaultConfigProvider struct {
	data string
}

func NewDefaultConfigProvider() *DefaultConfigProvider {
	d := new(DefaultConfigProvider)
	d.data = `
[Debug]
# Randomize adds a random factor +/- the give percent.  3.1 for 3.1%
  Randomize=0.1
# Turns on logging so the user can see the OPRs and mining balances as they update
  Logging=true
# Puts the logs in a file.  If not specified, logs are written to stdout
  LogFile=

[Miner]
  NetworkType=LOCAL
  NumberOfMiners=15
  Protocol=PegNet 
  Network=TestNet

  # For LOCAL network testing, EC private key is
  # Es2XT3jSxi1xqrDvS5JERM3W3jh1awRHuyoahn3hbQLyfEi1jvbq
  ECAddress=EC3TsJHUs8bzbbVnratBafub6toRYdgzgbR7kWwCW4tqbmyySRmg

  # For LOCAL network testing, FCT private key is
  # Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK
  FCTAddress=FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q

  CoinbasePNTAddress=tPNT_mEU1i4M5rn7bnrxNKdVVf4HXLG15Q798oaVAMrXq7zdbhQ9pv
  IdentityChain=prototype
[Oracle]
  APILayerKey=f6c9765ef81279e8891d40e34ef7ab01
  CoinCap=1
  APILayer=1
  ExchangeRatesAPI=0
  OpenExchangeRates=0
  Kitco=1
`
	return d
}

func (this *DefaultConfigProvider) Load() (map[string]string, error) {
	settings := map[string]string{}

	file, err := ini.Load([]byte(this.data))
	if err != nil {
		return nil, err
	}

	for _, section := range file.Sections() {
		for _, key := range section.Keys() {
			token := fmt.Sprintf("%s.%s", section.Name(), key.Name())
			settings[token] = key.String()
		}
	}

	return settings, nil
}
