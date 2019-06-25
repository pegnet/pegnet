// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package opr

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/FactomProject/btcutil/base58"
	"github.com/FactomProject/factom"
	"github.com/pegnet/LXR256"
	"github.com/pegnet/pegnet/polling"
	"github.com/pegnet/pegnet/support"
	"github.com/zpatrick/go-config"
	"strings"
	"time"
)

type OraclePriceRecord struct {
	// These fields are not part of the OPR, but track values associated with the OPR.
	Config     *config.Config    `json:"-"` //  The config of the miner using the record
	Difficulty uint64            `json:"-"` // The difficulty of the given nonce
	Grade      float64           `json:"-"` // The grade when OPR records are compared
	OPRHash    []byte            `json:"-"` // The hash of the OPR record (used by pegnetMining)
	EC         *factom.ECAddress `json:"-"` // Entry Credit Address used by a miner
	Entry      *factom.Entry     `json:"-"` // Entry to record this record
	EntryHash  string            `json:"-"` // Entry Hash is communicated here in base58
	StopMining chan int          `json:"-"` // Bool that stops pegnetMining this OPR

	// These values define the context of the OPR, and they go into the PegNet OPR record, and are mined.
	OPRChainID         string     `json:"oprchainid"` // [base58]  Chain ID of the chain used by the Oracle Miners
	Dbht               int32      `json:"dbht"`       //           The Directory Block Height of the OPR.
	WinningPreviousOPR [10]string `json:"winners"`    // [base58]  Winning OPR entries in the previous block
	CoinbasePNTAddress string     `json:"coinbase"`   // [base58]  PNT Address to pay PNT
	FactomDigitalID    []string   `json:"did"`        // [unicode] Digital Identity of the miner

	// The Oracle values of the OPR, they are the meat of the OPR record, and are mined.
	PNT float64
	USD float64
	EUR float64
	JPY float64
	GBP float64
	CAD float64
	CHF float64
	INR float64
	SGD float64
	CNY float64
	HKD float64
	XAU float64
	XAG float64
	XPD float64
	XPT float64
	XBT float64
	ETH float64
	LTC float64
	XBC float64
	FCT float64
}

var LX lxr.LXRHash
var OPRChainID string

func init() {
	LX.Init(0xfafaececfafaecec, 25, 256, 5)
}

