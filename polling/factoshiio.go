package polling

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pegnet/pegnet/common"
)

// FactoshiioDataSource is the datasource at https://pegapi.factoshi.io/
type FactoshiioDataSource struct {
}

func NewFactoshiioDataSource() (*FactoshiioDataSource, error) {
	s := new(FactoshiioDataSource)

	return s, nil
}

func (d *FactoshiioDataSource) Name() string {
	return "Factoshiio"
}

func (d *FactoshiioDataSource) Url() string {
	return "https://factoshi.io/"
}

func (d *FactoshiioDataSource) ApiUrl() string {
	return "https://pegapi.factoshi.io/"
}

func (d *FactoshiioDataSource) SupportedPegs() []string {
	return common.PEGAsset
}

func (d *FactoshiioDataSource) FetchPegPrices() (peg PegAssets, err error) {
	resp, err := d.CallFactoshiio()
	if err != nil {
		return nil, err
	}
	var _ = resp

	peg = make(map[string]PegItem)
	// Only PEG Supported
	timestamp := time.Unix(resp.UpdatedAt, 0)

	peg[d.SupportedPegs()[0]] = PegItem{Value: resp.Price, When: timestamp, WhenUnix: timestamp.Unix()}

	return
}

func (d *FactoshiioDataSource) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}

func (d *FactoshiioDataSource) CallFactoshiio() (*FactoshiioDataResponse, error) {
	var resp *FactoshiioDataResponse

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

func (d *FactoshiioDataSource) ParseFetchedPrices(data []byte) (*FactoshiioDataResponse, error) {
	var resp FactoshiioDataResponse
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (d *FactoshiioDataSource) FetchPeggedPrices() ([]byte, error) {
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

type FactoshiioDataResponse struct {
	Price     float64 `json:"price"`
	Volume    float64 `json:"volume"`
	Quote     string  `json:"quote"`
	Base      string  `json:"base"`
	UpdatedAt int64   `json:"updated_at"`
}
