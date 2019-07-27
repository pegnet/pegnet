// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package testutils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

// GetClientWithFixedResp will return a client that no matter what the request is,
// will always respond with the 'resp' as the body.
// From http://hassansin.github.io/Unit-Testing-http-client-in-Go
func GetClientWithFixedResp(resp []byte) *http.Client {
	client := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBuffer(resp)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	return client
}

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

// NewHTTPServerWithFixedResp will only return the fixed reponse, always
func NewHTTPServerWithFixedResp(port int, resp []byte) *http.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write(resp)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	srv := http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	return &srv
}