type Token struct {
	code  string
	value float64
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// This function cannot validate the winners of the previous block, but it can do some sanity
// checking of the structure and values of the OPR record.
func (opr *OraclePriceRecord) Validate(c *config.Config) bool {

	protocol, err1 := c.String("Miner.Protocol")
	network, err2 := c.String("Miner.Network")
	if err1 != nil || err2 != nil {
		return false
	}

	if len(OPRChainID) == 0 {
		OPRChainID = base58.Encode(support.ComputeChainIDFromStrings([]string{protocol, network, "Oracle Price Records"}))
	}

	if opr.OPRChainID != OPRChainID {
		return false
	}

	ntype := support.INVALID
	switch network {
	case "MainNet":
		ntype = support.MAIN_NETWORK
	case "TestNet":
		ntype = support.TEST_NETWORK
	default:
		return false
	}

	pre, _, err := support.ConvertPegAddrToRaw(ntype, opr.CoinbasePNTAddress)
	if err != nil || pre != "tPNT" {
		return false
	}

	if opr.USD == 0 ||
		opr.EUR == 0 ||
		opr.JPY == 0 ||
		opr.GBP == 0 ||
		opr.CAD == 0 ||
		opr.CHF == 0 ||
		opr.INR == 0 ||
		opr.SGD == 0 ||
		opr.CNY == 0 ||
		opr.HKD == 0 ||
		opr.XAU == 0 ||
		opr.XAG == 0 ||
		opr.XPD == 0 ||
		opr.XPT == 0 ||
		opr.XBT == 0 ||
		opr.ETH == 0 ||
		opr.LTC == 0 ||
		opr.XBC == 0 ||
		opr.FCT == 0 {
		return false
	}

	return true
}

func (opr *OraclePriceRecord) GetTokens() (tokens []Token) {
	tokens = append(tokens, Token{"PNT", opr.PNT})
	tokens = append(tokens, Token{"USD", opr.USD})
	tokens = append(tokens, Token{"EUR", opr.EUR})
	tokens = append(tokens, Token{"JPY", opr.JPY})
	tokens = append(tokens, Token{"GBP", opr.GBP})
	tokens = append(tokens, Token{"CAD", opr.CAD})
	tokens = append(tokens, Token{"CHF", opr.CHF})
	tokens = append(tokens, Token{"INR", opr.INR})
	tokens = append(tokens, Token{"SGD", opr.SGD})
	tokens = append(tokens, Token{"CNY", opr.CNY})
	tokens = append(tokens, Token{"HKD", opr.HKD})
	tokens = append(tokens, Token{"XAU", opr.XAU})
	tokens = append(tokens, Token{"XAG", opr.XAG})
	tokens = append(tokens, Token{"XPD", opr.XPD})
	tokens = append(tokens, Token{"XPT", opr.XPT})
	tokens = append(tokens, Token{"XBT", opr.XBT})
	tokens = append(tokens, Token{"ETH", opr.ETH})
	tokens = append(tokens, Token{"LTC", opr.LTC})
	tokens = append(tokens, Token{"XBC", opr.XBC})
	tokens = append(tokens, Token{"FCT", opr.FCT})
	return tokens
}

func (opr *OraclePriceRecord) GetHash() []byte {
	data, err := json.Marshal(opr)
	check(err)
	oprHash := LX.Hash(data)
	return oprHash
}

// ComputeDifficulty()
// Difficulty the high order 8 bytes of the hash( hash(OPR record) + nonce)
func (opr *OraclePriceRecord) ComputeDifficulty(nonce []byte) (difficulty uint64) {
	no := append(opr.OPRHash, nonce...)
	h := LX.Hash(no)

	// The high eight bytes of the hash(hash(entry.Content) + nonce) is the difficulty.
	// Because we don't have a difficulty bar, we can define difficulty as the greatest
	// value, rather than the minimum value.  Our bar is the greatest difficulty found
	// within a 10 minute period.  We compute difficulty as Big Endian.
	for i := uint64(0); i < 8; i++ {
		difficulty = difficulty<<8 + uint64(h[i])
	}
	return difficulty
}

// Mine()
// Mine the OraclePriceRecord for a given number of seconds
func (opr *OraclePriceRecord) Mine(verbose bool) {

	// Pick a new nonce as a starting point.  Take time + last best nonce and hash that.
	nonce := []byte{0,0}
	if verbose {
		fmt.Printf("OPRHash %x\n", opr.OPRHash)
	}

miningloop:
	for i := 0; ; i++ {
		select {
		case <-opr.StopMining:
			break miningloop

		default:
		}
		nonce = nonce[:0]
		for j := i; j > 0 ; j = j >> 8 {
			nonce = append(nonce,byte(j))
		}
		diff := opr.ComputeDifficulty(nonce)

		if diff > opr.Difficulty {
			opr.Difficulty = diff
			// Copy over the previous nonce
			opr.Entry.ExtIDs[0] = append(opr.Entry.ExtIDs[0][:0], nonce...)
			if verbose {
				fmt.Printf("%15v OPR Difficulty %016x on opr hash: %x nonce: %x\n",
					time.Now().Format("15:04:05.000"), diff, opr.OPRHash, nonce)
			}
		}

	}
}

func (opr *OraclePriceRecord) ShortString() string {

	fdid := ""
	for i, v := range opr.FactomDigitalID {
		if i > 0 {
			fdid = fdid + " --- "
		}
		fdid = fdid + v
	}

	str := fmt.Sprintf("DID %30x OPRHash %30x Nonce %33x Difficulty %15x Grade %20f",
		fdid,
		opr.OPRHash,
		opr.Entry.ExtIDs[0],
		opr.Difficulty,
		opr.Grade)
	return str
}

// String
// Returns a human readable string for the Oracle Record
func (opr *OraclePriceRecord) String() (str string) {
	str = fmt.Sprintf("%14sField%14sValue\n", "", "")
	str = fmt.Sprintf("%s%32s %v\n", str, "OPRChainID", opr.OPRChainID)
	str = str + fmt.Sprintf("%32s %v\n", "Difficulty", opr.Difficulty)
	str = str + fmt.Sprintf("%32s %v\n", "Directory Block Height", opr.Dbht)
	str = fmt.Sprintf("%s%32s %v\n", str, "WinningPreviousOPRs", "")
	for i, v := range opr.WinningPreviousOPR {
		str = fmt.Sprintf("%s%32s %2d, %s\n", str, "", i+1, v)
	}
	str = str + fmt.Sprintf("%32s %s\n", "Coinbase PNT", opr.CoinbasePNTAddress)

	// Make a display string out of the Digital Identity.
	did := ""
	for i, t := range opr.FactomDigitalID {
		if i > 0 {
			did = did + " --- "
		}
		did = did + t
	}

	str = fmt.Sprintf("%s%32s %v\n", str, "FactomDigitalID", did)
	str = fmt.Sprintf("%s%32s %v\n", str, "PNT", opr.PNT)
	str = fmt.Sprintf("%s%32s %v\n", str, "USD", opr.USD)
	str = fmt.Sprintf("%s%32s %v\n", str, "EUR", opr.EUR)
	str = fmt.Sprintf("%s%32s %v\n", str, "JPY", opr.JPY)
	str = fmt.Sprintf("%s%32s %v\n", str, "GBP", opr.GBP)
	str = fmt.Sprintf("%s%32s %v\n", str, "CAD", opr.CAD)
	str = fmt.Sprintf("%s%32s %v\n", str, "CHF", opr.CHF)
	str = fmt.Sprintf("%s%32s %v\n", str, "INR", opr.INR)
	str = fmt.Sprintf("%s%32s %v\n", str, "SGD", opr.SGD)
	str = fmt.Sprintf("%s%32s %v\n", str, "CNY", opr.CNY)
	str = fmt.Sprintf("%s%32s %v\n", str, "HKD", opr.HKD)
	str = fmt.Sprintf("%s%32s %v\n", str, "XAU", opr.XAU)
	str = fmt.Sprintf("%s%32s %v\n", str, "XAG", opr.XAG)
	str = fmt.Sprintf("%s%32s %v\n", str, "XPD", opr.XPD)
	str = fmt.Sprintf("%s%32s %v\n", str, "XPT", opr.XPT)
	str = fmt.Sprintf("%s%32s %v\n", str, "XBT", opr.XBT)
	str = fmt.Sprintf("%s%32s %v\n", str, "ETH", opr.ETH)
	str = fmt.Sprintf("%s%32s %v\n", str, "LTC", opr.LTC)
	str = fmt.Sprintf("%s%32s %v\n", str, "XBC", opr.XBC)
	str = fmt.Sprintf("%s%32s %v\n", str, "FCT", opr.FCT)
	return str
}

func (opr *OraclePriceRecord) SetPegValues(assets polling.PegAssets) {

	opr.PNT = assets.PNT.Value
	opr.USD = assets.USD.Value
	opr.EUR = assets.EUR.Value
	opr.JPY = assets.JPY.Value
	opr.GBP = assets.GBP.Value
	opr.CAD = assets.CAD.Value
	opr.CHF = assets.CHF.Value
	opr.INR = assets.INR.Value
	opr.SGD = assets.SGD.Value
	opr.CNY = assets.CNY.Value
	opr.HKD = assets.HKD.Value
	opr.XAU = assets.XAU.Value
	opr.XAG = assets.XAG.Value
	opr.XPD = assets.XPD.Value
	opr.XPT = assets.XPT.Value
	opr.XBT = assets.XBT.Value
	opr.ETH = assets.ETH.Value
	opr.LTC = assets.LTC.Value
	opr.XBC = assets.XBC.Value
	opr.FCT = assets.FCT.Value

}

// NewOpr()
// collects all the information unique to this miner and its configuration, and also
// goes and gets the oracle data.  Also collects the winners from the prior block and
// puts their entry hashes (base58) into this OPR
func NewOpr(minerNumber int, dbht int32, c *config.Config, alert chan *OPRs) (*OraclePriceRecord, error) {
	opr := new(OraclePriceRecord)

	// create the channel to stop pegnetMining
	opr.StopMining = make(chan int, 1)

	// Save the config object
	opr.Config = c

	// Get the Entry Credit Address that we need to write our OPR records.
	if ecadrStr, err := c.String("Miner.ECAddress"); err != nil {
		return nil, err
	} else {
		ecAdr, err := factom.FetchECAddress(ecadrStr)
		if err != nil {
			return nil, err
		}
		opr.EC = ecAdr
	}

	// Get the Identity Chain Specification
	if chainID58, err := c.String("Miner.IdentityChain"); err != nil {
		return nil, errors.New("config file has no Miner.IdentityChain specified")
	} else {
		fields := strings.Split(chainID58, ",")
		if minerNumber > 0 {
			fields = append(fields, fmt.Sprintf("miner%03d", minerNumber))
		}
		for i, v := range fields {
			if i > 0 {
				fmt.Print(" --- ")
			}
			fmt.Print(v)
		}
		fmt.Println()

		opr.FactomDigitalID = fields

	}

	// Get the protocol chain to be used for pegnetMining records
	protocol, err1 := c.String("Miner.Protocol")
	network, err2 := c.String("Miner.Network")
	if err1 != nil {
		return nil, errors.New("config file has no Miner.Protocol specified")
	}
	if err2 != nil {
		return nil, errors.New("config file has no Miner.Network specified")
	}
	opr.OPRChainID = base58.Encode(support.ComputeChainIDFromStrings([]string{protocol, network, "Oracle Price Records"}))

	opr.Dbht = dbht

	if str, err := c.String("Miner.CoinbasePNTAddress"); err != nil {
		return nil, errors.New("config file has no Coinbase PNT Address")
	} else {
		opr.CoinbasePNTAddress = str
	}

	winners := <-alert
	for i, w := range winners.ToBePaid {
		opr.WinningPreviousOPR[i] = w.EntryHash
	}

	opr.GetOPRecord(c)

	return opr, nil
}

func (opr *OraclePriceRecord) GetOPRecord(c *config.Config) {
	opr.Config = c
	//get asset values
	var Peg polling.PegAssets
	Peg = polling.PullPEGAssets(c)
	Peg.FillPriceBytes()
	opr.SetPegValues(Peg)

	var err error
	opr.Entry = new(factom.Entry)
	opr.Entry.ChainID = hex.EncodeToString(base58.Decode(opr.OPRChainID))
	opr.Entry.ExtIDs = [][]byte{{}}
	opr.Entry.Content, err = json.Marshal(opr)
	if err != nil {
		panic(err)
	}
	opr.OPRHash = LX.Hash(opr.Entry.Content)
}
