// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package polling_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/pegnet/pegnet/modules/opr"

	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/polling"
	"github.com/pegnet/pegnet/testutils"
	"github.com/zpatrick/go-config"
)

// TestFixedPegnetdSourcePeggedAssets
func TestFixedPegnetdSourcePeggedAssets(t *testing.T) {
	defer func() { http.DefaultClient = &http.Client{} }() // Don't leave http broken

	c := config.NewConfig([]config.Provider{common.NewUnitTestConfigProvider()})

	var fixed = []byte(`{"jsonrpc":"2.0","result":{"sync-status":{"syncheight":213433,"factomheight":213433},"issuance":{"PEG":3506000000000000,"pADA":9174741510,"pBNB":23964155,"pBRL":1553769559,"pCAD":505757599,"pCHF":378030672,"pCNY":2716169709,"pDASH":5317841,"pDCR":21694146,"pETH":2129855,"pEUR":1204132125,"pFCT":4841496454198,"pGBP":308496745,"pHKD":2980001159,"pINR":26981732013,"pJPY":40694137315,"pKRW":454720374337,"pLTC":6640888,"pMXN":7428080364,"pPHP":19717349003,"pRVN":11888226875,"pSGD":524774284,"pUSD":2461994629,"pXAG":21842549,"pXAU":25218137,"pXBC":1623103,"pXBT":326931234,"pXLM":6158039443,"pXMR":6690407,"pZEC":10101202}},"id":662}`)

	// Set default http client to return what we expect from apilayer
	cl := testutils.GetClientWithFixedResp(fixed)
	http.DefaultClient = cl
	polling.NewHTTPClient = func() *http.Client {
		return testutils.GetClientWithFixedResp(fixed)
	}

	s, err := polling.NewDataSource("PegnetdSource", c)
	if err != nil {
		t.Error(err)
	}

	ps := s.(*polling.PegNetIssuanceSource)
	assets := make(polling.PegAssets)
	now := time.Now()
	for _, asset := range opr.V2Assets {
		assets[asset] = polling.PegItem{
			Value:    1,
			When:     now,
			WhenUnix: now.Unix(),
		}
	}

	// All priced at 1 USD
	pegPrice, err := ps.PullPEGPrice(assets)
	if err != nil {
		t.Error(err)
	}

	exp := 0.00154915699 * 1e8
	if pegPrice != uint64(exp) {
		t.Errorf("PEG Quote is incorrect")
	}
}
