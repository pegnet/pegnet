// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package api

import (
  "net/http"
  "encoding/json"
  log "github.com/sirupsen/logrus"
)

// Error struct returns a code and it's associated message
// 0: <Success>
// 1: Method Not Found
// 2: Parameter Not Found
// 3: Error Decoding JSON
type Error struct {
  Code    int                   `json:"code"`
  Reason  string                `json:"reason"`
}

// errorResponse is a wrapper around all errors to be served
func errorResponse(w http.ResponseWriter, err Error) {
  json.NewEncoder(w).Encode(PostResponse{Err: err})
}

// methodNotAllowed returns a 405 status when an invalid HTTP request methid is used
func methodNotAllowed(w http.ResponseWriter) {
  log.Error("Invalid HTTP Request Method")
  http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
}
  
// invalidParameterError returns when the method is valid but the parameter is not
func invalidParameterError(w http.ResponseWriter, params Parameters) {
  log.WithFields(log.Fields{"Params": params}).Error("Post Parameters Error")
  errorResponse(w,Error{Code: 2, Reason: "Parameter Not Found"})
}
  
  
// jsonDecodingError returns when the request body is unable to be parsed
func jsonDecodingError(w http.ResponseWriter) {
  log.Error("Error Decoding JSON request")
  errorResponse(w,Error{Code: 3, Reason: "Unable to parse JSON body"})
}