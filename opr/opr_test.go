// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr_test

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/FactomProject/btcutil/base58"
	. "github.com/pegnet/pegnet/opr"
	"github.com/pegnet/pegnet/polling"
)

func TestOPR_JSON_Marshal(t *testing.T) {
	LX.Init(0x123412341234, 25, 256, 5)
	opr := NewOraclePriceRecord()

	opr.Difficulty = 1
	opr.Grade = 1
	//opr.Nonce = base58.Encode(LX.Hash([]byte("a Nonce")))
	//opr.ChainID = base58.Encode(LX.Hash([]byte("a chainID")))
	opr.Dbht = 1901232
	opr.WinPreviousOPR = [10]string{
		base58.Encode(LX.Hash([]byte("winner number 1"))),
		base58.Encode(LX.Hash([]byte("winner number 2"))),
		base58.Encode(LX.Hash([]byte("winner number 3"))),
		base58.Encode(LX.Hash([]byte("winner number 4"))),
		base58.Encode(LX.Hash([]byte("winner number 5"))),
		base58.Encode(LX.Hash([]byte("winner number 6"))),
		base58.Encode(LX.Hash([]byte("winner number 7"))),
		base58.Encode(LX.Hash([]byte("winner number 8"))),
		base58.Encode(LX.Hash([]byte("winner number 9"))),
		base58.Encode(LX.Hash([]byte("winner number 10"))),
	}
	opr.CoinbasePNTAddress = "PNT4wBqpZM9xaShSYTABzAf1i1eSHVbbNk2xd1x6AkfZiy366c620f"
	opr.FactomDigitalID = "minerone"
	opr.Assets["PNT"] = 2
	opr.Assets["USD"] = 20
	opr.Assets["EUR"] = 200
	opr.Assets["JPY"] = 11
	opr.Assets["GBP"] = 12
	opr.Assets["CAD"] = 13
	opr.Assets["CHF"] = 14
	opr.Assets["INR"] = 15
	opr.Assets["SGD"] = 16
	opr.Assets["CNY"] = 17
	opr.Assets["HKD"] = 18
	opr.Assets["XAU"] = 19
	opr.Assets["XAG"] = 101
	opr.Assets["XPD"] = 1012
	opr.Assets["XPT"] = 10123
	opr.Assets["XBT"] = 10124
	opr.Assets["ETH"] = 10125
	opr.Assets["LTC"] = 10126
	opr.Assets["XBC"] = 10127
	opr.Assets["FCT"] = 10128

	v, _ := json.Marshal(opr)
	fmt.Println("len of entry", len(string(v)), "\n\n", string(v))
	opr2 := NewOraclePriceRecord()
	err := json.Unmarshal(v, &opr2)
	if err != nil {
		t.Fail()
	}
	v2, _ := json.Marshal(opr2)
	fmt.Println("\n\n", string(v2))
	if string(v2) != string(v) {
		t.Error("JSON is different")
	}
}

type ExpPath struct {
	Path string
	Exp  string
}

func TestShortenPegnetFilePath(t *testing.T) {
	testPath(ExpPath{
		"/home/billy/go/src/github.com/pegnet/pegnet/opr.go",
		"pegnet/opr.go",
	}, t)

	testPath(ExpPath{
		"/home/billy/go/src/github.com/notpegnet/notpegnet/opr.go",
		"/home/billy/go/src/github.com/notpegnet/notpegnet/opr.go",
	}, t)

	testPath(ExpPath{
		"opr.go",
		"opr.go",
	}, t)

	testPath(ExpPath{
		"/home/steven/go/src/github.com/pegnet/pegnet/opr/oneminer.go",
		"pegnet/opr/oneminer.go",
	}, t)

	testPath(ExpPath{
		"pegnet",
		"pegnet",
	}, t)

	testPath(ExpPath{
		"pegnet/test_pegnet.go",
		"pegnet/test_pegnet.go",
	}, t)

	testPath(ExpPath{
		"/home/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec",
		"/home/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec/rec",
	}, t)
}

func testPath(e ExpPath, t *testing.T) {
	if f := ShortenPegnetFilePath(e.Path, "", 0); f != e.Exp {
		t.Errorf("Exp %s, found %s", e.Exp, f)
	}
}

