// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pegnet/pegnet/balances"
	"github.com/pegnet/pegnet/mining"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

// APIServer as the base handler
type APIServer struct {
	Statistics *mining.GlobalStatTracker
	Server     *http.Server
	Grader     *opr.QuickGrader
	Balances   *balances.BalanceTracker
	Mux        *http.ServeMux
	config     *config.Config
}

func NewApiServer(grader *opr.QuickGrader, balances *balances.BalanceTracker, config *config.Config) *APIServer {
	s := new(APIServer)
	s.Server = &http.Server{}
	mux := http.NewServeMux()
	mux.Handle("/v1", s)
	s.Server.Handler = corsHeader(mux)
	s.Mux = mux
	s.Grader = grader
	s.Balances = balances
	s.config = config

	return s
}

func corsHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Our middleware logic goes here...
		next.ServeHTTP(w, r)
	})
}

func (s *APIServer) Listen(port int) {
	log.Infof("Launching api on port :%d", port)
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
		result, apiError = h.getPerformance(request.Params)

	case "all-oprs":
		// TODO: This is not thread safe. This call could be exceedingly large too
		// 		I think it should be tossed
		result = &GenericResult{OPRBlocks: h.Grader.GetBlocks()}
	case "balance":
		result, apiError = h.getBalance(request.Params)

	case "chainid":
		result = &GenericResult{ChainID: opr.OPRChainID}

	case "current-oprs":
		result, apiError = h.getCurrentOPRs()

	case "leaderheight":
		result = &GenericResult{LeaderHeight: getLeaderHeight()}

	case "oprs-by-height":
		result, apiError = h.getOPRsByHeight(request.Params)

	case "oprs-by-id":
		result, apiError = h.getOprsByDigitalID(request.Params)

	case "opr-by-hash":
		result, apiError = h.getOprByHash(request.Params)

	case "opr-by-shorthash":
		result, apiError = h.getOprByShortHash(request.Params)

	case "winners":
		winners := h.getWinners()
		result = &GenericResult{Winners: winners[:]}

	case "winner":
		winner := h.getWinner()
		result = &GenericResult{Winner: winner}

	// Failing method - shorthash needs to be fixed
	case "winning-opr":
		winner := h.getWinner()
		winningOPR := h.Grader.OprByShortHash(winner)
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
