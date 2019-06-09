package utils_test

import (
	"bytes"
	"errors"
	"fmt"
	. "github.com/pegnet/OracleRecord/utils"
	"testing"
)

func TestConvertRawAddrToPegT(t *testing.T) {
	var RawAddress = [32]byte{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
		1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
		1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
		1, 2}

	var HumanAdr string
	var err error

	ConvertToHuman := func(prefix string) error {
		HumanAdr, err = ConvertRawAddrToPegT(MAIN_NETWORK, prefix, RawAddress)
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

	RawAddress = [32]byte{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0,
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

	RawAddress = [32]byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF,
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

}
