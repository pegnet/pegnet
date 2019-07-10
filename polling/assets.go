// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package polling

import (
	"sync"
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

func Round(v float64) float64 {
	return float64(int64(v*10000)) / 10000
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

	// digital currencies
	CoinCapResponse, err := CallCoinCap(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to access CoinCap")
	} else {
		HandleCoinCap(CoinCapResponse, &Peg)
	}

	// currency rates
	// this is a temp switch for Sources
	apikey, _ := config.String("Oracle.APILayerKey")
	if len(apikey) > 5 {
		// API Layer
		APILayerResponse, err := CallAPILayer(config)
		if err != nil {
			log.WithError(err).Fatal("Failed to access APILayer")
		} else {
			HandleAPILayer(APILayerResponse, &Peg)
		}
	} else {
		// ExchangeRates API
		ExchangeRatesApiResponse, err := CallExchangeRatesAPI(config)
		if err != nil {
			log.WithError(err).Fatal("Failed to access ExchangeRatesAPI")
		} else {
			HandleExchangeRatesAPI(ExchangeRatesApiResponse, &Peg)
		}
	}

	/*
		// Open Exchange Rates
		OpenExchangeRatesResponse, err := CallOpenExchangeRates(config)
		if err != nil {
			log.WithError(err).Fatal("Failed to access OpenExchangesRates")
		} else {
			HandleOpenExchangeRates(OpenExchangeRatesResponse, &Peg)
		}
	*/

	// precious metals
	KitcoResponse, err := CallKitcoWeb()
	if err != nil {
		log.WithError(err).Fatal("Failed to access Kitco Website")
	} else {
		HandleKitcoWeb(KitcoResponse, &Peg)
	}

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
