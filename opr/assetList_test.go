package opr_test

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/pegnet/pegnet/common"
	. "github.com/pegnet/pegnet/opr"
	"github.com/pegnet/pegnet/polling"
	"github.com/zpatrick/go-config"
)

// The correct marshal format
// {"PNT":0, "USD":0,"EUR":0,"JPY":0,"GBP":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"TWD":0,"KRW":0,"ARS":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZEC":0,"DCR":0}

func TestVerifyFunction(t *testing.T) {
	badOrder := `{"PNT":0, "DCR":0, "USD":0,"EUR":0,"JPY":0,"GBP":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"TWD":0,"KRW":0,"ARS":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZEC":0}`

	errs := verifyAssetStringOrder(badOrder)
	if len(errs) != 1 {
		t.Errorf("Expected 1 err, found %d", len(errs))
	}

	badOrder = `{"PNT":0, "GBP":0, "DCR":0, "USD":0,"EUR":0,"JPY":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"TWD":0,"KRW":0,"ARS":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZEC":0}`

	errs = verifyAssetStringOrder(badOrder)
	if len(errs) != 2 {
		t.Errorf("Expected 2 err, found %d", len(errs))
	}
}

func TestAssetListJSONMarshal(t *testing.T) {
	a := make(OraclePriceRecordAssetList)
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
	delete(a, "PNT")
	if a.ContainsExactly(common.AllAssets) {
		t.Error("Should fail but did not")
	}

}

