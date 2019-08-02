// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr_test

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
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
		FromRate   float64 // USD rate of from currency
		ToRate     float64 // USD rate of to currency
		Have       int64   // The fixed FROM input
		Get        int64   // The amount we receive from the fixed
		Need       int64   // The amount we expect to need to get 'Get'
		Difference int64   // the difference in the expected amounts
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

			if amt != a.Get {
				t.Errorf("Exp to get %d, got %d", a.Get, amt)
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
	t.Run("random", func(t *testing.T) {
		// Create random prices, and check the exchage from matches the to
		for i := 0; i < 100000; i++ {
			assets := make(OraclePriceRecordAssetList)
			// Random prices up to 100K usd per coin
			assets["from"] = rand.Float64() * float64(rand.Int63n(100e3))
			assets["to"] = rand.Float64() * float64(rand.Int63n(100e3))

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

			d, _ := json.Marshal(&ConversionTest{
				assets["from"],
				assets["to"],
				have,
				amt,
				need,
				have - need,
			})

			// A 400 'sat' tolerance. I'm not sure how else to test and know the expected error.
			// If you turn down this tolerance, you can get some more vector tests.
			if math.Abs(float64(have-need)) > 500 {
				t.Errorf(string(d))
				t.Errorf("Precision err. have %d, exp %d. Diff %d", have, need, have-need)
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

	t.Run("random-Float64", func(t *testing.T) {
		// Create random prices, and check the exchage from matches the to
		for i := 0; i < 100000; i++ {
			assets := make(OraclePriceRecordAssetList)
			// Random prices up to 100K usd per coin
			assets["from"] = rand.Float64() * float64(rand.Int63n(100e2))
			assets["to"] = rand.Float64() * float64(rand.Int63n(100e2))

			assets["from"] = polling.Round(assets["from"])
			assets["to"] = polling.Round(assets["to"])

			iAssets := RegularFloats(assets)

			// Random amount up to 100K*1e8
			have := float64(rand.Int63n(1000 * 1e8))
			amt, err := iAssets.ExchangeFrom("from", have, "to")
			if err != nil {
				t.Error(err)
			}

			// The amt you get should match the amount to
			need, err := iAssets.ExchangeTo("from", "to", amt)
			if err != nil {
				t.Error(err)
			}

			d, _ := json.Marshal(&ConversionTest{
				assets["from"],
				assets["to"],
				int64(have),
				int64(amt),
				int64(need),
				int64(have - need),
			})

			diff := math.Abs(have - need)

			// A 400 'sat' tolerance. I'm not sure how else to test and know the expected error.
			// If you turn down this tolerance, you can get some more vector tests.
			if diff > 500 {
				t.Errorf(string(d))
				t.Errorf("Precision err. have %f, exp %f. Diff %f", have, need, have-need)
			}

			// To ints

			hI := int64(have)
			nI := int64(need)
			// A 400 'sat' tolerance. I'm not sure how else to test and know the expected error.
			// If you turn down this tolerance, you can get some more vector tests.
			if math.Abs(float64(nI-hI)) > 500 {
				t.Errorf("Precision err. have %d, exp %d. Diff %d", hI, nI, hI-nI)
			}
		}
	})

	t.Run("random-bigFloats", func(t *testing.T) {
		// Create random prices, and check the exchage from matches the to
		for i := 0; i < 100000; i++ {
			assets := make(OraclePriceRecordAssetList)
			// Random prices up to 100K usd per coin
			assets["from"] = rand.Float64() * float64(rand.Int63n(100e3))
			assets["to"] = rand.Float64() * float64(rand.Int63n(100e3))

			assets["from"] = polling.Round(assets["from"])
			assets["to"] = polling.Round(assets["to"])

			iAssets := BigFloats(assets)

			// Random amount up to 100K*1e8
			have := big.NewFloat(float64(rand.Int63n(1000 * 1e8)))
			amt, err := iAssets.ExchangeFrom("from", have, "to")
			if err != nil {
				t.Error(err)
			}

			// The amt you get should match the amount to
			need, err := iAssets.ExchangeTo("from", "to", amt)
			if err != nil {
				t.Error(err)
			}

			//d, _ := json.Marshal(&ConversionTest{
			//	assets["from"],
			//	assets["to"],
			//	have,
			//	amt,
			//	need,
			//	have - need,
			//})

			diff := big.NewFloat(0).Sub(have, need)
			aDiff := big.NewFloat(0).Abs(diff)

			// A 400 'sat' tolerance. I'm not sure how else to test and know the expected error.
			// If you turn down this tolerance, you can get some more vector tests.
			if aDiff.Cmp(big.NewFloat(500)) == 1 {
				//t.Errorf(string(d))
				t.Errorf("Precision err. have %d, exp %d. Diff %d", have, need, diff)
			}

			// To ints

			hI, _ := have.Int64()
			nI, _ := need.Int64()
			// A 400 'sat' tolerance. I'm not sure how else to test and know the expected error.
			// If you turn down this tolerance, you can get some more vector tests.
			if math.Abs(float64(nI-hI)) > 500 {
				t.Errorf("Precision err. have %d, exp %d. Diff %d", hI, nI, hI-nI)
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

const conversionVector = `[{"FromRate":7401.2406,"ToRate":12554.2132,"Have":83311134652,"Get":49111913877,"Need":83311134651,"Difference":1},
{"FromRate":971.9422,"ToRate":58891.8338,"Have":60983331750,"Get":1006224974,"Need":60983331758,"Difference":-8},
{"FromRate":7840.4901,"ToRate":59324.0861,"Have":77372812453,"Get":10228685806,"Need":77372812451,"Difference":2},
{"FromRate":32570.8436,"ToRate":38708.8151,"Have":81437551529,"Get":68521555857,"Need":81437551530,"Difference":-1},
{"FromRate":1247.1401,"ToRate":8919.7313,"Have":72122742700,"Get":10082759429,"Need":72122742697,"Difference":3},
{"FromRate":363.0485,"ToRate":21587.7816,"Have":60614313135,"Get":1018320461,"Need":60614313155,"Difference":-20},
{"FromRate":879.7282,"ToRate":26859.0372,"Have":1432260190,"Get":46978134,"Need":1432260183,"Difference":7},
{"FromRate":7475.8643,"ToRate":19388.1319,"Have":39528292143,"Get":15242109450,"Need":39528292142,"Difference":1},
{"FromRate":349.895,"ToRate":57311.615,"Have":48154650895,"Get":293743370,"Need":48154650820,"Difference":75},
{"FromRate":10691.0146,"ToRate":76022.3443,"Have":578986641,"Get":81405522,"Need":578986643,"Difference":-2},
{"FromRate":3762.2321,"ToRate":41476.6813,"Have":29564889229,"Get":2681535453,"Need":29564889228,"Difference":1},
{"FromRate":33879.3138,"ToRate":64319.8167,"Have":68471983631,"Get":36064193778,"Need":68471983630,"Difference":1},
{"FromRate":7250.799,"ToRate":28106.2177,"Have":42639110966,"Get":11000890629,"Need":42639110965,"Difference":1},
{"FromRate":3199.0859,"ToRate":70556.5703,"Have":32527976373,"Get":1473517330,"Need":32527976380,"Difference":-7},
{"FromRate":6820.1745,"ToRate":16517.0497,"Have":22127172750,"Get":9136309628,"Need":22127172749,"Difference":1},
{"FromRate":889.5809,"ToRate":21820.7544,"Have":91911834437,"Get":3750002845,"Need":91911834436,"Difference":1},
{"FromRate":16434.0195,"ToRate":41816.5483,"Have":52175362199,"Get":20504917344,"Need":52175362198,"Difference":1},
{"FromRate":10916.4125,"ToRate":61329.0005,"Have":65879719289,"Get":11726590033,"Need":65879719287,"Difference":2},
{"FromRate":4773.2617,"ToRate":39510.0051,"Have":83423015425,"Get":10077500263,"Need":83423015422,"Difference":3},
{"FromRate":884.851,"ToRate":50361.801,"Have":34938452019,"Get":614916756,"Need":34938452045,"Difference":-26},
{"FromRate":7837.9481,"ToRate":39901.8513,"Have":33983192302,"Get":6674298968,"Need":33983192301,"Difference":1},
{"FromRate":2793.6353,"ToRate":40680.3513,"Have":68712918070,"Get":4720577471,"Need":68712918064,"Difference":6}
]
`
