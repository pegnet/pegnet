package polling

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pegnet/pegnet/common"
	"github.com/zpatrick/go-config"
)

// PegnetMarketCap is the datasource at https://pegnetmarketcap.com/
type PegnetMarketCapDataSource struct {
	config *config.Config
}

func NewPegnetMarketCapDataSource(config *config.Config) (*PegnetMarketCapDataSource, error) {
	s := new(PegnetMarketCapDataSource)
	s.config = config

	return s, nil
}

func (d *PegnetMarketCapDataSource) Name() string {
	return "PegnetMarketCap"
}

func (d *PegnetMarketCapDataSource) Url() string {
	return "https://pegnetmarketcap.com/"
}

func (d *PegnetMarketCapDataSource) ApiUrl() string {
	return "https://pegnetmarketcap.com/api/asset/PEG"
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
	timestamp := time.Unix(resp.ExchangePriceUpdatedDateline, 0)
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
	TickerSymbol             string      `json:"ticker_symbol"`
	Title                    string      `json:"title"`
	IconFile                 string      `json:"icon_file"`
	Price                    string      `json:"price"`
	ExchangePrice            string      `json:"exchange_price"`
	PriceChange              string      `json:"price_change"`
	ExchangePriceChange      string      `json:"exchange_price_change"`
	ExchangePriceUpdatedDateline int64       `json:"exchange_price_dateline"`
	Volume                   string      `json:"volume"`
	ExchangeVolume           string      `json:"exchange_volume"`
	VolumePrice              string      `json:"volume_price"`
	VolumeIn                 string      `json:"volume_in"`
	VolumeInPrice            string      `json:"volume_in_price"`
	VolumeTx                 string      `json:"volume_tx"`
	VolumeTxPrice            string      `json:"volume_tx_price"`
	VolumeOut                string      `json:"volume_out"`
	VolumeOutPrice           string      `json:"volume_out_price"`
	Supply                   string      `json:"supply"`
	SupplyChange             string      `json:"supply_change"`
	Height                   int         `json:"height"`
	UpdatedAt                int64       `json:"updated_at"`
	DeletedAt                interface{} `json:"deleted_at"`
	History                  []struct {
		TickerSymbol string `json:"ticker_symbol"`
		Height       int    `json:"height"`
		Price        string `json:"price"`
		Volume       string `json:"volume"`
		VolumeIn     string `json:"volume_in"`
		VolumeTx     string `json:"volume_tx"`
		VolumeOut    string `json:"volume_out"`
		Supply       string `json:"supply"`
		Dateline     int64  `json:"dateline"`
		UpdatedAt    string `json:"updated_at"`
	} `json:"history"`
	ExchangePriceHistory []struct {
		TickerSymbol string `json:"ticker_symbol"`
		QuoteSymbol  string `json:"quote_symbol"`
		Dateline     int64  `json:"dateline"`
		Price        string `json:"price"`
		UpdatedAt    string `json:"updated_at"`
	} `json:"exchange_price_history"`
}
