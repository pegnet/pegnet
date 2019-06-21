package support_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	. "github.com/pegnet/pegnet/support"
	"testing"
)

//func TestConvertFctToPegAssets(t *testing.T) {
//
//	// Create a function that takes a string address, and makes sure we can run it back
//	// and forth through address conversions without a failure.
//	checkAdr := func(FctAddress string) {
//		rawFct := ConvertFctECUserStrToAddress(FctAddress)
//
//		addrs,err := ConvertUserFctToUserPegNetAssets(FctAddress)
//		if err != nil {
//			t.Error(err)
//		}
//		fmt.Println("FCT Address:\n", FctAddress)
//		fmt.Printf("FCT Raw:\n%x\n",rawFct)
//		for i, adr := range addrs {
//			if i&1 == 0 {
//				println()
//			}
//			fmt.Println(adr)
//			raw := ConvertPegTAddrToRaw()
//		}
//	}
//
//	checkAdr( "FA2L6Vng4jBMbbDZtYLsxKQbAAin4Rxg2CgvnyzXrwENSK1t2QUx")
//
//
//
//	adr,_ := hex.DecodeString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
//
//	fct := ConvertFctAddressToUser(adr)
//	addrs,err = ConvertUserFctToUserPegNetAssets(fct)
//	fmt.Println("FCT Address:\n",FctAddress)
//	for i,adr := range addrs {
//		if i & 1 == 0 {
//			println()
//		}
//		fmt.Println(adr)
//	}
//
//	adr,_ = hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
//
//	fct = ConvertFctAddressToUser(adr)
//	addrs,err = ConvertUserFctToUserPegNetAssets(fct)
//	fmt.Println("FCT Address:\n",FctAddress)
//	for i,adr := range addrs {
//		if i & 1 == 0 {
//			println()
//		}
//		fmt.Println(adr)
//	}
//
//}


func TestBurnECAddress(t *testing.T) {
	ecAdd := "EC1moooFCT2TESToooo1oooo1oooo1oooo1oooo1oooo1oooo1oo"
	raw := ConvertFctECUserStrToAddress(ecAdd)
	raw2, _ := hex.DecodeString(raw)
	burn := ConvertECAddressToUser(raw2)
	fmt.Printf("Suggested Address %s\n", ecAdd)
	fmt.Printf("Raw Address       %s\n", raw)
	fmt.Printf("Suggested+csum    %s\n", burn)
	raw = ConvertFctECUserStrToAddress(burn)
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
		pre, raw, err := ConvertPegTAddrToRaw(MAIN_NETWORK, HumanAdr)
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
