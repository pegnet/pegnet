package opr_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
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

	errs := verifyAssetStringOrder(badOrder, common.AllAssets)
	if len(errs) != 1 {
		t.Errorf("Expected 1 err, found %d", len(errs))
	}

	badOrder = `{"PEG":0, "GBP":0, "DCR":0, "USD":0,"EUR":0,"JPY":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"TWD":0,"KRW":0,"ARS":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZEC":0}`

	errs = verifyAssetStringOrder(badOrder, common.AllAssets)
	if len(errs) != 2 {
		t.Errorf("Expected 2 err, found %d", len(errs))
	}
}

func TestAssetListJSONMarshal(t *testing.T) {
	a := make(opr.OraclePriceRecordAssetList)
	// Add them in reverse order
	for i := len(common.VersionTwoAssets) - 1; i >= 0; i-- {
		asset := common.VersionTwoAssets[i]
		a[asset] = 0
	}

	a["version"] = 2
	data, err := json.Marshal(a)
	if err != nil {
		t.Error(err)
	}

	errs := verifyAssetStringOrder(string(data), common.VersionTwoAssets)
	for _, err := range errs {
		t.Error(err)
	}

	if !a.ContainsExactly(common.VersionTwoAssets) {
		t.Error("Missing items in the set")
	}

	// Test adding a new one
	a["random"] = 0
	if a.ContainsExactly(common.VersionTwoAssets) {
		t.Error("Should fail but did not")
	}

	// Test missing one
	delete(a, "random")
	delete(a, "PEG")
	if a.ContainsExactly(common.VersionTwoAssets) {
		t.Error("Should fail but did not")
	}

}

func TestAssetListUnmarshal(t *testing.T) {
	j := `{"PEG":0,"USD":0,"EUR":0,"JPY":0,"GBP":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"KRW":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZEC":0,"DCR":0}`

	m := make(opr.OraclePriceRecordAssetList)
	err := json.Unmarshal([]byte(j), &m)
	if err != nil {
		t.Error(err)
	}

	if !m.Contains(common.VersionTwoAssets) {
		t.Error("Missing asset in unmarshal")
	}

	m["version"] = 2
	data, err := json.Marshal(m)
	if err != nil {
		t.Error(err)
	}

	if string(data) != j {
		t.Error("Marshal is different than unmarshaled original")
		fmt.Println(string(data), "\n", j)
	}
}

func TestAssetListVersionOneUnmarshal(t *testing.T) {
	// Safe unmarshal switches PNT for PEG
	j := `{"PEG":0,"USD":0,"EUR":0,"JPY":0,"GBP":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"KRW":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZEC":0,"DCR":0}`

	m := make(opr.OraclePriceRecordAssetList)
	err := json.Unmarshal([]byte(j), &m)
	if err != nil {
		t.Error(err)
	}

	if !m.Contains(common.VersionOneAssets) {
		t.Error("Missing asset in unmarshal")
	}

	if _, ok := m["PEG"]; ok {
		// Swap peg and pnt, as we expect pnt in the output
		j = strings.Replace(j, "PEG", "PNT", -1)
	} else {
		t.Errorf("PEG not found")
	}

	m["version"] = 1
	data, err := json.Marshal(m)
	if err != nil {
		t.Error(err)
	}

	if string(data) != j {
		t.Error("Marshal is different than unmarshaled original")
		fmt.Println(string(data), "\n", j)
	}
}

