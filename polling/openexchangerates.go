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

// OpenExchangeRatesDataSource is the datasource at "https://openexchangerates.org/"
type OpenExchangeRatesDataSource struct {
	config *config.Config
}

func NewOpenExchangeRatesDataSource(config *config.Config) (*OpenExchangeRatesDataSource, error) {
	s := new(OpenExchangeRatesDataSource)
	s.config = config
	return s, nil
}

func (d *OpenExchangeRatesDataSource) Name() string {
	return "OpenExchangeRates"
}

func (d *OpenExchangeRatesDataSource) Url() string {
	return "https://openexchangerates.org/"
}

func (d *OpenExchangeRatesDataSource) SupportedPegs() []string {
	return common.MergeLists(common.CurrencyAssets, common.CommodityAssets, []string{"XBT"}, common.V4CurrencyAdditions, common.V5CurrencyAdditions)
}

func (d *OpenExchangeRatesDataSource) FetchPegPrices() (peg PegAssets, err error) {
	resp, err := CallOpenExchangeRates(d.config)
	if err != nil {
		return nil, err
	}

	peg = make(map[string]PegItem)

	timestamp := time.Unix(resp.Timestamp, 0)
	for _, currencyISO := range d.SupportedPegs() {
		// Price is inverted
		if v, ok := resp.Rates[currencyISO]; ok {
			peg[currencyISO] = PegItem{Value: 1 / v, When: timestamp, WhenUnix: timestamp.Unix()}
		}

		// Special case for btc
		if currencyISO == "XBT" {
			if v, ok := resp.Rates["BTC"]; ok {
				peg[currencyISO] = PegItem{Value: 1 / v, When: timestamp, WhenUnix: timestamp.Unix()}
			}
		}
	}

	return
}

func (d *OpenExchangeRatesDataSource) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}

// --

type OpenExchangeRatesResponse struct {
	Disclaimer  string             `json:"disclaimer"`
	License     string             `json:"license"`
	Timestamp   int64              `json:"timestamp"`
	Base        string             `json:"base"`
	Error       bool               `json:"error"`
	Status      int64              `json:"status"`
	Message     string             `json:"message"`
	Description string             `json:"description"`
	Rates       map[string]float64 `json:"rates"`
}

func CallOpenExchangeRates(c *config.Config) (response OpenExchangeRatesResponse, err error) {
	var openExchangeRatesResponse OpenExchangeRatesResponse
	var emptyResponse OpenExchangeRatesResponse

	var apikey string
	{
		apikey, err = c.String("Oracle.OpenExchangeRatesKey")
		check(err)
	}

	resp, err := http.Get("https://openexchangerates.org/api/latest.json?app_id=" + apikey)
	if err != nil {
		log.WithError(err).Warning("Failed to get response from OpenExchangeRates")
		return emptyResponse, err
	}

	defer resp.Body.Close()
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return emptyResponse, err
	} else if err = json.Unmarshal(body, &openExchangeRatesResponse); err != nil {
		return emptyResponse, err
	}

	// Price is inverted
	if err == nil {
		for k, v := range openExchangeRatesResponse.Rates {
			openExchangeRatesResponse.Rates[k] = v
		}
	}
	return openExchangeRatesResponse, err
}
