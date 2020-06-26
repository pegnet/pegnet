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
//      * Add Pegnet asset pUSD (exchange price for pUSD)
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
}

// V5Assets contains the following changes to V4:
//      * Add CryptoCurrency HBAR
//      * Add CryptoCurrency NEO
//      * Add Currency AED
//      * Add CryptoCurrency CRO
//      * Add CryptoCurrency ETC
//      * Add CryptoCurrency ONT
//      * Add CryptoCurrency DOGE
//      * Add CryptoCurrency VET
//      * Add CryptoCurrency HT
//      * Add CryptoCurrency ALGO
//      * Add Currency ARS
//      * Add Currency TWD
var V5Assets = []string{
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

	// New Assets
	"HBAR",
	"NEO",
	"AED",
	"CRO",
	"ETC",
	"ONT",
	"DOGE",
	"VET",
	"HT",
	"ALGO",
	"ARS",
	"TWD",
	"RWF",
	"KES",
	"UGX",
	"TZS",
	"BIF",
	"ETB",
	"DGB",
	"NGN",
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
