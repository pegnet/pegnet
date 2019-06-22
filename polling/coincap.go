// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package polling

import (
	"fmt"
	"github.com/zpatrick/go-config"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type CoinCapResponse struct {
	Data      []CoinCapRecord `json:"data"`
	Timestamp int64           `json:"timestamp"`
}

type CoinCapRecord struct {
	ID                string `json:"id"`
	Rank              string `json:"rank"`
	Symbol            string `json:"symbol"`
	Name              string `json:"name"`
	Supply            string `json:"supply"`
	MaxSupply         string `json:"maxSupply"`
	MarketCapUSD      string `json:"marketCapUsd"`
	VolumeUSD24Hr     string `json:"volumeUsd24Hr"`
	PriceUSD          string `json:"priceUsd"`
	ChangePercent24Hr string `json:"changePercent24Hr"`
	VWAP24Hr          string `json:"vwap24Hr"`
}

func CallCoinCap(config *config.Config) ([]byte, error) {

	resp, err := http.Get("http://api.coincap.io/v2/assets?limit=500")
	for i := 0; i < 10; i++ {
		if err != nil {
			time.Sleep(time.Second)
			fmt.Fprintf(os.Stderr, "Error %2d, retrying... %v\n", i+1, err)
			resp, err = http.Get("http://api.coincap.io/v2/assets?limit=500")
		} else {
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			return body, err
		}
	}
	return nil, err

}
