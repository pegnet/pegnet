package polling_test

import (
	"testing"
)

func TestActualAlternativePeggedAssets(t *testing.T) {
	ActualDataSourceTest(t, "AlternativeMe", "FCT")
}

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
			"circulating_supply": 18158625,
			"total_supply": 17418787,
			"max_supply": 21000000,
			"quotes": {
				"USD": {
					"price": 8120.24825443594,
					"volume_24h": 1478416849.95321,
					"market_cap": 149639228608.025,
					"percentage_change_1h": 0.39,
					"percentage_change_24h": -0.41,
					"percentage_change_7d": 7.52
				}
			},
			"last_updated": 1578936301
		},
		"1027": {
			"id": 1027,
			"name": "Ethereum",
			"symbol": "ETH",
			"website_slug": "ethereum",
			"rank": 2,
			"circulating_supply": 109262219.874,
			"total_supply": 103748085,
			"max_supply": 0,
			"quotes": {
				"USD": {
					"price": 143.724312794161,
					"volume_24h": 300560896.315221,
					"market_cap": 15887341355.9039,
					"percentage_change_1h": 0.38,
					"percentage_change_24h": -0.76,
					"percentage_change_7d": 1.84
				}
			},
			"last_updated": 1578936302
		},
		"1831": {
			"id": 1831,
			"name": "Bitcoin Cash",
			"symbol": "BCH",
			"website_slug": "bitcoin-cash",
			"rank": 4,
			"circulating_supply": 18220489.6469082,
			"total_supply": 17505963,
			"max_supply": 21000000,
			"quotes": {
				"USD": {
					"price": 266.042303334313,
					"volume_24h": 216801012.570592,
					"market_cap": 4930359312.65712,
					"percentage_change_1h": 1.03,
					"percentage_change_24h": 0.02,
					"percentage_change_7d": 12.91
				}
			},
			"last_updated": 1578936302
		},
		"2": {
			"id": 2,
			"name": "Litecoin",
			"symbol": "LTC",
			"website_slug": "litecoin",
			"rank": 6,
			"circulating_supply": 63850826.9606127,
			"total_supply": 59533817,
			"max_supply": 84000000,
			"quotes": {
				"USD": {
					"price": 49.5848660584091,
					"volume_24h": 156082074.343095,
					"market_cap": 3205831360.95155,
					"percentage_change_1h": 0.29,
					"percentage_change_24h": -1.81,
					"percentage_change_7d": 11.25
				}
			},
			"last_updated": 1578936302
		},
		"1765": {
			"id": 1765,
			"name": "EOS",
			"symbol": "EOS",
			"website_slug": "eos",
			"rank": 8,
			"circulating_supply": 959517000.1331,
			"total_supply": 1006245120,
			"max_supply": 0,
			"quotes": {
				"USD": {
					"price": 3.11271097203748,
					"volume_24h": 219389466.926393,
					"market_cap": 3031104202.88016,
					"percentage_change_1h": 0.71,
					"percentage_change_24h": -1.74,
					"percentage_change_7d": 11.91
				}
			},
			"last_updated": 1578936302
		},
		"1839": {
			"id": 1839,
			"name": "Binance Coin",
			"symbol": "BNB",
			"website_slug": "binance-coin",
			"rank": 9,
			"circulating_supply": 153474825,
			"total_supply": 190799315,
			"max_supply": 0,
			"quotes": {
				"USD": {
					"price": 15.3029545030955,
					"volume_24h": 47195290.748898,
					"market_cap": 2348618264.34555,
					"percentage_change_1h": 0.12,
					"percentage_change_24h": 0.01,
					"percentage_change_7d": 3.38
				}
			},
			"last_updated": 1578936302
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
					"price": 0.0373576049082207,
					"volume_24h": 8845962.35534335,
					"market_cap": 1162287909.11835,
					"percentage_change_1h": 0.16,
					"percentage_change_24h": 0.35,
					"percentage_change_7d": 2.45
				}
			},
			"last_updated": 1578936302
		},
		"328": {
			"id": 328,
			"name": "Monero",
			"symbol": "XMR",
			"website_slug": "monero",
			"rank": 11,
			"circulating_supply": 17398253.0964224,
			"total_supply": 16638928,
			"max_supply": 0,
			"quotes": {
				"USD": {
					"price": 57.3404931192238,
					"volume_24h": 22605287.9938739,
					"market_cap": 1005686467.00577,
					"percentage_change_1h": -0.81,
					"percentage_change_24h": -2.43,
					"percentage_change_7d": 2.76
				}
			},
			"last_updated": 1578936302
		},
		"512": {
			"id": 512,
			"name": "Stellar",
			"symbol": "XLM",
			"website_slug": "stellar",
			"rank": 13,
			"circulating_supply": 19975881672.4764,
			"total_supply": 104542893473,
			"max_supply": 0,
			"quotes": {
				"USD": {
					"price": 0.0482712919760941,
					"volume_24h": 7770633.45594148,
					"market_cap": 964261616.692015,
					"percentage_change_1h": 0.01,
					"percentage_change_24h": -1.31,
					"percentage_change_7d": -2.97
				}
			},
			"last_updated": 1578936302
		},
		"2011": {
			"id": 2011,
			"name": "Tezos",
			"symbol": "XTZ",
			"website_slug": "tezos",
			"rank": 14,
			"circulating_supply": 697049026.500784,
			"total_supply": 763306930,
			"max_supply": 0,
			"quotes": {
				"USD": {
					"price": 1.27937022131689,
					"volume_24h": 8571783.90556064,
					"market_cap": 893306201.938749,
					"percentage_change_1h": 0.96,
					"percentage_change_24h": -1.95,
					"percentage_change_7d": -2.95
				}
			},
			"last_updated": 1578936303
		},
		"1975": {
			"id": 1975,
			"name": "ChainLink",
			"symbol": "LINK",
			"website_slug": "chainlink",
			"rank": 15,
			"circulating_supply": 364409568.99,
			"total_supply": 1000000000,
			"max_supply": 0,
			"quotes": {
				"USD": {
					"price": 2.19199382238974,
					"volume_24h": 15776187.3401717,
					"market_cap": 806321849.686625,
					"percentage_change_1h": 0.49,
					"percentage_change_24h": -2.08,
					"percentage_change_7d": 16.32
				}
			},
			"last_updated": 1578936302
		},
		"131": {
			"id": 131,
			"name": "Dash",
			"symbol": "DASH",
			"website_slug": "dash",
			"rank": 19,
			"circulating_supply": 9270729.14908187,
			"total_supply": 8500642,
			"max_supply": 18900000,
			"quotes": {
				"USD": {
					"price": 66.6610740457883,
					"volume_24h": 38412085.8981144,
					"market_cap": 625262582.804733,
					"percentage_change_1h": 2.1,
					"percentage_change_24h": 3.42,
					"percentage_change_7d": 18.09
				}
			},
			"last_updated": 1578936302
		},
		"1437": {
			"id": 1437,
			"name": "Zcash",
			"symbol": "ZEC",
			"website_slug": "zcash",
			"rank": 24,
			"circulating_supply": 8514743.75,
			"total_supply": 5432031,
			"max_supply": 0,
			"quotes": {
				"USD": {
					"price": 37.014783384815,
					"volume_24h": 33807479.773635,
					"market_cap": 320210745.865712,
					"percentage_change_1h": 0.28,
					"percentage_change_24h": 5.49,
					"percentage_change_7d": 16.03
				}
			},
			"last_updated": 1578936302
		},
		"1697": {
			"id": 1697,
			"name": "Basic Attention Token",
			"symbol": "BAT",
			"website_slug": "basic-attention-token",
			"rank": 27,
			"circulating_supply": 1423336562.3463,
			"total_supply": 1500000000,
			"max_supply": 0,
			"quotes": {
				"USD": {
					"price": 0.189106494277062,
					"volume_24h": 1857428.86816813,
					"market_cap": 269162187.481674,
					"percentage_change_1h": 0.42,
					"percentage_change_24h": -1.75,
					"percentage_change_7d": -1.46
				}
			},
			"last_updated": 1578936302
		},
		"1168": {
			"id": 1168,
			"name": "Decred",
			"symbol": "DCR",
			"website_slug": "decred",
			"rank": 31,
			"circulating_supply": 10949926.6823207,
			"total_supply": 8966763,
			"max_supply": 21000000,
			"quotes": {
				"USD": {
					"price": 17.2177811930058,
					"volume_24h": 741352.041017123,
					"market_cap": 188533441.695653,
					"percentage_change_1h": 0.49,
					"percentage_change_24h": 0.47,
					"percentage_change_7d": -5.12
				}
			},
			"last_updated": 1578936302
		},
		"2577": {
			"id": 2577,
			"name": "Ravencoin",
			"symbol": "RVN",
			"website_slug": "ravencoin",
			"rank": 36,
			"circulating_supply": 5282540000,
			"total_supply": 2439940000,
			"max_supply": 21000000000,
			"quotes": {
				"USD": {
					"price": 0.0237228196871413,
					"volume_24h": 1476357.07045128,
					"market_cap": 125316743.910111,
					"percentage_change_1h": 0.32,
					"percentage_change_24h": -1.22,
					"percentage_change_7d": -0.03
				}
			},
			"last_updated": 1578936303
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
		"timestamp": 1578936303,
		"num_cryptocurrencies": 488,
		"error": null
	}
}

`
