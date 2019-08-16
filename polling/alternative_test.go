package polling_test

import (
	"testing"
)

//func TestActualAlternativePeggedAssets(t *testing.T) {
//	ActualDataSourceTest(t, "AlternativeMe")
//}

func TestFixedAlternativePeggedAssets(t *testing.T) {
	FixedDataSourceTest(t, "AlternativeMe", []byte(alternativeResp))
}

var alternativeResp = `
{
	"data": {
		"1": {
			"id": 1,
			"name": "Bitcoin",
			"symbol": "BTC",
			"website_slug": "bitcoin",
			"rank": 1,
			"circulating_supply": 17880475,
			"total_supply": 17418787,
			"max_supply": 21000000,
			"quotes": {
				"USD": {
					"price": 10450.1431242258,
					"volume_24h": 3708826379.50524,
					"market_cap": 186688608673.694,
					"percentage_change_1h": 0.21,
					"percentage_change_24h": 0.75,
					"percentage_change_7d": -12.17
				}
			},
			"last_updated": 1565994301
		},
		"1027": {
			"id": 1027,
			"name": "Ethereum",
			"symbol": "ETH",
			"website_slug": "ethereum",
			"rank": 2,
			"circulating_supply": 107349512.1865,
			"total_supply": 103748085,
			"max_supply": 0,
			"quotes": {
				"USD": {
					"price": 186.778961246059,
					"volume_24h": 694906316.737733,
					"market_cap": 20069106747.7562,
					"percentage_change_1h": 0.8,
					"percentage_change_24h": -0.85,
					"percentage_change_7d": -11.34
				}
			},
			"last_updated": 1565994301
		},
		"52": {
			"id": 52,
			"name": "Ripple",
			"symbol": "XRP",
			"website_slug": "ripple",
			"rank": 3,
			"circulating_supply": 42890708341,
			"total_supply": 99991757426,
			"max_supply": 100000000000,
			"quotes": {
				"USD": {
					"price": 0.26147471468326,
					"volume_24h": 208613485.614034,
					"market_cap": 11214835726.0259,
					"percentage_change_1h": 3,
					"percentage_change_24h": 2.76,
					"percentage_change_7d": -11.26
				}
			},
			"last_updated": 1565994302
		},
		"1831": {
			"id": 1831,
			"name": "Bitcoin Cash",
			"symbol": "BCH",
			"website_slug": "bitcoin-cash",
			"rank": 4,
			"circulating_supply": 17951102.1469082,
			"total_supply": 17505963,
			"max_supply": 21000000,
			"quotes": {
				"USD": {
					"price": 313.651054155512,
					"volume_24h": 284296663.789982,
					"market_cap": 5639927800.1669,
					"percentage_change_1h": -0.32,
					"percentage_change_24h": -1.13,
					"percentage_change_7d": 0.08
				}
			},
			"last_updated": 1565994302
		},
		"2": {
			"id": 2,
			"name": "Litecoin",
			"symbol": "LTC",
			"website_slug": "litecoin",
			"rank": 5,
			"circulating_supply": 63050968.3000049,
			"total_supply": 59533817,
			"max_supply": 84000000,
			"quotes": {
				"USD": {
					"price": 74.992386000606,
					"volume_24h": 270007746.519876,
					"market_cap": 4737714488.23085,
					"percentage_change_1h": -0.77,
					"percentage_change_24h": -2.62,
					"percentage_change_7d": -10.64
				}
			},
			"last_updated": 1565994302
		},
		"1839": {
			"id": 1839,
			"name": "Binance Coin",
			"symbol": "BNB",
			"website_slug": "binance-coin",
			"rank": 6,
			"circulating_supply": 155536713,
			"total_supply": 190799315,
			"max_supply": 0,
			"quotes": {
				"USD": {
					"price": 27.7065630102369,
					"volume_24h": 64963051.6950017,
					"market_cap": 4309387739.13963,
					"percentage_change_1h": -0.02,
					"percentage_change_24h": -1.3,
					"percentage_change_7d": -7.15
				}
			},
			"last_updated": 1565994302
		},
		"328": {
			"id": 328,
			"name": "Monero",
			"symbol": "XMR",
			"website_slug": "monero",
			"rank": 11,
			"circulating_supply": 17158233.5629789,
			"total_supply": 16638928,
			"max_supply": 0,
			"quotes": {
				"USD": {
					"price": 82.4207631052468,
					"volume_24h": 85290584.1456297,
					"market_cap": 1413556667.44168,
					"percentage_change_1h": 0.71,
					"percentage_change_24h": 0.69,
					"percentage_change_7d": -12.8
				}
			},
			"last_updated": 1565994302
		},
		"512": {
			"id": 512,
			"name": "Stellar",
			"symbol": "XLM",
			"website_slug": "stellar",
			"rank": 12,
			"circulating_supply": 19633175221.2173,
			"total_supply": 104542893473,
			"max_supply": 0,
			"quotes": {
				"USD": {
					"price": 0.0685452569269779,
					"volume_24h": 12250147.3467232,
					"market_cap": 1345761039.83072,
					"percentage_change_1h": -0.07,
					"percentage_change_24h": -2.05,
					"percentage_change_7d": -4.82
				}
			},
			"last_updated": 1565994302
		},
		"131": {
			"id": 131,
			"name": "Dash",
			"symbol": "DASH",
			"website_slug": "dash",
			"rank": 17,
			"circulating_supply": 8988407.2182673,
			"total_supply": 8500642,
			"max_supply": 18900000,
			"quotes": {
				"USD": {
					"price": 93.9584452523325,
					"volume_24h": 13795814.8477689,
					"market_cap": 845105099.432812,
					"percentage_change_1h": 0.13,
					"percentage_change_24h": -0.43,
					"percentage_change_7d": -9.29
				}
			},
			"last_updated": 1565994302
		},
		"1437": {
			"id": 1437,
			"name": "Zcash",
			"symbol": "ZEC",
			"website_slug": "zcash",
			"rank": 24,
			"circulating_supply": 7206481.25,
			"total_supply": 5432031,
			"max_supply": 0,
			"quotes": {
				"USD": {
					"price": 49.8772129846135,
					"volume_24h": 15986128.6923724,
					"market_cap": 357637265.447102,
					"percentage_change_1h": 0.49,
					"percentage_change_24h": -1.09,
					"percentage_change_7d": -15.62
				}
			},
			"last_updated": 1565994302
		},
		"1168": {
			"id": 1168,
			"name": "Decred",
			"symbol": "DCR",
			"website_slug": "decred",
			"rank": 27,
			"circulating_supply": 10245150.097245,
			"total_supply": 8966763,
			"max_supply": 21000000,
			"quotes": {
				"USD": {
					"price": 25.6534988058625,
					"volume_24h": 1024853.38634182,
					"market_cap": 262823945.785557,
					"percentage_change_1h": 0.94,
					"percentage_change_24h": 0.4,
					"percentage_change_7d": -4.4
				}
			},
			"last_updated": 1565994302
		},
		"2577": {
			"id": 2577,
			"name": "Ravencoin",
			"symbol": "RVN",
			"website_slug": "ravencoin",
			"rank": 41,
			"circulating_supply": 4211105000,
			"total_supply": 2439940000,
			"max_supply": 21000000000,
			"quotes": {
				"USD": {
					"price": 0.0331484585054899,
					"volume_24h": 4403913.75065542,
					"market_cap": 139591639.354761,
					"percentage_change_1h": 0.11,
					"percentage_change_24h": -5.36,
					"percentage_change_7d": -14.14
				}
			},
			"last_updated": 1565994303
		},
		"2010": {
			"id": 2010,
			"name": "Cardano",
			"symbol": "ADA",
			"website_slug": "cardano",
			"rank": 10,
			"circulating_supply": 31112484646,
			"total_supply": 31112483745,
			"max_supply": 45000000000,
			"quotes": {
				"USD": {
					"price": 0.0462589865511383,
					"volume_24h": 30790290.5445737,
					"market_cap": 1439232008.81181,
					"percentage_change_1h": -0.19,
					"percentage_change_24h": -2.93,
					"percentage_change_7d": -2.68
				}
			},
			"last_updated": 1565994603
		},
		"1087": {
			"id": 1087,
			"name": "Factom",
			"symbol": "FCT",
			"website_slug": "factom",
			"rank": 117,
			"circulating_supply": 9699162.17,
			"total_supply": 8745102,
			"max_supply": 0,
			"quotes": {
				"USD": {
					"price": 3.16071637196172,
					"volume_24h": 38729.8590292983,
					"market_cap": 30656300.6650308,
					"percentage_change_1h": 0.89,
					"percentage_change_24h": -3.62,
					"percentage_change_7d": -14.67
				}
			},
			"last_updated": 1565994302
		}
	},
	"metadata": {
		"timestamp": 1565994303,
		"num_cryptocurrencies": 687,
		"error": null
	}
}
   
`
