// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package polling_test

import (
	"testing"
)

// Need an api key to run this
func TestActualPegnetMarketCapRatesPeggedAssets(t *testing.T) {
	ActualDataSourceTest(t, "PegnetMarketCap")
}

// TestFixedOpenExchangeRatesPeggedAssets tests all the crypto assets are found on OpenExchangeRates from fixed
func TestPegnetMarketCapPeggedAssets(t *testing.T) {
	FixedDataSourceTest(t, "PegnetMarketCap", []byte(pegnetMarketCapData))
}

var pegnetMarketCapData = []byte(`
{
  "ticker_symbol": "PEG",
  "title": "PegNet",
  "icon_file": "peg.png",
  "price": "0.00431459",
  "exchange_price": "0.00512609",
  "price_change": "2.57",
  "exchange_price_change": "-7.99",
  "volume": "19227434.62056507",
  "exchange_volume": "8338581.79170000",
  "volume_price": "81069.09104965",
  "volume_in": "6923201.95237133",
  "volume_in_price": "29532.26212911",
  "volume_tx": "11304232.66819374",
  "volume_tx_price": "47527.28892053",
  "volume_out": "1000000.00000000",
  "volume_out_price": "4009.54000000",
  "supply": "172427384.12301612",
  "supply_change": "3.55",
  "height": 220570,
  "updated_at": 1574888700,
  "deleted_at": null,
  "exchange_price_updated_at": "2019-11-27 21:15:12",
  "exchange_price_dateline": 1574889312
}
`)
