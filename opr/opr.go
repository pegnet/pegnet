package opr

// These

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/FactomProject/btcutil/base58"
	"github.com/FactomProject/factom"
	"github.com/pegnet/LXR256"
	"github.com/zpatrick/go-config"
	"strings"
	"time"
	"github.com/pegnet/OracleRecord/polling"
	"github.com/pegnet/OracleRecord/support"
)

type OraclePriceRecord struct {
	// These fields are not part of the OPR, but track values associated with the OPR.
	Config     *config.Config    `json:"-"` //  The config of the miner using the record
	Difficulty uint64            `json:"-"` // The difficulty of the given nonce
	Grade      float64           `json:"-"` // The grade when OPR records are compared
	BestNonce  []byte            `json:"-"` // nonce created by mining;
	OPRHash    []byte            `json:"-"` // The hash of the OPR record (used by mining)
	EC         *factom.ECAddress `json:"-"` // Entry Credit Address used by a miner
	Entry      *factom.Entry     `json:"-"` // Entry to record this record
	StopMining chan int          `json:"-"` // Bool that stops mining this OPR

	// These values define the context of the OPR, and they go into the PegNet OPR record, and are mined.
	OPRChainID         string     `json:oprchainid` // [base58]  Chain ID of the chain used by the Oracle Miners
	Dbht               int32      `json:dbht`       //           The Directory Block Height of the OPR.
	WinningPreviousOPR [10]string `json:winners`    // [base58]  Winning OPR entries in the previous block
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

func init() {
	LX.Init(0x123412341234, 25, 256, 5)
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
func (opr *OraclePriceRecord) ComputeDifficulty(oprHash []byte, nonce []byte) (difficulty uint64) {
	no := append(oprHash, nonce...)
	h := LX.Hash(no)
	difficulty = 0
	for i := uint64(0); i < 8; i++ {
		difficulty = difficulty<<8 + uint64(h[i])
	}
	return difficulty
}

// Mine()
// Mine the OraclePriceRecord for a given number of seconds
func (opr *OraclePriceRecord) Mine(seed int64, verbose bool) {

	// Pick a new nonce as a starting point.  Take time + last best nonce and hash that.
	t := []byte(time.Now().Format("10:10:10.0000000000"))
	nonce := LX.Hash(append(opr.BestNonce, t...))
	nonce = LX.Hash(append(nonce,
		byte(seed), byte(seed>>8),byte(seed>>16),byte(seed>>24),
		byte(seed>>32),byte(seed>>40),byte(seed>>48),byte(seed>>56)))

	// Set the OPRHash of the content of the Oracle Record
	js, err := json.Marshal(opr)
	if err != nil {
		panic(err)
	}
	opr.OPRHash = LX.Hash(js)
	if verbose {
		fmt.Printf("OPRHash %x\n",opr.OPRHash)
	}


	for i := 0;i<5;i++{
		nonce[i]=0
	}
miningloop:
	for {
		select {
		case <-opr.StopMining:
			break miningloop

		default:
		}

		for i := 0; i < 100; i++ {
			for j := 0; ; j++ {
				nonce[j]++
				if nonce[j] > 0 {
					break
				}
			}

			diff := opr.ComputeDifficulty(opr.OPRHash, nonce)
			if diff > opr.Difficulty {
				opr.Difficulty = diff
				opr.BestNonce = append(opr.BestNonce[:0], nonce...)
				if verbose {
					fmt.Printf("%15v OPR Difficulty %016x on opr hash: %x nonce: %x\n",
						time.Now().Format("15:04:05.000"), diff, opr.OPRHash, nonce)
				}
			}
		}
	}
}

func (opr *OraclePriceRecord) ShortString() string {

	hash := []byte{0}
	if opr.Entry != nil {
		hash = opr.Entry.Hash()
	}

	fdid := ""
	for i,v := range opr.FactomDigitalID {
		if i>0 {
			fdid = fdid+" --- "
		}
		fdid = fdid + v
	}

	str := fmt.Sprintf("DID %30x EntryHash %30x Nonce %33x Difficulty %15d Grade %20f",
		fdid,
		hash,
		opr.BestNonce,
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

func (opr *OraclePriceRecord) GetOPRecord(c *config.Config) {
	opr.Config = c
	//get asset values
	var Peg polling.PegAssets
	Peg = polling.PullPEGAssets(c)
	Peg.FillPriceBytes()

	opr.SetPegValues(Peg)

}

// Set the chainID; assumes a base58 string
func (opr *OraclePriceRecord) SetOPRChainID(chainID string) {
	opr.OPRChainID = chainID
}

// Sets one of the winning OPR records from the previous block.  Expects a base58 string
func (opr *OraclePriceRecord) SetWinningPreviousOPR(index int, winning string) {
	opr.WinningPreviousOPR[index] = winning
}

// Sets the PNT Address in human/wallet format
func (opr *OraclePriceRecord) SetCoinbasePNTAddress(coinbaseAddress string) {
	opr.CoinbasePNTAddress = coinbaseAddress
}

// Sets the DigitalID for the miner.  Expects the ExtIDs of the Identity chain.
// Miner IDs are expected to be in unicode
func (opr *OraclePriceRecord) SetFactomDigitalID(factomDigitalID []string) {
	opr.FactomDigitalID = factomDigitalID

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

// GetEntry
// Given a particular chain to write this entry, compute a proper entry
// for this OraclePriceRecord
func (opr *OraclePriceRecord) GetEntry(chainID string) *factom.Entry {
	// An OPR record only has the nonce as an external ID

	entryExtIDs := append([][]byte{}, opr.BestNonce)
	// The body Data is the marshal of the OPR
	bodyData, err := json.Marshal(opr)
	check(err)
	// Create the Entry struct
	assetEntry := factom.Entry{ChainID: chainID, ExtIDs: entryExtIDs, Content: bodyData}
	opr.Entry = &assetEntry
	return opr.Entry
}

func NewOpr(minerNumber int, dbht int32, c *config.Config) (*OraclePriceRecord, error) {
	opr := new(OraclePriceRecord)

	// create the channel to stop mining
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
		if len(fields) == 1 && string(fields[0]) == "prototype" {
			fields = append(fields, fmt.Sprintf("miner%03d", minerNumber))
		}
		opr.FactomDigitalID = fields
	}

	// Get the protocol chain to be used for mining records
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

	return opr, nil
}
