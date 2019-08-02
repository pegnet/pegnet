// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full
package common_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/FactomProject/factoid"
	. "github.com/pegnet/pegnet/common"
)

func TestConvertFctToPegAssets(t *testing.T) {

	fctAddrs := []string{
		"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
		"0000000000000000000000000000000000000000000000000000000000000000",
		"1235134346267276876846779695807808006ababababababababdfdfdfeeeeF",
		"abcdef1256267584690ab3472accffeeeeeee154134628697025428890000002",
		"1234651367286539567abababefefefefefefef6228365940abcdefedededede",
	}

	for _, fa := range fctAddrs {
		fmt.Println("\n", fa)
		raw, err := hex.DecodeString(fa)
		if err != nil {
			t.Fatal(fa, " ", err)
		}
		fct := ConvertRawToFCT(raw)
		fmt.Println(fct)
		raw = factoid.ConvertUserStrToAddress(fct)
		fmt.Printf("%x\n", raw)
		pnt, err := ConvertRawToPegNetAsset("PNT", raw)
		if err != nil {
			t.Fatal(fa, " ", err)
		}
		fmt.Println(pnt)
		pre, raw2, err := ConvertPegNetAssetToRaw(pnt)
		if pre != "PNT" || !bytes.Equal(raw, raw2) || err != nil {
			t.Fatal("Round trip failed with PNT")
		}
		fmt.Printf("%x\n", raw2)
		tpnt, err := ConvertRawToPegNetAsset("tPNT", raw)
		if err != nil {
			t.Fatal(fa, " ", err)
		}
		fmt.Println(tpnt)
		pre, raw2, err = ConvertPegNetAssetToRaw(tpnt)
		if pre != "tPNT" || !bytes.Equal(raw, raw2) || err != nil {
			t.Fatal("Round trip failed with tPNT")
		}
		fmt.Printf("%x\n", raw2)

		assets, err := ConvertFCTtoAllPegNetAssets(fct)
		if err != nil {
			t.Fatal(fa, " ", err)
		}
		for _, asset := range assets {
			fmt.Println(asset)
		}
	}
}

func TestBurnECAddress(t *testing.T) {
	ecAdd := "EC1moooFCT2TESToooo1oooo1oooo1oooo1oooo1oooo1oooo1oo"
	raw, err := ConvertAnyFactomAdrToRaw(ecAdd)
	if err != nil {
		t.Fatal(err)
	}
	burn := ConvertRawToEC(raw)
	fmt.Printf("Suggested Address %s\n", ecAdd)
	fmt.Printf("Raw Address       %s\n", hex.EncodeToString(raw))
	fmt.Printf("Suggested+csum    %s\n", burn)
	raw, err = ConvertAnyFactomAdrToRaw(burn)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Back again        %s\n", hex.EncodeToString(raw))
}

