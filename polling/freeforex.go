package polling

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff"

	"github.com/pegnet/pegnet/common"
	"github.com/zpatrick/go-config"
)

// FreeForexAPIDataSource is the datasource at https://www.freeforexapi.com
type FreeForexAPIDataSource struct {
	config  *config.Config
	lastPeg PegAssets
	apikey  string
}

func NewFreeForexAPIDataSource(config *config.Config) (*FreeForexAPIDataSource, error) {
	var err error
	s := new(FreeForexAPIDataSource)
	s.config = config

	// Load api key
	s.apikey, err = s.config.String(common.ConfigCoinMarketCapKey)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (d *FreeForexAPIDataSource) Name() string {
	return "FreeForexAPI"
}

func (d *FreeForexAPIDataSource) Url() string {
	return "https://www.freeforexapi.com"
}

func (d *FreeForexAPIDataSource) ApiUrl() string {
	return "https://www.freeforexapi.com/api/"
}

func (d *FreeForexAPIDataSource) SupportedPegs() []string {
	return common.MergeLists(common.CurrencyAssets, []string{"XAU", "XAG"})
}

func (d *FreeForexAPIDataSource) FetchPegPrices() (peg PegAssets, err error) {
	resp, err := d.CallFreeForexAPI()
	if err != nil {
		return nil, err
	}

	peg = make(map[string]PegItem)

	// Look for each asset we support
	for _, asset := range d.SupportedPegs() {
		index := fmt.Sprintf("USD%s", asset)
		currency, ok := resp.Rates[index]
		if !ok {
			continue
		}

		timestamp := time.Unix(currency.Timestamp, 0)
		// The USD price is 1/rate
		peg[asset] = PegItem{Value: currency.Rate, WhenUnix: timestamp.Unix(), When: timestamp}
	}

	return
}

func (d *FreeForexAPIDataSource) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}

func (d *FreeForexAPIDataSource) CallFreeForexAPI() (*FreeForexAPIDataSourceResponse, error) {
	var resp *FreeForexAPIDataSourceResponse

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

func (d *FreeForexAPIDataSource) ParseFetchedPrices(data []byte) (*FreeForexAPIDataSourceResponse, error) {
	var resp FreeForexAPIDataSourceResponse
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (d *FreeForexAPIDataSource) FetchPeggedPrices() ([]byte, error) {
	client := NewHTTPClient()
	req, err := http.NewRequest("GET", d.ApiUrl()+"live", nil)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	var ids []string
	for _, asset := range d.SupportedPegs() {
		ids = append(ids, "USD"+asset)
	}

	q := url.Values{}
	q.Add("pairs", strings.Join(ids, ","))
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

type FreeForexAPIDataSourceResponse struct {
	Code  int                                   `json:"code"`
	Rates map[string]FreeForexAPIDataSourceRate `json:"rates"`
}

type FreeForexAPIDataSourceRate struct {
	Rate      float64 `json:"rate"`
	Timestamp int64   `json:"timestamp"`
}
