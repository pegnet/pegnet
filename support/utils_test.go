package support_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	. "github.com/pegnet/pegnet/support"
	"testing"
	"github.com/FactomProject/factoid"
)


func TestConvertFctToPegAssets(t *testing.T) {

	fctAddrs := []string  {
		"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
		"0000000000000000000000000000000000000000000000000000000000000000",
		"1235134346267276876846779695807808006ababababababababdfdfdfeeeeF",
		"abcdef1256267584690ab3472accffeeeeeee154134628697025428890000002",
		"1234651367286539567abababefefefefefefef6228365940abcdefedededede",
	}

	for _,fa := range fctAddrs {
		fmt.Println("\n",fa)
		raw,err := hex.DecodeString(fa)
		if err != nil {
			t.Fatal(fa," ",err)
		}
		fct := ConvertFctAddressToUser(raw)
		fmt.Println(fct)
		raw = factoid.ConvertUserStrToAddress(fct)
		fmt.Printf("%x\n",raw)
		pnt,err := ConvertRawAddrToPeg(MAIN_NETWORK,"pPNT",raw)
		if err != nil {
			t.Fatal(fa," ",err)
		}
		fmt.Println(pnt)
		pre,raw2,err := ConvertPegAddrToRaw(MAIN_NETWORK,pnt)
		if pre != "pPNT" || bytes.Compare(raw,raw2)!=0 || err != nil {
			t.Fatal("Round trip failed with pPNT")
		}
		fmt.Printf("%x\n",raw2)
		tpnt,err := ConvertRawAddrToPeg(TEST_NETWORK,"tPNT",raw)
		if err != nil {
			t.Fatal(fa," ",err)
		}
		fmt.Println(tpnt)
		pre,raw2,err = ConvertPegAddrToRaw(TEST_NETWORK,tpnt)
		if pre != "tPNT" || bytes.Compare(raw,raw2)!=0 || err != nil {
			t.Fatal("Round trip failed with tPNT")
		}
		fmt.Printf("%x\n",raw2)

		assets , err:= ConvertUserFctToUserPegNetAssets(fct)
		if err != nil {
			t.Fatal(fa," ",err)
		}
		for _,asset := range assets {
			fmt.Println(asset)
		}
	}
}


func TestBurnECAddress(t *testing.T) {
	ecAdd := "EC1moooFCT2TESToooo1oooo1oooo1oooo1oooo1oooo1oooo1oo"
	raw,err := ConvertUserStrFctEcToAddress(ecAdd)
	if err != nil {
		t.Fatal(err)
	}
	raw2, _ := hex.DecodeString(raw)
	burn := ConvertECAddressToUser(raw2)
	fmt.Printf("Suggested Address %s\n", ecAdd)
	fmt.Printf("Raw Address       %s\n", raw)
	fmt.Printf("Suggested+csum    %s\n", burn)
	raw,err = ConvertUserStrFctEcToAddress(burn)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Back again        %s\n", raw)
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
		HumanAdr, err = ConvertRawAddrToPeg(MAIN_NETWORK, prefix, RawAddress[:])
		if err != nil {
			return err
		}
		fmt.Printf("%5s %15s,%x\n%5s %15s,%s, len %d\n",
			prefix, "Raw Address:", RawAddress, "", "HumanAddress", HumanAdr, len(HumanAdr))
		return nil
	}

	ConvertToRaw := func() error {
		pre, raw, err := ConvertPegAddrToRaw(MAIN_NETWORK, HumanAdr)
		if err != nil {
			return err
		}
		if CheckPrefix(MAIN_NETWORK, pre) != true {
			return errors.New("The Prefix " + pre + " returned by ConvertTo Raw is invalid")
		}
		if bytes.Compare(raw, RawAddress[:]) != 0 {
			return errors.New(fmt.Sprintf("Expected Raw address %x and got %x",
				RawAddress, raw))
		}
		return nil
	}

	if err := ConvertToHuman("pPNT"); err != nil {
		t.Error(err)
	}
	if err := PegTAdrIsValid(MAIN_NETWORK, HumanAdr); err != nil {
		t.Error(err)
	}
	if err := ConvertToRaw(); err != nil {
		t.Error(err)
	}

	if err := ConvertToHuman("pUSD"); err != nil {
		t.Error(err)
	}
	if err := PegTAdrIsValid(MAIN_NETWORK, HumanAdr); err != nil {
		t.Error(err)
	}
	if err := ConvertToRaw(); err != nil {
		t.Error(err)
	}

	if err := ConvertToHuman("pEUR"); err != nil {
		t.Error(err)
	}
	if err := PegTAdrIsValid(MAIN_NETWORK, HumanAdr); err != nil {
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
	if err := PegTAdrIsValid(MAIN_NETWORK, HumanAdr); err != nil {
		t.Error(err)
	}
	if err := ConvertToRaw(); err != nil {
		t.Error(err)
	}

	if err := ConvertToHuman("pPNT"); err != nil {
		t.Error(err)
	}
	if err := PegTAdrIsValid(MAIN_NETWORK, HumanAdr); err != nil {
		t.Error(err)
	}
	if err := ConvertToRaw(); err != nil {
		t.Error(err)
	}

	if err := ConvertToHuman("pFCT"); err != nil {
		t.Error(err)
	}
	if err := PegTAdrIsValid(MAIN_NETWORK, HumanAdr); err != nil {
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
	if err := ConvertToHuman("PNT"); err == nil {
		t.Error(err)
	}

	setAdr("2222222222222222222222222222222222222222222222222222222222222222")

	if err := ConvertToHuman("pPNT"); err != nil {
		t.Error(err)
	}
	if err := PegTAdrIsValid(MAIN_NETWORK, HumanAdr); err != nil {
		t.Error(err)
	}
	if err := ConvertToRaw(); err != nil {
		t.Error(err)
	}

	setAdr("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")

	if err := ConvertToHuman("pPNT"); err != nil {
		t.Error(err)
	}
	if err := PegTAdrIsValid(MAIN_NETWORK, HumanAdr); err != nil {
		t.Error(err)
	}
	if err := ConvertToRaw(); err != nil {
		t.Error(err)
	}

}