func TestAssetListUnmarshal(t *testing.T) {
	j := `{"PNT":0,"USD":0,"EUR":0,"JPY":0,"GBP":0,"CAD":0,"CHF":0,"INR":0,"SGD":0,"CNY":0,"HKD":0,"KRW":0,"BRL":0,"PHP":0,"MXN":0,"XAU":0,"XAG":0,"XPD":0,"XPT":0,"XBT":0,"ETH":0,"LTC":0,"RVN":0,"XBC":0,"FCT":0,"BNB":0,"XLM":0,"ADA":0,"XMR":0,"DASH":0,"ZEC":0,"DCR":0}`

	m := new(OraclePriceRecordAssetList)
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
	o := NewOraclePriceRecord()
	for _, asset := range common.AllAssets {
		o.Assets[asset] = rand.Float64()
	}

	c := config.NewConfig([]config.Provider{common.NewUnitTestConfigProvider()})
	o.CoinbaseAddress = common.ConvertRawToFCT(common.RandomByteSliceOfLen(32))

	for _, asset := range common.AllAssets {
		o.Assets[asset] = rand.Float64() * 1000
	}

	if !o.Validate(c, int64(o.Dbht)) {
		t.Error("Should be valid")
	}

	data, err := json.Marshal(o)
	if err != nil {
		t.Error(err)
	}

	o2 := NewOraclePriceRecord()
	err = json.Unmarshal(data, o2)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(o, o2) {
		t.Errorf("did not marshal into the same")
	}

	o2.Assets["PNT"] = 0.123
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

func TestPriceConversions(t *testing.T) {
	// Simple round numbers
	t.Run("simple numbers", func(t *testing.T) {
		type Conv struct {
			Amt int64
			Res int64
		}
		type Vector struct {
			FCTRate float64
			BTCRate float64

			FctToBTC Conv
			BTCToFCT Conv
		}

		vectors := []Vector{
			{FCTRate: 100, BTCRate: 10000, FctToBTC: Conv{100, 1}, BTCToFCT: Conv{1, 100}},
			{FCTRate: 5, BTCRate: 10000, FctToBTC: Conv{250, 0}, BTCToFCT: Conv{1, 2000}}, // 250 factoshis not enough for 1 satoshi
			{FCTRate: 5, BTCRate: 10000, FctToBTC: Conv{250e3, 0.125e3}, BTCToFCT: Conv{1, 2000}},
		}

		for _, vec := range vectors {
			assets := make(OraclePriceRecordAssetList)
			assets["FCT"] = vec.FCTRate
			assets["BTC"] = vec.BTCRate

			if a, err := assets.ExchangeFrom("FCT", vec.FctToBTC.Amt, "BTC"); err != nil || a != vec.FctToBTC.Res {
				t.Errorf("amt is incorrect. found %d, exp %d", a, vec.FctToBTC.Res)
			}

			if a, err := assets.ExchangeFrom("BTC", vec.BTCToFCT.Amt, "FCT"); err != nil || a != vec.BTCToFCT.Res {
				t.Errorf("amt is incorrect. found %d, exp %d", a, vec.BTCToFCT.Res)
			}
		}

	})

	type ConversionTest struct {
		FromRate    float64 // USD rate of from currency
		ToRate      float64 // USD rate of to currency
		ConvertRate float64 // Rate from -> to
		Have        int64   // The fixed FROM input
		Amt         int64   // The amount we receive from the fixed
		Need        int64   // The amount we expect to need to get 'amt'
		Difference  int64   // the difference in the expected amounts
		WhenNeeded  int64
	}

	t.Run("vectored", func(t *testing.T) {
		var arr []ConversionTest
		err := json.Unmarshal([]byte(conversionVector), &arr)
		if err != nil {
			t.Error(err)
		}
		// Create random prices, and check the exchage from matches the to
		for _, a := range arr {
			assets := make(OraclePriceRecordAssetList)
			// Random prices up to 100K usd per coin
			assets["from"] = a.FromRate
			assets["to"] = a.ToRate

			assets["from"] = polling.RoundRate(assets["from"])
			assets["to"] = polling.RoundRate(assets["to"])

			// Random amount up to 100K*1e8
			have := a.Have
			amt, err := assets.ExchangeFrom("from", have, "to")
			if err != nil {
				t.Error(err)
			}

			if amt != a.Amt {
				t.Errorf("Exp to get %d, got %d", a.Amt, amt)
			}

			// The amt you get should match the amount to
			need, err := assets.ExchangeTo("from", "to", amt)
			if err != nil {
				t.Error(err)
			}

			if need != a.Need {
				t.Errorf("Exp to get %d, got %d", a.Need, need)
			}
		}
	})

	// This unit test is checking that the conversions where you choose the FROM or the TO.
	// The math should work out, that the tx amounts should be the same, no matter which way you choose.
	// The equation:
	//		f = from amt
	//		t = to amt
	//		r = from->to exchrate
	//		Solve for t or for f
	//		t = f * r
	//	We expect some errors. This test has some tolerance allowed. It can be used
	// 	to create reference vectors.
	//
	// | TX | From Currency (int64) | Conversion Direction (float64 ratio) | To Currency (int64) |
	// |----|-----------------------|:------------------------------------:|---------------------|
	// | 1  | Given `have`          |                 --->                 | Computed `amt`      |
	// | 2  | Computed `need`       |                 --->                 | Given `amt`         |
	// | 3  | Computed `whenNeeded` |                 <---                 | Given `amt`         |
	t.Run("random", func(t *testing.T) {
		// Create random prices, and check the exchage from matches the to
		for i := 0; i < 50000; i++ {
			assets := make(OraclePriceRecordAssetList)
			// Random prices up to 100 usd per coin
			assets["from"] = rand.Float64() * float64(rand.Int63n(100))
			assets["to"] = rand.Float64() * float64(rand.Int63n(100))

			if assets["from"] < 0.0001 || assets["to"] < 0.0001 {
				continue // This won't work. 0 rates are not valid
			}

			cr, _ := assets.ExchangeRate("from", "to")
			if cr < 0.0001 || cr == math.Inf(1) {
				// When doing random rates, we will get precision errors.
				// This unit test is good at finding vectors, and checking
				// general conversion logic. But other testing should be done
				// and this should serve as a tool to verify precision
				// continue
			}

			if cr == math.NaN() {
				continue
			}

			assets["from"] = polling.RoundRate(assets["from"])
			assets["to"] = polling.RoundRate(assets["to"])

			// Random amount up to 100K*1e8
			have := rand.Int63n(1000 * 1e8)
			amt, err := assets.ExchangeFrom("from", have, "to")
			if err != nil {
				t.Error(err)
			}

			// The amt you get should match the amount to
			need, err := assets.ExchangeTo("from", "to", amt)
			if err != nil {
				t.Error(err)
			}

			// The amt you get should match the amount to
			whenNeed, err := assets.ExchangeFrom("to", amt, "from")
			if err != nil {
				t.Error(err)
			}

			d, err := json.Marshal(&ConversionTest{
				FromRate:    assets["from"],
				ToRate:      assets["to"],
				ConvertRate: cr,
				Have:        have,
				Amt:         amt,
				Need:        need,
				Difference:  have - need,
				WhenNeeded:  whenNeed,
			})

			if err != nil {
				t.Error(err)
				continue
			}

			// This checks if you can ever create tokens out of think air.
			// There should only be allowed a loss of tokens
			if float64(have-need) < 0 {
				t.Errorf(string(d))
				t.Errorf("Precision err. have %d, exp %d. Diff %d", have, need, have-need)
			}

			// Checking the reverse. We truncate, so a diff of 1 can happen
			if math.Abs(float64(whenNeed-need)) > 1 {
				t.Errorf(string(d))
				t.Errorf("Precision err going opposit. whenNeed %d, exp %d. Diff %d", whenNeed, need, whenNeed-need)
			}

			assets["USD"] = assets["from"]
			assets["XBT"] = assets["to"]
			ad, err := json.Marshal(&assets)
			if err != nil {
				t.Error(err)
			}

			// Verify our values are the same coming out of json
			assets2 := make(OraclePriceRecordAssetList)
			err = json.Unmarshal(ad, &assets2)
			if err != nil {
				t.Error(err)
			}

			if assets2["USD"] != assets["USD"] {
				t.Errorf("Json output does not match in for %f. Found %f", assets["USD"], assets2["USD"])
			}
			if assets2["XBT"] != assets["XBT"] {
				t.Errorf("Json output does not match in for %f. Found %f", assets["XBT"], assets2["XBT"])
			}
		}
	})

	// Verify the numbers we write to chain are the same we calculate from source
	t.Run("Test float json rounding", func(t *testing.T) {
		for i := float64(0); i < 2; i += float64(1) / 10000 {
			a := make(OraclePriceRecordAssetList)
			a["USD"] = polling.RoundRate(i)
			d, _ := json.Marshal(a)

			a2 := make(OraclePriceRecordAssetList)
			err := json.Unmarshal(d, &a2)
			if err != nil {
				t.Error(err)
			}

			if a2["USD"] != a["USD"] {
				t.Errorf("Pre RoundRate : Exp %f, found %f", a["USD"], a2["USD"])
			}

			p := a2["USD"]
			p2 := polling.RoundRate(polling.RoundRate(p))
			var _, _ = p, p2
			a2["USD"] = polling.RoundRate(a2["USD"])
			if a2["USD"] != a["USD"] {
				t.Errorf("Post RoundRate : Exp %f, found %f", a["USD"], a2["USD"])
			}
		}
	})

}

const conversionVector = `
[

]
`
