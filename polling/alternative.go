package polling

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/pegnet/pegnet/common"
	"github.com/zpatrick/go-config"
)

// AlternativeMeDataSource is the datasource at https://alternative.me/crypto/api/
//	Notes:
//		- Price quotes are every 5 minutes
type AlternativeMeDataSource struct {
	config *config.Config
}

func NewAlternativeMeDataSource(config *config.Config) (*AlternativeMeDataSource, error) {
	s := new(AlternativeMeDataSource)
	s.config = config

	return s, nil
}

func (d *AlternativeMeDataSource) Name() string {
	return "AlternativeMe"
}

func (d *AlternativeMeDataSource) Url() string {
	return "https://alternative.me/crypto/api/"
}

func (d *AlternativeMeDataSource) ApiUrl() string {
	return "https://api.alternative.me/v2/"
}

func (d *AlternativeMeDataSource) SupportedPegs() []string {
	// Does not have all the currencies, commodities, or crypto
	return common.MergeLists(common.CryptoAssets, []string{"EOS", "LINK", "BAT", "NEO", "ETC", "ONT", "DOGE", "HT"})
}

// AssetMapping changes some asset symbols to others to match 1forge
func (d *AlternativeMeDataSource) AssetMapping() map[string]int {
	return map[string]int{
		"XBT":  1,
		"ETH":  1027,
		"LTC":  2,
		"RVN":  2577,
		"XBC":  1831,
		"FCT":  1087,
		"BNB":  1839,
		"XLM":  512,
		"ADA":  2010,
		"XMR":  328,
		"DASH": 131,
		"ZEC":  1437,
		"DCR":  1168,

		//"PEG": NO PEG,
		"EOS":  1765,
		"LINK": 1975,
		"XTZ":  2011,
		"BAT":  1697,
		//"ATOM": NO ATOM,

		"NEO":  1376,
		"ETC":  1321,
		"ONT":  2566,
		"DOGE": 74,
		"HT":   2502,
	}
}

func (d *AlternativeMeDataSource) FetchPegPrices() (peg PegAssets, err error) {
	resp, err := d.CallAlternativeMe()
	if err != nil {
		return nil, err
	}

	peg = make(map[string]PegItem)
	mapping := d.AssetMapping()

	// Look for each asset we support
	for _, asset := range d.SupportedPegs() {
		id := mapping[asset]
		index := fmt.Sprintf("%d", id)
		currency, ok := resp.Data[index]
		if !ok {
			continue
		}

		// Find us quote
		usdQuote, ok := currency.Quotes["USD"]
		if !ok {
			continue
		}

		timestamp := time.Unix(currency.LastUpdated, 0)
		peg[asset] = PegItem{Value: usdQuote.Price, WhenUnix: timestamp.Unix(), When: timestamp}
	}

	return
}

func (d *AlternativeMeDataSource) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}

func (d *AlternativeMeDataSource) CallAlternativeMe() (*AlternativeMeDataSourceResponse, error) {
	var resp *AlternativeMeDataSourceResponse

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

func (d *AlternativeMeDataSource) ParseFetchedPrices(data []byte) (*AlternativeMeDataSourceResponse, error) {
	var resp AlternativeMeDataSourceResponse
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (d *AlternativeMeDataSource) FetchPeggedPrices() ([]byte, error) {
	client := NewHTTPClient()
	req, err := http.NewRequest("GET", d.ApiUrl()+"ticker", nil)
	if err != nil {
		return nil, err
	}

	mapping := d.AssetMapping()
	var ids []string
	for _, asset := range d.SupportedPegs() {
		ids = append(ids, fmt.Sprintf("%d", mapping[asset]))
	}

	q := url.Values{}
	//q.Add("id", strings.Join(ids, ","))
	q.Add("convert", "USD") // We want usd prices
	q.Add("limit", "300")

	req.Header.Set("Accepts", "application/json")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

type AlternativeMeDataSourceResponse struct {
	Data     map[string]AlternativeMeDataSourceRate `json:"data"`
	Metadata AlternativeMeDataSourceMetadata        `json:"metadata"`
}

type AlternativeMeDataSourceMetadata struct {
	Timestamp           int         `json:"timestamp"`
	NumCryptocurrencies int         `json:"num_cryptocurrencies"`
	Error               interface{} `json:"error"`
}

type AlternativeMeDataSourceRate struct {
	ID                int                                     `json:"id"`
	Name              string                                  `json:"name"`
	Symbol            string                                  `json:"symbol"`
	WebsiteSlug       string                                  `json:"website_slug"`
	Rank              int                                     `json:"rank"`
	CirculatingSupply float64                                 `json:"circulating_supply"`
	TotalSupply       float64                                 `json:"total_supply"`
	MaxSupply         float64                                 `json:"max_supply"`
	Quotes            map[string]AlternativeMeDataSourceQuote `json:"quotes"`
	LastUpdated       int64                                   `json:"last_updated"`
}

type AlternativeMeDataSourceQuote struct {
	Price               float64 `json:"price"`
	Volume24H           float64 `json:"volume_24h"`
	MarketCap           float64 `json:"market_cap"`
	PercentageChange1H  float64 `json:"percentage_change_1h"`
	PercentageChange24H float64 `json:"percentage_change_24h"`
	PercentageChange7D  float64 `json:"percentage_change_7d"`
}
