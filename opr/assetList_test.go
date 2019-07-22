package opr_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr"
)

// The correct marshal format
// {"PNT":0, "USD":0,"EUR":0,"JPY":0,"GBP":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"TWD":0,"KRW":0,"ARS":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZCASH":0,"DCR":0}

func TestVerifyFunction(t *testing.T) {
	badOrder := `{"PNT":0, "DCR":0, "USD":0,"EUR":0,"JPY":0,"GBP":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"TWD":0,"KRW":0,"ARS":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZCASH":0}`

	errs := verifyAssetStringOrder(badOrder)
	if len(errs) != 1 {
		t.Errorf("Expected 1 err, found %d", len(errs))
	}

	badOrder = `{"PNT":0, "GBP":0, "DCR":0, "USD":0,"EUR":0,"JPY":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"TWD":0,"KRW":0,"ARS":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZCASH":0}`

	errs = verifyAssetStringOrder(badOrder)
	if len(errs) != 2 {
		t.Errorf("Expected 2 err, found %d", len(errs))
	}
}

func TestAssetListJSONMarshal(t *testing.T) {
	a := make(opr.OraclePriceRecordAssetList)
	for _, asset := range common.AllAssets {
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

	if !a.Contains(common.AllAssets) {
		t.Error("Missing items in the set")
	}
}

func TestAssetListUnmarshal(t *testing.T) {
	j := `{"PNT":0,"USD":0,"EUR":0,"JPY":0,"GBP":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"KRW":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZCASH":0,"DCR":0}`

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
