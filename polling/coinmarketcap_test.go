package polling_test

import (
	"testing"
)

// Need an api key to run this
//func TestCoinMarketCapPeggedAssets(t *testing.T) {
//	ActualDataSourceTest(t, "CoinMarketCap")
//}

func TestFixedCoinMarketCapPeggedAssets(t *testing.T) {
	FixedDataSourceTest(t, "CoinMarketCap", []byte(coinMarketCapResp), "AED")
}

var coinMarketCapResp = `
{
    "status": {
        "timestamp": "2020-01-13T17:18:03.635Z",
        "error_code": 0,
        "error_message": null,
        "elapsed": 11,
        "credit_count": 1,
        "notice": null
    },
    "data": {
        "1": {
            "id": 1,
            "name": "Bitcoin",
            "symbol": "BTC",
            "slug": "bitcoin",
            "num_market_pairs": 7601,
            "date_added": "2013-04-28T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": 21000000,
            "circulating_supply": 18158637,
            "total_supply": 18158637,
            "platform": null,
            "cmc_rank": 1,
            "last_updated": "2020-01-13T17:17:33.000Z",
            "quote": {
                "USD": {
                    "price": 8118.5153313,
                    "volume_24h": 22314130067.5031,
                    "percent_change_1h": 0.250108,
                    "percent_change_24h": -0.430186,
                    "percent_change_7d": 6.83503,
                    "market_cap": 147421172880.01144,
                    "last_updated": "2020-01-13T17:17:33.000Z"
                }
            }
        },
        "2": {
            "id": 2,
            "name": "Litecoin",
            "symbol": "LTC",
            "slug": "litecoin",
            "num_market_pairs": 537,
            "date_added": "2013-04-28T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": 84000000,
            "circulating_supply": 63851039.4606127,
            "total_supply": 63851039.4606127,
            "platform": null,
            "cmc_rank": 6,
            "last_updated": "2020-01-13T17:17:02.000Z",
            "quote": {
                "USD": {
                    "price": 49.7031141606,
                    "volume_24h": 3392360676.0803,
                    "percent_change_1h": 0.701209,
                    "percent_change_24h": -1.77641,
                    "percent_change_7d": 10.1183,
                    "market_cap": 3173595503.5838084,
                    "last_updated": "2020-01-13T17:17:02.000Z"
                }
            }
        },
        "131": {
            "id": 131,
            "name": "Dash",
            "symbol": "DASH",
            "slug": "dash",
            "num_market_pairs": 259,
            "date_added": "2014-02-14T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": 18900000,
            "circulating_supply": 9265936.13467026,
            "total_supply": 9265936.13467026,
            "platform": null,
            "cmc_rank": 21,
            "last_updated": "2020-01-13T17:17:02.000Z",
            "quote": {
                "USD": {
                    "price": 66.3161053105,
                    "volume_24h": 414562947.895995,
                    "percent_change_1h": 1.74612,
                    "percent_change_24h": 1.97213,
                    "percent_change_7d": 19.0273,
                    "market_cap": 614480796.5071603,
                    "last_updated": "2020-01-13T17:17:02.000Z"
                }
            }
        },
		"2011": {
 			"id": 2011,
            "name": "Tezos",
            "symbol": "XTZ",
            "slug": "tezos",
            "num_market_pairs": 125,
            "date_added": "2014-05-21T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": null,
            "circulating_supply": 17398281.0938049,
            "total_supply": 17398281.0938049,
            "platform": null,
            "cmc_rank": 10,
            "last_updated": "2020-01-13T17:17:01.000Z",
            "quote": {
                "USD": {
					"price": 1.27937022131689,
                    "volume_24h": 414562947.895995,
                    "percent_change_1h": -0.521103,
                    "percent_change_24h": -2.85593,
                    "percent_change_7d": 1.80022,
                    "market_cap": 995090665.470944,
                    "last_updated": "2020-01-13T17:17:01.000Z"
                }
            }
		},
        "328": {
            "id": 328,
            "name": "Monero",
            "symbol": "XMR",
            "slug": "monero",
            "num_market_pairs": 125,
            "date_added": "2014-05-21T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": null,
            "circulating_supply": 17398281.0938049,
            "total_supply": 17398281.0938049,
            "platform": null,
            "cmc_rank": 10,
            "last_updated": "2020-01-13T17:17:01.000Z",
            "quote": {
                "USD": {
                    "price": 57.1947688456,
                    "volume_24h": 56144410.9540942,
                    "percent_change_1h": -0.521103,
                    "percent_change_24h": -2.85593,
                    "percent_change_7d": 1.80022,
                    "market_cap": 995090665.470944,
                    "last_updated": "2020-01-13T17:17:01.000Z"
                }
            }
        },
        "512": {
            "id": 512,
            "name": "Stellar",
            "symbol": "XLM",
            "slug": "stellar",
            "num_market_pairs": 293,
            "date_added": "2014-08-05T00:00:00.000Z",
            "tags": [],
            "max_supply": null,
            "circulating_supply": 19975881672.1288,
            "total_supply": 50001803905.9717,
            "platform": null,
            "cmc_rank": 13,
            "last_updated": "2020-01-13T17:17:02.000Z",
            "quote": {
                "USD": {
                    "price": 0.0479596513494,
                    "volume_24h": 217332009.88552,
                    "percent_change_1h": 0.0708115,
                    "percent_change_24h": -0.674386,
                    "percent_change_7d": -1.31276,
                    "market_cap": 958036320.3921667,
                    "last_updated": "2020-01-13T17:17:02.000Z"
                }
            }
        },
        "1027": {
            "id": 1027,
            "name": "Ethereum",
            "symbol": "ETH",
            "slug": "ethereum",
            "num_market_pairs": 5163,
            "date_added": "2015-08-07T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": null,
            "circulating_supply": 109262509.874,
            "total_supply": 109262509.874,
            "platform": null,
            "cmc_rank": 2,
            "last_updated": "2020-01-13T17:17:25.000Z",
            "quote": {
                "USD": {
                    "price": 143.715553655,
                    "volume_24h": 8302435696.09797,
                    "percent_change_1h": 0.304515,
                    "percent_change_24h": -0.898311,
                    "percent_change_7d": 1.07741,
                    "market_cap": 15702722100.276815,
                    "last_updated": "2020-01-13T17:17:25.000Z"
                }
            }
        },
        "1087": {
            "id": 1087,
            "name": "Factom",
            "symbol": "FCT",
            "slug": "factom",
            "num_market_pairs": 5,
            "date_added": "2015-10-06T00:00:00.000Z",
            "tags": [],
            "max_supply": null,
            "circulating_supply": 8827592.73,
            "total_supply": 8827592.73,
            "platform": null,
            "cmc_rank": 160,
            "last_updated": "2020-01-13T17:17:01.000Z",
            "quote": {
                "USD": {
                    "price": 1.89996831065,
                    "volume_24h": 2368730.30701913,
                    "percent_change_1h": 0.41237,
                    "percent_change_24h": -7.59701,
                    "percent_change_7d": -7.08537,
                    "market_cap": 16772146.446324322,
                    "last_updated": "2020-01-13T17:17:01.000Z"
                }
            }
        },
        "1168": {
            "id": 1168,
            "name": "Decred",
            "symbol": "DCR",
            "slug": "decred",
            "num_market_pairs": 47,
            "date_added": "2016-02-10T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": 21000000,
            "circulating_supply": 10786830.8782284,
            "total_supply": 10786830.8782284,
            "platform": null,
            "cmc_rank": 35,
            "last_updated": "2020-01-13T17:17:02.000Z",
            "quote": {
                "USD": {
                    "price": 17.0535955567,
                    "volume_24h": 81338725.3464959,
                    "percent_change_1h": 0.745915,
                    "percent_change_24h": -0.594447,
                    "percent_change_7d": -6.38474,
                    "market_cap": 183954251.1358302,
                    "last_updated": "2020-01-13T17:17:02.000Z"
                }
            }
        },
        "1437": {
            "id": 1437,
            "name": "Zcash",
            "symbol": "ZEC",
            "slug": "zcash",
            "num_market_pairs": 210,
            "date_added": "2016-10-29T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": 21000000,
            "circulating_supply": 8515056.25,
            "total_supply": 8515056.25,
            "platform": null,
            "cmc_rank": 28,
            "last_updated": "2020-01-13T17:17:05.000Z",
            "quote": {
                "USD": {
                    "price": 37.0972269833,
                    "volume_24h": 235906615.196192,
                    "percent_change_1h": 0.786555,
                    "percent_change_24h": 4.69247,
                    "percent_change_7d": 14.45,
                    "market_cap": 315884974.4818173,
                    "last_updated": "2020-01-13T17:17:05.000Z"
                }
            }
        },
        "1697": {
            "id": 1697,
            "name": "Basic Attention Token",
            "symbol": "BAT",
            "slug": "basic-attention-token",
            "num_market_pairs": 158,
            "date_added": "2017-06-01T00:00:00.000Z",
            "tags": [],
            "max_supply": null,
            "circulating_supply": 1421086562.3463,
            "total_supply": 1500000000,
            "platform": {
                "id": 1027,
                "name": "Ethereum",
                "symbol": "ETH",
                "slug": "ethereum",
                "token_address": "0x0d8775f648430679a709e98d2b0cb6250d2887ef"
            },
            "cmc_rank": 32,
            "last_updated": "2020-01-13T17:17:05.000Z",
            "quote": {
                "USD": {
                    "price": 0.187890514251,
                    "volume_24h": 44484978.7388953,
                    "percent_change_1h": 0.136503,
                    "percent_change_24h": -2.00274,
                    "percent_change_7d": -2.66033,
                    "market_cap": 267008684.99443206,
                    "last_updated": "2020-01-13T17:17:05.000Z"
                }
            }
        },
        "1765": {
            "id": 1765,
            "name": "EOS",
            "symbol": "EOS",
            "slug": "eos",
            "num_market_pairs": 379,
            "date_added": "2017-07-01T00:00:00.000Z",
            "tags": [],
            "max_supply": null,
            "circulating_supply": 948524809.1964,
            "total_supply": 1045224820.7914,
            "platform": null,
            "cmc_rank": 8,
            "last_updated": "2020-01-13T17:17:08.000Z",
            "quote": {
                "USD": {
                    "price": 3.10819450044,
                    "volume_24h": 2235864924.03032,
                    "percent_change_1h": 0.840596,
                    "percent_change_24h": -2.00384,
                    "percent_change_7d": 10.6831,
                    "market_cap": 2948199595.4751506,
                    "last_updated": "2020-01-13T17:17:08.000Z"
                }
            }
        },
        "1831": {
            "id": 1831,
            "name": "Bitcoin Cash",
            "symbol": "BCH",
            "slug": "bitcoin-cash",
            "num_market_pairs": 416,
            "date_added": "2017-07-23T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": 21000000,
            "circulating_supply": 18220512.5,
            "total_supply": 18220512.5,
            "platform": null,
            "cmc_rank": 4,
            "last_updated": "2020-01-13T17:17:09.000Z",
            "quote": {
                "USD": {
                    "price": 266.205857122,
                    "volume_24h": 1851744900.62652,
                    "percent_change_1h": 1.12405,
                    "percent_change_24h": 0.112444,
                    "percent_change_7d": 12.0885,
                    "market_cap": 4850407147.264615,
                    "last_updated": "2020-01-13T17:17:09.000Z"
                }
            }
        },
        "1839": {
            "id": 1839,
            "name": "Binance Coin",
            "symbol": "BNB",
            "slug": "binance-coin",
            "num_market_pairs": 293,
            "date_added": "2017-07-25T00:00:00.000Z",
            "tags": [],
            "max_supply": 187536713,
            "circulating_supply": 155536713,
            "total_supply": 187536713,
            "platform": null,
            "cmc_rank": 9,
            "last_updated": "2020-01-13T17:17:06.000Z",
            "quote": {
                "USD": {
                    "price": 15.0649617227,
                    "volume_24h": 228255212.829298,
                    "percent_change_1h": 0.231931,
                    "percent_change_24h": -1.28296,
                    "percent_change_7d": 2.14569,
                    "market_cap": 2343154627.8195753,
                    "last_updated": "2020-01-13T17:17:06.000Z"
                }
            }
        },
        "1975": {
            "id": 1975,
            "name": "Chainlink",
            "symbol": "LINK",
            "slug": "chainlink",
            "num_market_pairs": 124,
            "date_added": "2017-09-20T00:00:00.000Z",
            "tags": [],
            "max_supply": null,
            "circulating_supply": 350000000,
            "total_supply": 1000000000,
            "platform": {
                "id": 1027,
                "name": "Ethereum",
                "symbol": "ETH",
                "slug": "ethereum",
                "token_address": "0x514910771af9ca656af840dff83e8264ecf986ca"
            },
            "cmc_rank": 17,
            "last_updated": "2020-01-13T17:17:05.000Z",
            "quote": {
                "USD": {
                    "price": 2.1832636942,
                    "volume_24h": 86427224.3542053,
                    "percent_change_1h": 0.468359,
                    "percent_change_24h": -2.93745,
                    "percent_change_7d": 14.6197,
                    "market_cap": 764142292.9699999,
                    "last_updated": "2020-01-13T17:17:05.000Z"
                }
            }
        },
        "2010": {
            "id": 2010,
            "name": "Cardano",
            "symbol": "ADA",
            "slug": "cardano",
            "num_market_pairs": 112,
            "date_added": "2017-10-01T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": 45000000000,
            "circulating_supply": 25927070538,
            "total_supply": 31112483745,
            "platform": null,
            "cmc_rank": 12,
            "last_updated": "2020-01-13T17:17:05.000Z",
            "quote": {
                "USD": {
                    "price": 0.0369991092437,
                    "volume_24h": 57954241.7460133,
                    "percent_change_1h": 0.304765,
                    "percent_change_24h": -0.801979,
                    "percent_change_7d": 1.79183,
                    "market_cap": 959278515.2045777,
                    "last_updated": "2020-01-13T17:17:05.000Z"
                }
            }
        },
        "2577": {
            "id": 2577,
            "name": "Ravencoin",
            "symbol": "RVN",
            "slug": "ravencoin",
            "num_market_pairs": 39,
            "date_added": "2018-03-10T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": 21000000000,
            "circulating_supply": 5282660000,
            "total_supply": 5282660000,
            "platform": null,
            "cmc_rank": 40,
            "last_updated": "2020-01-13T17:17:06.000Z",
            "quote": {
                "USD": {
                    "price": 0.0236078448595,
                    "volume_24h": 4723684.89736945,
                    "percent_change_1h": 0.149641,
                    "percent_change_24h": -1.7186,
                    "percent_change_7d": -0.804446,
                    "market_cap": 124712217.72548626,
                    "last_updated": "2020-01-13T17:17:06.000Z"
                }
            }
        },
        "3794": {
            "id": 3794,
            "name": "Cosmos",
            "symbol": "ATOM",
            "slug": "cosmos",
            "num_market_pairs": 97,
            "date_added": "2019-03-14T00:00:00.000Z",
            "tags": [],
            "max_supply": null,
            "circulating_supply": 190688439.2,
            "total_supply": 237928230.821588,
            "platform": null,
            "cmc_rank": 16,
            "last_updated": "2020-01-13T17:17:12.000Z",
            "quote": {
                "USD": {
                    "price": 4.17773287417,
                    "volume_24h": 107670600.528639,
                    "percent_change_1h": 0.284184,
                    "percent_change_24h": -1.81861,
                    "percent_change_7d": 1.2855,
                    "market_cap": 796645361.1700072,
                    "last_updated": "2020-01-13T17:17:12.000Z"
                }
            }
        },
        "4979": {
            "id": 4979,
            "name": "PegNet",
            "symbol": "PEG",
            "slug": "pegnet",
            "num_market_pairs": 6,
            "date_added": "2019-12-16T00:00:00.000Z",
            "tags": [],
            "max_supply": null,
            "circulating_supply": 2044910921.89098,
            "total_supply": 2044910921.89098,
            "platform": null,
            "cmc_rank": 506,
            "last_updated": "2020-01-13T17:17:14.000Z",
            "quote": {
                "USD": {
                    "price": 0.00169050122722,
                    "volume_24h": 42984.7094987263,
                    "percent_change_1h": -2.49792,
                    "percent_change_24h": -19.4651,
                    "percent_change_7d": -13.3923,
                    "market_cap": 3456924.423012283,
                    "last_updated": "2020-01-13T17:17:14.000Z"
                }
            }
        },
		"74": {
            "id": 74,
            "name": "Dogecoin",
            "symbol": "DOGE",
            "slug": "dogecoin",
            "num_market_pairs": 233,
            "date_added": "2013-12-15T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": null,
            "circulating_supply": 124481731530.374,
            "total_supply": 124481731530.374,
            "is_active": 1,
            "platform": null,
            "cmc_rank": 29,
            "is_fiat": 0,
            "last_updated": "2020-05-08T15:30:03.000Z",
            "quote": {
                "USD": {
                    "price": 0.00262333753849,
                    "volume_24h": 293655177.524912,
                    "percent_change_1h": 0.0291953,
                    "percent_change_24h": 2.13404,
                    "percent_change_7d": 4.78736,
                    "market_cap": 326557599.1798643,
                    "last_updated": "2020-05-08T15:30:03.000Z"
                }
            }
        },
        "1321": {
            "id": 1321,
            "name": "Ethereum Classic",
            "symbol": "ETC",
            "slug": "ethereum-classic",
            "num_market_pairs": 245,
            "date_added": "2016-07-24T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": 210700000,
            "circulating_supply": 116313299,
            "total_supply": 116313299,
            "is_active": 1,
            "platform": null,
            "cmc_rank": 19,
            "is_fiat": 0,
            "last_updated": "2020-05-08T15:30:05.000Z",
            "quote": {
                "USD": {
                    "price": 7.06374667054,
                    "volume_24h": 2290155738.91049,
                    "percent_change_1h": 0.27964,
                    "percent_change_24h": 1.49811,
                    "percent_change_7d": 6.57505,
                    "market_cap": 821607678.5507735,
                    "last_updated": "2020-05-08T15:30:05.000Z"
                }
            }
        },
        "1376": {
            "id": 1376,
            "name": "Neo",
            "symbol": "NEO",
            "slug": "neo",
            "num_market_pairs": 242,
            "date_added": "2016-09-08T00:00:00.000Z",
            "tags": [],
            "max_supply": 100000000,
            "circulating_supply": 70538831,
            "total_supply": 100000000,
            "is_active": 1,
            "platform": null,
            "cmc_rank": 22,
            "is_fiat": 0,
            "last_updated": "2020-05-08T15:30:05.000Z",
            "quote": {
                "USD": {
                    "price": 9.97290591759,
                    "volume_24h": 844903805.85821,
                    "percent_change_1h": 0.389377,
                    "percent_change_24h": 5.13224,
                    "percent_change_7d": 8.50539,
                    "market_cap": 703477125.0997809,
                    "last_updated": "2020-05-08T15:30:05.000Z"
                }
            }
        },
        "2502": {
            "id": 2502,
            "name": "Huobi Token",
            "symbol": "HT",
            "slug": "huobi-token",
            "num_market_pairs": 137,
            "date_added": "2018-02-03T00:00:00.000Z",
            "tags": [],
            "max_supply": null,
            "circulating_supply": 222668092.971921,
            "total_supply": 500000000,
            "platform": {
                "id": 1027,
                "name": "Ethereum",
                "symbol": "ETH",
                "slug": "ethereum",
                "token_address": "0x6f259637dcd74c767781e37bc6133cd6a68aa161"
            },
            "is_active": 1,
            "cmc_rank": 18,
            "is_fiat": 0,
            "last_updated": "2020-05-08T15:30:08.000Z",
            "quote": {
                "USD": {
                    "price": 4.19968497626,
                    "volume_24h": 160251274.110309,
                    "percent_change_1h": -0.0208536,
                    "percent_change_24h": 1.52108,
                    "percent_change_7d": -1.3644,
                    "market_cap": 935135844.7466416,
                    "last_updated": "2020-05-08T15:30:08.000Z"
                }
            }
        },
        "2566": {
            "id": 2566,
            "name": "Ontology",
            "symbol": "ONT",
            "slug": "ontology",
            "num_market_pairs": 121,
            "date_added": "2018-03-08T00:00:00.000Z",
            "tags": [],
            "max_supply": null,
            "circulating_supply": 656746573,
            "total_supply": 1000000000,
            "is_active": 1,
            "platform": null,
            "cmc_rank": 31,
            "is_fiat": 0,
            "last_updated": "2020-05-08T15:30:08.000Z",
            "quote": {
                "USD": {
                    "price": 0.484778206767,
                    "volume_24h": 111281484.808273,
                    "percent_change_1h": 1.24602,
                    "percent_change_24h": 4.54492,
                    "percent_change_7d": -3.71465,
                    "market_cap": 318376425.9593127,
                    "last_updated": "2020-05-08T15:30:08.000Z"
                }
            }
        },
        "3077": {
            "id": 3077,
            "name": "VeChain",
            "symbol": "VET",
            "slug": "vechain",
            "num_market_pairs": 99,
            "date_added": "2017-08-22T00:00:00.000Z",
            "tags": [],
            "max_supply": null,
            "circulating_supply": 55454734800,
            "total_supply": 86712634466,
            "is_active": 1,
            "platform": null,
            "cmc_rank": 36,
            "is_fiat": 0,
            "last_updated": "2020-05-08T15:30:10.000Z",
            "quote": {
                "USD": {
                    "price": 0.00458254973749,
                    "volume_24h": 207076472.620976,
                    "percent_change_1h": 0.0871861,
                    "percent_change_24h": 5.26156,
                    "percent_change_7d": 1.60808,
                    "market_cap": 254124080.40031755,
                    "last_updated": "2020-05-08T15:30:10.000Z"
                }
            }
        },
        "3635": {
            "id": 3635,
            "name": "Crypto.com Coin",
            "symbol": "CRO",
            "slug": "crypto-com-coin",
            "num_market_pairs": 45,
            "date_added": "2018-12-14T00:00:00.000Z",
            "tags": [],
            "max_supply": null,
            "circulating_supply": 16603196347.032,
            "total_supply": 100000000000,
            "platform": {
                "id": 1027,
                "name": "Ethereum",
                "symbol": "ETH",
                "slug": "ethereum",
                "token_address": "0xa0b73e1ff0b80914ab6fe0444e65848c4c34450b"
            },
            "is_active": 1,
            "cmc_rank": 14,
            "is_fiat": 0,
            "last_updated": "2020-05-08T15:30:11.000Z",
            "quote": {
                "USD": {
                    "price": 0.0685169026857,
                    "volume_24h": 7612155.78256728,
                    "percent_change_1h": 0.973498,
                    "percent_change_24h": 5.81626,
                    "percent_change_7d": 17.8782,
                    "market_cap": 1137599588.3811612,
                    "last_updated": "2020-05-08T15:30:11.000Z"
                }
            }
        },
        "4642": {
            "id": 4642,
            "name": "Hedera Hashgraph",
            "symbol": "HBAR",
            "slug": "hedera-hashgraph",
            "num_market_pairs": 20,
            "date_added": "2019-09-17T00:00:00.000Z",
            "tags": [],
            "max_supply": null,
            "circulating_supply": 4077684788,
            "total_supply": 50000000000,
            "is_active": 1,
            "platform": null,
            "cmc_rank": 44,
            "is_fiat": 0,
            "last_updated": "2020-05-08T15:30:14.000Z",
            "quote": {
                "USD": {
                    "price": 0.0385473143175,
                    "volume_24h": 9277591.2046873,
                    "percent_change_1h": -0.35206,
                    "percent_change_24h": 1.30547,
                    "percent_change_7d": -4.07461,
                    "market_cap": 157183797.21072435,
                    "last_updated": "2020-05-08T15:30:14.000Z"
                }
            }
        },
		"4030": {
            "id": 4030,
            "name": "Algorand",
            "symbol": "ALGO",
            "slug": "algorand",
            "num_market_pairs": 83,
            "date_added": "2019-06-20T00:00:00.000Z",
            "tags": [
                "mineable"
            ],
            "max_supply": null,
            "circulating_supply": 708570138.818224,
            "total_supply": 3239841981.81822,
            "is_active": 1,
            "platform": null,
            "cmc_rank": 49,
            "is_fiat": 0,
            "last_updated": "2020-05-13T19:52:13.000Z",
            "quote": {
                "USD": {
                    "price": 0.194391876846,
                    "volume_24h": 31300045.720028,
                    "percent_change_1h": 0.384152,
                    "percent_change_24h": -0.768105,
                    "percent_change_7d": -5.3003,
                    "market_cap": 137740279.16190532,
                    "last_updated": "2020-05-13T19:52:13.000Z"
                }
            }
        }
    }
}
`
