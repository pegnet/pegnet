package common

import (
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	"github.com/FactomProject/btcutil/base58"
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

	if err := CheckAdr("MainNet", "EC2BURNFCT2PEGNETooo1oooo1oooo1oooo1oooo1oooo1CCCCCC"); err != nil {
		t.Error(err)
	}
	if err := CheckAdr("TestNet", "EC2BURNFCT2TESTxoooo1oooo1oooo1oooo1oooo1oooo1CCCCCC"); err != nil {
		t.Error(err)
	}

	if err := CheckAdr("MainNet", "EC3BURNFCT2PEGNETooo1oooo1oooo1oooo1oooo1oooo1CCCCCC"); err == nil {
		t.Error(err)
	} else {
		fmt.Println("        no match")
	}
	if err := CheckAdr("TestNet", "EC2BURNFCT2TESTxooooiooooiooooiooooiooooiooooiCCCCCC"); err == nil {
		t.Error(err)
	} else {
		fmt.Println("        no match")
	}

}
