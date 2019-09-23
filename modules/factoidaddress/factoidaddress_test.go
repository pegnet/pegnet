package factoidaddress_test

import (
	"bytes"
	"testing"

	. "github.com/pegnet/pegnet/modules/factoidaddress"
)

var valids = []string{
	"FA3VAprcnL8pdgDmoSvJbMjsX4DoHx814jeen31n7S9Md8aGwiKX",
	"FA3WWfn5mjDMY1ueTgLouJ6rCBUwATmB1hodBExF3Pu2qYsS4FG8",
	"FA36vN8pVt11RXkvqKqME3GJMkcXZPqgX74vgAqG86dvFvnTFV33",
	"FA3PG5HjCD51Vt28PYrVgcqDD9thbCiuZ6of5caPLA1xcWuvcDjJ",
	"FA32i2GjZWtDv5qT5iK2k1GFja3GRSkwqyyNixx1uVmQ7zmxmFJn",
	"FA2VxtLRw5FqQ7xYoTmg14VRsuZC5zZSPeCsLWbVWrLdpF2abt21",
}

func TestValid(t *testing.T) {
	for _, addr := range valids {
		if err := Valid(addr); err != nil {
			t.Errorf("%s is valid, but found %s", addr, err.Error())
		}
	}

	invalids := []string{
		"FA3VAprcnL8pdgDmoSvJbMjsX4DoHx814jeen31n7S9Md8aGwiKa",  // Bad checksum
		"FA3WWfn5mjDMY1ueTgLouJ6rCBUATmB1hodBExF3Pu2qYsS4FG8",   // Bad length : Too short
		"FA36vN8pVt11RXkvqKqME3GJMkcXZPqgX74vgAqG86dvFvnTFV333", // Bad length : Too long
		"FA2PG5HjCD51Vt28PYrVgcqDD9thbCiuZ6of5caPLA1xcWuvcDjJ",  // Bad checksum
		"FA", // Bad length
		"FA_2VxtLRw5FqQ7xYoTmg14VRsuZC5zZSPeCsLWbVWrLdpF2abt21", // Invalid characters
		"Fs2aStyP44soHQrwRyjEnRsH37tJc9fAPUWwCbAyyooEaS9NNorN",  // Bad Prefix
		"xyz", // Not even close
		"Fs2aStyP44soHQrwRyjEnRsH37tJc9fAPUWwCbAyyooEaS9NNorN--", // Invalid characters
		"", // Not anything
		"EC3TsJHUs8bzbbVnratBafub6toRYdgzgbR7kWwCW4tqbmyySRmg",
		"Es2XT3jSxi1xqrDvS5JERM3W3jh1awRHuyoahn3hbQLyfEi1jvbq",
	}

	for _, addr := range invalids {
		if err := Valid(addr); err == nil {
			t.Errorf("%s is invalid, but found to be valid", addr)
		}
	}

}

func TestRandom(t *testing.T) {
	for i := 0; i < 100; i++ {
		if err := Valid(Random()); err != nil {
			t.Error(err)
		}
	}
}

// TestEncode checks encoding against a set of valid factoid addresses
func TestEncode(t *testing.T) {
	for _, v := range valids {
		data := Base58Decode(v)
		if addr, err := Encode(data[2:34]); err != nil {
			t.Error(err)
		} else if addr != v {
			t.Errorf("exp %s, found %s", v, addr)
		}
	}

	// Might as well check the byte length check
	for i := 0; i < 40; i++ {
		if i == 32 { // RCD length
			continue
		}

		rcd := make([]byte, i)
		_, err := Encode(rcd)
		if err == nil {
			t.Errorf("exp data with length %d to fail, but it did not", i)
		}
	}
}

// TestChecksum checks the checksum against a set of valid factoid addresses
func TestChecksum(t *testing.T) {
	for _, v := range valids {
		data := Base58Decode(v)
		if checksum, err := Checksum(data[:34]); err != nil {
			t.Error(err)
		} else if bytes.Compare(data[34:], checksum) != 0 {
			t.Errorf("exp %s, found %s", Base58Encode(data[34:]), Base58Encode(checksum))
		}
	}

	// Might as well check the byte length check
	for i := 0; i < 40; i++ {
		if i == 34 { // Addr length without checksum
			continue
		}

		data := make([]byte, i)
		_, err := Checksum(data)
		if err == nil {
			t.Errorf("exp data with length %d to fail, but it did not", i)
		}
	}
}
