package opr_test

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/pegnet/pegnet/common"

	. "github.com/pegnet/pegnet/modules/opr"
)

// TestRandomOPR verifies random oprs that are created are valid
//	Otherwise all tests using these random oprs might be failing
//	because this function is broken.
func TestRandomOPR(t *testing.T) {
	t.Run("version 1", func(t *testing.T) {
		testRandomOPR(t, 1)
	})
	t.Run("version 2", func(t *testing.T) {
		testRandomOPR(t, 2)
	})
}

func testRandomOPR(t *testing.T, version uint8) {
	for i := 0; i < 20; i++ {
		opr := RandomOPR(version)
		if err := opr.Validate(opr.Dbht); err != nil {
			t.Error(err)
		}

		// Verify it can be invalid.
		if err := opr.Validate(opr.Dbht + 1); err == nil {
			t.Error("Wrong dbht, we should error")
		}

		// Verify we can grab the shorthash if we want the winners for it
		if len(opr.ShortEntryHash()) != 16 {
			t.Errorf("exp a short hash")
		}
	}
}

// Test the marshal function returns the same opr after an unmarshal
func TestOPR_SafeMarshal(t *testing.T) {
	t.Run("version 1", func(t *testing.T) {
		testOPR_SafeMarshal(t, 1)
	})

	t.Run("version 2", func(t *testing.T) {
		testOPR_SafeMarshal(t, 2)
	})
}

func testOPR_SafeMarshal(t *testing.T, version uint8) {
	// Test valid marshalling is consistent with itself
	for i := 0; i < 20; i++ {
		o := RandomOPR(version)
		data, err := o.SafeMarshal()
		if err != nil {
			t.Error(err)
		}

		o2 := new(OPR)

		// Set the unmarshaled fields
		o2.EntryHash = o.EntryHash
		o2.Nonce = o.Nonce
		o2.SelfReportedDifficulty = o.SelfReportedDifficulty
		o2.Version = o.Version
		o2.OPRHash = o.OPRHash
		o2.Grade = o.Grade
		o2.Difficulty = o.Difficulty

		// Unmarshal
		err = o2.SafeUnmarshal(data)
		if err != nil {
			t.Error(err)
		}

		data2, err := o2.SafeMarshal()
		if err != nil {
			t.Error(err)
		}

		if bytes.Compare(data, data2) != 0 {
			t.Error("Marshalled content changed")
		}

		if !reflect.DeepEqual(o, o2) {
			for k, v := range o.Assets.AssetList {
				if o2.Assets.AssetList[k] != v {
					fmt.Println("\t Diff: ", k, v, o2.Assets.AssetList[k])
				}
			}
			t.Error(common.DetailErrorCallStack(fmt.Errorf("resulting opr is different"), 2))
		}
	}

	// --------------
	// Testing Marshal fail conditions

	// Testing bad Unmarshals
	// Verify invalid data does not crash
	shouldErrUnmarshal := func(data []byte, reason string, version uint8) {
		o := new(OPR)
		o.Version = version // Must be set before unmarshal
		err := o.SafeUnmarshal(data)
		if err == nil {
			t.Error(common.DetailErrorCallStack(fmt.Errorf("expected error, got none: %s", reason), 2))
		}
	}

	// Get some good content we can use
	tmp := RandomOPR(version)
	goodContent, err := tmp.SafeMarshal()
	if err != nil {
		t.Error(err)
	}

	o := new(OPR)
	o.Version = version
	err = o.SafeUnmarshal(goodContent)
	if err != nil {
		t.Error(err)
	}

	shouldErrUnmarshal(nil, "nil content", version)
	shouldErrUnmarshal([]byte{0x00}, "bad content", version)
	shouldErrUnmarshal(append(goodContent, 0x00), "bad extra content", version)
	shouldErrUnmarshal(goodContent[:len(goodContent)/2], "incomplete content", version)
	shouldErrUnmarshal(goodContent, "bad version", 0)
	shouldErrUnmarshal(goodContent, "bad version", 3)

	// Testing bad Marshal conditions
	shouldErrMarshal := func(o *OPR, reason string) {
		_, err := o.SafeMarshal()
		if err == nil {
			t.Error(common.DetailErrorCallStack(fmt.Errorf("expected error, got none: %s", reason), 2))
		}
	}

	bV := RandomOPR(version)
	bV.Version = 0 // BadVersion

	nA := RandomOPR(version)
	nA.Assets = nil // NilAssets

	bA := RandomOPR(version)
	bA.Assets.AssetList["PNT"] = 0 // BadAsset

	shouldErrMarshal(bV, "bad version")
	bV.Version = 3
	shouldErrMarshal(bV, "bad version")
	shouldErrMarshal(nA, "nil assets")
	shouldErrMarshal(bA, "pnt asset in assets")
}

