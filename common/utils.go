// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full

package common

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/FactomProject/btcutil/base58"
	log "github.com/sirupsen/logrus"
)

var PegAssetNames []string

var TestPegAssetNames []string

func init() {
	for _, asset := range AllAssets {
		if asset != "PEG" {
			PegAssetNames = append(PegAssetNames, "p"+asset)
		} else {
			PegAssetNames = append(PegAssetNames, asset)
		}
		TestPegAssetNames = append(TestPegAssetNames, "t"+asset)
	}
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

func ValidIdentity(identity string) error {
	valid, _ := regexp.MatchString("^[a-zA-Z0-9,]+$", identity)
	if !valid {
		return fmt.Errorf("only alphanumeric characters and commas are allowed in the identity")
	}
	return nil
}

// CheckPrefix()
// Check the prefix for either network type.
func CheckPrefix(name string) bool {
	for _, v := range PegAssetNames {
		if v == name {
			return true
		}
	}
	for _, v := range TestPegAssetNames {
		if v == name {
			return true
		}
	}
	return false

}

// ConvertRawToPegNetAsset()
// Converts a raw RCD1 address into a wallet friendly address that can be used to
// convert assets, check balances, and send tokens.  While the underlying private key can be
// used to hold Factoids or any token in the PegNet, users need addresses that create a
// barrier to mistakes that can lead to sending the wrong tokens to the wrong addresses
func ConvertRawToPegNetAsset(prefix string, adr []byte) (string, error) {

	// Make sure the prefix is valid.
	if !CheckPrefix(prefix) {
		return "", errors.New(prefix + " is not a valid PegNet prefix")
	}

	h := sha256.Sum256([]byte(append(append([]byte(prefix), '_'), adr...)))
	hash := sha256.Sum256(h[:])

	// Append the prefix to the base 58 representation of the raw address
	b58 := prefix + "_" + base58.Encode(append(adr, hash[:4]...))

	return b58, nil
}

func GetPrefix(address string) (length int, prefix string) {
	idx := strings.Index(address, "_")
	if idx < 0 {
		return -1, ""
	}
	return idx, address[:idx]
}

// ConvertPegTAddrToRaw()
// Convert a human/wallet address to the raw underlying address.  Verifies the checksum and
// the validity of the prefix.  Returns the prefix, the raw address, and error.
//
func ConvertPegNetAssetToRaw(adr string) (prefix string, rawAdr []byte, err error) {
	adrLen := len(adr)
	if adrLen < 42 || len(adr) > 56 {
		return "", nil,
			fmt.Errorf("valid pegNet token addresses are 44 to 56 characters in length. len(adr)=%d ", adrLen)
	}
	var prefixLen int
	prefixLen, prefix = GetPrefix(adr)
	if !CheckPrefix(prefix) {
		return "", nil, errors.New(prefix + " is not a valid PegNet prefix")
	}

	b58 := adr[prefixLen+1:]
	raw := base58.Decode(b58)
	if len(raw) == 0 {
		return "", nil, errors.New("invalid base58 encoding")
	}
	rawAdr = raw[:len(raw)-4]
	chksum := raw[len(raw)-4:]

	hash := sha256.Sum256(append(append([]byte(prefix), '_'), rawAdr...))
	hash = sha256.Sum256(hash[:])
	if !bytes.Equal(hash[:4], chksum) {
		return "", nil, errors.New("checksum failure")
	}

	return prefix, rawAdr, nil

}

// ValidatePegNetAssetAddress()
// Check that the given human/wallet PegNet address is valid.
func ValidatePegNetAssetAddress(adr string) error {
	_, _, err := ConvertPegNetAssetToRaw(adr)
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

//  Convert Factoid and Entry Credit addresses to their more user
//  friendly and human readable formats.
//
//  Creates the binary form.  Just needs the conversion to base58
//  for display.
func ConvertRawToFCT(addr []byte) string {
	dat := make([]byte, 0, 64)
	dat = append(dat, fcPubPrefix...)
	dat = append(dat, addr...)
	hash := sha256.Sum256(dat)
	sha256d := sha256.Sum256(hash[:])
	userd := append(dat, sha256d[:4]...)
	return base58.Encode(userd)
}

//  Convert Factoid and Entry Credit addresses to their more user
//  friendly and human readable formats.
//
//  Creates the binary form.  Just needs the conversion to base58
//  for display.
func ConvertRawToEC(addr []byte) string {
	dat := make([]byte, 0, 64)
	dat = append(dat, ecPubPrefix...)
	dat = append(dat, addr...)
	hash := sha256.Sum256(dat)
	sha256d := sha256.Sum256(hash[:])
	userd := append(dat, sha256d[:4]...)
	return base58.Encode(userd)
}

// Convert a User facing Factoid address
// to the raw form.  We do what validation we can here, and
// return an error if the Factoid address is not valid
func ConvertFCTtoRaw(userFAddr string) (raw []byte, err error) {
	if len(userFAddr) != 52 {
		return nil, errors.New("invalid length of a factoid address")
	}
	v := base58.Decode(userFAddr)
	switch {
	case bytes.Equal(v[:2], fcPubPrefix):
	default:
		return nil, errors.New("wrong format for a factoid address")
	}
	rcd := v[:34]
	hash := sha256.Sum256(rcd)
	hash = sha256.Sum256(hash[:])
	cksum := v[34:]
	if !bytes.Equal(hash[:4], cksum[:]) {
		return nil, errors.New("")
	}
	return v[2:34], nil
}

// Convert a User facing Factoid or Entry Credit address
// or their Private Key representations
// to the regular form.  Note validation must be done
// separately!
func ConvertAnyFactomAdrToRaw(userFAddr string) ([]byte, error) {
	v := base58.Decode(userFAddr)
	switch {
	case bytes.Equal(v[:2], fcPubPrefix):
	case bytes.Equal(v[:2], fcSecPrefix):
	case bytes.Equal(v[:2], ecPubPrefix):
	case bytes.Equal(v[:2], ecSecPrefix):
	default:
		return nil, errors.New("unknown prefix")
	}
	return v[2:34], nil
}

// Convert a User facing FCT address to all of its PegNet
// asset token User facing forms.
func ConvertFCTtoAllPegNetAssets(userFctAddr string) (assets []string, err error) {
	raw := base58.Decode(userFctAddr)[2:34]
	cvt := func(asset string) (passet string) {
		passet, err = ConvertRawToPegNetAsset(asset, raw)
		if err != nil {
			panic(err)
		}
		return passet
	}

	for _, asset := range AllAssets {
		pAsset := "p" + asset
		if asset == "PEG" {
			pAsset = "PEG"
		}

		assets = append(assets, cvt(pAsset))
		assets = append(assets, cvt("t"+asset))
	}

	return assets, nil
}

func ConvertFCTtoPegNetAsset(network string, asset string, userFAdr string) (PegNetAsset string, err error) {
	raw, err := ConvertFCTtoRaw(userFAdr)
	if err != nil {
		return "", err
	}

	switch network {
	case TestNetwork:
		PegNetAsset, err = ConvertRawToPegNetAsset("t"+asset, raw)
	case MainNetwork:
		if asset != "PEG" {
			PegNetAsset, err = ConvertRawToPegNetAsset("p"+asset, raw)
		} else {
			PegNetAsset, err = ConvertRawToPegNetAsset(asset, raw)
		}
	}
	if err != nil {
		log.Errorf("Invalid RCD, could not create PEG address")
	}
	return
}

func Abs(v int) int {
	if v < 0 {
		return v * -1
	}
	return v
}

func CheckAndPanic(e error) {
	if e != nil {
		_, file, line, _ := runtime.Caller(1) // The line that called this function
		shortFile := ShortenPegnetFilePath(file, "", 0)
		log.WithField("caller", fmt.Sprintf("%s:%d", shortFile, line)).WithError(e).Fatal("An error was encountered")
	}
}

func DetailError(e error) error {
	_, file, line, _ := runtime.Caller(1) // The line that called this function
	shortFile := ShortenPegnetFilePath(file, "", 0)
	return fmt.Errorf("%s:%d %s", shortFile, line, e.Error())
}

// ShortenPegnetFilePath takes a long path url to pegnet, and shortens it:
//	"/home/billy/go/src/github.com/pegnet/pegnet/opr.go" -> "pegnet/opr.go"
//	This is nice for errors that print the file + line number
//
// 		!! Only use for error printing !!
//
func ShortenPegnetFilePath(path, acc string, depth int) (trimmed string) {
	if depth > 5 || path == "." {
		// Recursive base case
		// If depth > 5 probably no pegnet dir exists
		return filepath.ToSlash(filepath.Join(path, acc))
	}
	dir, base := filepath.Split(path)
	if strings.ToLower(base) == "pegnet" { // Used to be named PegNet. Not everyone changed I bet
		return filepath.ToSlash(filepath.Join(base, acc))
	}

	return ShortenPegnetFilePath(filepath.Clean(dir), filepath.Join(base, acc), depth+1)
}

func FindIndexInStringArray(haystack []string, needle string) int {
	for i, v := range haystack {
		if v == needle {
			return i
		}
	}
	return -1
}
