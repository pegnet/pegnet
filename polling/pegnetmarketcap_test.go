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
	FixedDataSourceTest(t, "PegnetMarketCap", []byte(openExchangeRateResponse))
}
