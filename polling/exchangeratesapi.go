// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package polling

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pegnet/pegnet/common"

	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

// ExchangeRatesDataSource is the datasource at "https://exchangeratesapi.io"
type ExchangeRatesDataSource struct {
	config *config.Config
}

func NewExchangeRatesDataSource(config *config.Config) (*ExchangeRatesDataSource, error) {
	s := new(ExchangeRatesDataSource)
	s.config = config
	return s, nil
}

func (d *ExchangeRatesDataSource) Name() string {
	return "ExchangeRates"
}

func (d *ExchangeRatesDataSource) Url() string {
	return "https://exchangeratesapi.io"
}

func (d *ExchangeRatesDataSource) SupportedPegs() []string {
	return common.CurrencyAssets
}

func (d *ExchangeRatesDataSource) FetchPegPrices() (peg PegAssets, err error) {
	resp, err := CallExchangeRatesAPI(d.config)
	if err != nil {
		return nil, err
	}

	peg = make(map[string]PegItem)

	timestamp, err := time.Parse("2006-01-02", resp.Date)
	if err != nil {
		return nil, err
	}

	for _, currencyISO := range d.SupportedPegs() {
		if v, ok := resp.Rates[currencyISO]; ok {
			peg[currencyISO] = PegItem{Value: 1 / v, When: timestamp, WhenUnix: timestamp.Unix()}
		}
	}

	return
}

func (d *ExchangeRatesDataSource) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}

// ------

type ExchangeRatesAPIResponse struct {
	Date  string             `json:"date"`
	Base  string             `json:"base"`
	Rates map[string]float64 `json:"rates"`
}

func CallExchangeRatesAPI(c *config.Config) (ExchangeRatesAPIResponse, error) {
	var exchangeRatesAPIResponse ExchangeRatesAPIResponse
	var emptyAPIResponse ExchangeRatesAPIResponse

	resp, err := http.Get("https://api.exchangeratesapi.io/latest?base=USD")
	if err != nil {
		log.WithError(err).Warning("Failed to get response from ExchangeRatesAPI")
		return emptyAPIResponse, err
	}

	defer resp.Body.Close()
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return emptyAPIResponse, err
	} else if err = json.Unmarshal(body, &exchangeRatesAPIResponse); err != nil {
		return emptyAPIResponse, err
	}

	return exchangeRatesAPIResponse, err
}
