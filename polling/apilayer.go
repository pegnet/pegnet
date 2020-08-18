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

// APILayerDataSource is the datasource at http://www.apilayer.net
type APILayerDataSource struct {
	config *config.Config
}

func NewAPILayerDataSource(config *config.Config) (*APILayerDataSource, error) {
	s := new(APILayerDataSource)
	s.config = config
	return s, nil
}

func (d *APILayerDataSource) Name() string {
	return "APILayer"
}

func (d *APILayerDataSource) Url() string {
	return "https://apilayer.com/"
}

func (d *APILayerDataSource) SupportedPegs() []string {
	return common.MergeLists(common.CurrencyAssets, common.V4CurrencyAdditions, common.V5CurrencyAdditions)
}

func (d *APILayerDataSource) FetchPegPrices() (peg PegAssets, err error) {
	resp, err := CallAPILayer(d.config)
	if err != nil {
		return nil, err
	}

	peg = make(map[string]PegItem)

	timestamp := time.Unix(resp.Timestamp, 0)
	for _, currencyISO := range d.SupportedPegs() {
		// Search for USDXXX pairs
		if v, ok := resp.Quotes["USD"+currencyISO]; ok {
			peg[currencyISO] = PegItem{Value: 1 / v, When: timestamp, WhenUnix: timestamp.Unix()}
		}
	}

	return
}

func (d *APILayerDataSource) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}

// ----

type APILayerResponse struct {
	Success   bool               `json:"success"`
	Terms     string             `json:"terms"`
	Privacy   string             `json:"privacy"`
	Timestamp int64              `json:"timestamp"`
	Source    string             `json:"source"`
	Quotes    map[string]float64 `json:"quotes"`
	Error     APILayerError      `json:"error"`
}

type APILayerError struct {
	Code int64
	Type string
	Info string
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func CallAPILayer(c *config.Config) (response APILayerResponse, err error) {
	var apiLayerResponse APILayerResponse
	var emptyResponse APILayerResponse

	var apikey string
	{
		apikey, err = c.String("Oracle.APILayerKey")
		check(err)
	}

	resp, err := http.Get("http://www.apilayer.net/api/live?access_key=" + apikey)
	if err != nil {
		log.WithError(err).Warning("Failed to get response from API Layer")
	}

	defer resp.Body.Close()
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return emptyResponse, err
	} else if err = json.Unmarshal(body, &apiLayerResponse); err != nil {
		return emptyResponse, err
	}

	return apiLayerResponse, err
}