func TestPriceConversions(t *testing.T) {
	// Simple round numbers
	t.Run("simple numbers", func(t *testing.T) {
		assets := make(OraclePriceRecordAssetList)
		assets["FCT"] = 100   // $100 per FCT. Woah
		assets["BTC"] = 10000 // $10,000 per BTC

		if r, err := assets.ExchangeRate("FCT", "BTC"); err != nil || r != 0.01 {
			t.Errorf("rate is incorrect. found %f", r)
		}

		if a, err := assets.ExchangeFrom("FCT", 100, "BTC"); err != nil || a != 1 {
			t.Errorf("amt is incorrect. found %d", a)
		}

		if a, err := assets.ExchangeTo("FCT", "BTC", 1); err != nil || a != 100 {
			t.Errorf("amt is incorrect. found %d", a)
		}
	})

	type ConversionTest struct {
		FromRate    float64 // USD rate of from currency
		ToRate      float64 // USD rate of to currency
		ConvertRate float64 // Rate from -> to
		Have        int64   // The fixed FROM input
		amt         int64   // The amount we receive from the fixed
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

			assets["from"] = polling.Round(assets["from"])
			assets["to"] = polling.Round(assets["to"])

			// Random amount up to 100K*1e8
			have := a.Have
			amt, err := assets.ExchangeFrom("from", have, "to")
			if err != nil {
				t.Error(err)
			}

			if amt != a.amt {
				t.Errorf("Exp to get %d, got %d", a.amt, amt)
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
		for i := 0; i < 100000; i++ {
			assets := make(OraclePriceRecordAssetList)
			// Random prices up to 100 usd per coin
			assets["from"] = rand.Float64() * float64(rand.Int63n(100))
			assets["to"] = rand.Float64() * float64(rand.Int63n(100))

			if assets["from"] < 0.0001 || assets["to"] < 0.0001 {
				continue // This won't work. 0 rates are not valid
			}

			cr, _ := assets.ExchangeRate("from", "to")
			if cr < 0.01 || cr == math.Inf(1) {
				// When doing random rates, we will get precision errors.
				// This unit test is good at finding vectors, and checking
				// general conversion logic. But other testing should be done
				// and this should serve as a tool to verify precision
				continue
			}

			if cr == math.NaN() {
				continue
			}

			assets["from"] = polling.Round(assets["from"])
			assets["to"] = polling.Round(assets["to"])

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
				assets["from"],
				assets["to"],
				cr,
				have,
				amt,
				need,
				have - need,
				whenNeed,
			})

			if err != nil {
				t.Error(err)
				continue
			}

			// A 400 'sat' tolerance. I'm not sure how else to test and know the expected error.
			// If you turn down this tolerance, you can get some more vector tests.
			if math.Abs(float64(have-need)) > 400 {
				t.Errorf(string(d))
				t.Errorf("Precision err. have %d, exp %d. Diff %d", have, need, have-need)
			}

			// Checking the reverse
			if math.Abs(float64(whenNeed-need)) > 400 {
				t.Errorf(string(d))
				t.Errorf("Precision err going opposit. whenNeed %d, exp %d. Diff %d", whenNeed, need, whenNeed-need)
			}

			// Verify our values are the same coming out of json
			var c ConversionTest
			err = json.Unmarshal(d, &c)
			if err != nil {
				t.Error(err)
			}

			assets2 := make(OraclePriceRecordAssetList)
			// Random prices up to 100K usd per coin
			assets2["from"] = c.FromRate
			assets["to"] = c.ToRate
			assets2["from"] = polling.Round(assets["from"])
			assets2["to"] = polling.Round(assets["to"])

			if assets2["from"] != assets["from"] {
				t.Errorf("Json output does not match in for %f", assets["from"])
			}
			if assets2["to"] != assets["to"] {
				t.Errorf("Json output does not match in for %f", assets["to"])
			}
		}
	})

	// Verify the numbers we write to chain are the same we calculate from source
	t.Run("Test float json rounding", func(t *testing.T) {
		for i := float64(0); i < 2; i += float64(1) / 10000 {
			c := polling.Round(i)
			d, _ := json.Marshal(c)

			var nf float64
			err := json.Unmarshal(d, &nf)
			if err != nil {
				t.Error(err)
			}

			if nf != c {
				t.Errorf("Pre Round : Exp %f, found %f", i, nf)
			}

			nf = polling.Round(nf)

			if nf != c {
				t.Errorf("Post Round : Exp %f, found %f", i, nf)
			}
		}
	})

}

func TestIntCast(t *testing.T) {
	type Expected struct {
		V   float64
		Exp int64
	}

	testingfunc := func(t *testing.T, e Expected) {
		if i := Int64RoundedCast(e.V); i != e.Exp {
			t.Errorf("Exp %d, found %d with %f", i, e.Exp, e.V)
		}
	}

	testingfunc(t, Expected{V: 0.1, Exp: 0})
	testingfunc(t, Expected{V: 0.2, Exp: 0})
	testingfunc(t, Expected{V: 0.3, Exp: 0})
	testingfunc(t, Expected{V: 0.4, Exp: 0})
	testingfunc(t, Expected{V: 0.5, Exp: 1})
	testingfunc(t, Expected{V: 0.6, Exp: 1})
	testingfunc(t, Expected{V: 0.7, Exp: 1})
	testingfunc(t, Expected{V: 0.8, Exp: 1})
	testingfunc(t, Expected{V: 0.9, Exp: 1})

	testingfunc(t, Expected{V: 15.6, Exp: 16})
	testingfunc(t, Expected{V: 22.49, Exp: 22})
}

const conversionVector = `
[
{"FromRate":0.0056,"ToRate":28.4376,"ConvertRate":0.0001957678336190335,"Have":49530497071,"Need":49530499173,"Difference":-2102,"WhenNeeded":49530499173},
{"FromRate":0.0194,"ToRate":24.9232,"ConvertRate":0.0007776509548873975,"Have":14220790757,"Need":14220791225,"Difference":-468,"WhenNeeded":14220791225},
{"FromRate":0.032,"ToRate":82.2916,"ConvertRate":0.00038912927643798784,"Have":75673685781,"Need":75673686383,"Difference":-602,"WhenNeeded":75673686383},
{"FromRate":0.0005,"ToRate":48.4376,"ConvertRate":0.000010732782363517055,"Have":54662636938,"Need":54662606602,"Difference":30336,"WhenNeeded":54662606602}
]
`
