package polling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"

	"github.com/cenkalti/backoff"

	"github.com/pegnet/pegnet/common"
	"github.com/zpatrick/go-config"
)

// PegNetIssuanceSource is the datasource that can retrieve PegNet supply data
type PegNetIssuanceSource struct {
	// url location
	location string

	config *config.Config
}

func NewPegNetIssuanceSource(config *config.Config) (*PegNetIssuanceSource, error) {
	s := new(PegNetIssuanceSource)
	s.config = config

	var err error
	// Load pegnet url location
	s.location, err = s.config.String(common.ConfigPegnetdSourceUrl)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (d *PegNetIssuanceSource) Name() string {
	return "PegnetdSource"
}

// Url is just for display purposes for the cli on
func (d *PegNetIssuanceSource) Url() string {
	if d.location == "" {
		return "http://pegnetd/v1"
	}
	return d.location
}

func (d *PegNetIssuanceSource) ApiUrl() string {
	return d.Url()
}

func (d *PegNetIssuanceSource) SupportedPegs() []string {
	// Does not have all the currencies, commodities, or crypto
	return common.MergeLists(common.PEGAsset)
}

func (d *PegNetIssuanceSource) PullPEGPrice(assets PegAssets, dbht int32) (pegQuote uint64, err error) {
	issuance, err := d.FetchIssuance(dbht)
	if err != nil {
		return
	}

	// Market cap in pUSD
	cap := new(big.Int)
	// Calculate market cap
	for asset, quote := range assets {
		if asset != "PEG" {
			asset = "p" + asset
		} else {
			continue // We don't add the PEG to the market cap
		}
		// TODO: Probably should reuse the same function instead of a copy paste
		// TODO: Change all price quotes to return uint64, and toss floats from the polling
		// The quotes are provided in USD, and the issuance is in the pAsset. To get the total market cap,
		// you take the issuance * price quote, where the quote is in pUSD and the issuance is in the
		// pAsset.
		pUSDQuote := quote.Value
		// Supply * price
		add := new(big.Int).Mul(new(big.Int).SetUint64(issuance.Issuance[asset]), new(big.Int).SetUint64(pUSDQuote))
		cap.Add(cap, add)
	}

	// TODO: Should we validate all assets were returned in the issuance?

	pegSupply, ok := issuance.Issuance["PEG"]
	if !ok {
		return 0, fmt.Errorf("no PEG supply given")
	}

	// Now set peg Price
	if pegSupply == 0 { // No divide by 0 error
		return 0, nil
	}
	peg := new(big.Int).Div(cap, new(big.Int).SetUint64(pegSupply))
	return peg.Uint64(), nil
}

func (d *PegNetIssuanceSource) FetchIssuance(dbht int32) (*PegNetSourceIssuance, error) {
	var resp *PegNetSourceIssuance

	operation := func() error {
		data, err := d.CallIssuance()
		if err != nil {
			return err
		}

		resp, err = d.ParseIssuance(data)
		if err != nil {
			return err
		}

		if resp.SyncStatus.SyncHeight != dbht-1 {
			return fmt.Errorf("PEG datasource is not synced, at %d/%d",
				resp.SyncStatus.SyncHeight, dbht)
		}

		return nil
	}

	err := backoff.Retry(operation, PollingExponentialBackOff())
	return resp, err
}

func (d *PegNetIssuanceSource) CallIssuance() ([]byte, error) {
	client := NewHTTPClient()
	var buf bytes.Buffer
	// get-pegnet-issuance method on api
	buf.Write([]byte(`{"jsonrpc":"2.0","method":"get-pegnet-issuance","id":0}`))
	req, err := http.NewRequest("GET", d.ApiUrl(), &buf)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func (d *PegNetIssuanceSource) ParseIssuance(data []byte) (*PegNetSourceIssuance, error) {
	var rpcResp JsonRPC
	err := json.Unmarshal(data, &rpcResp)
	if err != nil {
		return nil, err
	}
	if rpcResp.Error.Code != 0 {
		return nil, fmt.Errorf("%v", rpcResp.Error)
	}

	var issuance PegNetSourceIssuance
	err = json.Unmarshal(rpcResp.Result, &issuance)
	if err != nil {
		return nil, err
	}

	return &issuance, nil
}

func (d *PegNetIssuanceSource) FetchPegPrices() (peg PegAssets, err error) {
	// Do not return anything. The PEG price will be 0 if used as a regular data-source.
	peg = make(PegAssets)
	peg["PEG"] = PegItem{}
	return
}

func (d *PegNetIssuanceSource) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}

type JsonRPC struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Method string `json:"method"`
		} `json:"data"`
	} `json:"error"`
}

type PegNetSourceIssuance struct {
	SyncStatus struct {
		SyncHeight   int32
		FactomHeight int32
	} `json:"sync-status"`
	Issuance map[string]uint64 `json:"issuance"`
}
