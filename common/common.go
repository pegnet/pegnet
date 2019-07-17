// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full
package common

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/FactomProject/factom"
)

var PointMultiple float64 = 100000000

type NetworkType int

const (
	INVALID NetworkType = iota + 1
	MAIN_NETWORK
	TEST_NETWORK
)

var (
	// Pegnet Burn Addresses
	BurnAddresses = map[string]string{
		"main": "EC1moooFCT2mainoooo1oooo1oooo1oooo1oooo1oooo1opfDJqF",
		"test": "EC1moooFCT2TESToooo1oooo1oooo1oooo1oooo1oooo1on1iNDk",
	}
)

func PegnetBurnAddress(network string) string {
	return BurnAddresses[strings.ToLower(network)]
}

// TODO: Remove this, when the final main and test burn addresses are determined.
//		Once they are determined, then we don't need to check if they are valid on init.
//  	This code is put here to ensure the chosen burn addresses are actually valid.
//		If the address is invalid, it will panic with a corrected checksum, so you can
//		update the burn address you put here. So it makes changing the burn addresses easy.
func init() {
	for network, add := range BurnAddresses {
		if !factom.IsValidAddress(add) {
			// If it is not valid, could be checksum related
			newAddr, err := ConvertUserStrFctEcToAddress(add)
			if err == nil {
				// Try and fix it, then suggest the new checksum
				raw, _ := hex.DecodeString(newAddr)
				burn := ConvertECAddressToUser(raw)
				if factom.IsValidAddress(burn) {
					panic(fmt.Sprintf("[%s] %s is not a valid address, but %s is", network, add, burn))
				}
			}
			panic(fmt.Sprintf("[%s] %s is not a valid address.", network, add))
		}
	}
}

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
