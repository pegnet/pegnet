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

// lessChecks checks if an asset is less than the given map value. If it is
// found to be less than the map value, then the token price is likely incorrect.
var lessChecks = map[string]float64{
	"XBT":  1000, // Bitcoin
	"XAU":  500,  // Gold
	"XAG":  5,    // Silver
	"ETH":  20,
	"LINK": 0.1,
	"USD":  0.99,
	"DCR":  1,
	"BAT":  0.01,
	"FCT":  0.10,
	"TRY":  0.01,
	"PEG":  0.0000001,
	"pUSD": 0.15,
}

// greaterChecks checks if an asset is greater than the given map value. If it is
// found to be greater than the map value, then the token price is likely incorrect.
var greaterChecks = map[string]float64{
	"XBT":  50000, // Bitcoin
	"XAU":  5000,  // Gold
	"XAG":  100,   // Silver
	"ETH":  1000,
	"LINK": 100,
	"USD":  1.01,
	"MXN":  0.30,
	"INR":  0.30,
	"JPY":  0.30,
	"DCR":  500,
	"BAT":  2,
	"FCT":  25,
	"TRY":  0.90,
	"PEG":  0.10,
	"pUSD": 2,
}

// PriceCheck checks if the price is "reasonable" to see if we inverted the prices
func PriceCheck(asset string, rate float64) error {
	if lt, ok := lessChecks[asset]; ok {
		if rate < lt {
			return fmt.Errorf("%s found to be $%.2f, less than $%.2f, this seems wrong", asset, rate, lt)
		}
	}

	if gt, ok := greaterChecks[asset]; ok {
		if rate > gt {
			return fmt.Errorf("%s found to be $%.2f, greater than $%.2f, this seems wrong", asset, rate, gt)
		}
	}

	return nil
}
