package polling_test

import (
	"testing"
)

func TestCoinGeckoAssets(t *testing.T) {
	ActualDataSourceTest(t, "CoinGecko", "AED")
}

// TestFixedCoinGeckoAssets tests all the crypto assets are found on CoinGecko from fixed
func TestFixedCoinGeckoAssets(t *testing.T) {
	FixedDataSourceTest(t, "CoinGecko", []byte(coinGeckoData), "AED")
}

var coinGeckoData = `
{"monero":{"usd":57.81,"last_updated_at":1578961668},"binancecoin":{"usd":15.19,"last_updated_at":1578961673},"ethereum":{"usd":143.88,"last_updated_at":1578961559},"cosmos":{"usd":4.26,"last_updated_at":1578961801},"litecoin":{"usd":49.92,"last_updated_at":1578961538},"tezos":{"usd":1.28,"last_updated_at":1578961792},"eos":{"usd":3.12,"last_updated_at":1578961560},"cardano":{"usd":0.03709022,"last_updated_at":1578961796},"bitcoin":{"usd":8122.16,"last_updated_at":1578961554},"chainlink":{"usd":2.2,"last_updated_at":1578961801},"dash":{"usd":69.87,"last_updated_at":1578961792},"stellar":{"usd":0.04817891,"last_updated_at":1578961674},"basic-attention-token":{"usd":0.186364,"last_updated_at":1578961799},"bitcoin-cash":{"usd":269.29,"last_updated_at":1578961555},"pegnet":{"usd":0.00170656,"last_updated_at":1578961757},"zcash":{"usd":38.23,"last_updated_at":1578961553},"factom":{"usd":1.9,"last_updated_at":1578961778},"decred":{"usd":16.09,"last_updated_at":1578961802},"ravencoin":{"usd":0.02356596,"last_updated_at":1578961807},
	"hedera-hashgraph":{"usd":0.03862941,"last_updated_at":1578961668},
	"neo":{"usd":9.98,"last_updated_at":1578961668},
	"crypto-com-chain":{"usd":0.069124,"last_updated_at":1578961668},
	"ethereum-classic":{"usd":7.09,"last_updated_at":1578961668},
	"ontology":{"usd":0.483145,"last_updated_at":1578961668},
	"dogecoin":{"usd":0.00266085,"last_updated_at":1578961668},
	"vechain":{"usd":0.0045895,"last_updated_at":1578961668},
	"huobi-token":{"usd":4,"last_updated_at":15789616682},
	"algorand":{"usd":0.19,"last_updated_at":15789616682},
	"digibyte": {"usd": 0.01868084,"last_updated_at": 1593191599}
}
`
