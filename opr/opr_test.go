// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr_test

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"testing"

	"github.com/FactomProject/btcutil/base58"
	"github.com/pegnet/pegnet/common"
	. "github.com/pegnet/pegnet/opr"
)

func TestOPRTokens(t *testing.T) {
	opr := NewOraclePriceRecord()
	opr.Version = 1

	if len(opr.GetTokens()) != len(common.AssetsV1) {
		t.Errorf("exp %d tokens, found %d", len(common.AssetsV1), len(opr.GetTokens()))
	}
	for i, token := range opr.GetTokens() {
		if token.Code != common.AssetsV1[i] {
			t.Errorf("exp %s got %s", token.Code, common.AssetsV1[i])
		}
	}

	opr.Version = 2
	if len(opr.GetTokens()) != len(common.AssetsV2) {
		t.Errorf("exp %d tokens, found %d", len(common.AssetsV2), len(opr.GetTokens()))
	}
	for i, token := range opr.GetTokens() {
		if token.Code != common.AssetsV2[i] {
			t.Errorf("exp %s got %s", token.Code, common.AssetsV2[i])
		}
	}

	opr.Version = 3
	if len(opr.GetTokens()) != len(common.AssetsV2) {
		t.Errorf("exp %d tokens, found %d", len(common.AssetsV2), len(opr.GetTokens()))
	}
	for i, token := range opr.GetTokens() {
		if token.Code != common.AssetsV2[i] {
			t.Errorf("exp %s got %s", token.Code, common.AssetsV2[i])
		}
	}

	opr.Version = 4
	if len(opr.GetTokens()) != len(common.AssetsV4) {
		t.Errorf("exp %d tokens, found %d", len(common.AssetsV4), len(opr.GetTokens()))
	}
	for i, token := range opr.GetTokens() {
		if token.Code != common.AssetsV4[i] {
			t.Errorf("exp %s got %s", token.Code, common.AssetsV4[i])
		}
	}

	opr.Version = 5
	if len(opr.GetTokens()) != len(common.AssetsV5) {
		t.Errorf("exp %d tokens, found %d", len(common.AssetsV5), len(opr.GetTokens()))
	}
	for i, token := range opr.GetTokens() {
		if token.Code != common.AssetsV5[i] {
			t.Errorf("exp %s got %s", token.Code, common.AssetsV5[i])
		}
	}
}

