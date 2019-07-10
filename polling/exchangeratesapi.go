// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package polling

import (
	"encoding/json"
	"github.com/cenkalti/backoff"
	"github.com/zpatrick/go-config"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type ExchangeRatesAPIResponse struct {
	Date    string
	Base    string
	Rates   ExchangeRatesAPIRecord
}

type ExchangeRatesAPIRecord struct {
	CAD float64
	HKD float64
	ISK float64
	PHP float64
	DKK float64
	HUF float64
	CZK float64
	GBP float64
	RON float64
	SEK float64
	IDR float64
	INR float64
	BRL float64
	RUB float64
	HRK float64
	JPY float64
	THB float64
	CHF float64
	EUR float64
	MYR float64
	BGN float64
	TRY float64
	CNY float64
	NOK float64
	NZD float64
	ZAR float64
	USD float64
	MXN float64
	SGD float64
	AUD float64
	ILS float64
	KRW float64
	PLN float64
}

func CallExchangeRatesAPI(c *config.Config) (ExchangeRatesAPIResponse, error) {
	var ExchangeRatesAPIResponse ExchangeRatesAPIResponse

	operation := func() error {
		resp, err := http.Get("https://api.exchangeratesapi.io/latest?base=USD")
		if err != nil {
			log.WithError(err).Warning("Failed to get response from ExchangeRatesAPI")
			return err
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(body, &ExchangeRatesAPIResponse)
		return nil
	}

	err := backoff.Retry(operation, PollingExponentialBackOff())
	return ExchangeRatesAPIResponse, err
}

func HandleExchangeRatesAPI(response ExchangeRatesAPIResponse, peg *PegAssets) {

	// Exchange rates api does not return timestamp.
	var timestamp = ConverToUnix("2006-01-02", response.Date)

	peg.USD.Value = Round(response.Rates.USD)
	peg.USD.When = timestamp
	peg.EUR.Value = Round(response.Rates.EUR)
	peg.EUR.When = timestamp
	peg.JPY.Value = Round(response.Rates.JPY)
	peg.JPY.When = timestamp
	peg.GBP.Value = Round(response.Rates.GBP)
	peg.GBP.When = timestamp
	peg.CAD.Value = Round(response.Rates.CAD)
	peg.CAD.When = timestamp
	peg.CHF.Value = Round(response.Rates.CHF)
	peg.CHF.When = timestamp
	peg.INR.Value = Round(response.Rates.INR)
	peg.INR.When = timestamp
	peg.SGD.Value = Round(response.Rates.SGD)
	peg.SGD.When = timestamp
	peg.CNY.Value = Round(response.Rates.CNY)
	peg.CNY.When = timestamp
	peg.HKD.Value = Round(response.Rates.HKD)
	peg.HKD.When = timestamp

}
