// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package polling

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pegnet/pegnet/common"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

// CoinCapDataSource is the datasource at https://coincap.io/
type CoinCapDataSource struct {
	config *config.Config
}

func NewCoinCapDataSource(config *config.Config) (*CoinCapDataSource, error) {
	s := new(CoinCapDataSource)
	s.config = config
	return s, nil
}

func (d *CoinCapDataSource) Name() string {
	return "CoinCap"
}

func (d *CoinCapDataSource) Url() string {
	return "https://coincap.io/"
}

func (d *CoinCapDataSource) SupportedPegs() []string {
	return common.CryptoAssets
}

func (d *CoinCapDataSource) FetchPegPrices() (peg PegAssets, err error) {
	resp, err := CallCoinCap(d.config)
	if err != nil {
		return nil, err
	}

	peg = make(map[string]PegItem)

	var UnixTimestamp = resp.Timestamp
	timestamp := time.Unix(resp.Timestamp, 0)

	for _, currency := range resp.Data {
		switch currency.Symbol {
		case "BTC", "XBT":
			value, err := strconv.ParseFloat(currency.PriceUSD, 64)
			peg["XBT"] = PegItem{Value: value, WhenUnix: UnixTimestamp, When: timestamp}
			if err != nil {
				return nil, err
			}
		case "BCH", "XBC":
			value, err := strconv.ParseFloat(currency.PriceUSD, 64)
			peg["XBC"] = PegItem{Value: value, WhenUnix: UnixTimestamp, When: timestamp}
			if err != nil {
				return nil, err
			}
		case "ZCASH", "ZEC":
			value, err := strconv.ParseFloat(currency.PriceUSD, 64)
			peg["ZEC"] = PegItem{Value: value, WhenUnix: UnixTimestamp, When: timestamp}
			if err != nil {
				return nil, err
			}
		default:
			// See if the ticker is in our crypto currency list
			if common.AssetListContains(common.CryptoAssets, currency.Symbol) {
				value, err := strconv.ParseFloat(currency.PriceUSD, 64)
				peg[currency.Symbol] = PegItem{Value: value, WhenUnix: UnixTimestamp, When: timestamp}
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return
}

func (d *CoinCapDataSource) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}

// -----

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

// CoinCapIOCryptoAssetNames is used by coincapio to query for the crypto we care about
var CoinCapIOCryptoAssetNames = map[string]string{
	"XBT":  "bitcoin",
	"ETH":  "ethereum",
	"LTC":  "litecoin",
	"RVN":  "ravencoin",
	"XBC":  "bitcoin-cash",
	"FCT":  "factom",
	"BNB":  "binance-coin",
	"XLM":  "stellar",
	"ADA":  "cardano",
	"XMR":  "monero",
	"DASH": "dash",
	"ZEC":  "zcash",
	"DCR":  "decred",
}

func CallCoinCap(config *config.Config) (CoinCapResponse, error) {
	var CoinCapResponse CoinCapResponse

	var ids []string
	// Need to append all the ids we care about for the call
	for _, a := range common.CryptoAssets {
		ids = append(ids, CoinCapIOCryptoAssetNames[a])
	}

	operation := func() error {
		url := "http://api.coincap.io/v2/assets?ids=" + strings.Join(ids, ",")
		resp, err := http.Get(url)
		if err != nil {
			log.WithError(err).Warning("Failed to get response from CoinCap")
			return err
		}

		defer resp.Body.Close()
		if body, err := ioutil.ReadAll(resp.Body); err != nil {
			return err
		} else if err = json.Unmarshal(body, &CoinCapResponse); err != nil {
			return err
		}
		return nil
	}

	err := backoff.Retry(operation, PollingExponentialBackOff())
	return CoinCapResponse, err
}
