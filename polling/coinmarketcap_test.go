package polling_test

import (
	"testing"
)

// Need an api key to run this
//func TestCoinMarketCapPeggedAssets(t *testing.T) {
//	ActualDataSourceTest(t, "CoinMarketCap")
//}

func TestFixedCoinMarketCapPeggedAssets(t *testing.T) {
	FixedDataSourceTest(t, "CoinMarketCap", []byte(coinMarketCapResp))
}

var coinMarketCapResp = `
{
    "status": {
        "timestamp": "2019-08-06T23:21:35.689Z",
        "error_code": 0,
        "error_message": null,
        "elapsed": 4,
        "credit_count": 1
    },
    "data": {
        "1": {
            "id": 1,
            "name": "Bitcoin",
            "symbol": "BTC",
            "slug": "bitcoin",
            "num_market_pairs": 7807,
            "date_added": "2013-04-28T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": 21000000,
            "circulating_supply": 17861975,
            "total_supply": 17861975,
            "platform": null,
            "cmc_rank": 1,
            "last_updated": "2019-08-06T23:20:32.000Z",
            "quote": {
                "USD": {
                    "price": 11372.9660929,
                    "volume_24h": 23249957361.4667,
                    "percent_change_1h": -0.0376081,
                    "percent_change_24h": -3.52478,
                    "percent_change_7d": 17.8769,
                    "market_cap": 203143636027.22748,
                    "last_updated": "2019-08-06T23:20:32.000Z"
                }
            }
        },
        "2": {
            "id": 2,
            "name": "Litecoin",
            "symbol": "LTC",
            "slug": "litecoin",
            "num_market_pairs": 577,
            "date_added": "2013-04-28T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": 84000000,
            "circulating_supply": 62984680.8333857,
            "total_supply": 62984680.8333857,
            "platform": null,
            "cmc_rank": 5,
            "last_updated": "2019-08-06T23:21:04.000Z",
            "quote": {
                "USD": {
                    "price": 92.4017774128,
                    "volume_24h": 3226488261.93387,
                    "percent_change_1h": 0.201713,
                    "percent_change_24h": -5.01277,
                    "percent_change_7d": 1.42675,
                    "market_cap": 5819896458.782756,
                    "last_updated": "2019-08-06T23:21:04.000Z"
                }
            }
        },
        "131": {
            "id": 131,
            "name": "Dash",
            "symbol": "DASH",
            "slug": "dash",
            "num_market_pairs": 252,
            "date_added": "2014-02-14T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": 18900000,
            "circulating_supply": 8966891.78168277,
            "total_supply": 8966891.78168277,
            "platform": null,
            "cmc_rank": 16,
            "last_updated": "2019-08-06T23:21:02.000Z",
            "quote": {
                "USD": {
                    "price": 105.666043739,
                    "volume_24h": 168682633.837007,
                    "percent_change_1h": 0.668357,
                    "percent_change_24h": -5.45741,
                    "percent_change_7d": -0.554694,
                    "market_cap": 947495979.2061713,
                    "last_updated": "2019-08-06T23:21:02.000Z"
                }
            }
        },
        "328": {
            "id": 328,
            "name": "Monero",
            "symbol": "XMR",
            "slug": "monero",
            "num_market_pairs": 131,
            "date_added": "2014-05-21T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": null,
            "circulating_supply": 17140458.8442542,
            "total_supply": 17140458.8442542,
            "platform": null,
            "cmc_rank": 10,
            "last_updated": "2019-08-06T23:21:02.000Z",
            "quote": {
                "USD": {
                    "price": 88.9643074829,
                    "volume_24h": 104972936.785045,
                    "percent_change_1h": -0.0101745,
                    "percent_change_24h": -4.73024,
                    "percent_change_7d": 12.961,
                    "market_cap": 1524889051.0182233,
                    "last_updated": "2019-08-06T23:21:02.000Z"
                }
            }
        },
        "512": {
            "id": 512,
            "name": "Stellar",
            "symbol": "XLM",
            "slug": "stellar",
            "num_market_pairs": 277,
            "date_added": "2014-08-05T00:00:00.000Z",
            "tags": [],
            "max_supply": null,
            "circulating_supply": 19618463974.9053,
            "total_supply": 105222940988.294,
            "platform": null,
            "cmc_rank": 11,
            "last_updated": "2019-08-06T23:21:02.000Z",
            "quote": {
                "USD": {
                    "price": 0.0776468795892,
                    "volume_24h": 95790611.2339985,
                    "percent_change_1h": 0.253015,
                    "percent_change_24h": -5.85508,
                    "percent_change_7d": -7.45122,
                    "market_cap": 1523312509.9845297,
                    "last_updated": "2019-08-06T23:21:02.000Z"
                }
            }
        },
        "1027": {
            "id": 1027,
            "name": "Ethereum",
            "symbol": "ETH",
            "slug": "ethereum",
            "num_market_pairs": 5534,
            "date_added": "2015-08-07T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": null,
            "circulating_supply": 107214195.999,
            "total_supply": 107214195.999,
            "platform": null,
            "cmc_rank": 2,
            "last_updated": "2019-08-06T23:21:20.000Z",
            "quote": {
                "USD": {
                    "price": 224.607482422,
                    "volume_24h": 7594992359.22728,
                    "percent_change_1h": 0.0264132,
                    "percent_change_24h": -3.66574,
                    "percent_change_7d": 6.11582,
                    "market_cap": 24081110643.234257,
                    "last_updated": "2019-08-06T23:21:20.000Z"
                }
            }
        },
        "1087": {
            "id": 1087,
            "name": "Factom",
            "symbol": "FCT",
            "slug": "factom",
            "num_market_pairs": 6,
            "date_added": "2015-10-06T00:00:00.000Z",
            "tags": [],
            "max_supply": null,
            "circulating_supply": 9687914.715,
            "total_supply": 9687914.715,
            "platform": null,
            "cmc_rank": 117,
            "last_updated": "2019-08-06T23:21:02.000Z",
            "quote": {
                "USD": {
                    "price": 3.9886823891,
                    "volume_24h": 183159.456062266,
                    "percent_change_1h": 0.0208209,
                    "percent_change_24h": -0.607465,
                    "percent_change_7d": 0.0627718,
                    "market_cap": 38642014.81082325,
                    "last_updated": "2019-08-06T23:21:02.000Z"
                }
            }
        },
        "1168": {
            "id": 1168,
            "name": "Decred",
            "symbol": "DCR",
            "slug": "decred",
            "num_market_pairs": 41,
            "date_added": "2016-02-10T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": 21000000,
            "circulating_supply": 10196380.1158702,
            "total_supply": 10196380.1158702,
            "platform": null,
            "cmc_rank": 30,
            "last_updated": "2019-08-06T23:21:02.000Z",
            "quote": {
                "USD": {
                    "price": 30.8881029433,
                    "volume_24h": 5259210.72453707,
                    "percent_change_1h": -1.25465,
                    "percent_change_24h": -0.526161,
                    "percent_change_7d": 21.8435,
                    "market_cap": 314946838.66801596,
                    "last_updated": "2019-08-06T23:21:02.000Z"
                }
            }
        },
        "1437": {
            "id": 1437,
            "name": "Zcash",
            "symbol": "ZEC",
            "slug": "zcash",
            "num_market_pairs": 202,
            "date_added": "2016-10-29T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": null,
            "circulating_supply": 7135068.75,
            "total_supply": 7135068.75,
            "platform": null,
            "cmc_rank": 26,
            "last_updated": "2019-08-06T23:21:04.000Z",
            "quote": {
                "USD": {
                    "price": 61.9457529334,
                    "volume_24h": 195482779.520805,
                    "percent_change_1h": -0.0626494,
                    "percent_change_24h": -7.77709,
                    "percent_change_7d": -7.52006,
                    "market_cap": 441987205.95032316,
                    "last_updated": "2019-08-06T23:21:04.000Z"
                }
            }
        },
        "1831": {
            "id": 1831,
            "name": "Bitcoin Cash",
            "symbol": "BCH",
            "slug": "bitcoin-cash",
            "num_market_pairs": 367,
            "date_added": "2017-07-23T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": 21000000,
            "circulating_supply": 17933187.5,
            "total_supply": 17933187.5,
            "platform": null,
            "cmc_rank": 4,
            "last_updated": "2019-08-06T23:21:05.000Z",
            "quote": {
                "USD": {
                    "price": 331.293846446,
                    "volume_24h": 1593974727.92442,
                    "percent_change_1h": -0.343715,
                    "percent_change_24h": -4.54232,
                    "percent_change_7d": 3.32586,
                    "market_cap": 5941154665.912326,
                    "last_updated": "2019-08-06T23:21:05.000Z"
                }
            }
        },
        "1839": {
            "id": 1839,
            "name": "Binance Coin",
            "symbol": "BNB",
            "slug": "binance-coin",
            "num_market_pairs": 235,
            "date_added": "2017-07-25T00:00:00.000Z",
            "tags": [],
            "max_supply": 187536713,
            "circulating_supply": 155536713,
            "total_supply": 187536713,
            "platform": null,
            "cmc_rank": 6,
            "last_updated": "2019-08-06T23:21:05.000Z",
            "quote": {
                "USD": {
                    "price": 27.5037301858,
                    "volume_24h": 192464823.056188,
                    "percent_change_1h": 0.0324754,
                    "percent_change_24h": -1.69279,
                    "percent_change_7d": 1.57349,
                    "market_cap": 4277839788.338211,
                    "last_updated": "2019-08-06T23:21:05.000Z"
                }
            }
        },
        "2010": {
            "id": 2010,
            "name": "Cardano",
            "symbol": "ADA",
            "slug": "cardano",
            "num_market_pairs": 97,
            "date_added": "2017-10-01T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": 45000000000,
            "circulating_supply": 25927070538,
            "total_supply": 31112483745,
            "platform": null,
            "cmc_rank": 13,
            "last_updated": "2019-08-06T23:21:05.000Z",
            "quote": {
                "USD": {
                    "price": 0.0530665994885,
                    "volume_24h": 48347261.6724406,
                    "percent_change_1h": -0.0281195,
                    "percent_change_24h": -6.57116,
                    "percent_change_7d": -11.9828,
                    "market_cap": 1375861468.1501343,
                    "last_updated": "2019-08-06T23:21:05.000Z"
                }
            }
        },
        "2577": {
            "id": 2577,
            "name": "Ravencoin",
            "symbol": "RVN",
            "slug": "ravencoin",
            "num_market_pairs": 43,
            "date_added": "2018-03-10T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": 21000000000,
            "circulating_supply": 4139940000,
            "total_supply": 4139940000,
            "platform": null,
            "cmc_rank": 42,
            "last_updated": "2019-08-06T23:21:07.000Z",
            "quote": {
                "USD": {
                    "price": 0.0402544160468,
                    "volume_24h": 14288948.2216103,
                    "percent_change_1h": 1.1378,
                    "percent_change_24h": -4.3106,
                    "percent_change_7d": -7.16623,
                    "market_cap": 166650867.1687892,
                    "last_updated": "2019-08-06T23:21:07.000Z"
                }
            }
        }
    }
}
`
