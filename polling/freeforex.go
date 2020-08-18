package polling

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pegnet/pegnet/common"
	"github.com/zpatrick/go-config"
)

// FreeForexAPIDataSource is the datasource at https://www.freeforexapi.com
//	Notes:
//		- Price quotes are every minute
type FreeForexAPIDataSource struct {
	config *config.Config
}

func NewFreeForexAPIDataSource(config *config.Config) (*FreeForexAPIDataSource, error) {
	s := new(FreeForexAPIDataSource)
	s.config = config

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
	// Does not have all the commodities
	return common.MergeLists(common.CurrencyAssets, []string{"XAU", "XAG"}, common.V4CurrencyAdditions, common.V5CurrencyAdditions)
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
		peg[asset] = PegItem{Value: 1 / currency.Rate, WhenUnix: timestamp.Unix(), When: timestamp}
	}

	return
}

func (d *FreeForexAPIDataSource) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}

func (d *FreeForexAPIDataSource) CallFreeForexAPI() (*FreeForexAPIDataSourceResponse, error) {
	var resp *FreeForexAPIDataSourceResponse

	data, err := d.FetchPeggedPrices()
	if err != nil {
		return nil, err
	}

	resp, err = d.ParseFetchedPrices(data)
	if err != nil {
		// Try the other variation
		return nil, err
	}

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

// ParseFetchedPricesVariation2 is when the api returns a different variation. For some reason this variation is missing
// XAG and XAU. For now, let's just let the original parse fail and exponentially retry.
// TODO: Figure out what the heck is going on when the response format is changed.
// 		Also the date was more than 2 weeks old. So I think this variation is.... bad
func (d *FreeForexAPIDataSource) ParseFetchedPricesVariation2(data []byte) (*FreeForexAPIDataSourceResponse, error) {
	var resp FreeForexAPIDataSourceResponseVariation2
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}

	// Convert to the normal
	var orig FreeForexAPIDataSourceResponse
	orig.Rates = make(map[string]FreeForexAPIDataSourceRate)
	orig.Code = resp.Code
	timestamp, err := time.Parse("2006-01-02", resp.Date)
	if err != nil {
		return nil, err
	}
	for k, v := range resp.Rates {
		orig.Rates[resp.Base+k] = FreeForexAPIDataSourceRate{Rate: v, Timestamp: timestamp.Unix()}
	}
	return &orig, nil
}

func (d *FreeForexAPIDataSource) FetchPeggedPrices() ([]byte, error) {
	client := NewHTTPClient()
	req, err := http.NewRequest("GET", d.ApiUrl()+"live", nil)
	if err != nil {
		return nil, err
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

type FreeForexAPIDataSourceResponseVariation2 struct {
	Code  int                `json:"code"`
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

type FreeForexAPIDataSourceResponse struct {
	Code  int                                   `json:"code"`
	Rates map[string]FreeForexAPIDataSourceRate `json:"rates"`
}

type FreeForexAPIDataSourceRate struct {
	Rate      float64 `json:"rate"`
	Timestamp int64   `json:"timestamp"`
}
