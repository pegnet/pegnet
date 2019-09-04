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
			// Must be set to non-nil Value or it panics
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

// PriceCheck checks if the price is "reasonable" to see if we inverted the prices
func PriceCheck(asset string, rate float64) error {
	switch asset {
	case "XBT":
		// BTC < $500? That sounds wrong
		if rate < 500 {
			return fmt.Errorf("bitcoin(%s) found to be %.2f, less than $500, this seems wrong", asset, rate)
		}
	case "XAU":
		// Gold < $50? That sounds wrong
		if rate < 50 {
			return fmt.Errorf("gold(%s) found to be %.2f, less than $50, this seems wrong", asset, rate)
		}
	case "XPD":
		// Silver < $5? That sounds wrong
		if rate < 5 {
			return fmt.Errorf("%s found to be %.2f, less than $5, this seems wrong", asset, rate)
		}
	case "MXN":
		if rate > 1 {
			return fmt.Errorf("the peso(%s) found to be %.2f, greater than $1, this seems wrong", asset, rate)
		}
	case "ETH":
		if rate < 20 {
			return fmt.Errorf("%s found to be %.2f, less than $20, this seems wrong", asset, rate)
		}
	}
	return nil
}
