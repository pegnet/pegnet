package polling_test

import (
	"testing"
)

func TestCoinGeckoAssets(t *testing.T) {
	ActualDataSourceTest(t, "CoinGecko")
}

// TestFixedCoinGeckoAssets tests all the crypto assets are found on CoinGecko from fixed
func TestFixedCoinGeckoAssets(t *testing.T) {
	FixedDataSourceTest(t, "CoinGecko", []byte(coinGeckoData))
}

var coinGeckoData = `
{"bitcoin":{"usd":7734.54,"last_updated_at":1575040666},"monero":{"usd":55.79,"last_updated_at":1575040737},"litecoin":{"usd":48.73,"last_updated_at":1575040671},"cardano":{"usd":0.04154594,"last_updated_at":1575040732},"ethereum":{"usd":156.08,"last_updated_at":1575040671},"binancecoin":{"usd":16.02,"last_updated_at":1575040667},"pegnet":{"usd":0.0044042,"last_updated_at":1575022668},"stellar":{"usd":0.059243,"last_updated_at":1575040656},"factom":{"usd":2.71,"last_updated_at":1575040484},"dash":{"usd":58.31,"last_updated_at":1575040663},"bitcoin-cash":{"usd":225.02,"last_updated_at":1575040672},"zcash":{"usd":29.56,"last_updated_at":1575040534},"decred":{"usd":19.17,"last_updated_at":1575040732},"ravencoin":{"usd":0.02333279,"last_updated_at":1575040669}}
`
