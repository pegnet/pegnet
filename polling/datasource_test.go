package polling_test

import (
	"fmt"
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

	pegs, err := s.FetchPegPrices()
	if err != nil {
		t.Error(err)
	}

	for _, asset := range s.SupportedPegs() {
		_, ok := pegs[asset]
		if !ok {
			t.Errorf("Missing %s", asset)
		}
	}
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

	pegs, err := s.FetchPegPrices()
	if err != nil {
		t.Error(err)
	}

	for _, asset := range s.SupportedPegs() {
		r, ok := pegs[asset]
		if !ok {
			t.Errorf("Missing %s", asset)
		}

		err := PriceCheck(asset, r.Value)
		if err != nil {
			t.Error(err)
		}
	}
}

// PriceCheck checks if the price is "reasonable" to see if we inverted the prices
func PriceCheck(asset string, rate float64) error {
	switch asset {
	case "XBT":
		if rate < 1 {
			return fmt.Errorf("bitcoin(%s) found to be %.2f, less than $1, this seems wrong", asset, rate)
		}
	case "XAU":
		if rate < 1 {
			return fmt.Errorf("gold(%s) found to be %.2f, less than $1, this seems wrong", asset, rate)
		}
	case "MXN":
		if rate > 1 {
			return fmt.Errorf("the peso(%s) found to be %.2f, greater than $1, this seems wrong", asset, rate)
		}
	}
	return nil
}
