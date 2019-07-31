// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package api

import (
	"bytes"
	"encoding/hex"
	"strconv"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/opr"
)

// -------------------------------------------------------------
// Required for M1

func getPerformance(params interface{}) (*PerformanceResult, *Error) {
	performanceParams := new(PerformanceParameters)
	err := MapToObject(params, performanceParams)
	if err != nil {
		return nil, NewJSONDecodingError()
	}

	// Parameter validation
	if performanceParams.DigitalID == "" || performanceParams.BlockRange.Start == nil {
		return nil, NewInvalidParametersError()
	}

	start := *performanceParams.BlockRange.Start
	var end int64
	var leaderHeight int64
	if start < 0 || performanceParams.BlockRange.End == nil {
		leaderHeight = getLeaderHeight()
		if start < 0 {
			start = leaderHeight + start
			if start < 0 {
				return nil, NewInvalidParametersError() // Computed a negative height from relative start
			}
		}
	}

	if performanceParams.BlockRange.End == nil {
		end = leaderHeight
	} else if start < 0 {
		return nil, NewInvalidParametersError() // Relative start cannot be mixed with absolute end
	} else {
		end = *performanceParams.BlockRange.End
	}

	if start > end {
		return nil, NewInvalidParametersError()
	}

	// Aggregate the stats
	submissions := int64(0)
	rewards := int64(0)
	difficultyPlacementsCount := int64(0)
	difficultyPlacementsSum := int64(0)
	difficultyPlacements := map[int]int64{
		1:  0,
		5:  0,
		10: 0,
		25: 0,
		50: 0,
	}

	gradingPlacementsCount := int64(0)
	gradingPlacementsSum := int64(0)
	gradingPlacements := map[int]int64{}
	for i := 1; i <= 50; i++ {
		gradingPlacements[i] = 0
	}
	for i := start; i <= end; i++ {
		block := oprBlockByHeight(i)
		if block == nil {
			continue
		}
		// Difficulty stats for this block
		for i, record := range block.OPRs {
			if record.FactomDigitalID == performanceParams.DigitalID {
				submissions += 1
				if i <= 50 {
					difficultyPlacementsCount += 1
					difficultyPlacementsSum += int64(i + 1)
					for k := range difficultyPlacements {
						if i <= k {
							difficultyPlacements[k] += 1
						}
					}
				}
			}
		}
		// Grading and reward stats for this block
		for i, record := range block.GradedOPRs {
			if record.FactomDigitalID == performanceParams.DigitalID {
				rewards += int64(opr.GetRewardFromPlace(i))
				gradingPlacementsCount += 1
				gradingPlacementsSum += int64(i + 1)
				for k := range gradingPlacements {
					if i+1 <= k {
						gradingPlacements[k] += 1
					}
				}
			}
		}
	}

	fullDifficultyPlacements := map[string]int64{
		"count": difficultyPlacementsCount,
		"sum":   difficultyPlacementsSum,
	}
	for k, v := range difficultyPlacements {
		fullDifficultyPlacements[strconv.Itoa(k)] = v
	}
	fullGradingPlacements := map[string]int64{
		"count": gradingPlacementsCount,
		"sum":   gradingPlacementsSum,
	}
	for k, v := range gradingPlacements {
		fullGradingPlacements[strconv.Itoa(k)] = v
	}

	result := &PerformanceResult{
		BlockRange:           BlockRange{Start: &start, End: &end},
		Submissions:          submissions,
		Rewards:              rewards,
		DifficultyPlacements: fullDifficultyPlacements,
		GradingPlacements:    fullGradingPlacements,
	}
	return result, nil
}

// -------------------------------------------------------------
// Somewhat temporary, might not remain

func getCurrentOPRs() (*GenericResult, *Error) {
	height := getLeaderHeight()
	records := oprBlockByHeight(height)
	return &GenericResult{OPRBlock: records}, nil
}

// getOprByHash handler to return the opr by full hash
func getOprByHash(params interface{}) (*GenericResult, *Error) {
	genericParams := new(GenericParameters)
	err := MapToObject(params, genericParams)
	if err != nil {
		return nil, NewInvalidParametersError()
	} else if genericParams.Hash == "" {
		return nil, NewInvalidParametersError()
	}
	record := oprByHash(genericParams.Hash)
	return &GenericResult{OPR: &record}, nil
}

