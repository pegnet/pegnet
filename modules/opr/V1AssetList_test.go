package opr_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	. "github.com/pegnet/pegnet/modules/opr"
)

// The correct marshal format
// {"PEG":0, "USD":0,"EUR":0,"JPY":0,"GBP":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"TWD":0,"KRW":0,"ARS":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZEC":0,"DCR":0}

func TestVerifyFunction(t *testing.T) {
	badOrder := `{"PNT":0, "DCR":0, "USD":0,"EUR":0,"JPY":0,"GBP":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"TWD":0,"KRW":0,"ARS":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZEC":0}`

	errs := verifyAssetStringOrder(badOrder, V1Assets)
	if len(errs) != 1 {
		t.Errorf("Expected 1 err, found %d", len(errs))
	}

	badOrder = `{"PNT":0, "GBP":0, "DCR":0, "USD":0,"EUR":0,"JPY":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"TWD":0,"KRW":0,"ARS":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZEC":0}`

	errs = verifyAssetStringOrder(badOrder, V1Assets)
	if len(errs) != 2 {
		t.Errorf("Expected 2 err, found %d", len(errs))
	}
}

// TestAssetListUnmarshal verifies not only does the v1 unmarshal from json correctly, but also the
// resulting json that is marshaled is marshaled in the correct order.
func TestAssetListUnmarshal(t *testing.T) {
	j := `{"PNT":0,"USD":0,"EUR":0,"JPY":0,"GBP":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"KRW":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZEC":0,"DCR":0}`

	m := make(V1AssetList)
	err := json.Unmarshal([]byte(j), &m)
	if err != nil {
		t.Error(err)
	}

	for _, asset := range V1Assets {
		if _, ok := m[asset]; !ok {
			t.Errorf("Missing asset %s in unmarshal", asset)
		}
	}

	data, err := m.MarshalJSON()
	if err != nil {
		t.Error(err)
	}

	if string(data) != j {
		t.Error("Marshal is different than unmarshaled original")
		fmt.Println(string(data), "\n", j)
	}

	if errs := verifyAssetStringOrder(string(data), V1Assets); len(errs) != 0 {
		t.Error("marshalled order is wrong")
	}
}

// verifyAssetStringOrder will verify the resulting string has the assets in the same order as the global order
func verifyAssetStringOrder(str string, list []string) []error {
	index := 0
	errs := []error{}
	for _, asset := range list {
		i := strings.Index(str, asset)
		if i < index {
			errs = append(errs, fmt.Errorf("asset %s in the wrong order", asset))
		}
		index = i
	}
	return errs
}
