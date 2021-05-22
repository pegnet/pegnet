package common

import (
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	"github.com/FactomProject/btcutil/base58"
	"github.com/FactomProject/factom"
)

func TestBurnAddresses(t *testing.T) {

	// Note that this check takes the arbitrary part of a burn address, calculates the
	// proper checksum, ensures everything encodes properly, and makes sure the BurnAddresses
	// Map has the right burn address, and the right underlying RCDs.
	CheckAdr := func(key, address string) error {
		b58 := base58.Decode(address)
		rcd := b58[2:34]
		ec := ConvertRawToEC(rcd)
		fmt.Printf("\n%6s: EC  address %s\n", key, ec)
		fmt.Printf("%6s  RCD address %x\n", "", rcd)
		if BurnAddresses[key] != ec {
			return errors.New(key + " burn address is not correct")
		}
		if BurnAddresses[key+"RCD"] != hex.EncodeToString(rcd) {
			return errors.New(key + " rcd is not correct")
		}
		return nil
	}

	if err := CheckAdr(MainNetwork, "EC2BURNFCT2PEGNETooo1oooo1oooo1oooo1oooo1oooo1CCCCCC"); err != nil {
		t.Error(err)
	}
	if err := CheckAdr(TestNetwork, "EC2BURNFCT2TESTxoooo1oooo1oooo1oooo1oooo1oooo1CCCCCC"); err != nil {
		t.Error(err)
	}

	if err := CheckAdr(MainNetwork, "EC3BURNFCT2PEGNETooo1oooo1oooo1oooo1oooo1oooo1CCCCCC"); err == nil {
		t.Error(err)
	} else {
		fmt.Println("        no match")
	}
	if err := CheckAdr(TestNetwork, "EC2BURNFCT2TESTxooooiooooiooooiooooiooooiooooiCCCCCC"); err == nil {
		t.Error(err)
	} else {
		fmt.Println("        no match")
	}
}

// TestLoadConfigNetwork tests when we change the testnet title, that using "TestNet" still works
func TestLoadConfigNetwork(t *testing.T) {
	c := NewUnitTestConfig()
	if n, err := LoadConfigNetwork(c); err != nil || n != UnitTestNetwork {
		t.Error("LoadConfigNetwork Failed")
	}

	type NetTest struct {
		Vector string
		Exp    string
	}
	vects := []NetTest{
		{Vector: "TestNet", Exp: TestNetwork},
		{Vector: "testnet", Exp: TestNetwork},
		{Vector: "TESTNET", Exp: TestNetwork},

		{Vector: "MainNet", Exp: MainNetwork},
		{Vector: "mainnet", Exp: MainNetwork},
		{Vector: "MAINNET", Exp: MainNetwork},
	}

	for _, v := range vects {
		if n, err := GetNetwork(v.Vector); err != nil || n != v.Exp {
			t.Error("GetNetwork Failed")
		}
	}

	// Test an invalid
	_, err := GetNetwork("random")
	if err == nil {
		t.Error("getNetwork expected an error")
	}
}

func TestPegnetBurnAddress(t *testing.T) {
	addr := PegnetBurnAddress(MainNetwork)
	if !factom.IsValidAddress(addr) {
		t.Error("Burn addr is not valid")
	}

	addr = PegnetBurnAddress(TestNetwork)
	if !factom.IsValidAddress(addr) {
		t.Error("Burn addr is not valid")
	}
}

func TestFormatDiff(t *testing.T) {
	type Format struct {
		Vector    uint64
		Precision uint
		Exp       string
	}
	vects := []Format{
		{1e8, 0, "1e+08"},
		{1e8, 1, "1.0e+08"},
		{1e8, 2, "1.00e+08"},
		{123456789, 1, "1.2e+08"},
		{123456789, 2, "1.23e+08"},
		{123456789, 3, "1.235e+08"}, // Rounds up
		{123456789, 8, "1.23456789e+08"},
		{1, 8, "1.00000000e+00"},
	}

	for _, v := range vects {
		if s := FormatDiff(v.Vector, v.Precision); s != v.Exp {
			t.Errorf("Exp %s, got %s", v.Exp, s)
		}
	}
}

func TestFormatGrade(t *testing.T) {
	type Format struct {
		Vector    float64
		Precision uint
		Exp       string
	}
	vects := []Format{
		{1e8, 0, "1e+08"},
		{1e8, 1, "1.0e+08"},
		{1e8, 2, "1.00e+08"},
		{123456789, 1, "1.2e+08"},
		{123456789, 2, "1.23e+08"},
		{123456789, 3, "1.235e+08"}, // Rounds up
		{123456789, 8, "1.23456789e+08"},
		{1, 8, "1.00000000e+00"},
		{0.1, 8, "1.00000000e-01"},
		{0.01, 8, "1.00000000e-02"},
		{0.001, 8, "1.00000000e-03"},
	}

	for _, v := range vects {
		if s := FormatGrade(v.Vector, v.Precision); s != v.Exp {
			t.Errorf("Exp %s, got %s", v.Exp, s)
		}
	}
}
