// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package common

import (
	"fmt"

	"github.com/go-ini/ini"
	"github.com/zpatrick/go-config"
)

// Config names
const (
	ConfigCoordinatorListen            = "Miner.MiningCoordinatorPort"
	ConfigCoordinatorLocation          = "Miner.MiningCoordinatorHost"
	ConfigCoordinatorSecret            = "Miner.CoordinatorSecret"
	ConfigCoordinatorUseAuthentication = "Miner.UseCoordinatorAuthentication"
	ConfigSubmissionCutOff             = "Miner.SubmissionCutOff"

	ConfigMinerDBPath      = "Database.MinerDatabase"
	ConfigMinerDBType      = "Database.MinerDatabaseType"
	ConfigPegnetNodeDBPath = "Database.NodeDatabase"

	ConfigAPIPort          = "API.APIPort"
	ConfigControlPanelPort = "API.ControlPanelPort"

	ConfigCoinbaseAddress = "Miner.CoinbaseAddress"
	ConfigPegnetNetwork   = "Miner.Network"

	ConfigCoinbaseStakeAddress = "Staker.CoinbaseAddress"
	ConfigPegnetStakeNetwork   = "Staker.Network"

	ConfigCoinMarketCapKey = "Oracle.CoinMarketCapKey"
	Config1ForgeKey        = "Oracle.1ForgeKey"

	// ConfigStaleDuration determines how old a quote is allowed to be and still be
	// acceptable
	ConfigStaleDuration = "Oracle.StaleQuoteDuration"
)

// DefaultConfigOptions gives us the ability to add configurable settings that really
// should not be tinkered with often. Making the config file long and overly complex
// is just daunting to new users. Many of the settings that will likely never be touched
// can be inclued in here.
type DefaultConfigOptions struct {
}

func NewDefaultConfigOptionsProvider() *DefaultConfigOptions {
	d := new(DefaultConfigOptions)

	return d
}

func (c *DefaultConfigOptions) Load() (map[string]string, error) {
	settings := map[string]string{}
	// Include default settings here
	settings[ConfigSubmissionCutOff] = "200"
	settings[ConfigAPIPort] = "8099"
	settings[ConfigMinerDBPath] = "$PEGNETHOME/data_$PEGNETNETWORK/miner.ldb"
	settings[ConfigMinerDBType] = "ldb"
	settings[ConfigPegnetNodeDBPath] = "$PEGNETHOME/data_$PEGNETNETWORK/node.sqlite"
	settings[ConfigControlPanelPort] = "8080"
	settings[ConfigStaleDuration] = "30m"

	return settings, nil
}

func NewUnitTestConfig() *config.Config {
	return config.NewConfig([]config.Provider{NewDefaultConfigOptionsProvider(), NewUnitTestConfigProvider()})
}

// UnitTestConfigProvider is only used in unit tests.
//	This way we don't have to deal with pathing to find the
//	`defaultconfig.ini`.
type UnitTestConfigProvider struct {
	Data string
}

func NewUnitTestConfigProvider() *UnitTestConfigProvider {
	d := new(UnitTestConfigProvider)
	d.Data = `
[Debug]
# Randomize adds a random factor +/- the give percent.  3.1 for 3.1%
  Randomize=0.1
# Turns on logging so the user can see the OPRs and mining balances as they update
  Logging=true
# Puts the logs in a file.  If not specified, logs are written to stdout
  LogFile=

[Miner]
  FactomdLocation="localhost:8088"
  WalletdLocation="localhost:8089"
  NetworkType=LOCAL
  NumberOfMiners=15
# The number of records to submit per block. The top N records are chosen, where N is the config value
  RecordsPerBlock=10
  Protocol=PegNet 
  Network=unit-test

  # For LOCAL network testing, EC private key is
  # Es2XT3jSxi1xqrDvS5JERM3W3jh1awRHuyoahn3hbQLyfEi1jvbq
  ECAddress=EC3TsJHUs8bzbbVnratBafub6toRYdgzgbR7kWwCW4tqbmyySRmg

  # For LOCAL network testing, FCT private key is
  # Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK

  CoinbaseAddress=FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q
  IdentityChain=prototype
[Oracle]
  APILayerKey=CHANGEME
  OpenExchangeRatesKey=CHANGEME
  CoinMarketCapKey=CHANGEME
  1ForgeKey=CHANGEME
  StaleQuoteDuration=10m


[OracleDataSources]
  FreeForexAPI=-1
  APILayer=-1
  ExchangeRates=-1
  OpenExchangeRates=-1
  1Forge=-1

  # Crypto
  CoinMarketCap=-1
  CoinCap=-1

  # Commodities
  Kitco=-1

`
	return d
}

func (this *UnitTestConfigProvider) Load() (map[string]string, error) {
	settings := map[string]string{}

	file, err := ini.Load([]byte(this.Data))
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
