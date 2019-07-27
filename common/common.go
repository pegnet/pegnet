// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full
package common

import (
	"fmt"
	"strings"
	"sync"
)

var PointMultiple float64 = 100000000

type NetworkType int

// Stats contains the hashrate and difficulty of the last mined block
var Stats MiningStats

const (
	INVALID NetworkType = iota + 1
	MAIN_NETWORK
	TEST_NETWORK
)

const (
	TransactionChainTag = "Transactions"
	MinerChainTag       = "Miners"
	OPRChainTag         = "OraclePriceRecords"
)

var (
	// Pegnet Burn Addresses
	BurnAddresses = map[string]string{
		"main":    "EC2BURNFCT2PEGNETooo1oooo1oooo1oooo1oooo1oooo19wthin",
		"test":    "EC2BURNFCT2TESTxoooo1oooo1oooo1oooo1oooo1oooo1EoyM6d",
		"mainRCD": "37399721298d77984585040ea61055377039a4c3f3e2cd48c46ff643d50fd64f",
		"testRCD": "37399721298d8b92934b4f767a56be38ad8a30cf0b7ed9d9fd2eb0919905c4af",
	}
)

func PegnetBurnAddress(network string) string {
	return BurnAddresses[strings.ToLower(network)]
}

var (
	fcPubPrefix = []byte{0x5f, 0xb1}
	fcSecPrefix = []byte{0x64, 0x78}
	ecPubPrefix = []byte{0x59, 0x2a}
	ecSecPrefix = []byte{0x5d, 0xb6}
)

// MiningStats contains the hashrate and difficulty of the last mined block
type MiningStats struct {
	Mux        sync.Mutex
	HashRate   uint64
	Difficulty uint64
	Miners     uint
}

// Update adds info from an individual miner
func (stats *MiningStats) Update(hashrate uint64, diff uint64) {
	stats.Mux.Lock()
	defer stats.Mux.Unlock()
	stats.HashRate += hashrate
	if diff > stats.Difficulty {
		stats.Difficulty = diff
	}
	stats.Miners++
}

// Clear resets the stats struct
func (stats *MiningStats) Clear() {
	stats.Mux.Lock()
	defer stats.Mux.Unlock()
	stats.HashRate = 0
	stats.Difficulty = 0
	stats.Miners = 0
}

// GetHashRate returns the hashes per second mined in the last block
func (stats *MiningStats) GetHashRate() uint64 {
	stats.Mux.Lock()
	defer stats.Mux.Unlock()
	return stats.HashRate / 480 // 480 seconds in 8 minutes
}

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
