package polling

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pegnet/pegnet/common"
)

// PegnetMarketCap is the datasource at https://pegnetmarketcap.com/
type PegnetMarketCapDataSource struct {
}

func NewPegnetMarketCapDataSource() (*PegnetMarketCapDataSource, error) {
	s := new(PegnetMarketCapDataSource)

	return s, nil
}

func (d *PegnetMarketCapDataSource) Name() string {
	return "PegnetMarketCap"
}

func (d *PegnetMarketCapDataSource) Url() string {
	return "https://pegnetmarketcap.com/"
}

func (d *PegnetMarketCapDataSource) ApiUrl() string {
	return "https://pegnetmarketcap.com/api/asset/PEG?columns=ticker_symbol,exchange_price,exchange_price_dateline"
}

func (d *PegnetMarketCapDataSource) SupportedPegs() []string {
	// Does not have all the currencies, commodities, or crypto
	return common.PEGAsset
}

func (d *PegnetMarketCapDataSource) FetchPegPrices() (peg PegAssets, err error) {
	resp, err := d.CallPegnetMarketCap()
	if err != nil {
		return nil, err
	}
	var _ = resp

	peg = make(map[string]PegItem)
	// Only PEG Supported
	timestamp := time.Unix(resp.ExchangePriceDateline, 0)
	price, err := strconv.ParseFloat(resp.ExchangePrice, 64)
	if err != nil {
		return
	}

	peg[d.SupportedPegs()[0]] = PegItem{Value: price, When: timestamp, WhenUnix: timestamp.Unix()}

	return
}

func (d *PegnetMarketCapDataSource) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}

func (d *PegnetMarketCapDataSource) CallPegnetMarketCap() (*PegnetMarketCapResponse, error) {
	var resp *PegnetMarketCapResponse

	operation := func() error {
		data, err := d.FetchPeggedPrices()
		if err != nil {
			return err
		}

		resp, err = d.ParseFetchedPrices(data)
		if err != nil {
			return err
		}
		return nil
	}

	err := backoff.Retry(operation, PollingExponentialBackOff())
	return resp, err
}

func (d *PegnetMarketCapDataSource) ParseFetchedPrices(data []byte) (*PegnetMarketCapResponse, error) {
	var resp PegnetMarketCapResponse
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (d *PegnetMarketCapDataSource) FetchPeggedPrices() ([]byte, error) {
	client := NewHTTPClient()
	req, err := http.NewRequest("GET", d.ApiUrl(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

type PegnetMarketCapResponse struct {
	TickerSymbol          string `json:"ticker_symbol"`
	ExchangePrice         string `json:"exchange_price"`
	ExchangePriceDateline int64  `json:"exchange_price_dateline"`
}