// Testing some basic validation rules
// TODO: Should probably make these tests a bit more thorough, as these come from the
// 		blockchain, so we should never panic
func TestOPR_Validate(t *testing.T) {
	t.Run("version 1", func(t *testing.T) {
		testOPR_Validate(t, 1)
	})

	t.Run("version 2", func(t *testing.T) {
		testOPR_Validate(t, 2)
	})
}

func testOPR_Validate(t *testing.T, version uint8) {
	// Just make sure it's actually valid before we modify it
	randomValidOPR := func() *OPR {
		o := RandomOPR(version)
		// If it is not valid before we change it, fail immediately
		if err := o.Validate(o.Dbht); err != nil {
			t.Error(err)
			t.FailNow()
		}
		return o
	}

	shouldErrValidate := func(o *OPR, dbht int32, reason string) {
		err := o.Validate(dbht)
		if err == nil {
			t.Error(common.DetailErrorCallStack(fmt.Errorf("expected error, got none: %s", reason), 2))
		}
	}

	// Test bad asset list
	o := randomValidOPR()
	delete(o.Assets.AssetList, "PEG")
	shouldErrValidate(o, o.Dbht, "missing PEG")

	// Test nil asset list
	o.Assets = nil
	shouldErrValidate(o, o.Dbht, "missing assetlist")

	// Test a 0 asset
	o = randomValidOPR()
	o.Assets.AssetList["XBT"] = 0
	shouldErrValidate(o, o.Dbht, "XBT is 0")

	// Test nil winners
	o = randomValidOPR()
	o.WinPreviousOPR = nil
	shouldErrValidate(o, o.Dbht, "missing winners")

	// Test bad length winners
	o = randomValidOPR()
	o.WinPreviousOPR = o.WinPreviousOPR[1:]
	shouldErrValidate(o, o.Dbht, "bad length winners")

	// Bad version
	o = randomValidOPR()
	o.Version = 0
	shouldErrValidate(o, o.Dbht, "bad version")
	o.Version = 3
	shouldErrValidate(o, o.Dbht, "bad version")

	// Bad Height
	o = randomValidOPR()
	shouldErrValidate(o, o.Dbht+1, "bad version")
	shouldErrValidate(o, o.Dbht-1, "bad version")
	shouldErrValidate(o, 0, "bad version")
	shouldErrValidate(o, o.Dbht*-1, "bad version")

	if version == 2 {
		o = randomValidOPR()
		o.FactomDigitalID = "-_"
		shouldErrValidate(o, o.Dbht, "bad identity")

		o = randomValidOPR()
		o.CoinbaseAddress = ""
		shouldErrValidate(o, o.Dbht, "bad coinbase address")

		o.CoinbaseAddress = "Fs2zso56oYwpoYBipxuW6mCrmNmzdduDGcY6xdTy7Rmrpbo54mSZ"
		shouldErrValidate(o, o.Dbht, "bad coinbase address")

		// TODO: Add more improper address checking when the full address check is in place
	}

}

// TestOPR_GetTokens is fairly trivial. It is not verifying much beyond getting the right set of
// assets for a version
func TestOPR_GetTokens(t *testing.T) {
	t.Run("version 1", func(t *testing.T) {
		testOPR_GetTokens(t, 1)
	})

	t.Run("version 2", func(t *testing.T) {
		testOPR_GetTokens(t, 2)
	})
}

