package opr_test

import (
	"bytes"
	"encoding/hex"
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

	errs := verifyAssetStringOrder(badOrder, common.AssetsV1)
	if len(errs) != 1 {
		t.Errorf("Expected 1 err, found %d", len(errs))
	}

	badOrder = `{"PEG":0, "GBP":0, "DCR":0, "USD":0,"EUR":0,"JPY":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"TWD":0,"KRW":0,"ARS":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZEC":0}`

	errs = verifyAssetStringOrder(badOrder, common.AssetsV1)
	if len(errs) != 2 {
		t.Errorf("Expected 2 err, found %d", len(errs))
	}
}

func TestAssetListJSONMarshal(t *testing.T) {
	a := make(opr.OraclePriceRecordAssetList)
	// Add them in reverse order
	for i := len(common.AssetsV2) - 1; i >= 0; i-- {
		asset := common.AssetsV2[i]
		a[asset] = 0
	}

	a["version"] = 2
	data, err := json.Marshal(a)
	if err != nil {
		t.Error(err)
	}

	errs := verifyAssetStringOrder(string(data), common.AssetsV2)
	for _, err := range errs {
		t.Error(err)
	}

	if !a.ContainsExactly(common.AssetsV2) {
		t.Error("Missing items in the set")
	}

	// Test adding a new one
	a["random"] = 0
	if a.ContainsExactly(common.AssetsV2) {
		t.Error("Should fail but did not")
	}

	// Test missing one
	delete(a, "random")
	delete(a, "PEG")
	if a.ContainsExactly(common.AssetsV2) {
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

	if !m.Contains(common.AssetsV2) {
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

	if !m.Contains(common.AssetsV1) {
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

// TestSingleMarshal will ensure the marshalling can go both ways
func TestSingleMarshal(t *testing.T) {
	var err error
	o := opr.NewOraclePriceRecord()
	for _, asset := range common.AssetsV2 {
		o.Assets.SetValue(asset, rand.Float64())
	}

	common.SetTestingVersion(2)
	//c := config.NewConfig([]config.Provider{common.NewUnitTestConfigProvider()})
	o.CoinbaseAddress = common.ConvertRawToFCT(common.RandomByteSliceOfLen(32))
	o.FactomDigitalID = "random"
	o.Network = common.UnitTestNetwork
	o.Version = common.OPRVersion(o.Network, int64(o.Dbht))
	o.WinPreviousOPR = []string{}

	for i, asset := range common.AssetsV2 {
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

// testOPRMarshaling will ensure the marshalling can go both ways
func TestOPRMarshaling(t *testing.T) {
	t.Run("version 1", func(t *testing.T) {
		testOPRMarshaling(t, 1)
	})
	t.Run("version 2", func(t *testing.T) {
		testOPRMarshaling(t, 2)
	})
	t.Run("version 3", func(t *testing.T) {
		testOPRMarshaling(t, 3)
	})
	t.Run("version 4", func(t *testing.T) {
		testOPRMarshaling(t, 4)
	})
	t.Run("version 5", func(t *testing.T) {
		testOPRMarshaling(t, 5)
	})
}

func testOPRMarshaling(t *testing.T, version uint8) {
	common.SetTestingVersion(version)

	// Test randoms
	for i := 0; i < 10; i++ {
		o := RandomOPROfVersion(version)
		o.Network = common.UnitTestNetwork
		o.Version = common.OPRVersion(o.Network, int64(o.Dbht))

		testMarshal(o, t)
	}

	// Test with winners
	for i := 0; i < 10; i++ {
		o := RandomOPROfVersion(version)
		o.Network = common.UnitTestNetwork
		o.Version = common.OPRVersion(o.Network, int64(o.Dbht))
		for i := range o.WinPreviousOPR {
			win := make([]byte, 8)
			rand.Read(win)
			o.WinPreviousOPR[i] = hex.EncodeToString(win)
		}

		testMarshal(o, t)
	}

}

func testMarshal(o *opr.OraclePriceRecord, t *testing.T) {
	data, err := o.SafeMarshal()
	if err != nil {
		t.Error(err)
	}

	o2 := opr.NewOraclePriceRecord()
	// These not set by json
	o2.Network = o.Network
	o2.Version = o.Version
	o2.SelfReportedDifficulty = o.SelfReportedDifficulty
	o2.Nonce = o.Nonce
	o2.Difficulty = o.Difficulty
	o2.EntryHash = o.EntryHash
	o2.OPRHash = o.OPRHash

	err = o2.SafeUnmarshal(data)
	if err != nil {
		t.Error(err)
	}

	data2, err := o2.SafeMarshal()
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(data, data2) != 0 {
		t.Error("marshaled data different after re-marshal")
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
