// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package polling

import (
	"sync"
	"math/rand"
	"time"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

const qlimit = 580 // Limit queries to once just shy of 10 minutes (600 seconds)

type PegAssets struct {
	PNT PegItems
	USD PegItems
	EUR PegItems
	JPY PegItems
	GBP PegItems
	CAD PegItems
	CHF PegItems
	INR PegItems
	SGD PegItems
	CNY PegItems
	HKD PegItems
	XAU PegItems
	XAG PegItems
	XPD PegItems
	XPT PegItems
	XBT PegItems
	ETH PegItems
	LTC PegItems
	XBC PegItems
	FCT PegItems
}

func (p *PegAssets) Clone(randomize float64) PegAssets {
	np := new(PegAssets)
	np.PNT = p.PNT.Clone(randomize)
	np.USD = p.USD.Clone(randomize)
	np.EUR = p.EUR.Clone(randomize)
	np.JPY = p.JPY.Clone(randomize)
	np.GBP = p.GBP.Clone(randomize)
	np.CAD = p.CAD.Clone(randomize)
	np.CHF = p.CHF.Clone(randomize)
	np.INR = p.INR.Clone(randomize)
	np.SGD = p.SGD.Clone(randomize)
	np.CNY = p.CNY.Clone(randomize)
	np.HKD = p.HKD.Clone(randomize)
	np.XAU = p.XAU.Clone(randomize)
	np.XAG = p.XAG.Clone(randomize)
	np.XPD = p.XPD.Clone(randomize)
	np.XPT = p.XPT.Clone(randomize)
	np.XBT = p.XBT.Clone(randomize)
	np.ETH = p.ETH.Clone(randomize)
	np.LTC = p.LTC.Clone(randomize)
	np.XBC = p.XBC.Clone(randomize)
	np.FCT = p.FCT.Clone(randomize)
	return *np
}

type PegItems struct {
	Value float64
	When  int64 // unix timestamp
}

func (p *PegItems) Clone(randomize float64) PegItems {
	np := new(PegItems)
	np.Value = p.Value + p.Value*(randomize/2*rand.Float64()) - p.Value*(randomize/2*rand.Float64())
	np.Value = Round(np.Value)
	np.When = p.When
	return *np
}

var lastMutex sync.Mutex
var lastAnswer PegAssets //
var lastTime int64       // In seconds

var defaultDigitalAsset = "CoinCap"
var availableDigitalAssets = map[string]func(config *config.Config, peg *PegAssets){
	"CoinCap": CoinCapInterface,
}

var defaultCurrencyAsset = "ExchangeRatesAPI"
var availableCurrencyAssets = map[string]func(config *config.Config, peg *PegAssets){
	"APILayer": APILayerInterface,
	"ExchangeRatesAPI": ExchangeRatesAPIInterface, 
	"OpenExchangeRates": OpenExchangeRatesInterface,
}

var defaultMetalAsset = "Kitco"
var availableMetalAssets = map[string]func(config *config.Config, peg *PegAssets){
	"Kitco": KitcoInterface,
}

func GetAssetsByWieght(config *config.Config, assets map[string]func(config *config.Config, peg *PegAssets), default_asset string) []string {
	var result = []string{}
	for key := range assets {
		weight, _ := config.Int("Oracle." + key)
		for w := 0; w < weight; w++ {
			result = append(result, key)
		}
	}
	if len(result) == 0 {
		result = append(result, default_asset)
	}
	return result
}

func GetAvailableAssetsByWieght(config *config.Config) (string, string, string) {
	rand.Seed(time.Now().Unix())

	var digital_currencies = GetAssetsByWieght(config, availableDigitalAssets, defaultDigitalAsset)
	var currency_rates = GetAssetsByWieght(config, availableCurrencyAssets, defaultCurrencyAsset)
	var precious_metals = GetAssetsByWieght(config, availableMetalAssets, defaultMetalAsset)

	var digital_currencies_asset = digital_currencies[rand.Intn(len(digital_currencies))]
	var currency_rates_asset = currency_rates[rand.Intn(len(currency_rates))]
	var precious_metals_asset = precious_metals[rand.Intn(len(precious_metals))]

	return digital_currencies_asset, currency_rates_asset, precious_metals_asset
}

func PullPEGAssets(config *config.Config) (pa PegAssets) {

	// Prevent pounding of external APIs
	lastMutex.Lock()
	defer lastMutex.Unlock()
	now := time.Now().Unix()
	delta := now - lastTime

	// For testing, you can specify a randomization of the values returned by the oracles.
	// If the value specified isn't reasonable, then randomize is zero, and the values returned
	// are not changed.
	randomize, err := config.Float("Debug.Randomize")
	if err != nil && lastTime == 0 {
		log.WithError(err).Fatal(fmt.Sprintf("the config file doesn't have a valid Randomize value. %v", err))
	}

	if delta < qlimit && lastTime != 0 {
		pa := lastAnswer.Clone(randomize)
		return pa
	}

	lastTime = now
	log.WithFields(log.Fields{
		"delta_time": delta,
	}).Debug("Pulling PEG Asset data")

	var Peg PegAssets

	
	digital_currencies, currency_rates, precious_metals := GetAvailableAssetsByWieght(config)

	// digital currencies
	availableDigitalAssets[digital_currencies](config, &Peg)

	// currency rates
	availableCurrencyAssets[currency_rates](config, &Peg)

	// precious metals
	availableMetalAssets[precious_metals](config, &Peg)
	
	// debug
	log.WithFields(log.Fields{
		"XBT": Peg.XBT.Value,
		"ETH": Peg.ETH.Value,
		"LTC": Peg.LTC.Value,
		"XBC": Peg.XBC.Value,
		"FCT": Peg.FCT.Value,
		"USD": Peg.USD.Value,
		"EUR": Peg.EUR.Value,
		"JPY": Peg.JPY.Value,
		"GBP": Peg.GBP.Value,
		"CAD": Peg.CAD.Value,
		"CHF": Peg.CHF.Value,
		"INR": Peg.INR.Value,
		"SGD": Peg.SGD.Value,
		"CNY": Peg.CNY.Value,
		"HKD": Peg.HKD.Value,
		"XAU": Peg.XAU.Value,
		"XAG": Peg.XAG.Value,
		"XPD": Peg.XPD.Value,
		"XPT": Peg.XPT.Value,
	}).Debug("Pulling PEG Asset data Result")

	lastAnswer = Peg

	return Peg.Clone(randomize)
}
