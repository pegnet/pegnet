// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package api

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Error struct returns a code and it's associated message
// 0: <Success>
// 1: Method Not Found
// 2: Parameter Not Found
// 3: Error Decoding JSON
// 4: Internal Error
type Error struct {
	Code   int         `json:"code"`
	Reason string      `json:"reason"`
	Data   interface{} `json:"data"`
}

// errorResponse is a wrapper around all errors to be served
func errorResponse(w http.ResponseWriter, err Error) {
	json.NewEncoder(w).Encode(PostResponse{Err: &err})
}

// methodNotAllowed returns a 405 status when an invalid HTTP request methid is used
func methodNotAllowed(w http.ResponseWriter) {
	log.Error("Invalid HTTP Request Method")
	http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
}

// NewMethodNotFoundError returns when the specified RPC method is not found
func NewMethodNotFoundError() *Error {
	return &Error{Code: 1, Reason: "Method Not Found"}
}

// NewInvalidParameterError returns when the method is valid but the parameter is not
func NewInvalidParametersError() *Error {
	return &Error{Code: 2, Reason: "Invalid parameters"}
}

// NewJSONDecodingError returns when the request body is unable to be parsed
func NewJSONDecodingError() *Error {
	return &Error{Code: 3, Reason: "Unable to parse JSON body"}
}

// NewInternalError returns when a bug creeps in
func NewInternalError() *Error {
	return &Error{Code: 4, Reason: "Internal error"}
}
