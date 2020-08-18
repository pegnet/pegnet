package polling

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pegnet/pegnet/common"
)

// https://api.coingecko.com/api/v3/simple/price?ids=pegnet&vs_currencies=usd

// CoinGeckoDataSource is the datasource at https://www.coingecko.com
type CoinGeckoDataSource struct {
}

func NewCoinGeckoDataSource() (*CoinGeckoDataSource, error) {
	s := new(CoinGeckoDataSource)

	return s, nil
}

func (d *CoinGeckoDataSource) Name() string {
	return "CoinGecko"
}

func (d *CoinGeckoDataSource) Url() string {
	return "https://www.coingecko.com/"
}

func (d *CoinGeckoDataSource) ApiUrl() string {
	return "https://api.coingecko.com/api/v3/"
}

func (d *CoinGeckoDataSource) SupportedPegs() []string {
	return common.MergeLists(common.PEGAsset, common.CryptoAssets, common.V4CryptoAdditions, common.V5CryptoAdditions)
}

func (d *CoinGeckoDataSource) FetchPegPrices() (peg PegAssets, err error) {
	resp, err := d.CallCoinGecko()
	if err != nil {
		return nil, err
	}

	peg = make(map[string]PegItem)
	mapping := d.CurrencyIDMapping()
	for _, asset := range d.SupportedPegs() {
		v, ok := resp[mapping[asset]]
		if ok {
			timestamp := time.Unix(v.UpdatedAt, 0)
			peg[asset] = PegItem{Value: v.UsdQuote, When: timestamp, WhenUnix: timestamp.Unix()}
		}
	}

	return
}

func (d *CoinGeckoDataSource) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}

func (d *CoinGeckoDataSource) CallCoinGecko() (map[string]CoinGeckoDataSourceResponse, error) {
	resp := make(map[string]CoinGeckoDataSourceResponse)

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

func (d *CoinGeckoDataSource) ParseFetchedPrices(data []byte) (map[string]CoinGeckoDataSourceResponse, error) {
	resp := make(map[string]CoinGeckoDataSourceResponse)
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (d *CoinGeckoDataSource) FetchPeggedPrices() ([]byte, error) {
	client := NewHTTPClient()
	req, err := http.NewRequest("GET", d.ApiUrl()+"simple/price", nil)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Add("vs_currencies", "usd")
	q.Add("include_last_updated_at", "true")

	mapping := d.CurrencyIDMapping()
	ids := make([]string, len(d.SupportedPegs()))
	for i, cur := range d.SupportedPegs() {
		ids[i] = mapping[cur]
	}
	q.Add("ids", strings.Join(ids, ","))

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

// CurrencyIDMapping finds the coingecko name for each currency vs using the syms.
func (d *CoinGeckoDataSource) CurrencyIDMapping() map[string]string {
	return map[string]string{
		"PEG":  "pegnet",
		"XBT":  "bitcoin",
		"ETH":  "ethereum",
		"LTC":  "litecoin",
		"RVN":  "ravencoin",
		"XBC":  "bitcoin-cash",
		"FCT":  "factom",
		"BNB":  "binancecoin",
		"XLM":  "stellar",
		"ADA":  "cardano",
		"XMR":  "monero",
		"DASH": "dash",
		"ZEC":  "zcash",
		"DCR":  "decred",
		// V4 Adds
		"EOS":  "eos",
		"LINK": "chainlink",
		"ATOM": "cosmos",
		"BAT":  "basic-attention-token",
		"XTZ":  "tezos",
		// V5 Adds
		"HBAR": "hedera-hashgraph",
		"NEO":  "neo",
		"CRO":  "crypto-com-chain",
		"ETC":  "ethereum-classic",
		"ONT":  "ontology",
		"DOGE": "dogecoin",
		"VET":  "vechain",
		"HT":   "huobi-token",
		"ALGO": "algorand",
		"DGB":  "digibyte",
	}
}

type CoinGeckoDataSourceResponse struct {
	UsdQuote  float64 `json:"usd"`
	UpdatedAt int64   `json:"last_updated_at"`
}
