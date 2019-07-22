package polling_test

import (
	"bytes"
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
