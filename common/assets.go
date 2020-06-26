// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package common

import "strings"

/*          All the assets on pegnet
 *
 *          PegNet,                 PEG,        PEG
 *
 *          US Dollar,              USD,        pUSD
 *          Euro,                   EUR,        pEUR
 *          Japanese Yen,           JPY,        pJPY
 *          Pound Sterling,         GBP,        pGBP
 *          Canadian Dollar,        CAD,        pCAD
 *          Swiss Franc,            CHF,        pCHF
 *          Indian Rupee,           INR,        pINR
 *          Singapore Dollar,       SGD,        pSGD
 *          Chinese Yuan,           CNY,        pCNY
 *          Hong Kong Dollar,       HKD,        pHKD
 * DROPPED  Tiawanese Dollar,       TWD,        pTWD
 *          Korean Won,             KRW,        pKRW
 * DROPPED  Argentine Peso,         ARS,        pARS
 *          Brazil Real,            BRL,        pBRL
 *          Philippine Peso         PHP,        pPHP
 *          Mexican Peso            MXN,        pMXN
 *
 *          Gold Troy Ounce,        XAU,        pXAU
 *          Silver Troy Ounce,      XAG,        pXAG
 *          Palladium Troy Ounce,   XPD,        pXPD
 *          Platinum Troy Ounce,    XPT,        pXPT
 *
 *          Bitcoin,                XBT,        pXBT
 *          Ethereum,               ETH,        pETH
 *          Litecoin,               LTC,        pLTC
 *          Ravencoin,              RVN,        pRVN
 *          Bitcoin Cash,           XBC,        pXBC
 *          Factom,                 FCT,        pFCT
 *          Binance Coin            BNB,        pBNB
 *          Stellar                 XLM,        pXLM
 *          Cardano                 ADA,        pADA
 *          Monero                  XMR,        pXMR
 *          Dash                    DASH,       pDASH
 *          Zcash                   ZEC,      	pZEC
 *          Decred                  DCR,        pDCR
 */

var (
	PEGAsset = []string{
		"PEG",
	}

	CurrencyAssets = []string{
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
		//"TWD",
		"KRW",
		//"ARS",
		"BRL",
		"PHP",
		"MXN",
	}

	CommodityAssets = []string{
		"XAU",
		"XAG",
		"XPD",
		"XPT",
	}

	CryptoAssets = []string{
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

	V4CurrencyAdditions = []string{
		"AUD",
		"NZD",
		"SEK",
		"NOK",
		"RUB",
		"ZAR",
		"TRY",
	}

	V4CryptoAdditions = []string{
		"EOS",
		"LINK",
		"ATOM",
		"BAT",
		"XTZ",
	}

	V5CryptoAdditions = []string{
		"HBAR",
		"NEO",
		"CRO",
		"ETC",
		"ONT",
		"DOGE",
		"VET",
		"HT",
		"ALGO",
		"DGB",
	}

	V5CurrencyAdditions = []string{
		"AED",
		"ARS",
		"TWD",
		"RWF",
		"KES",
		"UGX",
		"TZS",
		"BIF",
		"ETB",
		"NGN",
	}

	AllAssets = MergeLists(PEGAsset, CurrencyAssets, CommodityAssets, CryptoAssets, V4CurrencyAdditions, V4CryptoAdditions, V5CryptoAdditions, V5CurrencyAdditions)
	AssetsV1  = MergeLists(PEGAsset, CurrencyAssets, CommodityAssets, CryptoAssets)
	// This is with the PNT instead of PEG. Should never be used unless absolutely necessary.
	//
	// Deprecated: Was used for version 1 before PNT -> PEG
	AssetsV1WithPNT = MergeLists([]string{"PNT"}, SubtractFromSet(AssetsV1, "PEG"))
	// Version One, subtract 2 assets
	AssetsV2 = SubtractFromSet(AssetsV1, "XPD", "XPT")

	// Additional assets to V2 set
	AssetsV4 = MergeLists(AssetsV2, V4CurrencyAdditions, V4CryptoAdditions)

	// Additional assets to V4 set
	AssetsV5 = MergeLists(AssetsV4, V5CryptoAdditions, V5CurrencyAdditions)
)

// AssetListContainsCaseInsensitive is for when using user input. It's helpful for the
// cmd line.
func AssetListContainsCaseInsensitive(assetList []string, asset string) bool {
	for _, a := range assetList {
		if strings.ToLower(asset) == strings.ToLower(a) {
			return true
		}
	}
	return false
}

func AssetListContains(assetList []string, asset string) bool {
	for _, a := range assetList {
		if asset == a {
			return true
		}
	}
	return false
}

func SubtractFromSet(set []string, sub ...string) []string {
	var result []string
	for _, r := range set {
		if !AssetListContains(sub, r) {
			result = append(result, r)
		}
	}
	return result
}

func MergeLists(assets ...[]string) []string {
	acc := []string{}
	for _, list := range assets {
		acc = append(acc, list...)
	}
	return acc
}
