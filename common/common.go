// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full
package common

var PointMultiple float64 = 100000000

// Alert from the Factomd monitor
type FDStatus struct {
	Minute int64
	Dbht   int32
}

type NetworkType int

const (
	INVALID NetworkType = iota + 1
	MAIN_NETWORK
	TEST_NETWORK
)

var AssetNames = []string{
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
	"XAU",
	"XAG",
	"XPD",
	"XPT",
	"XBT",
	"ETH",
	"LTC",
	"XBC",
	"FCT",
}

var (
	fcPubPrefix = []byte{0x5f, 0xb1}
	fcSecPrefix = []byte{0x64, 0x78}
	ecPubPrefix = []byte{0x59, 0x2a}
	ecSecPrefix = []byte{0x5d, 0xb6}
)
