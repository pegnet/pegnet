package polling

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

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
	return "https://pegnetmarketcap.com/api/asset/all?columns=ticker_symbol,exchange_price,exchange_price_dateline"
	//return "https://pegnetmarketcap.com/api/asset/PEG?columns=ticker_symbol,exchange_price,exchange_price_dateline"
}

func (d *PegnetMarketCapDataSource) SupportedPegs() []string {
	// Does not have all the currencies, commodities, or crypto
	return common.MergeLists(common.PEGAsset)
}

func (d *PegnetMarketCapDataSource) FetchPegPrices() (peg PegAssets, err error) {
	resp, err := d.CallPegnetMarketCap()
	if err != nil {
		return nil, err
	}
	var _ = resp

	peg = make(map[string]PegItem)
	for _, price := range resp {
		switch price.TickerSymbol {
		case "PEG":
			timestamp := time.Unix(price.ExchangePriceDateline, 0)
			rate, err := strconv.ParseFloat(price.ExchangePrice, 64)
			if err != nil {
				continue
			}
			peg[price.TickerSymbol] = PegItem{Value: rate, When: timestamp, WhenUnix: timestamp.Unix()}
		}
	}

	return
}

func (d *PegnetMarketCapDataSource) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}

func (d *PegnetMarketCapDataSource) CallPegnetMarketCap() (map[string]PegnetMarketCapResponse, error) {
	var resp map[string]PegnetMarketCapResponse

	data, err := d.FetchPeggedPrices()
	if err != nil {
		return nil, err
	}

	resp, err = d.ParseFetchedPrices(data)
	if err != nil {
		return nil, err
	}

	return resp, err
}

func (d *PegnetMarketCapDataSource) ParseFetchedPrices(data []byte) (map[string]PegnetMarketCapResponse, error) {
	var resp map[string]PegnetMarketCapResponse
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
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
