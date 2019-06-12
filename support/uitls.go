package support

import (
	"fmt"
	"strings"

	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/FactomProject/btcutil/base58"
)

type NetworkType int

const (
	MAIN_NETWORK NetworkType = iota + 1
	TEST_NETWORK
)

var PegAssetNames = []string{
	"pPNT",
	"pUSD",
	"pEUR",
	"pJPY",
	"pGBP",
	"pCAD",
	"pCHF",
	"pINR",
	"pSGD",
	"pCNY",
	"pHKD",
	"pXAU",
	"pXAG",
	"pXPD",
	"pXPT",
	"pXBT",
	"pETH",
	"pLTC",
	"pXBC",
	"pFCT",
}

var TestPegAssetNames = []string{
	"tPNT",
	"tUSD",
	"tEUR",
	"tJPY",
	"tGBP",
	"tCAD",
	"tCHF",
	"tINR",
	"tSGD",
	"tCNY",
	"tHKD",
	"tXAU",
	"tXAG",
	"tXPD",
	"tXPT",
	"tXBT",
	"tETH",
	"tLTC",
	"tXBC",
	"tFCT",
}

func PullValue(line string, howMany int) string {
	i := 0
	//fmt.Println(line)
	var pos int
	for i < howMany {
		//find the end of the howmany-th tag
		pos = strings.Index(line, ">")
		line = line[pos+1:]
		//fmt.Println(line)
		i = i + 1
	}
	//fmt.Println("line:", line)
	pos = strings.Index(line, "<")
	//fmt.Println("POS:", pos)
	line = line[0:pos]
	//fmt.Println(line)
	return line
}

// CheckPrefix()
// Check the prefix for either network type.
func CheckPrefix(network NetworkType, name string) bool {
	if network == MAIN_NETWORK {
		for _, v := range PegAssetNames {
			if v == name {
				return true
			}
		}
	} else {
		for _, v := range TestPegAssetNames {
			if v == name {
				return true
			}
		}
	}
	return false

}

// ConvertRawAddrToPegT()
// Converts a raw RCD1 address into a wallet friendly address that can be used to
// convert assets, check balances, and send tokens.  While the underlying private key can be
// used to hold Factoids or any token in the PegNet, users need addresses that create a
// barrier to mistakes that can lead to sending the wrong tokens to the wrong addresses
func ConvertRawAddrToPegT(network NetworkType, prefix string, adr [32]byte) (string, error) {

	// Make sure the prefix is valid.
	if !CheckPrefix(network, prefix) {
		return "", errors.New(prefix + " is not a valid PegNet prefix")
	}

	// Append the prefix to the base 58 representation of the raw address
	b58 := prefix + base58.Encode(adr[:])
	// Compute the double sha258 of the resulting string
	hash := sha256.Sum256([]byte(b58))
	hash = sha256.Sum256(hash[:])
	fmt.Printf("SHA256d %x\n", hash)
	// Use the high order 4 bytes of the hash as a checksum, convert that 4 bytes to a string
	chksum := hex.EncodeToString(hash[:4])
	// Add the checksum to the end, and that is the human readable address
	b58 = b58 + chksum

	return b58, nil
}

// ConvertPegTAddrToRaw()
// Convert a human/wallet address to the raw underlying address.  Verifies the checksum and
// the validity of the prefix.  Returns the prefix, the raw address, and error.
//
func ConvertPegTAddrToRaw(network NetworkType, adr string) (prefix string, rawAdr []byte, err error) {
	adrLen := len(adr)
	if adrLen < 44 || len(adr) > 56 {
		return "", nil, errors.New(
			fmt.Sprintf("valid pegNet token addresses are 44 to 56 characters in length. len(adr)=%d ", adrLen))
	}

	prefix = adr[:4]
	if !CheckPrefix(network, prefix) {
		return "", nil, errors.New(prefix + " is not a valid PegNet prefix")
	}
	b58 := adr[4 : adrLen-8]
	raw := base58.Decode(b58)
	if len(raw) == 0 {
		return "", nil, errors.New("invalid base58 encoding")
	}
	hash := sha256.Sum256([]byte(adr[:adrLen-8]))
	hash = sha256.Sum256(hash[:])
	chksum := hex.EncodeToString(hash[:4])
	if chksum != adr[adrLen-8:] {
		return "", nil, errors.New("checksum failure")
	}

	rawAdr = base58.Decode(adr[4 : adrLen-8])

	return prefix, rawAdr, nil

}

// PegTAdrIsValid()
// Check that the given human/wallet PegNet address is valid.
func PegTAdrIsValid(network NetworkType, adr string) error {
	_, _, err := ConvertPegTAddrToRaw(network, adr)
	return err
}

// RandomByteSliceOfLen()
// Returns a random set of bytes of a given length
func RandomByteSliceOfLen(sliceLen int) []byte {
	if sliceLen <= 0 {
		return nil
	}
	answer := make([]byte, sliceLen)
	_, err := rand.Read(answer)
	if err != nil {
		return nil
	}
	return answer
}