// TestOPRJsonMarshal will ensure the json marshalling can go both ways
func TestOPRJsonMarshal(t *testing.T) {
	var err error
	o := opr.NewOraclePriceRecord()
	for _, asset := range common.VersionTwoAssets {
		o.Assets.SetValue(asset, rand.Float64())
	}

	common.SetTestingVersion(2)
	//c := config.NewConfig([]config.Provider{common.NewUnitTestConfigProvider()})
	o.CoinbaseAddress = common.ConvertRawToFCT(common.RandomByteSliceOfLen(32))
	o.FactomDigitalID = "random"
	o.Network = common.UnitTestNetwork
	o.Version = common.OPRVersion(o.Network, int64(o.Dbht))

	for i, asset := range common.VersionTwoAssets {
		o.Assets.SetValue(asset, rand.Float64()*1000)

		// Test truncate does not affect json
		if i%3 == 0 {
			o.Assets.SetValue(asset, polling.TruncateTo4(o.Assets.Value(asset)))
		} else if i%3 == 1 {
			o.Assets.SetValue(asset, polling.TruncateTo8(o.Assets.Value(asset)))
		}
	}

	data, err := o.SafeMarshal()
	if err != nil {
		t.Error(err)
	}

	o2 := opr.NewOraclePriceRecord()
	// These two not set by json
	o2.Network = common.UnitTestNetwork
	o2.Version = common.OPRVersion(o.Network, int64(o.Dbht))
	err = o2.SafeUnmarshal(data)
	if err != nil {
		t.Error(err)
	}

	data2, err := o2.SafeMarshal()
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(data, data2) != 0 {
		o.Assets["version"] = 2
		o2.Assets["version"] = 2
		jData, err1 := json.Marshal(o)
		jData2, err2 := json.Marshal(o2)
		delete(o.Assets, "version")
		delete(o2.Assets, "version")
		fmt.Println(err1, err2)
		fmt.Println(" ", string(jData), "\n", string(jData2))
		fmt.Println(hex.EncodeToString(data), "\n", hex.EncodeToString(data2))
		t.Error("Json different after remarshal")
	}

	if !reflect.DeepEqual(o, o2) {
		t.Errorf("did not marshal into the same")
	}

	o2.Assets.SetValue("PEG", 0.123)
	// Ensure not just a deep equal oddity
	if reflect.DeepEqual(o, o2) {
		t.Errorf("I changed it, they should not be different")
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

func TestValueSetting(t *testing.T) {
	list := make(opr.OraclePriceRecordAssetList)

	t.Run("float only", func(t *testing.T) {
		for i := 0; i < 10000; i++ {
			v := polling.TruncateTo8(rand.Float64())
			list.SetValue("test", v)
			v2 := list.Value("test")
			// 1e-8 diff is truncation
			if v != v2 && math.Abs(v-v2) > float64(1/1e8) {
				t.Errorf("exp %.8f, got %.8f", v, v2)
			}
		}
	})

	t.Run("float -> uint64 -> float", func(t *testing.T) {
		for i := 0; i < 10000; i++ {
			v := polling.TruncateTo8(rand.Float64())
			// Set value from float
			list.SetValue("test", v)

			// Get the uint64 val
			uv := list.Uint64Value("test")
			// Set the value as uint64
			list.SetValueFromUint64("test", uv)

			// Try it as a float again
			v2 := list.Value("test")
			if v != v2 && math.Abs(v-v2) > float64(1/1e8) {
				t.Errorf("exp %.8f, got %.8f", v, v2)
			}

			// This test is kinda pointless. Just checking the float always has the same uint out
			list.SetValue("test", v)
			uv2 := list.Uint64Value("test")
			if uv != uv2 {
				t.Errorf("exp %d, got %d", uv, uv2)
			}
		}
	})

	t.Run("uint64 -> float -> uint64", func(t *testing.T) {
		for i := 0; i < 10000; i++ {
			v := rand.Uint64() / 1e8
			list.SetValueFromUint64("test", v)
			v2 := list.Uint64Value("test")
			if v != v2 {
				t.Errorf("exp %d, got %d", v, v2)
			}

			f := list.Value("test")
			list.SetValue("test", f)
			if v3 := list.Uint64Value("test"); v3 != v {
				t.Errorf("exp %d, got %d, diff %d", v3, v, v3-v)
			}
		}
	})
}
