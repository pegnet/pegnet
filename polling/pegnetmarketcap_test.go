// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package polling_test

import (
	"testing"
)

func TestActualPegnetMarketCapRatesPeggedAssets(t *testing.T) {
	ActualDataSourceTest(t, "PegnetMarketCap")
}

// TestFixedPegnetMarketCapPeggedAssets tests all the crypto assets are found on PegnetMarketCap from fixed
func TestFixedPegnetMarketCapPeggedAssets(t *testing.T) {
	FixedDataSourceTest(t, "PegnetMarketCap", []byte(pegnetMarketCapData))
}

var pegnetMarketCapData = []byte(`
{"0":{"ticker_symbol":"pADA","exchange_price":"0.00000000","exchange_price_dateline":0},"14":{"ticker_symbol":"pGBP","exchange_price":"0.00000000","exchange_price_dateline":0},"2":{"ticker_symbol":"pBNB","exchange_price":"0.00000000","exchange_price_dateline":0},"3":{"ticker_symbol":"pBRL","exchange_price":"0.00000000","exchange_price_dateline":0},"4":{"ticker_symbol":"pBTC","exchange_price":"0.00000000","exchange_price_dateline":1579019297},"5":{"ticker_symbol":"pCAD","exchange_price":"0.00000000","exchange_price_dateline":0},"6":{"ticker_symbol":"pCHF","exchange_price":"0.00000000","exchange_price_dateline":0},"7":{"ticker_symbol":"pCNY","exchange_price":"0.00000000","exchange_price_dateline":0},"8":{"ticker_symbol":"pDASH","exchange_price":"0.00000000","exchange_price_dateline":0},"9":{"ticker_symbol":"pDCR","exchange_price":"0.00000000","exchange_price_dateline":0},"10":{"ticker_symbol":"PEG","exchange_price":"0.00158254","exchange_price_dateline":1579019357},"11":{"ticker_symbol":"pETH","exchange_price":"0.00000000","exchange_price_dateline":1579019297},"12":{"ticker_symbol":"pEUR","exchange_price":"0.00000000","exchange_price_dateline":0},"13":{"ticker_symbol":"pFCT","exchange_price":"2.11774881","exchange_price_dateline":1579019357},"1":{"ticker_symbol":"pBCH","exchange_price":"0.00000000","exchange_price_dateline":0},"16":{"ticker_symbol":"pHKD","exchange_price":"0.00000000","exchange_price_dateline":0},"15":{"ticker_symbol":"pGOLD","exchange_price":"0.00000000","exchange_price_dateline":1579019297},"17":{"ticker_symbol":"pINR","exchange_price":"0.00000000","exchange_price_dateline":0},"18":{"ticker_symbol":"pJPY","exchange_price":"0.00000000","exchange_price_dateline":0},"19":{"ticker_symbol":"pKRW","exchange_price":"0.00000000","exchange_price_dateline":0},"20":{"ticker_symbol":"pLTC","exchange_price":"0.00000000","exchange_price_dateline":0},"21":{"ticker_symbol":"pMXN","exchange_price":"0.00000000","exchange_price_dateline":0},"22":{"ticker_symbol":"pPHP","exchange_price":"0.00000000","exchange_price_dateline":0},"23":{"ticker_symbol":"pRVN","exchange_price":"0.00000000","exchange_price_dateline":0},"24":{"ticker_symbol":"pSGD","exchange_price":"0.00000000","exchange_price_dateline":0},"25":{"ticker_symbol":"pSILVER","exchange_price":"0.00000000","exchange_price_dateline":0},"26":{"ticker_symbol":"pUSD","exchange_price":"0.77289211","exchange_price_dateline":1579019357},"27":{"ticker_symbol":"pXLM","exchange_price":"0.00000000","exchange_price_dateline":0},"28":{"ticker_symbol":"pXMR","exchange_price":"0.00000000","exchange_price_dateline":0},"29":{"ticker_symbol":"pZEC","exchange_price":"0.00000000","exchange_price_dateline":0}}
`)
