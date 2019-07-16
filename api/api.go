// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package api

import (
	"bytes"
	"reflect"
	"net/http"
	"encoding/json"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"github.com/pegnet/pegnet/opr"
	"github.com/FactomProject/factom"
)

func apiHandler(w http.ResponseWriter, r *http.Request) {
	var request PostRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		jsonDecodingError(w)
		return 
		}
	params, _ := json.Marshal(request.Params)
	log.WithFields(log.Fields{
		"API Method": request.Method,
		"Params": string(params)}).Info("API Request")

	switch request.Method {
		// All oprs in the OPRBlocks struct (large!)
		case "all-oprs":
			response(w, Result{OPRBlocks: opr.OPRBlocks})
		
		// Balance of a particular pegnet address
		case "balance":
			balance(w, request.Params)

		// Factom chainid that oprs are entered into
		case "chainid":
			response(w, Result{ChainID: opr.OPRChainID})

		// The OPR of the last winning miner. Contains all conversion rates
		case "current-rates":
			rates := CurrentRates()
			response(w, Result{OPR: &rates})

		// Returns the conversion rate for a particular ticker
		case "conversion-rate":
			conversionRate(w, request.Params)
			
		// Highest current block
		case "leaderheight":
			response(w, Result{LeaderHeight: LeaderHeight()})

		// Returns the difficulty of an OPR given its shorthash
		case "opr-difficulty":
			oprDifficulty(w, request.Params)

		// Returns the full factom entryhash of an OPR given its shorthash
		case "opr-entryhash":
			oprEntry(w, request.Params)

		// Gets all the OPRs given a particular height
		case "oprs-by-height":
			byHeight(w, request.Params)

		// Gets all the OPRs given a particular Digital ID (Multiple IDs possible per miner)
		case "oprs-by-id":
			byDigitalID(w, request.Params)

		// Returns the full OPR when given a valid OPR Hash
		case "opr-by-hash":
			byHash(w, request.Params)

		// Returns the full OPR when given a valid entry Hash
		case "opr-by-entryhash":
			byEntryHash(w, request.Params)

		// Returns the full OPR when given a valid short entry Hash
		case "opr-by-shorthash":
			byShortHash(w, request.Params)

		// List of shorthash strings of current winners
		case "winners":
			winners := GetWinners()
			response(w, Result{Winners: winners[:]})

		// Single shorthash string of the current highest graded OPR 
		case "winner":
			response(w, Result{Winner: GetWinner()}) 

		// Full OPR of the current highest graded OPR
		case "winning-opr":
			winningOPR := OprByShortHash(GetWinner())
			response(w, Result{OPR: &winningOPR})

		default:
			methodNotFound(w, request.Method)
	}
}

// response is a wrapper around all responses to be served
func response(w http.ResponseWriter, res Result){
	json.NewEncoder(w).Encode(PostResponse{Res: res})
}

func conversionRate(w http.ResponseWriter, params Parameters) {
	if params.Ticker != "" {
		rates := CurrentRates()
		rate := reflect.ValueOf(rates).FieldByName(params.Ticker).Float()
		response(w, Result{Rate: rate})
	} else {
		invalidParameterError(w, params)
	}
}

// byHash handler to return the opr by full hash
func byHash(w http.ResponseWriter, params Parameters) {
	if params.Hash != "" {
		opr := OprByHash(params.Hash)
		response(w, Result{OPR: &opr})
	} else {
		invalidParameterError(w, params)
	}
}

// getOprByHash handler to return the opr by full hash
func byEntryHash(w http.ResponseWriter, params Parameters) {
	if params.Hash != "" {
		opr := OprByEntryHash(params.Hash)
		response(w, Result{OPR: &opr})
	} else {
		invalidParameterError(w, params)
	}
}

// byShortHash handler to return the opr by the short 8 byte hash
func byShortHash(w http.ResponseWriter, params Parameters) {
	if params.Hash != "" {
		opr := OprByShortHash(params.Hash)
		response(w, Result{OPR: &opr})
	} else {
		invalidParameterError(w, params)
	}
}

// getOprsByDigitalID handler to return all oprs based on Digital ID
func byDigitalID(w http.ResponseWriter, params Parameters) {
	if params.DigitalID != "" {
		oprs := OprsByDigitalID(params.DigitalID)
		response(w, Result{OPRs: oprs})
	} else {
		invalidParameterError(w, params)
	}
}

