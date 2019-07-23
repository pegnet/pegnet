// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
)

func apiHandler(w http.ResponseWriter, r *http.Request) {
	var request PostRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		jsonDecodingError(w)
		return
	}
	log.WithFields(log.Fields{
		"API Method": request.Method,
		"Params":     request.Params}).Info("API Request")

	// enable cors
	w.Header().Set("Access-Control-Allow-Origin", "*")

	switch request.Method {
	case "all-oprs":
		response(w, Result{OPRBlocks: opr.OPRBlocks})

	case "balances":
		response(w, Result{Balances: opr.Balances})

	case "balance":
		getBalance(w, request.Params)

	case "chainid":
		response(w, Result{ChainID: opr.OPRChainID})

	case "current-oprs":
		getCurrentOPRs(w)

	case "leaderheight":
		response(w, Result{LeaderHeight: leaderHeight()})

	case "oprs-by-height":
		getOPRsByHeight(w, request.Params)

	case "oprs-by-id":
		getOprsByDigitalID(w, request.Params)

	case "opr-by-hash":
		getOprByHash(w, request.Params)

	case "opr-by-shorthash":
		getOprByShortHash(w, request.Params)

	case "winners":
		winners := getWinners()
		response(w, Result{Winners: winners[:]})

	case "winner":
		winner := getWinner()
		response(w, Result{Winner: winner})

	// Failing method - shorthash needs to be fixed
	case "winning-opr":
		winner := getWinner()
		winningOPR := oprByShortHash(winner)
		response(w, Result{OPR: &winningOPR})

	default:
		errorResponse(w, Error{Code: 1, Reason: "Method Not Found"})
	}
}

func getCurrentOPRs(w http.ResponseWriter) {
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

// getOPRsByHeight handler will return all OPR's at any height except the current block.
// Will only return local OPR's for the current chainhead.
func getOPRsByHeight(w http.ResponseWriter, params Parameters) {
	if params.Height != nil {
		oprblock := oprsByHeight(*params.Height)
		response(w, Result{OPRBlock: oprblock})
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
	height := leaderHeight()
	currentOPRS := oprsByHeight(height)
	opr := currentOPRS.OPRs[0]
	return opr.WinPreviousOPR
}

// getWinner returns the highest graded entry shorthash from the last recorded block
func getWinner() string {
	return getWinners()[0]
}

// response is a wrapper around all responses to be served
func response(w http.ResponseWriter, res Result) {
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
		for _, opr := range block.OPRs {
			for _, digitalID := range opr.FactomDigitalID {
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
	for _, block := range opr.OPRBlocks {
		for _, opr := range block.OPRs {
			if hash == hex.EncodeToString(opr.OPRHash) {
				return *opr
			}
		}
	}
	return opr.OraclePriceRecord{}
}

// Failing tests. Need to grok how the short 8 byte winning oprhashes are done.
func oprByShortHash(shorthash string) opr.OraclePriceRecord {
	hashbytes, _ := hex.DecodeString(shorthash)
	// hashbytes = reverseBytes(hashbytes)
	for _, block := range opr.OPRBlocks {
		for _, opr := range block.OPRs {
			if bytes.Compare(hashbytes, opr.OPRHash[:8]) == 0 {
				return *opr
			}
		}
	}
	return opr.OraclePriceRecord{}
}