func TestConvertRawAddrToPegT(t *testing.T) {

	var RawAddress [32]byte
	setAdr := func(str string) {
		adr, err := hex.DecodeString(str)
		if err != nil {
			panic(err)
		}
		copy(RawAddress[:], adr)
	}

	setAdr("000102030405060708090001020304050607080900010203040506070809AABB")

	var HumanAdr string
	var err error

	ConvertToHuman := func(prefix string) error {
		HumanAdr, err = ConvertRawToPegNetAsset(prefix, RawAddress[:])
		if err != nil {
			return err
		}
		fmt.Printf("%5s %15s,%x\n%5s %15s,%s, len %d\n",
			prefix, "Raw Address:", RawAddress, "", "HumanAddress", HumanAdr, len(HumanAdr))
		return nil
	}

	ConvertToRaw := func() error {
		pre, raw, err := ConvertPegNetAssetToRaw(HumanAdr)
		if err != nil {
			return err
		}
		if CheckPrefix(pre) != true {
			return errors.New("The Prefix " + pre + " returned by ConvertTo Raw is invalid")
		}
		if !bytes.Equal(raw, RawAddress[:]) {
			return fmt.Errorf("Expected Raw address %x and got %x",
				RawAddress, raw)
		}
		return nil
	}

	if err := ConvertToHuman("PNT"); err != nil {
		t.Error(err)
	}
	if err := ValidatePegNetAssetAddress(HumanAdr); err != nil {
		t.Error(err)
	}
	if err := ConvertToRaw(); err != nil {
		t.Error(err)
	}

	if err := ConvertToHuman("pUSD"); err != nil {
		t.Error(err)
	}
	if err := ValidatePegNetAssetAddress(HumanAdr); err != nil {
		t.Error(err)
	}
	if err := ConvertToRaw(); err != nil {
		t.Error(err)
	}

	if err := ConvertToHuman("pEUR"); err != nil {
		t.Error(err)
	}
	if err := ValidatePegNetAssetAddress(HumanAdr); err != nil {
		t.Error(err)
	}
	if err := ConvertToRaw(); err != nil {
		t.Error(err)
	}

	if err := ConvertToHuman("pYEN"); err == nil {
		t.Error(err)
	}

	if err := ConvertToHuman("pJPY"); err != nil {
		t.Error(err)
	}
	if err := ValidatePegNetAssetAddress(HumanAdr); err != nil {
		t.Error(err)
	}
	if err := ConvertToRaw(); err != nil {
		t.Error(err)
	}

	if err := ConvertToHuman("PNT"); err != nil {
		t.Error(err)
	}
	if err := ValidatePegNetAssetAddress(HumanAdr); err != nil {
		t.Error(err)
	}
	if err := ConvertToRaw(); err != nil {
		t.Error(err)
	}

	if err := ConvertToHuman("pFCT"); err != nil {
		t.Error(err)
	}
	if err := ValidatePegNetAssetAddress(HumanAdr); err != nil {
		t.Error(err)
	}
	if err := ConvertToRaw(); err != nil {
		t.Error(err)
	}

	if err := ConvertToHuman("USD"); err == nil {
		t.Error(err)
	}
	if err := ConvertToHuman("EUR"); err == nil {
		t.Error(err)
	}
	if err := ConvertToHuman("YEN"); err == nil {
		t.Error(err)
	}
	if err := ConvertToHuman("pPNT"); err == nil {
		t.Error(err)
	}

	setAdr("2222222222222222222222222222222222222222222222222222222222222222")

	if err := ConvertToHuman("PNT"); err != nil {
		t.Error(err)
	}
	if err := ValidatePegNetAssetAddress(HumanAdr); err != nil {
		t.Error(err)
	}
	if err := ConvertToRaw(); err != nil {
		t.Error(err)
	}

	setAdr("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")

	if err := ConvertToHuman("PNT"); err != nil {
		t.Error(err)
	}
	if err := ValidatePegNetAssetAddress(HumanAdr); err != nil {
		t.Error(err)
	}
	if err := ConvertToRaw(); err != nil {
		t.Error(err)
	}
}

func TestFloatTruncate(t *testing.T) {
	type Expected struct {
		V   float64
		Exp int64
	}

	testingfunc := func(t *testing.T, e Expected) {
		if i := TruncateFloat(e.V); i != e.Exp {
			t.Errorf("Exp %d, found %d with %f", i, e.Exp, e.V)
		}
	}

	testingfunc(t, Expected{V: 0.1, Exp: 0})
	testingfunc(t, Expected{V: 0.2, Exp: 0})
	testingfunc(t, Expected{V: 0.3, Exp: 0})
	testingfunc(t, Expected{V: 0.4, Exp: 0})
	testingfunc(t, Expected{V: 0.5, Exp: 0})
	testingfunc(t, Expected{V: 0.6, Exp: 0})
	testingfunc(t, Expected{V: 0.7, Exp: 0})
	testingfunc(t, Expected{V: 0.8, Exp: 0})
	testingfunc(t, Expected{V: 0.9, Exp: 0})

	testingfunc(t, Expected{V: 15.6, Exp: 15})
	testingfunc(t, Expected{V: 22.49, Exp: 22})

	for i := 0; i < 100; i++ {
		testingfunc(t, Expected{V: rand.Float64(), Exp: 0})
	}
}
