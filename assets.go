package oprecord

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"

	"github.com/pegnet/OracleRecord/common"
	"github.com/pegnet/OracleRecord/utils"
	"github.com/zpatrick/go-config"
)

type PegAssets struct {
	PNT        PegItems
	USD        PegItems
	EUR        PegItems
	JPY        PegItems
	GBP        PegItems
	CAD        PegItems
	CHF        PegItems
	INR        PegItems
	SGD        PegItems
	CNY        PegItems
	HKD        PegItems
	XAU        PegItems
	XAG        PegItems
	XPD        PegItems
	XPT        PegItems
	XBT        PegItems
	ETH        PegItems
	LTC        PegItems
	XBC        PegItems
	FCT        PegItems
	PriceBytes [160]byte
}

type PegItems struct {
	Value int64
	When  string
}

func PullPEGAssets(config *config.Config) PegAssets {
	var Peg PegAssets
	Peg.USD.Value = int64(1 * common.PointMultiple)
	// digital currencies
	CoinCapResponseBytes, err := CallCoinCap(config)
	//fmt.Println(string(CoinCapResponseBytes))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		var CoinCapValues CoinCapResponse
		err = json.Unmarshal(CoinCapResponseBytes, &CoinCapValues)
		for _, currency := range CoinCapValues.Data {
			//fmt.Println(currency.Symbol + "-" + currency.PriceUSD)
			if currency.Symbol == "XBT" || currency.Symbol == "BTC" {
				Peg.XBT.Value = utils.FloatStringToInt(currency.PriceUSD)
				Peg.XBT.When = string(CoinCapValues.Timestamp)
			} else if currency.Symbol == "ETH" {
				Peg.ETH.Value = utils.FloatStringToInt(currency.PriceUSD)
				Peg.ETH.When = string(CoinCapValues.Timestamp)
			} else if currency.Symbol == "LTC" {
				Peg.LTC.Value = utils.FloatStringToInt(currency.PriceUSD)
				Peg.LTC.When = string(CoinCapValues.Timestamp)
			} else if currency.Symbol == "XBC" || currency.Symbol == "BCH" {
				Peg.XBC.Value = utils.FloatStringToInt(currency.PriceUSD)
				Peg.XBC.When = string(CoinCapValues.Timestamp)
			} else if currency.Symbol == "FCT" {
				Peg.FCT.Value = utils.FloatStringToInt(currency.PriceUSD)
				Peg.FCT.When = string(CoinCapValues.Timestamp)
			}
		}
	}

	//fiat option 1.  terms uf use seem tighter
	// has fiat and digital
	// https://currencylayer.com/product  <-- pricing
	// $10 a month will let you pull 10,000 but it is updated once an hour
	// $40 a month is 100,000 updated every 10 minutes.
	// 20% discount for annual payment

	//	fmt.Println("API LAYER:")

	APILayerBytes, err := CallAPILayer(config)
	//fmt.Println(string(APILayerBytes))

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		var APILayerResponse APILayerResponse
		err = json.Unmarshal(APILayerBytes, &APILayerResponse)
		//	fmt.Println(APILayerResponse)
		//	fmt.Println("UDS-GBP")
		//	fmt.Println(APILayerResponse.Quotes.USDGBP)
		Peg.EUR.Value = int64(1 / APILayerResponse.Quotes.USDEUR * common.PointMultiple)
		Peg.EUR.When = string(APILayerResponse.Timestamp)
		Peg.JPY.Value = int64(1 / APILayerResponse.Quotes.USDJPY * common.PointMultiple)
		Peg.JPY.When = string(APILayerResponse.Timestamp)
		Peg.GBP.Value = int64(1 / APILayerResponse.Quotes.USDGBP * common.PointMultiple)
		Peg.GBP.When = string(APILayerResponse.Timestamp)
		Peg.CAD.Value = int64(1 / APILayerResponse.Quotes.USDCAD * common.PointMultiple)
		Peg.CAD.When = string(APILayerResponse.Timestamp)
		Peg.CHF.Value = int64(1 / APILayerResponse.Quotes.USDCHF * common.PointMultiple)
		Peg.CHF.When = string(APILayerResponse.Timestamp)
		Peg.INR.Value = int64(1 / APILayerResponse.Quotes.USDINR * common.PointMultiple)
		Peg.INR.When = string(APILayerResponse.Timestamp)
		Peg.SGD.Value = int64(1 / APILayerResponse.Quotes.USDSGD * common.PointMultiple)
		Peg.SGD.When = string(APILayerResponse.Timestamp)
		Peg.CNY.Value = int64(1 / APILayerResponse.Quotes.USDCNY * common.PointMultiple)
		Peg.CNY.When = string(APILayerResponse.Timestamp)
		Peg.HKD.Value = int64(1 / APILayerResponse.Quotes.USDHKD * common.PointMultiple)
		Peg.HKD.When = string(APILayerResponse.Timestamp)

	}

	KitcoResponse, err := CallKitcoWeb()

	if err != nil {
		fmt.Println(err)
		//	os.Exit(1)
	} else {

		//fmt.Println("KitcoResponse:", KitcoResponse)
		Peg.XAU.Value = utils.FloatStringToInt(KitcoResponse.Silver.Bid)
		Peg.XAU.When = KitcoResponse.Silver.Date
		Peg.XAG.Value = utils.FloatStringToInt(KitcoResponse.Gold.Bid)
		Peg.XAG.When = KitcoResponse.Gold.Date
		Peg.XPD.Value = utils.FloatStringToInt(KitcoResponse.Palladium.Bid)
		Peg.XPD.When = KitcoResponse.Palladium.Date
		Peg.XPT.Value = utils.FloatStringToInt(KitcoResponse.Platinum.Bid)
		Peg.XPT.When = KitcoResponse.Platinum.Date

	}

	return Peg

}

func (peg *PegAssets) FillPriceBytes() {
	byteVal := make([]byte, 160)
	nextStart := 0
	byteLength := 8
	b := make([]byte, 8)

	binary.BigEndian.PutUint64(b, uint64(peg.PNT.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.USD.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.EUR.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.JPY.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.GBP.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.CAD.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.CHF.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.INR.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.SGD.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.CNY.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.HKD.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.XAU.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.XAG.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.XPD.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.XPT.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.XBT.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.ETH.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.LTC.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.XBC.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])
	nextStart = nextStart + byteLength
	binary.BigEndian.PutUint64(b, uint64(peg.FCT.Value))
	copy(byteVal[nextStart:nextStart+8], b[:])

	copy(peg.PriceBytes[0:], byteVal[:])
}
