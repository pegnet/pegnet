// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full
package common

import (
	"fmt"
	"strings"

	"github.com/zpatrick/go-config"
)

var PointMultiple float64 = 100000000

type NetworkType string

// LoadConfigNetwork handles the different casings of `MainNet`.
//	So: `mainnet`, `Mainnet`, and other stuff is all valid
func LoadConfigNetwork(c *config.Config) (string, error) {
	network, err := c.String(ConfigPegnetNetwork)
	if err != nil {
		return "", err
	}
	return GetNetwork(network)
}

func LoadConfigStakerNetwork(c *config.Config) (string, error) {
	network, err := c.String(ConfigPegnetStakeNetwork)
	if err != nil {
		return "", err
	}
	return GetNetwork(network)
}

func GetNetwork(network string) (string, error) {
	switch strings.ToLower(network) {
	case strings.ToLower(MainNetwork):
		return MainNetwork, nil
	case strings.ToLower(TestNetwork), strings.ToLower("TestNet"):
		return TestNetwork, nil
	case strings.ToLower(UnitTestNetwork), strings.ToLower("UnitTest"):
		return UnitTestNetwork, nil
	default:
		return "", fmt.Errorf("'%s' is not a valid network", network)
	}
}

const (
	ZeroHash = "0000000000000000000000000000000000000000000000000000000000000000"
)

const (
	MainNetwork     = "MainNet"
	TestNetwork     = "TestNet-pM2"
	UnitTestNetwork = "unit-test" // Only used for unit tests

	MainNetworkRCD = MainNetwork + "RCD"
	TestNetworkRCD = TestNetwork + "RCD"
)

const (
	TransactionChainTag = "Transactions"
	MinerChainTag       = "Miners"
	OPRChainTag         = "OraclePriceRecords"
	SPRChainTag         = "StakingPriceRecords"
)

var (
	// Pegnet Burn Addresses
	BurnAddresses = map[string]string{
		MainNetwork:    "EC2BURNFCT2PEGNETooo1oooo1oooo1oooo1oooo1oooo19wthin",
		TestNetwork:    "EC2BURNFCT2TESTxoooo1oooo1oooo1oooo1oooo1oooo1EoyM6d",
		MainNetworkRCD: "37399721298d77984585040ea61055377039a4c3f3e2cd48c46ff643d50fd64f",
		TestNetworkRCD: "37399721298d8b92934b4f767a56be38ad8a30cf0b7ed9d9fd2eb0919905c4af",
	}
)

func PegnetBurnAddress(network string) string {
	return BurnAddresses[network]
}

var (
	fcPubPrefix = []byte{0x5f, 0xb1}
	fcSecPrefix = []byte{0x64, 0x78}
	ecPubPrefix = []byte{0x59, 0x2a}
	ecSecPrefix = []byte{0x5d, 0xb6}
)

// FormatDiff returns a human readable string in scientific notation
func FormatDiff(diff uint64, precision uint) string {
	format := "%." + fmt.Sprint(precision) + "e"
	return fmt.Sprintf(format, float64(diff))
}

// FormatGrade returns a human readable string in scientific notation
func FormatGrade(grade float64, precision uint) string {
	format := "%." + fmt.Sprint(precision) + "e"
	return fmt.Sprintf(format, grade)
}
