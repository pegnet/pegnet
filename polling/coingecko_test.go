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
{"eos":{"usd":3.08,"last_updated_at":1578934508},"chainlink":{"usd":2.18,"last_updated_at":1578934522},"bitcoin":{"usd":8090.66,"last_updated_at":1578934526},"ethereum":{"usd":142.94,"last_updated_at":1578934529},"cardano":{"usd":0.03674716,"last_updated_at":1578934525},"cosmos":{"usd":4.15,"last_updated_at":1578934555},"stellar":{"usd":0.04779436,"last_updated_at":1578934547},"binancecoin":{"usd":15.02,"last_updated_at":1578934542},"basic-attention-token":{"usd":0.186574,"last_updated_at":1578934540},"bitcoin-cash":{"usd":263.05,"last_updated_at":1578934532},"litecoin":{"usd":49.41,"last_updated_at":1578934516},"pegnet":{"usd":0.0016529,"last_updated_at":1578934509},"monero":{"usd":57.61,"last_updated_at":1578934550},"dash":{"usd":65.06,"last_updated_at":1578934527},"zcash":{"usd":36.78,"last_updated_at":1578934553},"factom":{"usd":1.91,"last_updated_at":1578934362},"decred":{"usd":17.16,"last_updated_at":1578934549},"link":{"usd":4.76,"last_updated_at":1578933776},"ravencoin":{"usd":0.02356063,"last_updated_at":1578934520}}
`
