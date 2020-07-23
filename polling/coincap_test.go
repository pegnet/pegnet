// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package polling_test

import (
	"testing"
)

// TestCoinCapPeggedAssets tests all the crypto assets are found on coinmarket cap
func TestCoinCapPeggedAssets(t *testing.T) {
	ActualDataSourceTest(t, "CoinCap")
}

// TestFixedCoinCapRatesPeggedAssets with fixed resp
func TestFixedCoinCapRatesPeggedAssets(t *testing.T) {
	FixedDataSourceTest(t, "CoinCap", []byte(coincapResp), "HBAR", "AED", "CRO")
}

// Added some bad data in here too
var coincapResp = `{"data":[{"id":"bitcoin","rank":"1","symbol":"BTC","name":"Bitcoin","supply":"18159062.0000000000000000","maxSupply":"21000000.0000000000000000","marketCapUsd":"148668545611.5263838967057808","volumeUsd24Hr":"4207124490.5076897574319320","priceUsd":"8187.0167969868919384","changePercent24Hr":"-0.3396532749315348","vwap24Hr":"8152.5602443463322047"},{"id":"ethereum","rank":"2","symbol":"ETH","name":"Ethereum","supply":"109266463.9365000000000000","maxSupply":null,"marketCapUsd":"15881147723.3857246411158837","volumeUsd24Hr":"1555792423.6957113403276168","priceUsd":"145.3432933696383162","changePercent24Hr":"-0.8019725630374905","vwap24Hr":"144.4872191458879686"},{"id":"bitcoin-cash","rank":"4","symbol":"BCH","name":"Bitcoin Cash","supply":"18221137.5000000000000000","maxSupply":"21000000.0000000000000000","marketCapUsd":"4967098485.9985970396025150","volumeUsd24Hr":"245933945.7800796544167182","priceUsd":"272.6009002455854932","changePercent24Hr":"1.2571772717590811","vwap24Hr":"266.7210427895712847"},{"id":"litecoin","rank":"7","symbol":"LTC","name":"Litecoin","supply":"63853051.9606127000000000","maxSupply":"84000000.0000000000000000","marketCapUsd":"3224089460.3410892949194727","volumeUsd24Hr":"280409300.6037888998330834","priceUsd":"50.4923313975633600","changePercent24Hr":"-1.3755509580761724","vwap24Hr":"50.2102301877250892"},{"id":"eos","rank":"8","symbol":"EOS","name":"EOS","supply":"948563518.3699000000000000","maxSupply":null,"marketCapUsd":"2992516099.3591603049595676","volumeUsd24Hr":"416047653.2317845892050843","priceUsd":"3.1547872561046613","changePercent24Hr":"-0.8948392633638420","vwap24Hr":"3.1307846086015386"},{"id":"binance-coin","rank":"9","symbol":"BNB","name":"Binance Coin","supply":"155536713.0000000000000000","maxSupply":"187536713.0000000000000000","marketCapUsd":"2384116091.8913421610175340","volumeUsd24Hr":"42236487.0467411380570205","priceUsd":"15.3283173207559180","changePercent24Hr":"-0.3831046237738580","vwap24Hr":"15.1958151552810613"},{"id":"cosmos","rank":"10","symbol":"ATOM","name":"Cosmos","supply":"248453201.0000000000000000","maxSupply":null,"marketCapUsd":"1065776187.1067183778863058","volumeUsd24Hr":"8497018.9900516144346778","priceUsd":"4.2896456266897458","changePercent24Hr":"0.3860711572195862","vwap24Hr":"4.2265556693506192"},{"id":"monero","rank":"11","symbol":"XMR","name":"Monero","supply":"17398742.9426968000000000","maxSupply":null,"marketCapUsd":"1013490602.2194820549314932","volumeUsd24Hr":"18999509.0648267362502309","priceUsd":"58.2507946440406063","changePercent24Hr":"-1.7055210719932849","vwap24Hr":"58.1938614592866523"},{"id":"stellar","rank":"13","symbol":"XLM","name":"Stellar","supply":"19985458983.4742000000000000","maxSupply":null,"marketCapUsd":"972200775.3361537499423198","volumeUsd24Hr":"7275429.3168510824270986","priceUsd":"0.0486454064497622","changePercent24Hr":"-1.1256714313346928","vwap24Hr":"0.0484267511687369"},{"id":"cardano","rank":"14","symbol":"ADA","name":"Cardano","supply":"25927070538.0000000000000000","maxSupply":"45000000000.0000000000000000","marketCapUsd":"967014065.4232831535261928","volumeUsd24Hr":"8049516.1607084167714128","priceUsd":"0.0372974672941156","changePercent24Hr":"-1.6726868536292463","vwap24Hr":"0.0373040434671783"},{"id":"tezos","rank":"15","symbol":"XTZ","name":"Tezos","supply":"694191973.5305590000000000","maxSupply":null,"marketCapUsd":"870439798.5106668199800575","volumeUsd24Hr":"601018.5003859653230000","priceUsd":"1.2538891714401956","changePercent24Hr":"-0.6787710379899737","vwap24Hr":"1.2549189705034595"},{"id":"chainlink","rank":"16","symbol":"LINK","name":"Chainlink","supply":"350000000.0000000000000000","maxSupply":null,"marketCapUsd":"773817597.2177260750000000","volumeUsd24Hr":"5434215.1086896975982682","priceUsd":"2.2109074206220745","changePercent24Hr":"-1.0601570250790814","vwap24Hr":"2.1959081090890256"},{"id":"dash","rank":"19","symbol":"DASH","name":"Dash","supply":"9266404.5552159800000000","maxSupply":"18900000.0000000000000000","marketCapUsd":"651687447.1125658407718537","volumeUsd24Hr":"74182711.8280336477558675","priceUsd":"70.3279727567837022","changePercent24Hr":"5.6080870285804943","vwap24Hr":"67.4144625631858099"},{"id":"zcash","rank":"25","symbol":"ZEC","name":"Zcash","supply":"8519231.2500000000000000","maxSupply":"21000000.0000000000000000","marketCapUsd":"328701106.6710193473018581","volumeUsd24Hr":"91842527.0444833509292572","priceUsd":"38.5834234363598649","changePercent24Hr":"8.8855109033678991","vwap24Hr":"36.8997100602260872"},{"id":"basic-attention-token","rank":"29","symbol":"BAT","name":"Basic Attention Token","supply":"1421086562.3463000000000000","maxSupply":null,"marketCapUsd":"267890304.3412179445570755","volumeUsd24Hr":"1175687.9875231740053388","priceUsd":"0.1885108982375534","changePercent24Hr":"-0.9035359103226683","vwap24Hr":"0.1891809524312390"},{"id":"decred","rank":"31","symbol":"DCR","name":"Decred","supply":"10786830.8782284000000000","maxSupply":"21000000.0000000000000000","marketCapUsd":"175363923.7633690182827321","volumeUsd24Hr":"644095.9529970095552398","priceUsd":"16.2572238077186129","changePercent24Hr":"-4.4563974406003762","vwap24Hr":"16.6137979539881258"},{"id":"ravencoin","rank":"35","symbol":"RVN","name":"Ravencoin","supply":"5284655000.0000000000000000","maxSupply":"21000000000.0000000000000000","marketCapUsd":"125011783.4409349975805000","volumeUsd24Hr":"1201895.7858763144616225","priceUsd":"0.0236556186621331","changePercent24Hr":"-2.3737091158358697","vwap24Hr":"0.0237002502946506"},{"id":"factom","rank":"186","symbol":"FCT","name":"Factom","supply":"8828001.6600000000000000","maxSupply":null,"marketCapUsd":"16980322.3686381965789831","volumeUsd24Hr":"717343.5133958003785176","priceUsd":"1.9234616193579416","changePercent24Hr":"-9.0691418893170707","vwap24Hr":null}, 
	{
		"id": "neo",
		"rank": "21",
		"symbol": "NEO",
		"name": "Neo",
		"supply": "70538831.0000000000000000",
		"maxSupply": "100000000.0000000000000000",
		"marketCapUsd": "697691055.5663624728575940",
		"volumeUsd24Hr": "71298476.1164362772194702",
		"priceUsd": "9.8908791891711740",
		"changePercent24Hr": "6.1732375421269032",
		"vwap24Hr": "9.7562650566920200"
	},
	{
		"id": "ethereum-classic",
		"rank": "18",
		"symbol": "ETC",
		"name": "Ethereum Classic",
		"supply": "116313299.0000000000000000",
		"maxSupply": "210700000.0000000000000000",
		"marketCapUsd": "813985914.6299173547373237",
		"volumeUsd24Hr": "259054045.8260332826629240",
		"priceUsd": "6.9982187903544663",
		"changePercent24Hr": "0.1669216511001280",
		"vwap24Hr": "7.0577680334683030"
	},
	{
		"id": "ontology",
		"rank": "28",
		"symbol": "ONT",
		"name": "Ontology",
		"supply": "656746573.0000000000000000",
		"maxSupply": null,
		"marketCapUsd": "314510962.7926669499844657",
		"volumeUsd24Hr": "19040396.2180246242590037",
		"priceUsd": "0.4788924308443509",
		"changePercent24Hr": "2.4961559782590015",
		"vwap24Hr": "0.4760694708016068"
	},
	{
		"id": "dogecoin",
		"rank": "27",
		"symbol": "DOGE",
		"name": "Dogecoin",
		"supply": "124467938767.8020000000000000",
		"maxSupply": null,
		"marketCapUsd": "324583740.7340023232500040",
		"volumeUsd24Hr": "74165140.6567083389617276",
		"priceUsd": "0.0026077698718826",
		"changePercent24Hr": "2.5050034056238508",
		"vwap24Hr": "0.0026450422990080"
	},
	{
		"id": "vechain",
		"rank": "31",
		"symbol": "VET",
		"name": "VeChain",
		"supply": "55454734800.0000000000000000",
		"maxSupply": null,
		"marketCapUsd": "254733487.6095577863378000",
		"volumeUsd24Hr": "16664942.3631266861825977",
		"priceUsd": "0.0045935390102985",
		"changePercent24Hr": "6.0179263730689309",
		"vwap24Hr": "0.0044998773162273"
	},
	{
		"id": "huobi-token",
		"rank": "17",
		"symbol": "HT",
		"name": "Huobi Token",
		"supply": "222668092.9719210000000000",
		"maxSupply": null,
		"marketCapUsd": "933892762.6375205791299124",
		"volumeUsd24Hr": "49316948.0818501478438748",
		"priceUsd": "4.1941023079372525",
		"changePercent24Hr": "1.8552986021140738",
		"vwap24Hr": "4.1850297473325675"
	},
	{
		"id": "algorand",
		"rank": "40",
		"symbol": "ALGO",
		"name": "Algorand",
		"supply": "708514469.8441000000000000",
		"maxSupply": null,
		"marketCapUsd": "137059720.9009309905999386",
		"volumeUsd24Hr": "9817721.4790420683438338",
		"priceUsd": "0.1934466079868338",
		"changePercent24Hr": "-2.5180742989796971",
		"vwap24Hr": "0.1937739322255232"
	},
	{
		"id": "digibyte",
		"rank": "33",
		"symbol": "DGB",
		"name": "DigiByte",
		"supply": "13282775305.7105000000000000",
		"maxSupply": "21000000000.0000000000000000",
		"marketCapUsd": "248029371.3350586018937920",
		"volumeUsd24Hr": "2334611.0106594491009365",
		"priceUsd": "0.0186730081347101",
		"changePercent24Hr": "-7.7960430480711129",
		"vwap24Hr": "0.0192020828278987"
	}
],"timestamp":1578961769824}`
