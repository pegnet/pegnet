package opr

// V1Assets is the list of assets PegNet launched with
var V1Assets = []string{
	"PNT",
	"USD",
	"EUR",
	"JPY",
	"GBP",
	"CAD",
	"CHF",
	"INR",
	"SGD",
	"CNY",
	"HKD",
	"KRW",
	"BRL",
	"PHP",
	"MXN",
	"XAU",
	"XAG",
	"XPD",
	"XPT",
	"XBT",
	"ETH",
	"LTC",
	"RVN",
	"XBC",
	"FCT",
	"BNB",
	"XLM",
	"ADA",
	"XMR",
	"DASH",
	"ZEC",
	"DCR",
}

// V2Assets contains the following changes to V1:
//		* Rename PNT to PEG
//		* Drop XPD
// 		* Drop XPT
var V2Assets = []string{
	"PEG",
	"USD",
	"EUR",
	"JPY",
	"GBP",
	"CAD",
	"CHF",
	"INR",
	"SGD",
	"CNY",
	"HKD",
	"KRW",
	"BRL",
	"PHP",
	"MXN",
	"XAU",
	"XAG",
	"XBT",
	"ETH",
	"LTC",
	"RVN",
	"XBC",
	"FCT",
	"BNB",
	"XLM",
	"ADA",
	"XMR",
	"DASH",
	"ZEC",
	"DCR",
}

// V4Assets contains the following changes to V2:
//      * Add Currency AUD
//      * Add Currency NZD
//      * Add Currency SEK
//      * Add Currency NOK
//      * Add Currency RUB
//      * Add Currency ZAR
//      * Add Currency TRY
//      * Add CryptoCurrency EOS
//      * Add CryptoCurrency LINK
//      * Add CryptoCurrency ATOM
//      * Add CryptoCurrency BAT
//      * Add CryptoCurrency XTZ
var V4Assets = []string{
	"PEG",
	"USD",
	"EUR",
	"JPY",
	"GBP",
	"CAD",
	"CHF",
	"INR",
	"SGD",
	"CNY",
	"HKD",
	"KRW",
	"BRL",
	"PHP",
	"MXN",
	"XAU",
	"XAG",
	"XBT",
	"ETH",
	"LTC",
	"RVN",
	"XBC",
	"FCT",
	"BNB",
	"XLM",
	"ADA",
	"XMR",
	"DASH",
	"ZEC",
	"DCR",

	// New Assets
	"AUD",
	"NZD",
	"SEK",
	"NOK",
	"RUB",
	"ZAR",
	"TRY",
	"EOS",
	"LINK",
	"ATOM",
	"BAT",
	"XTZ",

	// Reference Tokens
	"pUSD",
}

// AssetFloat is an asset holding a float64 value
type AssetFloat struct {
	Name  string
	Value float64
}

// AssetUint is an asset holding a uint64 value
type AssetUint struct {
	Name  string
	Value uint64
}
