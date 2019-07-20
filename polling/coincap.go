// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package polling

import (
	"encoding/json"
	"github.com/cenkalti/backoff"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
	"io/ioutil"
	"net/http"
	"strconv"
)

type CoinCapResponse struct {
	Data      []CoinCapRecord `json:"data"`
	Timestamp int64           `json:"timestamp"`
}

type CoinCapRecord struct {
	ID                string `json:"id"`
	Rank              string `json:"rank"`
	Symbol            string `json:"symbol"`
	Name              string `json:"name"`
	Supply            string `json:"supply"`
	MaxSupply         string `json:"maxSupply"`
	MarketCapUSD      string `json:"marketCapUsd"`
	VolumeUSD24Hr     string `json:"volumeUsd24Hr"`
	PriceUSD          string `json:"priceUsd"`
	ChangePercent24Hr string `json:"changePercent24Hr"`
	VWAP24Hr          string `json:"vwap24Hr"`
}

func CallCoinCap(config *config.Config) (CoinCapResponse, error) {
	var CoinCapResponse CoinCapResponse

	operation := func() error {
		resp, err := http.Get("http://api.coincap.io/v2/assets?limit=500")
		if err != nil {
			log.WithError(err).Warning("Failed to get response from CoinCap")
			return err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(body, &CoinCapResponse)
		return nil
	}

	err := backoff.Retry(operation, PollingExponentialBackOff())
	return CoinCapResponse, err
}

func HandleCoinCap(response CoinCapResponse, peg *PegAssets) {

	var timestamp = response.Timestamp

	for _, currency := range response.Data {
		if currency.Symbol == "XBT" || currency.Symbol == "BTC" {
			value, err := strconv.ParseFloat(currency.PriceUSD, 64)
			peg.XBT.Value = Round(value)
			peg.XBT.When = timestamp
			if err != nil {
				continue
			}
		} else if currency.Symbol == "ETH" {
			value, err := strconv.ParseFloat(currency.PriceUSD, 64)
			peg.ETH.Value = Round(value)
			peg.ETH.When = timestamp
			if err != nil {
				continue
			}
		} else if currency.Symbol == "LTC" {
			value, err := strconv.ParseFloat(currency.PriceUSD, 64)
			peg.LTC.Value = Round(value)
			peg.LTC.When = timestamp
			if err != nil {
				continue
			}
		} else if currency.Symbol == "XBC" || currency.Symbol == "BCH" {
			value, err := strconv.ParseFloat(currency.PriceUSD, 64)
			peg.XBC.Value = Round(value)
			peg.XBC.When = timestamp
			if err != nil {
				continue
			}
		} else if currency.Symbol == "FCT" {
			value, err := strconv.ParseFloat(currency.PriceUSD, 64)
			peg.FCT.Value = Round(value)
			peg.FCT.When = timestamp
			if err != nil {
				continue
			}
		}
	}

}

func CoinCapInterface(config *config.Config, peg *PegAssets) {
	log.Debug("Pulling Asset data from CoinCap")
	CoinCapResponse, err := CallCoinCap(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to access CoinCap")
	} else {
		HandleCoinCap(CoinCapResponse, peg)
	}
}
