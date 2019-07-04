// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package polling

import (
	"encoding/json"
	"github.com/pegnet/pegnet/common"
	"math/rand"
	"os"

	"github.com/zpatrick/go-config"
	"strconv"
	"sync"
	"time"
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
	When  string
}

func (p *PegItems) Clone(randomize float64) PegItems {
	np := new(PegItems)
	np.Value = p.Value + p.Value*(randomize/2*rand.Float64()) - p.Value*(randomize/2*rand.Float64())
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
	if err != nil && lastTime==0 {
		common.Logf("error","the config file doesn't have a valid Randomize value. %v",err)
		randomize = 0
	}

	if delta < qlimit && lastTime != 0 {
		pa := lastAnswer.Clone(randomize)
		return pa
	}

	lastTime = now
	common.Logf("PullPEGAssets", "Make a call to get data. Seconds since last call: %d", delta)
	var Peg PegAssets
	// digital currencies
	CoinCapResponseBytes, err := CallCoinCap(config)
	if err != nil {
		common.Logf("error", "Error accessing CallCoinCap %v", err)
		os.Exit(1)
	} else {
		var CoinCapValues CoinCapResponse
		err = json.Unmarshal(CoinCapResponseBytes, &CoinCapValues)
		for _, currency := range CoinCapValues.Data {
			if currency.Symbol == "XBT" || currency.Symbol == "BTC" {
				Peg.XBT.Value, err = strconv.ParseFloat(currency.PriceUSD, 64)
				Peg.XBT.Value = Round(Peg.XBT.Value)
				if err != nil {
					continue
				}
				Peg.XBT.When = string(CoinCapValues.Timestamp)
			} else if currency.Symbol == "ETH" {
				Peg.ETH.Value, err = strconv.ParseFloat(currency.PriceUSD, 64)
				Peg.ETH.Value = Round(Peg.ETH.Value)
				if err != nil {
					continue
				}
				Peg.ETH.When = string(CoinCapValues.Timestamp)
			} else if currency.Symbol == "LTC" {
				Peg.LTC.Value, err = strconv.ParseFloat(currency.PriceUSD, 64)
				Peg.LTC.Value = Round(Peg.LTC.Value)
				if err != nil {
					continue
				}
				Peg.LTC.When = string(CoinCapValues.Timestamp)
			} else if currency.Symbol == "XBC" || currency.Symbol == "BCH" {
				Peg.XBC.Value, err = strconv.ParseFloat(currency.PriceUSD, 64)
				Peg.XBC.Value = Round(Peg.XBC.Value)
				if err != nil {
					continue
				}
				Peg.XBC.When = string(CoinCapValues.Timestamp)
			} else if currency.Symbol == "FCT" {
				Peg.FCT.Value, err = strconv.ParseFloat(currency.PriceUSD, 64)
				Peg.FCT.Value = Round(Peg.FCT.Value)
				if err != nil {
					continue
				}
				Peg.FCT.When = string(CoinCapValues.Timestamp)
			}
		}
	}

	APILayerBytes, err := CallAPILayer(config)

	if err != nil {
		common.Logf("error", "Error accessing CallAPILayer(): %v", err)
		os.Exit(1)
	} else {
		var APILayerResponse APILayerResponse
		err = json.Unmarshal(APILayerBytes, &APILayerResponse)

		Peg.USD.Value = Round(APILayerResponse.Quotes.USDUSD)
		Peg.USD.When = string(APILayerResponse.Timestamp)
		Peg.EUR.Value = Round(APILayerResponse.Quotes.USDEUR)
		Peg.EUR.When = string(APILayerResponse.Timestamp)
		Peg.JPY.Value = Round(APILayerResponse.Quotes.USDJPY)
		Peg.JPY.When = string(APILayerResponse.Timestamp)
		Peg.GBP.Value = Round(APILayerResponse.Quotes.USDGBP)
		Peg.GBP.When = string(APILayerResponse.Timestamp)
		Peg.CAD.Value = Round(APILayerResponse.Quotes.USDCAD)
		Peg.CAD.When = string(APILayerResponse.Timestamp)
		Peg.CHF.Value = Round(APILayerResponse.Quotes.USDCHF)
		Peg.CHF.When = string(APILayerResponse.Timestamp)
		Peg.INR.Value = Round(APILayerResponse.Quotes.USDINR)
		Peg.INR.When = string(APILayerResponse.Timestamp)
		Peg.SGD.Value = Round(APILayerResponse.Quotes.USDSGD)
		Peg.SGD.When = string(APILayerResponse.Timestamp)
		Peg.CNY.Value = Round(APILayerResponse.Quotes.USDCNY)
		Peg.CNY.When = string(APILayerResponse.Timestamp)
		Peg.HKD.Value = Round(APILayerResponse.Quotes.USDHKD)
		Peg.HKD.When = string(APILayerResponse.Timestamp)

	}

	KitcoResponse, err := CallKitcoWeb()

	for i := 0; i < 10; i++ {
		if err != nil {
			common.Logf("error", "Error %d so retrying.  Error %v", i+1, err)
			time.Sleep(time.Second)
			KitcoResponse, err = CallKitcoWeb()
		} else {
			break //	os.Exit(1)
		}
	}
	if err != nil {
		common.Logf("error", "Error, using old data.")
		pa := lastAnswer.Clone(randomize)
		return pa
	}
	Peg.XAU.Value, err = strconv.ParseFloat(KitcoResponse.Silver.Bid, 64)
	Peg.XAU.When = KitcoResponse.Silver.Date
	Peg.XAG.Value, err = strconv.ParseFloat(KitcoResponse.Gold.Bid, 64)
	Peg.XAG.When = KitcoResponse.Gold.Date
	Peg.XPD.Value, err = strconv.ParseFloat(KitcoResponse.Palladium.Bid, 64)
	Peg.XPD.When = KitcoResponse.Palladium.Date
	Peg.XPT.Value, err = strconv.ParseFloat(KitcoResponse.Platinum.Bid, 64)
	Peg.XPT.When = KitcoResponse.Platinum.Date

	lastAnswer = Peg

	return Peg.Clone(randomize)
}