func TestOPR_JSON_Marshal(t *testing.T) {
	LX.Init(0x123412341234, 25, 256, 5)
	opr := NewOraclePriceRecord()

	opr.Difficulty = 1
	opr.Grade = 1
	//opr.Nonce = base58.Encode(LX.Hash([]byte("a Nonce")))
	//opr.ChainID = base58.Encode(LX.Hash([]byte("a chainID")))
	opr.Dbht = 1901232
	opr.WinPreviousOPR = []string{
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
	opr.CoinbasePEGAddress = "PEG_2Gec4tfkeQ64xVPM1Rz2esDcy6XAW3kHEM1jvZLbTTWCDciiqN"
	opr.FactomDigitalID = "minerone"
	opr.Assets["PEG"] = 2
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

	opr.Version = 1
	v, _ := opr.SafeMarshal()
	fmt.Println("len of entry", len(string(v)), "\n\n", string(v))
	opr2 := NewOraclePriceRecord()
	opr2.Version = 1
	err := opr2.SafeUnmarshal(v)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	v2, _ := opr2.SafeMarshal()
	fmt.Println("\n\n", string(v2))
	if string(v2) != string(v) {
		t.Error("JSON is different")
	}
}

func TestValidFCTAddress(t *testing.T) {
	tfa := func(addr string, valid bool, reason string) {
		if v := ValidFCTAddress(addr); v != valid {
			t.Errorf("Valid: %t, exp %t: %s", v, valid, reason)
		}
	}

	tfa("FA2vP7vAyDBmBBhdWqRPyM9W2WGqPYeAoMcG7QtNQb2TY6MKpanu", true, "valid addr")
	tfa("FA2DSjsRoKEyHnmLg6BzCUg9tRpS1Hod62aEV8Gdf5sU9hesrRZc", true, "valid addr")
	tfa("FA2AvQRG58jPtGAkRiXsajWFQvWo5VWA31ds7neG95cLJtACiiw7", true, "valid addr")

	tfa("FA2vP7vAyDBmBBhdWqRPyM9W2WGqPYeAoMcG7QtNQb2TY6MKpana", false, "bad checksum")
	tfa("FA2DSjsRoKEyHnmLg6BzCUg9tRpS1Hod62aEV8Gdf5aU9hesrRZc", false, "bad checksum")

	tfa("Fs2Uk1vnk2JrHHQXTDvSW6LsRTFqfim4khBk2yKHU4MWSYSnQCcg", false, "not a FA key")
	tfa("Es2XT3jSxi1xqrDvS5JERM3W3jh1awRHuyoahn3hbQLyfEi1jvbq", false, "not a FA key")
	tfa("EC3TsJHUs8bzbbVnratBafub6toRYdgzgbR7kWwCW4tqbmyySRmg", false, "not a FA key")

	tfa("", false, "empty")
	tfa("FA", false, "not long enough")
	tfa("FAs", false, "not long enough")
	tfa("FAAvQRG58jPtGAkRiXsajWFQvWo5VWA31ds7neG95cLJtACiiw7", false, "missing a character")
}

func TestProtobufSize(t *testing.T) {
	opr := NewOraclePriceRecord()
	opr.Version = 2
	opr.CoinbaseAddress = "FA3bGeJUkzu6BnjqkcfxAqAcKXhu5dwygGnT6qfGLRy1otkEZqpd"
	opr.FactomDigitalID = "v2protobufmarshaltesting"
	opr.Dbht = 200000
	for _, asset := range common.AssetsV2 {
		opr.Assets.SetValueFromUint64(asset, rand.Uint64())
	}

	opr.WinPreviousOPR = make([]string, 25, 25)
	for i := range opr.WinPreviousOPR {
		opr.WinPreviousOPR[i] = "0001000200030004"
	}

	entry, err := opr.CreateOPREntry(make([]byte, 5, 5), rand.Uint64())
	if err != nil {
		t.Error(err)
	}

	data, err := entry.MarshalBinary()
	if len(data) > 1024 {
		t.Errorf("opr entry is over 1kb, found %d bytes", len(data))
	}
	fmt.Println(len(data))
}

func rstring(len int) string {
	r := make([]byte, len)
	rand.Read(r)
	return string(r)
}

func genJSONOPR() *OraclePriceRecord {
	opr := new(OraclePriceRecord)
	opr.CoinbaseAddress = rstring(56)
	opr.Dbht = rand.Int31()
	for i := range opr.WinPreviousOPR {
		opr.WinPreviousOPR[i] = rstring(8)
	}
	opr.FactomDigitalID = rstring(25)
	opr.Assets = make(OraclePriceRecordAssetList)
	for _, a := range common.AllAssets {
		opr.Assets.SetValue(a, rand.Float64()*5000)
	}
	return opr
}

func BenchmarkJSONMarshal(b *testing.B) {
	b.StopTimer()
	InitLX()
	data := make([]*OraclePriceRecord, 0)
	for i := 0; i < b.N; i++ {
		data = append(data, genJSONOPR())
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		data[i].SafeMarshal()
	}
}

func BenchmarkSingleOPRHash(b *testing.B) {
	b.StopTimer()
	InitLX()
	data := make([][]byte, 0)
	for i := 0; i < b.N; i++ {
		json, _ := genJSONOPR().SafeMarshal()
		data = append(data, json)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		LX.Hash(data[i])
	}
}

func BenchmarkSingleSha256(b *testing.B) {
	b.StopTimer()
	InitLX()
	data := make([][]byte, 0)
	for i := 0; i < b.N; i++ {
		json, _ := genJSONOPR().SafeMarshal()
		data = append(data, json)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sha256.Sum256(data[i])
	}
}
func BenchmarkComputeDifficulty(b *testing.B) {
	b.StopTimer()
	data := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = make([]byte, 37) // 32 oprhash + "average" 5 byte nonce
		rand.Read(data[i])
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ComputeDifficulty(data[i][0:32], data[i][32:])
	}
}
