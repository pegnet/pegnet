// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package api

import (
	"net/http"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
)

// RequestHandler as the base handler
type RequestHandler struct{}

// PostRequest struct to deserialise from request body
type PostRequest struct {
	Method string     `json:"method"`
	Params Parameters `json:"params"`
}

// Parameters contains all possible json inputs
type Parameters struct {
	Address   *string `json:"address,omitempty"`
	Height    *int64  `json:"height,omitempty"`
	DigitalID string  `json:"digital-id,omitempty"`
	Hash      string  `json:"hash,omitempty"`
}

// PostResponse to either contain a valid result or error
type PostResponse struct {
	Res Result `json:"result"`
	Err Error  `json:"error"`
}

// Result struct contains all potential json api responses
type Result struct {
	Balance      int64                         `json:"balance,omitempty"`
	Balances     map[string]map[[32]byte]int64 `json:"balances,omitempty"`
	ChainID      string                        `json:"chainid,omitempty"`
	LeaderHeight int64                         `json:"leaderheight,omitempty"`
	OPRBlocks    []*opr.OprBlock               `json:"oprblocks,omitempty"`
	Winners      []string                      `json:"winners,omitempty"`
	Winner       string                        `json:"winner,omitempty"`
	OPRBlock     *opr.OprBlock                 `json:"oprblock,omitempty"`
	OPRs         []opr.OraclePriceRecord       `json:"oprs,omitempty"`
	OPR          *opr.OraclePriceRecord        `json:"opr,omitempty"`
}

// Base handler of all requests
func (h RequestHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	log.WithFields(log.Fields{
		"IP":             req.RemoteAddr,
		"URL":						req.RequestURI,
		"Request Method": req.Method}).Debug("Server Request")
	if req.Method == "POST" {
		apiHandler(writer, req)
	} else {
		methodNotAllowed(writer)
	}
}
