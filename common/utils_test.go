// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full
package common_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
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
		peg, err := ConvertRawToPegNetAsset("PEG", raw)
		if err != nil {
			t.Fatal(fa, " ", err)
		}
		fmt.Println(peg)
		pre, raw2, err := ConvertPegNetAssetToRaw(peg)
		if pre != "PEG" || !bytes.Equal(raw, raw2) || err != nil {
			t.Fatal("Round trip failed with PEG")
		}
		fmt.Printf("%x\n", raw2)
		tpeg, err := ConvertRawToPegNetAsset("tPEG", raw)
		if err != nil {
			t.Fatal(fa, " ", err)
		}
		fmt.Println(tpeg)
		pre, raw2, err = ConvertPegNetAssetToRaw(tpeg)
		if pre != "tPEG" || !bytes.Equal(raw, raw2) || err != nil {
			t.Fatal("Round trip failed with tPEG")
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

	if err := ConvertToHuman("PEG"); err != nil {
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

	if err := ConvertToHuman("PEG"); err != nil {
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
	if err := ConvertToHuman("pPEG"); err == nil {
		t.Error(err)
	}

	setAdr("2222222222222222222222222222222222222222222222222222222222222222")

	if err := ConvertToHuman("PEG"); err != nil {
		t.Error(err)
	}
	if err := ValidatePegNetAssetAddress(HumanAdr); err != nil {
		t.Error(err)
	}
	if err := ConvertToRaw(); err != nil {
		t.Error(err)
	}

	setAdr("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")

	if err := ConvertToHuman("PEG"); err != nil {
		t.Error(err)
	}
	if err := ValidatePegNetAssetAddress(HumanAdr); err != nil {
		t.Error(err)
	}
	if err := ConvertToRaw(); err != nil {
		t.Error(err)
	}

}

func TestAbs(t *testing.T) {
	for i := -100; i < 100; i++ {
		if a := Abs(i); a < 0 {
			t.Errorf("a is less than 00")
		}
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

func TestFindIndexInStringArray(t *testing.T) {
	a := "abcdefghijklmnopqrstuvwxyz"
	var arr []string
	for _, l := range a {
		arr = append(arr, string(l))
	}

	for i, l := range a {
		if FindIndexInStringArray(arr, string(l)) != i {
			t.Error("did not find at right location")
		}
	}
}