func testOPR_GetTokens(t *testing.T, version uint8) {
	o := RandomOPR(version)
	assets := common.AssetsForVersion(version)
	tokens := o.GetTokens()
	if len(tokens) != len(assets) {
		t.Errorf("exp %d tokens, found %d", len(assets), len(tokens))
	}

	for i, token := range tokens {
		if token.Code != assets[i] {
			t.Errorf("tokens out of order")
		}
	}
}

// TODO: Make a ValidFCTAddress function more thorough
func TestValidFCTAddress(t *testing.T) {
	tfa := func(addr string, valid bool, reason string) {
		if v := ValidFCTAddress(addr); v != valid {
			t.Errorf("Valid: %t, exp %t: %s", v, valid, reason)
		}
	}

	tfa("FA2vP7vAyDBmBBhdWqRPyM9W2WGqPYeAoMcG7QtNQb2TY6MKpanu", true, "valid addr")
	tfa("FA2DSjsRoKEyHnmLg6BzCUg9tRpS1Hod62aEV8Gdf5sU9hesrRZc", true, "valid addr")
	tfa("FA2AvQRG58jPtGAkRiXsajWFQvWo5VWA31ds7neG95cLJtACiiw7", true, "valid addr")

	//tfa("FA2vP7vAyDBmBBhdWqRPyM9W2WGqPYeAoMcG7QtNQb2TY6MKpana", false, "bad checksum")
	//tfa("FA2DSjsRoKEyHnmLg6BzCUg9tRpS1Hod62aEV8Gdf5aU9hesrRZc", false, "bad checksum")

	tfa("Fs2Uk1vnk2JrHHQXTDvSW6LsRTFqfim4khBk2yKHU4MWSYSnQCcg", false, "not a FA key")
	tfa("Es2XT3jSxi1xqrDvS5JERM3W3jh1awRHuyoahn3hbQLyfEi1jvbq", false, "not a FA key")
	tfa("EC3TsJHUs8bzbbVnratBafub6toRYdgzgbR7kWwCW4tqbmyySRmg", false, "not a FA key")

	tfa("", false, "empty")
	//tfa("FA", false, "not long enough")
	//tfa("FAs", false, "not long enough")
	//tfa("FAAvQRG58jPtGAkRiXsajWFQvWo5VWA31ds7neG95cLJtACiiw7", false, "missing a character")
	//tfa("FA2DSjsRoKEyHnmLg6BzCUg9tRpS1Hod62aEV8Gdf5sU9hesrRZ_", false, "not base 58")
}

// Ensure we keep under 1EC
func TestUnder1EC(t *testing.T) {
	t.Run("version 1", func(t *testing.T) {
		test1EC(t, 1)
	})

	t.Run("version 2", func(t *testing.T) {
		test1EC(t, 2)
	})
}

func test1EC(t *testing.T, version uint8) {
	opr := RandomOPR(version)
	PopulateRandomWinners(opr)

	// Because this allows our prices to be wildly high, that will add bytes
	// Let's make the prices resonable
	mod := uint64(10000 * 1e8) // 10k Max
	for k, v := range opr.Assets.AssetList {
		opr.Assets.AssetList[k] = v % mod
	}

	tl := 0
	for _, ext := range opr.ExtIDs() {
		tl += len(ext) + 2
	}

	data, err := opr.SafeMarshal()
	if err != nil {
		t.Error(err)
	}

	tl += len(data)

	if tl > 1024 {
		t.Errorf("opr entry is over 1kb, found %d bytes", tl)
	}
	fmt.Println(tl)
}

func TestOPR_ShortEntryHash(t *testing.T) {
	o := new(OPR)
	if o.ShortEntryHash() != "" {
		t.Errorf("exp empty short hash")
	}

	o.EntryHash = make([]byte, 32, 32)
	if o.ShortEntryHash() != "0000000000000000" {
		t.Errorf("exp all 0 short hash")
	}
}
