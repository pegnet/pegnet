// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package api

import (
	"bytes"
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
		case "all-oprs":
			response(w, Result{OPRBlocks: opr.OPRBlocks})
		
		case "balance":
			getBalance(w, request.Params)

		case "chainid":
			response(w, Result{ChainID: opr.OPRChainID})

		// case "current-oprs":
		// 	var current []opr.OraclePriceRecord
		// 	for _, opr := range opr.CurrentOPRs {
		// 		current = append(current, *opr)
		// 	}
		// 	response(w, Result{OPRs: current})
    
		case "leaderheight":
			response(w, Result{LeaderHeight: leaderHeight()})

		case "oprs-by-height":
			getOPRsByHeight(w, request.Params)

		case "oprs-by-id":
			getOprsByDigitalID(w, request.Params)

		case "opr-by-hash":
			getOprByHash(w, request.Params)

		case "opr-by-entry-hash":
			getOprByHash(w, request.Params)

		case "opr-by-shorthash":
			getOprByShortHash(w, request.Params)

		case "winners":
			winners := getWinners()
			response(w, Result{Winners: winners[:]})

		case "winner":
			winner :=  getWinner()
			response(w, Result{Winner: winner}) 

		// Failing method - shorthash needs to be fixed
		case "winning-opr":
			winner :=  getWinner()
			winningOPR := oprByShortHash(winner)
			response(w, Result{OPR: &winningOPR})

		default:
			errorResponse(w, Error{Code: 1, Reason: "Method Not Found"})
	}
}

func getCurrentOPRs(w http.ResponseWriter){
	height := leaderHeight()
	oprs := oprsByHeight(height)
	response(w, Result{OPRBlock: oprs})
}
  
// getOprByHash handler to return the opr by full hash
func getOprByHash(w http.ResponseWriter, params Parameters) {
	if params.Hash != "" {
		opr := oprByHash(params.Hash)
		response(w, Result{OPR: &opr})
	} else {
		invalidParameterError(w, params)
	}
}

// getOprByHash handler to return the opr by full hash
func getOprByEntryHash(w http.ResponseWriter, params Parameters) {
	if params.Hash != "" {
		opr := oprByHash(params.Hash)
		response(w, Result{OPR: &opr})
	} else {
		invalidParameterError(w, params)
	}
}

// getOprByShortHash handler to return the opr by the short 8 byte hash
func getOprByShortHash(w http.ResponseWriter, params Parameters) {
	if params.Hash != "" {
		opr := oprByShortHash(params.Hash)
		response(w, Result{OPR: &opr})
	} else {
		invalidParameterError(w, params)
	}
}

// getOprsByDigitalID handler to return all oprs based on Digital ID
func getOprsByDigitalID(w http.ResponseWriter, params Parameters) {
	if params.DigitalID != "" {
		oprs := oprsByDigitalID(params.DigitalID)
		response(w, Result{OPRs: oprs})
	} else {
		invalidParameterError(w, params)
	}
}

// getOPRsByHeight handler will grab OPRs for a height except current block
func getOPRsByHeight(w http.ResponseWriter, params Parameters) {
	if params.Height !=nil {
		oprblock := oprsByHeight(*params.Height)
		if oprblock != nil {
			response(w, Result{OPRBlock: oprblock})
		}
		errorResponse(w, Error{Code: 4, Reason: "No OPRs found"})
	} else {
		invalidParameterError(w, params)
	}
}

// getBalance handler will get the balance of a pegnet address
func getBalance(w http.ResponseWriter, params Parameters) {
	if params.Address != nil {
		balance := opr.GetBalance(*params.Address)
		res := Result{Balance: balance}
		response(w, res)
	} else {
		invalidParameterError(w, params)
	}
}

// getWinners returns the current 10 winners entry shorthashes from the last recorded block
func getWinners() [10]string {
	var winners [10]string
	height := leaderHeight()
	currentOPRS := oprsByHeight(height - 1)
	if currentOPRS != nil {
		opr := currentOPRS.OPRs[0]
		return opr.WinPreviousOPR
	}
	return winners
}

// getWinner returns the highest graded entry shorthash from the last recorded block
func getWinner() string {
	return getWinners()[0]
}

// response is a wrapper around all responses to be served
func response(w http.ResponseWriter, res Result){
	json.NewEncoder(w).Encode(PostResponse{Res: res})
}

// getLeaderHeight helper function, cleaner than using the factom monitor
func leaderHeight() int64 {
	heights, err := factom.GetHeights()
	if err != nil {
		return 0
	}
	return heights.LeaderHeight
}

// OprsByHeight returns a single OPRBlock
func oprsByHeight(dbht int64) *opr.OprBlock {
	for _, opr := range opr.OPRBlocks {
		if opr.Dbht == dbht {
		return opr
		}
	}
	return nil
}

// OprsByDigitalID returns every OPR created by a given ID
// Multiple ID's per miner or single daemon are possible.
// This function searches through every possible ID and returns all.
func oprsByDigitalID(did string) []opr.OraclePriceRecord {
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
func oprByHash(hash string) opr.OraclePriceRecord {
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

// OprByHash returns the entire OPR based on it's Factom Entry hash 
func oprByEntryHash(hash string) opr.OraclePriceRecord {
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
func oprByShortHash(shorthash string) opr.OraclePriceRecord {
	hashbytes, _  := hex.DecodeString(shorthash)
	// hashbytes = reverseBytes(hashbytes)
	for _, block := range opr.OPRBlocks {
		for _, opr := range block.OPRs{
			if bytes.Compare(hashbytes, opr.Entry.Hash()[:8]) ==  0 {
			return *opr
			}
		}    
	}
	return opr.OraclePriceRecord{}
}