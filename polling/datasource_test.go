package polling_test

import (
	"net/http"
	"testing"

	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/polling"
	"github.com/pegnet/pegnet/testutils"
	"github.com/zpatrick/go-config"
)

// FixedDataSourceTest will test the parsing of the data source using the fixed response
func FixedDataSourceTest(t *testing.T, source string, fixed []byte, exceptions ...string) {
	defer func() { http.DefaultClient = &http.Client{} }() // Don't leave http broken

	c := config.NewConfig([]config.Provider{common.NewUnitTestConfigProvider()})

	// Set default http client to return what we expect from apilayer
	cl := testutils.GetClientWithFixedResp(fixed)
	http.DefaultClient = cl
	polling.NewHTTPClient = func() *http.Client {
		return testutils.GetClientWithFixedResp(fixed)
	}

	s, err := polling.NewDataSource(source, c)
	if err != nil {
		t.Error(err)
	}

	testDataSource(t, s, exceptions...)
}

// ActualDataSourceTest actually fetches the resp over the internet
func ActualDataSourceTest(t *testing.T, source string, exceptions ...string) {
	defer func() { http.DefaultClient = &http.Client{} }() // Don't leave http broken

	polling.NewHTTPClient = func() *http.Client {
		return &http.Client{}
	}

	c := config.NewConfig([]config.Provider{common.NewUnitTestConfigProvider()})
	http.DefaultClient = &http.Client{}

	s, err := polling.NewDataSource(source, c)
	if err != nil {
		t.Error(err)
	}

	testDataSource(t, s, exceptions...)
}

func testDataSource(t *testing.T, s polling.IDataSource, exceptions ...string) {
	pegs, err := s.FetchPegPrices()
	if err != nil {
		t.Error(err)
	}

	exceptionsMap := make(map[string]interface{})
	for _, e := range exceptions {
		exceptionsMap[e] = struct{}{}
	}

	for _, asset := range s.SupportedPegs() {
		r, ok := pegs[asset]
		if !ok {
			if _, except := exceptionsMap[asset]; except {
				continue // This asset is an exception from the check
			}
			t.Errorf("Missing %s", asset)
		}

		err := testutils.PriceCheck(asset, r.Value)
		if err != nil {
			t.Error(err)
		}
	}
}
