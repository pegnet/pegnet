package polling_test

import (
	"testing"
)

// TestActualFreeForexPeggedAssets tests all the crypto assets are found on exchangerates over the net
func TestActualFreeForexPeggedAssets(t *testing.T) {
	return // FreeForex is down right now. Has been for a few days
	// This sometimes fails because the data source sometimes returns a second variation response.
	ActualDataSourceTest(t, "FreeForexAPI")
}

func TestFixedFreeForexPeggedAssets(t *testing.T) {
	FixedDataSourceTest(t, "FreeForexAPI", []byte(freeforexRateResponse))
	//FixedDataSourceTest(t, "FreeForexAPI", []byte(freeforexRateResponse2))
}

// Yes they have more than 1 type of response... idk why. The docs specify 1, experimentation shows the second as well
var freeforexRateResponse = `
{"rates":{"USDUSD":{"rate":1,"timestamp":1578937865},"USDEUR":{"rate":0.897626,"timestamp":1578937865},"USDJPY":{"rate":109.914503,"timestamp":1578937865},"USDGBP":{"rate":0.769495,"timestamp":1578937865},"USDCAD":{"rate":1.30478,"timestamp":1578937865},"USDCHF":{"rate":0.97017,"timestamp":1578937865},"USDINR":{"rate":70.738039,"timestamp":1578937865},"USDSGD":{"rate":1.346605,"timestamp":1578937865},"USDCNY":{"rate":6.893701,"timestamp":1578937865},"USDHKD":{"rate":7.77185,"timestamp":1578937865},"USDKRW":{"rate":1155.57985,"timestamp":1578937865},"USDBRL":{"rate":4.1344,"timestamp":1578937865},"USDPHP":{"rate":50.430424,"timestamp":1578937865},"USDMXN":{"rate":18.79917,"timestamp":1578937865},"USDXAU":{"rate":0.000645,"timestamp":1578937865},"USDXAG":{"rate":0.055539,"timestamp":1578937865},"USDAUD":{"rate":1.446895,"timestamp":1578937865},"USDNZD":{"rate":1.507615,"timestamp":1578937865},"USDSEK":{"rate":9.466102,"timestamp":1578937865},"USDNOK":{"rate":8.894605,"timestamp":1578937865},"USDRUB":{"rate":61.232902,"timestamp":1578937865},"USDZAR":{"rate":14.42291,"timestamp":1578937865},"USDTRY":{"rate":5.866804,"timestamp":1578937865},"USDAED":{"rate":3.67,"timestamp":1578937865},"USDARS":{"rate":67.49,"timestamp":1578937865},"USDTWD":{"rate":29.92,"timestamp":1578937865},"USDRWF":{"rate":0.0011,"timestamp":1578937865},"USDKES":{"rate":0.0094,"timestamp":1578937865},"USDUGX":{"rate":0.00027,"timestamp":1578937865},"USDTZS":{"rate":0.00043,"timestamp":1578937865},"USDBIF":{"rate":0.00052,"timestamp":1578937865},"USDETB":{"rate":0.029,"timestamp":1578937865},"USDNGN":{"rate":0.0026,"timestamp":1578937865}},"code":200}
`

// The second source is missing gold and silver.
var freeforexRateResponse2 = `{"rates":{"CAD":1.3076237182,"HKD":7.8096299599,"ISK":124.7436469015,"PHP":51.1110120374,"DKK":6.6571555952,"HUF":289.7458760588,"CZK":22.7677218012,"GBP":0.8022113241,"RON":4.210878288,"SEK":9.4016941596,"IDR":13945.0022291574,"INR":68.9331252786,"BRL":3.7434685689,"RUB":62.9982166741,"HRK":6.5871600535,"JPY":107.9179670085,"THB":30.8452964779,"CHF":0.9814534106,"EUR":0.8916629514,"MYR":4.1129736959,"BGN":1.7439144004,"TRY":5.6818546589,"CNY":6.8807846634,"NOK":8.5844850646,"NZD":1.4750780205,"ZAR":13.8922871155,"USD":1.0,"MXN":19.0295140437,"SGD":1.3606776638,"AUD":1.4188140883,"ILS":3.5329469461,"KRW":1177.2982612572,"PLN":3.7877842176},"base":"USD","date":"2019-07-22"}
`
