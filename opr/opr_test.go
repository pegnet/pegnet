// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr_test

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/pegnet/pegnet/polling"
	"github.com/FactomProject/btcutil/base58"
	. "github.com/pegnet/pegnet/opr"
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
		for i := 0; i < 100; i++ {
			assets := make(OraclePriceRecordAssetList)
			// Random prices up to 100K usd per coin
			assets["from"] = rand.Float64() * float64(rand.Int63n(100e3))
			assets["to"] = rand.Float64() * float64(rand.Int63n(100e3))

			assets["from"] = polling.Round(assets["from"])
			assets["to"] = polling.Round(assets["to"])

			//iAssets := IntegerBasedOraclePriceRecordAssetList(assets)

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

			if math.Abs(float64(have-need)) > 200 {
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

	// Verify the numbers we write to chain are the same we calculate from source
	t.Run("Test float json rounding", func(t *testing.T) {
		for i := float64(0); i < 1; i += float64(1) / 10000 {
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

const conversionVector = `[{"FromRate":7401.2406,"ToRate":12554.2132,"Have":83311134652,"Get":49115443747,"Need":83311134651,"Difference":1},
{"FromRate":4185.0348,"ToRate":8681.0241,"Have":1706343734,"Get":822611229,"Need":1706343733,"Difference":1},
{"FromRate":971.9422,"ToRate":58891.8337,"Have":60983331750,"Get":1006459978,"Need":60983331776,"Difference":-26},
{"FromRate":7840.49,"ToRate":59324.0861,"Have":77372812453,"Get":10225876237,"Need":77372812456,"Difference":-3},
{"FromRate":17132.7037,"ToRate":28481.7391,"Have":64269878397,"Get":38660447648,"Need":64269878396,"Difference":1},
{"FromRate":1013.5703,"ToRate":12352.8003,"Have":3139208989,"Get":257577952,"Need":3139208995,"Difference":-6},
{"FromRate":8870.8259,"ToRate":44734.7877,"Have":63252345915,"Get":12542823544,"Need":63252345917,"Difference":-2},
{"FromRate":1247.14,"ToRate":8919.7312,"Have":72122742700,"Get":10084065911,"Need":72122742699,"Difference":1},
{"FromRate":363.0484,"ToRate":21587.7816,"Have":60614313135,"Get":1019369651,"Need":60614313120,"Difference":15},
{"FromRate":879.7282,"ToRate":26859.0372,"Have":1432260190,"Get":46911573,"Need":1432260196,"Difference":-6},
{"FromRate":7475.8643,"ToRate":19388.1319,"Have":39528292143,"Get":15241702996,"Need":39528292142,"Difference":1},
{"FromRate":16171.8423,"ToRate":48167.1005,"Have":77579800709,"Get":26046996595,"Need":77579800708,"Difference":1},
{"FromRate":349.8949,"ToRate":57311.6149,"Have":48154650895,"Get":293990438,"Need":48154650916,"Difference":-21},
{"FromRate":336.5728,"ToRate":6696.3206,"Have":26627778270,"Get":1338374673,"Need":26627778280,"Difference":-10},
{"FromRate":14241.8855,"ToRate":65514.1814,"Have":70914539462,"Get":15415849358,"Need":70914539460,"Difference":2},
{"FromRate":10691.0145,"ToRate":76022.3443,"Have":578986641,"Get":81422832,"Need":578986640,"Difference":1},
{"FromRate":4291.5717,"ToRate":14962.5347,"Have":10658231531,"Get":3057006432,"Need":10658231532,"Difference":-1},
{"FromRate":3199.0859,"ToRate":70556.5703,"Have":32527976373,"Get":1474841962,"Need":32527976374,"Difference":-1},
{"FromRate":24785.8749,"ToRate":70201.2528,"Have":11335423860,"Get":4002184954,"Need":11335423859,"Difference":1},
{"FromRate":6820.1745,"ToRate":16517.0497,"Have":22127172750,"Get":9136691000,"Need":22127172749,"Difference":1},
{"FromRate":889.5808,"ToRate":21820.7544,"Have":91911834437,"Get":3747029168,"Need":91911834433,"Difference":4},
{"FromRate":16434.0194,"ToRate":41816.5483,"Have":52175362199,"Get":20505061978,"Need":52175362200,"Difference":-1},
{"FromRate":10916.4124,"ToRate":61329.0005,"Have":65879719289,"Get":11726429237,"Need":65879719288,"Difference":1},
{"FromRate":4773.2617,"ToRate":39510.005,"Have":83423015425,"Get":10078456948,"Need":83423015421,"Difference":4},
{"FromRate":6103.3095,"ToRate":35477.0073,"Have":23730213087,"Get":4082442291,"Need":23730213085,"Difference":2},
{"FromRate":7837.9481,"ToRate":39901.8512,"Have":33983192302,"Get":6675341858,"Need":33983192301,"Difference":1},
{"FromRate":13091.4518,"ToRate":14054.2312,"Have":64640301018,"Get":60212143451,"Need":64640301017,"Difference":1},
{"FromRate":2793.6353,"ToRate":40680.3512,"Have":68712918070,"Get":4718711315,"Need":68712918077,"Difference":-7}]
`
