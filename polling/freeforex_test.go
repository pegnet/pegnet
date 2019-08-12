package polling_test

import (
	"testing"
)

// TestActualFreeForexPeggedAssets tests all the crypto assets are found on exchangerates over the net
func TestActualFreeForexPeggedAssets(t *testing.T) {
	ActualDataSourceTest(t, "FreeForexAPI")
}

func TestFixedFreeForexPeggedAssets(t *testing.T) {
	FixedDataSourceTest(t, "FreeForexAPI", []byte(freeforexRateResponse))
}

var freeforexRateResponse = `
{"rates":{"USDUSD":{"rate":1,"timestamp":1565629445},"USDEUR":{"rate":0.891862,"timestamp":1565629445},"USDJPY":{"rate":105.322974,"timestamp":1565629445},"USDGBP":{"rate":0.827835,"timestamp":1565629445},"USDCAD":{"rate":1.32346,"timestamp":1565629445},"USDCHF":{"rate":0.970365,"timestamp":1565629445},"USDINR":{"rate":71.252297,"timestamp":1565629445},"USDSGD":{"rate":1.386602,"timestamp":1565629445},"USDCNY":{"rate":7.058201,"timestamp":1565629445},"USDHKD":{"rate":7.846205,"timestamp":1565629445},"USDKRW":{"rate":1219.665007,"timestamp":1565629445},"USDBRL":{"rate":3.980703,"timestamp":1565629445},"USDPHP":{"rate":52.165024,"timestamp":1565629445},"USDMXN":{"rate":19.54305,"timestamp":1565629445},"USDXAU":{"rate":0.000664,"timestamp":1565629445},"USDXAG":{"rate":0.05855,"timestamp":1565629445}},"code":200}
`