// getOprByShortHash handler to return the opr by the short 8 byte hash
func getOprByShortHash(params interface{}) (*GenericResult, *Error) {
	genericParams := new(GenericParameters)
	err := MapToObject(params, genericParams)
	if err != nil {
		return nil, NewInvalidParametersError()
	} else if genericParams.Hash == "" {
		return nil, NewInvalidParametersError()
	}
	record := oprByShortHash(genericParams.Hash)
	return &GenericResult{OPR: &record}, nil
}

// getOprsByDigitalID handler to return all oprs based on Digital ID
func getOprsByDigitalID(params interface{}) (*GenericResult, *Error) {
	genericParams := new(GenericParameters)
	err := MapToObject(params, genericParams)
	if err != nil {
		return nil, NewInvalidParametersError()
	} else if genericParams.DigitalID == "" {
		return nil, NewInvalidParametersError()
	}
	records := oprsByDigitalID(genericParams.DigitalID)
	return &GenericResult{OPRs: records}, nil
}

// getOPRsByHeight handler will return all OPR's at any height except the current block.
// Will only return local OPR's for the current chainhead.
func getOPRsByHeight(params interface{}) (*GenericResult, *Error) {
	genericParams := new(GenericParameters)
	err := MapToObject(params, genericParams)
	if err != nil {
		return nil, NewInvalidParametersError()
	} else if genericParams.Height == nil {
		return nil, NewInvalidParametersError()
	}
	oprBlock := oprBlockByHeight(*genericParams.Height)
	return &GenericResult{OPRBlock: oprBlock}, nil
}

// getBalance handler will get the balance of a pegnet address
func getBalance(params interface{}) (*GenericResult, *Error) {
	genericParams := new(GenericParameters)
	err := MapToObject(params, genericParams)
	if err != nil {
		return nil, NewInvalidParametersError()
	} else if genericParams.Address == nil {
		return nil, NewInvalidParametersError()
	}
	balance := opr.GetBalance(*genericParams.Address)
	return &GenericResult{Balance: balance}, nil
}

// -------------------------------------------------------------
// Helpers

// getWinners returns the current 10 winners entry shorthashes from the last recorded block
func getWinners() [10]string {
	height := getLeaderHeight()
	currentOPRS := oprBlockByHeight(height)
	record := currentOPRS.OPRs[0]
	return record.WinPreviousOPR
}

// getWinner returns the highest graded entry shorthash from the last recorded block
func getWinner() string {
	return getWinners()[0]
}

// getLeaderHeight helper function, cleaner than using the factom monitor
func getLeaderHeight() int64 {
	heights, err := factom.GetHeights()
	if err != nil {
		return 0
	}
	return heights.LeaderHeight
}

// oprBlockByHeight returns a single OPRBlock
func oprBlockByHeight(dbht int64) *opr.OprBlock {
	for _, block := range opr.OPRBlocks {
		if block.Dbht == dbht {
			return block
		}
	}
	return nil
}

// oprsByDigitalID returns every OPR created by a given ID
// Multiple ID's per miner or single daemon are possible.
// This function searches through every possible ID and returns all.
func oprsByDigitalID(did string) []opr.OraclePriceRecord {
	var subset []opr.OraclePriceRecord
	for _, block := range opr.OPRBlocks {
		for _, record := range block.OPRs {
			if record.FactomDigitalID == did {
				subset = append(subset, *record)
			}
		}
	}
	return subset
}

// oprByHash returns the entire OPR based on it's hash
func oprByHash(hash string) opr.OraclePriceRecord {
	for _, block := range opr.OPRBlocks {
		for _, record := range block.OPRs {
			if hash == hex.EncodeToString(record.OPRHash) {
				return *record
			}
		}
	}
	return opr.OraclePriceRecord{}
}

// Failing tests. Need to grok how the short 8 byte winning oprhashes are done.
func oprByShortHash(shorthash string) opr.OraclePriceRecord {
	hashBytes, _ := hex.DecodeString(shorthash)
	// hashbytes = reverseBytes(hashbytes)
	for _, block := range opr.OPRBlocks {
		for _, record := range block.OPRs {
			if bytes.Compare(hashBytes, record.OPRHash[:8]) == 0 {
				return *record
			}
		}
	}
	return opr.OraclePriceRecord{}
}
