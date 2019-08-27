package opr_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr"
	"github.com/pegnet/pegnet/polling"
)

// The correct marshal format
// {"PEG":0, "USD":0,"EUR":0,"JPY":0,"GBP":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"TWD":0,"KRW":0,"ARS":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZEC":0,"DCR":0}

func TestVerifyFunction(t *testing.T) {
	badOrder := `{"PEG":0, "DCR":0, "USD":0,"EUR":0,"JPY":0,"GBP":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"TWD":0,"KRW":0,"ARS":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZEC":0}`

	errs := verifyAssetStringOrder(badOrder)
	if len(errs) != 1 {
		t.Errorf("Expected 1 err, found %d", len(errs))
	}

	badOrder = `{"PEG":0, "GBP":0, "DCR":0, "USD":0,"EUR":0,"JPY":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"TWD":0,"KRW":0,"ARS":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZEC":0}`

	errs = verifyAssetStringOrder(badOrder)
	if len(errs) != 2 {
		t.Errorf("Expected 2 err, found %d", len(errs))
	}
}

func TestAssetListJSONMarshal(t *testing.T) {
	a := make(opr.OraclePriceRecordAssetList)
	// Add them in reverse order
	for i := len(common.AllAssets) - 1; i >= 0; i-- {
		asset := common.AllAssets[i]
		a[asset] = 0
	}

	data, err := json.Marshal(a)
	if err != nil {
		t.Error(err)
	}

	errs := verifyAssetStringOrder(string(data))
	for _, err := range errs {
		t.Error(err)
	}

	if !a.ContainsExactly(common.AllAssets) {
		t.Error("Missing items in the set")
	}

	// Test adding a new one
	a["random"] = 0
	if a.ContainsExactly(common.AllAssets) {
		t.Error("Should fail but did not")
	}

	// Test missing one
	delete(a, "random")
	delete(a, "PEG")
	if a.ContainsExactly(common.AllAssets) {
		t.Error("Should fail but did not")
	}

}

func TestAssetListUnmarshal(t *testing.T) {
	j := `{"PEG":0,"USD":0,"EUR":0,"JPY":0,"GBP":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"KRW":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZEC":0,"DCR":0}`

	m := new(opr.OraclePriceRecordAssetList)
	err := json.Unmarshal([]byte(j), m)
	if err != nil {
		t.Error(err)
	}

	if !m.Contains(common.AllAssets) {
		t.Error("Missing asset in unmarshal")
	}

	data, err := m.MarshalJSON()
	if err != nil {
		t.Error(err)
	}

	if string(data) != j {
		t.Error("Marshal is different than unmarshaled original")
		fmt.Println(string(data))
	}
}

// TestOPRJsonMarshal will ensure the json marshalling can go both ways
func TestOPRJsonMarshal(t *testing.T) {
	var err error
	o := opr.NewOraclePriceRecord()
	for _, asset := range common.AllAssets {
		o.Assets[asset] = rand.Float64()
	}

	//c := config.NewConfig([]config.Provider{common.NewUnitTestConfigProvider()})
	o.CoinbaseAddress = common.ConvertRawToFCT(common.RandomByteSliceOfLen(32))
	o.FactomDigitalID = "random"
	o.Network = common.TestNetwork
	o.Version = common.OPRVersion(o.Network, int64(o.Dbht))

	for i, asset := range common.AllAssets {
		o.Assets[asset] = rand.Float64() * 1000

		// Test truncate does not affect json
		if i%3 == 0 {
			o.Assets[asset] = polling.TruncateTo4(o.Assets[asset])
		} else if i%3 == 1 {
			o.Assets[asset] = polling.TruncateTo8(o.Assets[asset])
		}
	}

	data, err := json.Marshal(o)
	if err != nil {
		t.Error(err)
	}

	o2 := opr.NewOraclePriceRecord()
	// These two not set by json
	o2.Network = common.TestNetwork
	o2.Version = common.OPRVersion(o.Network, int64(o.Dbht))
	err = json.Unmarshal(data, o2)
	if err != nil {
		t.Error(err)
	}

	data2, err := json.Marshal(o2)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(data, data2) != 0 {
		t.Error("Json different after remarshal")
	}

	if !reflect.DeepEqual(o, o2) {
		t.Errorf("did not marshal into the same")
	}

	o2.Assets["PEG"] = 0.123
	// Ensure not just a deep equal oddity
	if reflect.DeepEqual(o, o2) {
		t.Errorf("I changed it, they should not be different")
	}

}

// verifyAssetStringOrder will verify the resulting string has the assets in the same order as the global order
func verifyAssetStringOrder(str string) []error {
	index := 0
	errs := []error{}
	for _, asset := range common.AllAssets {
		i := strings.Index(str, asset)
		if i < index {
			errs = append(errs, fmt.Errorf("asset %s in the wrong order", asset))
		}
		index = i
	}
	return errs
}
