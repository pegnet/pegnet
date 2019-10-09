package polling

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pegnet/pegnet/common"
	"github.com/zpatrick/go-config"
)

// OneForgeDataSource is the datasource at https://1forge.com
type OneForgeDataSource struct {
	config *config.Config
	apikey string
}

func NewOneForgeDataSourceDataSource(config *config.Config) (*OneForgeDataSource, error) {
	s := new(OneForgeDataSource)
	s.config = config

	var err error
	// Load api key
	s.apikey, err = s.config.String(common.Config1ForgeKey)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (d *OneForgeDataSource) Name() string {
	return "1Forge"
}

func (d *OneForgeDataSource) Url() string {
	return "https://1forge.com"
}

func (d *OneForgeDataSource) ApiUrl() string {
	return "https://forex.1forge.com/1.0.3/"
}

func (d *OneForgeDataSource) SupportedPegs() []string {
	// Does not have all the currencies, commodities, or crypto
	return common.MergeLists([]string{"EUR", "JPY", "GBP", "CAD", "CHF", "SGD", "HKD", "MXN"}, []string{"XAU", "XAG"})
}

// AssetMapping changes some asset symbols to others to match 1forge
func (d *OneForgeDataSource) AssetMapping() map[string]string {
	return map[string]string{
		"XBT":  "BTC",
		"XBC":  "BCH",
		"DASH": "DSH",
	}
}

func (d *OneForgeDataSource) FetchPegPrices() (peg PegAssets, err error) {
	resp, err := d.Call1Forge()
	if err != nil {
		return nil, err
	}

	peg = make(map[string]PegItem)

	respRates := make(map[string]OneForgeDataSourceRate)
	for _, r := range resp {
		respRates[r.Symbol] = r
	}

	mapping := d.AssetMapping()

	// Look for each asset we support
	for _, asset := range d.SupportedPegs() {

		assetSym := asset
		if v, ok := mapping[asset]; ok {
			assetSym = v
		}

		index := fmt.Sprintf("%sUSD", assetSym)
		currency, ok := respRates[index]
		if !ok {
			continue
		}

		timestamp := time.Unix(currency.Timestamp, 0)
		peg[asset] = PegItem{Value: Uint64Value(currency.Price), WhenUnix: timestamp.Unix(), When: timestamp}
	}

	return
}

func (d *OneForgeDataSource) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}

func (d *OneForgeDataSource) Call1Forge() ([]OneForgeDataSourceRate, error) {
	var resp []OneForgeDataSourceRate

	operation := func() error {
		data, err := d.FetchPeggedPrices()
		if err != nil {
			return err
		}

		resp, err = d.ParseFetchedPrices(data)
		if err != nil {
			// Try the other variation
			return err
		}
		return nil
	}

	err := backoff.Retry(operation, PollingExponentialBackOff())
	return resp, err
}

func (d *OneForgeDataSource) ParseFetchedPrices(data []byte) ([]OneForgeDataSourceRate, error) {
	var resp []OneForgeDataSourceRate

	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (d *OneForgeDataSource) FetchPeggedPrices() ([]byte, error) {
	client := NewHTTPClient()
	req, err := http.NewRequest("GET", d.ApiUrl()+"quotes", nil)
	if err != nil {
		return nil, err
	}

	mapping := d.AssetMapping()

	var ids []string
	for _, asset := range d.SupportedPegs() {
		assetSym := asset
		if v, ok := mapping[asset]; ok {
			assetSym = v
		}
		ids = append(ids, assetSym+"USD")
	}

	q := url.Values{}
	q.Add("pairs", strings.Join(ids, ","))
	q.Add("api_key", d.apikey)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

type OneForgeDataSourceRate struct {
	Symbol    string  `json:"symbol"`
	Bid       float64 `json:"bid"`
	Ask       float64 `json:"ask"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
}
