package polling_test

import (
	"testing"
)

// TestActualExchangeRatesPeggedAssets tests all the crypto assets are found on exchangerates over the net
func TestActualExchangeRatesPeggedAssets(t *testing.T) {
	ActualDataSourceTest(t, "ExchangeRates")
}

func TestFixedExchangeRatesPeggedAssets(t *testing.T) {
	FixedDataSourceTest(t, "ExchangeRates", []byte(exchangeRateResponse))
}

var exchangeRateResponse = `{"rates":{"CAD":1.3076237182,"HKD":7.8096299599,"ISK":124.7436469015,"PHP":51.1110120374,"DKK":6.6571555952,"HUF":289.7458760588,"CZK":22.7677218012,"GBP":0.8022113241,"RON":4.210878288,"SEK":9.4016941596,"IDR":13945.0022291574,"INR":68.9331252786,"BRL":3.7434685689,"RUB":62.9982166741,"HRK":6.5871600535,"JPY":107.9179670085,"THB":30.8452964779,"CHF":0.9814534106,"EUR":0.8916629514,"MYR":4.1129736959,"BGN":1.7439144004,"TRY":5.6818546589,"CNY":6.8807846634,"NOK":8.5844850646,"NZD":1.4750780205,"ZAR":13.8922871155,"USD":1.0,"MXN":19.0295140437,"SGD":1.3606776638,"AUD":1.4188140883,"ILS":3.5329469461,"KRW":1177.2982612572,"PLN":3.7877842176},"base":"USD","date":"2019-07-22"}`