// byHeight handler will grab OPRs for a height except current block
func byHeight(w http.ResponseWriter, params Parameters) {
	if params.Height !=nil {
		oprblock := OprsByHeight(*params.Height)
		if oprblock != nil {
			response(w, Result{OPRBlock: oprblock})
		} else {
			oprLookupError(w, params)
		}
	} else {
		invalidParameterError(w, params)
	}
}

func oprDifficulty(w http.ResponseWriter, params Parameters) {
	if params.Hash != "" {
		oprblock := OprByShortHash(params.Hash)
		if oprblock.OPRChainID == "" {
			oprLookupError(w, params)
		}
		response(w, Result{Difficulty: oprblock.Difficulty})
	} else {
		invalidParameterError(w, params)
	}
}

func oprEntry(w http.ResponseWriter, params Parameters) {
	if params.Hash != "" {
		oprblock := OprByShortHash(params.Hash)
		if &oprblock == nil {
			oprLookupError(w, params)
		}
		response(w, Result{EntryHash: hex.EncodeToString(oprblock.Entry.Hash())})
	} else {
		invalidParameterError(w, params)
	}
}

// balance handler will get the balance of a pegnet address
func balance(w http.ResponseWriter, params Parameters) {
	if params.Address != nil {
		balance := opr.GetBalance(*params.Address)
		res := Result{Balance: balance}
		response(w, res)
	} else {
		invalidParameterError(w, params)
	}
}

// CurrentRates returns the last winning OPR Block
func CurrentRates() opr.OraclePriceRecord {
	winner := GetWinner()
	return OprByShortHash(winner)
}

// GetWinners returns the current 10 winners entry shorthashes from the last recorded block
func GetWinners() [10]string {
	var winners [10]string
	height := LeaderHeight()
	currentOPRS := OprsByHeight(height - 1)
	if currentOPRS != nil {
		opr := currentOPRS.OPRs[0]
		return opr.WinPreviousOPR
	}
	return winners
}

// GetWinner returns the highest graded entry shorthash from the last recorded block
func GetWinner() string {
	return GetWinners()[0]
}

// LeaderHeight helper function, returns current height
func LeaderHeight() int64 {
	heights, err := factom.GetHeights()
	if err != nil {
		return 0
	}
	return heights.LeaderHeight
}

// OprsByHeight returns a single OPRBlock
func OprsByHeight(dbht int64) *opr.OprBlock {
	for _, opr := range opr.OPRBlocks {
		if opr.Dbht == dbht {
		return opr
		}
	}
	return nil
}

// OprsByDigitalID returns every OPR created by a given ID
// Multiple ID's per miner are possible.
// This function searches through every possible ID and returns all.
func OprsByDigitalID(did string) []opr.OraclePriceRecord {
	var subset []opr.OraclePriceRecord
	for _, block := range opr.OPRBlocks {
		for _, opr := range block.OPRs{
			for _, digitalID := range opr.FactomDigitalID{
				if digitalID == did {
				subset = append(subset, *opr)
				}
			}
		}
	}
	return subset
}

// OprByHash returns the entire OPR based on it's hash 
func OprByHash(hash string) opr.OraclePriceRecord {
	oprhash, err := hex.DecodeString(hash)
	if err == nil {
		for _, block := range opr.OPRBlocks {
			for _, opr := range block.OPRs{
				if bytes.Compare(oprhash, opr.OPRHash) == 0 {
					return *opr
				}
			}    
		}
	}
	return opr.OraclePriceRecord{}
}

// OprByEntryHash returns the entire OPR based on it's Factom Entry hash 
func OprByEntryHash(hash string) opr.OraclePriceRecord {
	oprhash, err := hex.DecodeString(hash)
	if err == nil {
		for _, block := range opr.OPRBlocks {
			for _, opr := range block.OPRs{
				if bytes.Compare(oprhash, opr.Entry.Hash()) == 0 {
					return *opr
				}
			}    
		}
	}
	return opr.OraclePriceRecord{}
}

// OprByShortHash checks the truncated entry hash for listed OPR winners
func OprByShortHash(shorthash string) opr.OraclePriceRecord {
	hashbytes, _  := hex.DecodeString(shorthash)
	for _, block := range opr.OPRBlocks {
		for _, opr := range block.OPRs{
			if bytes.Compare(hashbytes, opr.Entry.Hash()[:8]) ==  0 {
			return *opr
			}
		}    
	}
	return opr.OraclePriceRecord{}
}