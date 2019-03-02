package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/PegNet/OracleRecord/structs"
)

var PointMultiple float64 = 100000000

func main() {
	var Peg structs.PegValues
	Peg = PullPEGValues()
	fmt.Println(Peg)
	//add error handling for padding bytes
	pBytes := FillPriceBytes(Peg)
	copy(Peg.PriceBytes[0:], pBytes[:])
	fmt.Println(Peg)
}

func PullPEGValues() structs.PegValues {
	var Peg structs.PegValues
	Peg.USD.Value = int64(1 * PointMultiple)
	// digital currencies
	CoinCapResponseBytes, err := structs.CallCoinCap()
	//fmt.Println(string(CoinCapResponseBytes))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		var CoinCapValues structs.CoinCapResponse
		err = json.Unmarshal(CoinCapResponseBytes, &CoinCapValues)
		for _, currency := range CoinCapValues.Data {
			//fmt.Println(currency.Symbol + "-" + currency.PriceUSD)
			if currency.Symbol == "XBT" || currency.Symbol == "BTC" {
				Peg.XBT.Value = FloatStringToInt(currency.PriceUSD)
				Peg.XBT.When = string(CoinCapValues.Timestamp)
			} else if currency.Symbol == "ETH" {
				Peg.ETH.Value = FloatStringToInt(currency.PriceUSD)
				Peg.ETH.When = string(CoinCapValues.Timestamp)
			} else if currency.Symbol == "LTC" {
				Peg.LTC.Value = FloatStringToInt(currency.PriceUSD)
				Peg.LTC.When = string(CoinCapValues.Timestamp)
			} else if currency.Symbol == "XBC" || currency.Symbol == "BCH" {
				Peg.XBC.Value = FloatStringToInt(currency.PriceUSD)
				Peg.XBC.When = string(CoinCapValues.Timestamp)
			} else if currency.Symbol == "FCT" {
				Peg.FCT.Value = FloatStringToInt(currency.PriceUSD)
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

	APILayerBytes, err := structs.CallAPILayer()
	//fmt.Println(string(APILayerBytes))

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		var APILayerResponse structs.APILayerResponse
		err = json.Unmarshal(APILayerBytes, &APILayerResponse)
		//	fmt.Println(APILayerResponse)
		//	fmt.Println("UDS-GBP")
		//	fmt.Println(APILayerResponse.Quotes.USDGBP)
		Peg.EUR.Value = int64(APILayerResponse.Quotes.USDEUR * PointMultiple)
		Peg.EUR.When = string(APILayerResponse.Timestamp)
		Peg.JPY.Value = int64(APILayerResponse.Quotes.USDJPY * PointMultiple)
		Peg.JPY.When = string(APILayerResponse.Timestamp)
		Peg.GBP.Value = int64(APILayerResponse.Quotes.USDGBP * PointMultiple)
		Peg.GBP.When = string(APILayerResponse.Timestamp)
		Peg.CAD.Value = int64(APILayerResponse.Quotes.USDCAD * PointMultiple)
		Peg.CAD.When = string(APILayerResponse.Timestamp)
		Peg.CHF.Value = int64(APILayerResponse.Quotes.USDCHF * PointMultiple)
		Peg.CHF.When = string(APILayerResponse.Timestamp)
		Peg.INR.Value = int64(APILayerResponse.Quotes.USDINR * PointMultiple)
		Peg.INR.When = string(APILayerResponse.Timestamp)
		Peg.SGD.Value = int64(APILayerResponse.Quotes.USDSGD * PointMultiple)
		Peg.SGD.When = string(APILayerResponse.Timestamp)
		Peg.CNY.Value = int64(APILayerResponse.Quotes.USDCNY * PointMultiple)
		Peg.CNY.When = string(APILayerResponse.Timestamp)
		Peg.HKD.Value = int64(APILayerResponse.Quotes.USDHKD * PointMultiple)
		Peg.HKD.When = string(APILayerResponse.Timestamp)

	}

	KitcoResponse, err := structs.CallKitcoWeb()

	if err != nil {
		fmt.Println(err)
		//	os.Exit(1)
	} else {

		//fmt.Println("KitcoResponse:", KitcoResponse)
		Peg.XAU.Value = FloatStringToInt(KitcoResponse.Silver.Bid)
		Peg.XAU.When = KitcoResponse.Silver.Date
		Peg.XAG.Value = FloatStringToInt(KitcoResponse.Gold.Bid)
		Peg.XAG.When = KitcoResponse.Gold.Date
		Peg.XPD.Value = FloatStringToInt(KitcoResponse.Palladium.Bid)
		Peg.XPD.When = KitcoResponse.Palladium.Date
		Peg.XPT.Value = FloatStringToInt(KitcoResponse.Platinum.Bid)
		Peg.XPT.When = KitcoResponse.Platinum.Date

	}

	return Peg

}

func FloatStringToInt(floatString string) int64 {
	//fmt.Println(floatString)
	if floatString == "-" {
		return 0
	}
	if strings.TrimSpace(floatString) == "" {
		return 0
	}
	floatValue, err := strconv.ParseFloat(floatString, 64)
	if err != nil {
		fmt.Println("ParseError:", floatString)
		return 0
	} else {
		return int64(floatValue * PointMultiple)
	}

}

/*   Not used right now.  structures are there if you want to use it
//   you will need to replace the values put into peg structure
func CallOpenExchangeRates() ([]byte, error) {
	resp, err := http.Get("https://openexchangerates.org/api/latest.json?app_id=<INSERT API KEY HERE>")
	if err != nil {
		return nil, err
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		return body, err
	}

}
*/

func FillPriceBytes(peg structs.PegValues) []byte {
	byteVal := make([]byte, 160)
	nextStart := 0
	byteLength := 8
	b := make([]byte, 8)

	binary.BigEndian.PutUint64(b, uint64(peg.Pegnet.Value))
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

	return byteVal[:]
}
