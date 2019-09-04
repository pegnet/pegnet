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
func FixedDataSourceTest(t *testing.T, source string, fixed []byte) {
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

	testDataSource(t, s)
}

// ActualDataSourceTest actually fetches the resp over the internet
func ActualDataSourceTest(t *testing.T, source string) {
	defer func() { http.DefaultClient = &http.Client{} }() // Don't leave http broken

	c := config.NewConfig([]config.Provider{common.NewUnitTestConfigProvider()})
	http.DefaultClient = &http.Client{}

	s, err := polling.NewDataSource(source, c)
	if err != nil {
		t.Error(err)
	}

	testDataSource(t, s)
}

func testDataSource(t *testing.T, s polling.IDataSource) {
	pegs, err := s.FetchPegPrices()
	if err != nil {
		t.Error(err)
	}

	for _, asset := range s.SupportedPegs() {
		r, ok := pegs[asset]
		if !ok {
			t.Errorf("Missing %s", asset)
		}

		err := testutils.PriceCheck(asset, r.Value)
		if err != nil {
			t.Error(err)
		}
	}
}
