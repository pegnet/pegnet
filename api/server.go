// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pegnet/pegnet/mining"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
)

// APIServer as the base handler
type APIServer struct {
	Statistics *mining.GlobalStatTracker
	Server     *http.Server
}

func NewApiServer() *APIServer {
	s := new(APIServer)
	s.Server = &http.Server{}
	s.Server.Handler = s

	return s
}

func (s *APIServer) Listen(port int) {
	s.Server.Addr = fmt.Sprintf(":%d", port)
	err := s.Server.ListenAndServe()
	if err != nil {
		log.WithError(err).Fatal("api server stopped")
	}
}

// Base handler of all requests
func (h *APIServer) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	log.WithFields(log.Fields{
		"IP":             req.RemoteAddr,
		"Request Method": req.Method}).Info("Server Request")
	if req.Method == "POST" {
		h.apiHandler(writer, req)
	} else {
		methodNotAllowed(writer)
	}
}

func (h *APIServer) apiHandler(w http.ResponseWriter, r *http.Request) {
	var request PostRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		Respond(w, PostResponse{Err: NewJSONDecodingError()})
		return
	}
	log.WithFields(log.Fields{
		"API Method": request.Method,
		"Params":     request.Params}).Info("API Request")

	// enable cors
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var result interface{}
	var apiError *Error
	switch request.Method {
	case "performance":
		result, apiError = getPerformance(request.Params)

	case "all-oprs":
		result = &GenericResult{OPRBlocks: opr.OPRBlocks}

	case "balances":
		result = &GenericResult{Balances: opr.Balances}

	case "balance":
		result, apiError = getBalance(request.Params)

	case "chainid":
		result = &GenericResult{ChainID: opr.OPRChainID}

	case "current-oprs":
		result, apiError = getCurrentOPRs()

	case "leaderheight":
		result = &GenericResult{LeaderHeight: getLeaderHeight()}

	case "oprs-by-height":
		result, apiError = getOPRsByHeight(request.Params)

	case "oprs-by-id":
		result, apiError = getOprsByDigitalID(request.Params)

	case "opr-by-hash":
		result, apiError = getOprByHash(request.Params)

	case "opr-by-shorthash":
		result, apiError = getOprByShortHash(request.Params)

	case "winners":
		winners := getWinners()
		result = &GenericResult{Winners: winners[:]}

	case "winner":
		winner := getWinner()
		result = &GenericResult{Winner: winner}

	// Failing method - shorthash needs to be fixed
	case "winning-opr":
		winner := getWinner()
		winningOPR := oprByShortHash(winner)
		result = &GenericResult{OPR: &winningOPR}

	default:
		apiError = NewMethodNotFoundError()
	}

	var response PostResponse
	if apiError != nil {
		response = PostResponse{Err: apiError}
	} else {
		response = PostResponse{Res: result}
	}
	Respond(w, response)
}

func Respond(w http.ResponseWriter, response PostResponse) {
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.WithError(err).Error("Failed to write response JSON")
		// Potential infinite recursion, but the error message should always encode properly:
		Respond(w, PostResponse{Err: NewInternalError()})
	}
}
