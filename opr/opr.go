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
	"github.com/dustin/go-humanize"
	"github.com/pegnet/LXR256"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/polling"
	"github.com/zpatrick/go-config"
	"strings"
	"time"
)

type OraclePriceRecord struct {
	// These fields are not part of the OPR, but track values associated with the OPR.
	Config     *config.Config    `json:"-"` //  The config of the miner using the record
	Difficulty uint64            `json:"-"` // The difficulty of the given nonce
	Grade      float64           `json:"-"` // The grade when OPR records are compared
	OPRHash    []byte            `json:"-"` // The hash of the OPR record (used by PegNet Mining)
	EC         *factom.ECAddress `json:"-"` // Entry Credit Address used by a miner
	Entry      *factom.Entry     `json:"-"` // Entry to record this record
	StopMining chan int          `json:"-"` // Bool that stops PegNet Mining this OPR

	// These values define the context of the OPR, and they go into the PegNet OPR record, and are mined.
	OPRChainID         string     `json:oprchainid` // [base58]  Chain ID of the chain used by the Oracle Miners
	Dbht               int32      `json:dbht`       //           The Directory Block Height of the OPR.
	WinPreviousOPR     [10]string `json:winners`    // First 8 bytes of the Entry Hashes of the previous winners
	CoinbasePNTAddress string     `json:coinbase`   // [base58]  PNT Address to pay PNT
	FactomDigitalID    []string   `did`             // [unicode] Digital Identity of the miner

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
		common.Logf("error", "An error has been encountered: %v", e)
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
		OPRChainID = base58.Encode(common.ComputeChainIDFromStrings([]string{protocol, network, "Oracle Price Records"}))
	}

	if opr.OPRChainID != OPRChainID {
		return false
	}

	pre, _, err := common.ConvertPegAddrToRaw(opr.CoinbasePNTAddress)
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
	nonce := []byte{0, 0}
	common.Logf("OPR", "OPRHash %x", opr.OPRHash)

miningloop:
	for i := 0; ; i++ {
		select {
		case <-opr.StopMining:
			break miningloop

		default:
		}
		nonce = nonce[:0]
		for j := i; j > 0; j = j >> 8 {
			nonce = append(nonce, byte(j))
		}
		diff := opr.ComputeDifficulty(nonce)

		if diff > opr.Difficulty {
			opr.Difficulty = diff
			// Copy over the previous nonce
			opr.Entry.ExtIDs[0] = append(opr.Entry.ExtIDs[0][:0], nonce...)
			common.Logf("OPR", "%15v OPR Difficulty %016x on opr hash: %x nonce: %x",
				time.Now().Format("15:04:05.000"), diff, opr.OPRHash, nonce)
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
	str = fmt.Sprintf("Nonce %x\n", opr.Entry.ExtIDs[0])
	str = str + fmt.Sprintf("%32s %v\n", "OPRChainID", opr.OPRChainID)
	str = str + fmt.Sprintf("%32s %v\n", "Difficulty", opr.Difficulty)
	str = str + fmt.Sprintf("%32s %v\n", "Directory Block Height", opr.Dbht)
	str = str + fmt.Sprintf("%32s %v\n", "WinningPreviousOPRs", "")
	for i, v := range opr.WinPreviousOPR {
		str = str + fmt.Sprintf("%32s %2d, %s\n", "", i+1, v)
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

	str = str + fmt.Sprintf("%32s %v\n", "FactomDigitalID", did)
	str = str + fmt.Sprintf("%32s %v\n", "PNT", opr.PNT)
	str = str + fmt.Sprintf("%32s %v\n", "USD", opr.USD)
	str = str + fmt.Sprintf("%32s %v\n", "EUR", opr.EUR)
	str = str + fmt.Sprintf("%32s %v\n", "JPY", opr.JPY)
	str = str + fmt.Sprintf("%32s %v\n", "GBP", opr.GBP)
	str = str + fmt.Sprintf("%32s %v\n", "CAD", opr.CAD)
	str = str + fmt.Sprintf("%32s %v\n", "CHF", opr.CHF)
	str = str + fmt.Sprintf("%32s %v\n", "INR", opr.INR)
	str = str + fmt.Sprintf("%32s %v\n", "SGD", opr.SGD)
	str = str + fmt.Sprintf("%32s %v\n", "CNY", opr.CNY)
	str = str + fmt.Sprintf("%32s %v\n", "HKD", opr.HKD)
	str = str + fmt.Sprintf("%32s %v\n", "XAU", opr.XAU)
	str = str + fmt.Sprintf("%32s %v\n", "XAG", opr.XAG)
	str = str + fmt.Sprintf("%32s %v\n", "XPD", opr.XPD)
	str = str + fmt.Sprintf("%32s %v\n", "XPT", opr.XPT)
	str = str + fmt.Sprintf("%32s %v\n", "XBT", opr.XBT)
	str = str + fmt.Sprintf("%32s %v\n", "ETH", opr.ETH)
	str = str + fmt.Sprintf("%32s %v\n", "LTC", opr.LTC)
	str = str + fmt.Sprintf("%32s %v\n", "XBC", opr.XBC)
	str = str + fmt.Sprintf("%32s %v\n", "FCT", opr.FCT)

	str = str + fmt.Sprintf("\nWinners\n\n")

	pwin := GetPreviousOPRs(opr.Dbht - 1)

	// If there were previous winners, we need to make sure this miner is running
	// the software to detect them, and that we agree with their conclusions.
	if pwin != nil {
		for i, v := range opr.WinPreviousOPR {
			fid := ""
			for j, field := range pwin[i].FactomDigitalID {
				if j > 0 {
					fid = fid + "---"
				}
				fid = fid + field
			}
			balance := GetBalance(pwin[i].CoinbasePNTAddress)
			hbal := humanize.Comma(balance)
			str = str + fmt.Sprintf("   %16s %16x %30s %-56s = %10s\n",
				v,
				pwin[i].Entry.Hash()[:8],
				fid,
				pwin[i].
					CoinbasePNTAddress, hbal)
		}
	}
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
		fid := fields[0]
		for _, v := range fields[1:] {
			fid = fid + " --- " + v
		}
		common.Logf("OPR", "New OPR miner %s", fid)
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
	opr.OPRChainID = base58.Encode(common.ComputeChainIDFromStrings([]string{protocol, network, "Oracle Price Records"}))

	opr.Dbht = dbht

	// If this is a test network, then give multiple miners their own tPNT address
	// because that is way more useful debugging than giving all miners the same
	// PNT address.  Otherwise, give all miners the same PNT address because most
	// users really doing mining will mostly be happen sending rewards to a single
	// address.
	if network == "TestNet" && minerNumber != 0 {
		fct := common.DebugFCTaddresses[minerNumber][1]
		sraw, err := common.ConvertUserStrFctEcToAddress(fct)
		if err != nil {
			return nil, err
		}
		raw, err := hex.DecodeString(sraw)
		if err != nil {
			return nil, err
		}
		opr.CoinbasePNTAddress, err = common.ConvertRawAddrToPeg("tPNT", raw)
		if err != nil {
			return nil, err
		}
	} else {
		if str, err := c.String("Miner.CoinbasePNTAddress"); err != nil {
			return nil, errors.New("config file has no Coinbase PNT Address")
		} else {
			opr.CoinbasePNTAddress = str
		}
	}
	winners := <-alert
	for i, w := range winners.ToBePaid {
		opr.WinPreviousOPR[i] = hex.EncodeToString(w.Entry.Hash()[:8])
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
