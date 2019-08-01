// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package polling

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/cenkalti/backoff"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

type ExchangeRatesAPIResponse struct {
	Date  string             `json:"date"`
	Base  string             `json:"base"`
	Rates map[string]float64 `json:"rates"`
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
		if body, err := ioutil.ReadAll(resp.Body); err != nil {
			return err
		} else if err = json.Unmarshal(body, &ExchangeRatesAPIResponse); err != nil {
			return err
		}
		return nil
	}

	err := backoff.Retry(operation, PollingExponentialBackOff())
	return ExchangeRatesAPIResponse, err
}

func HandleExchangeRatesAPI(response ExchangeRatesAPIResponse, peg PegAssets) {
	// Exchange rates api does not return timestamp.
	var timestamp = ConverToUnix("2006-01-02", response.Date)
	UpdatePegAssets(response.Rates, timestamp, peg)
}

func ExchangeRatesAPIInterface(config *config.Config, peg PegAssets) {
	log.Debug("Pulling Asset data from ExchangeRatesAPI")
	ExchangeRatesApiResponse, err := CallExchangeRatesAPI(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to access ExchangeRatesAPI")
	} else {
		HandleExchangeRatesAPI(ExchangeRatesApiResponse, peg)
	}
}
