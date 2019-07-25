// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr_test

import (
	"encoding/json"
	"fmt"
	"testing"

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
	opr.CoinbasePNTAddress = "pPNT4wBqpZM9xaShSYTABzAf1i1eSHVbbNk2xd1x6AkfZiy366c620f"
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
	json.Unmarshal(v, &opr2)
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

func TestMarshal(t *testing.T) {
	j := `{"oprchainid":"HtnKdX6KXNnMv77StmUPrrHCWTCrWoEAYSRpMida8z67","dbht":917,"winners":["","","","","","","","","",""],"coinbase":"tPNT_mEU1i4M5rn7bnrxNKdVVf4HXLG15Q798oaVAMrXq7zdbhQ9pv","FactomDigitalID":"prototyp","assets":{"PNT":0,"USD":1.0171,"EUR":0.8799,"JPY":105.7501,"GBP":0.8031,"CAD":1.3237,"CHF":0.9661,"INR":71.5292,"SGD":1.371,"CNY":6.7092,"HKD":7.6939,"KRW":1193.8721,"BRL":3.7445,"PHP":50.7479,"MXN":18.5671,"XAU":16.4795,"XAG":1367.5927,"XPD":1487.2841,"XPT":865.4626,"XBT":9845.1248,"ETH":217.676,"LTC":93.0899,"RVN":0.0487,"XBC":301.6768,"FCT":4.0716,"BNB":30.3459,"XLM":0.0849,"ADA":0.0586,"XMR":79.7383,"DASH":115.7181,"ZEC":70.4885,"DCR":27.3995}}`

	o := new(OraclePriceRecord)
	err := json.Unmarshal([]byte(j), o)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(o.Assets)
}
