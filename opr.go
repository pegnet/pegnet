package oprecord

// These

import (
	"github.com/pegnet/LXR256"
	"fmt"
	"github.com/FactomProject/factom"
	"github.com/zpatrick/go-config"
	"encoding/json"
	"github.com/FactomProject/btcutil/base58"
)

type OraclePriceRecord struct {
	// These fields are not part of the OPR, but track values associated with the OPR.
	Config     *config.Config    `json:"-"`//  The config of the miner using the record
	Difficulty uint64            `json:"-"`// The difficulty of the given nonce
	Grade      float64           `json:"-"`// The grade when OPR records are compared
	Nonce      string            `json:"-"`// [base58] - nonce created by mining;
	EC         *factom.ECAddress `json:"-"`// Entry Credit Address used by a miner
	Entry      *factom.Entry     `json:"-"`// Entry to record this record

	// These values define the context of the OPR, and they go into the PegNet OPR record, and are mined.
	ChainID            string     `json:chainid`  // [base58]  Chain ID of the chain used by the Oracle Miners
	Dbht               int32      `json:dbht`     //           The Directory Block Height of the OPR.
	VersionEntryHash   string     `json:version`  // [base58]  Entry hash for the PegNet version
	WinningPreviousOPR [10]string `json:winners`  // [base58]  Winning OPR entries in the previous block
	CoinbasePNTAddress string     `json:coinbase` // [base58]  PNT Address to pay PNT
	FactomDigitalID    []string   `did`           // [unicode] Digital Identity of the miner

	// The Oracle values of the OPR, they are the meat of the OPR record, and are mined.
	PNT                float64
	USD                float64
	EUR                float64
	JPY                float64
	GBP                float64
	CAD                float64
	CHF                float64
	INR                float64
	SGD                float64
	CNY                float64
	HKD                float64
	XAU                float64
	XAG                float64
	XPD                float64
	XPT                float64
	XBT                float64
	ETH                float64
	LTC                float64
	XBC                float64
	FCT                float64
}

type Token struct {
	code  string
	value float64
}

var LX lxr.LXRHash

func init() {
	LX.Init(0x123412341234, 20, 256, 5)
}

func (opr *OraclePriceRecord) GetTokens() []Token {
	tokens := []Token{}
	tokens = append(tokens, Token{"PNT", opr.PNT})
	return tokens
}

func (opr *OraclePriceRecord) GetHash() []byte {
	data, err := json.Marshal(opr)
	check(err)
	oprHash := LX.Hash(data)
	return oprHash
}

func (opr *OraclePriceRecord) GetNonceHash() []byte {
	no := append([]byte{}, opr.Nonce[:]...)
	oprHash := opr.GetHash()
	no = append(no, oprHash...)
	h := LX.Hash(no)
	return h
}

func (opr *OraclePriceRecord) ComputeDifficulty() uint64 {
	h := opr.GetNonceHash()
	opr.Difficulty = lxr.Difficulty(h) // Go calculate the difficulty, and cache in the opr
	return opr.Difficulty
}

func (opr *OraclePriceRecord) ShortString() string {

	hash := []byte{0}
	if opr.Entry != nil {
		hash = opr.Entry.Hash()
	}
	str := fmt.Sprintf("DID %6x EntryHash %70x Nonce %33x Difficulty %15d Grade %20f",
		opr.FactomDigitalID[:6],
		hash,
		opr.Nonce[:16],
		opr.Difficulty,
		opr.Grade)
	return str
}

// String
// Returns a human readable string for the Oracle Record
func (opr *OraclePriceRecord) String() (str string) {
	str = fmt.Sprintf("%14sField%14sValue\n", "", "")
	opr.ComputeDifficulty()
	str = fmt.Sprintf("%s%32s %v\n", str, "ChainID", opr.ChainID)
	str = str + fmt.Sprintf("%32s %v\n", "Difficulty", opr.Difficulty)
	str = str + fmt.Sprintf("%32s %v\n", "Directory Block Height", opr.Dbht)
	str = fmt.Sprintf("%s%32s %v\n", str, "VersionEntryHash", opr.VersionEntryHash)
	str = fmt.Sprintf("%s%32s %v\n", str, "WinningPreviousOPRs", "")
	for i, v := range opr.WinningPreviousOPR {
		str = fmt.Sprintf("%s%32s     %2d, %s\n", str, "", i+1, v)
	}
	str = str + fmt.Sprintf("%32s %v\n", opr.CoinbasePNTAddress)

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
	var Peg PegAssets
	Peg = PullPEGAssets(c)
	Peg.FillPriceBytes()

	opr.SetPegValues(Peg)

}

// Set the chainID; assumes a base58 string
func (opr *OraclePriceRecord) SetChainID(chainID string) {
	opr.ChainID = chainID
}

// Set the VersionEntryHash; assumes a base58 string
func (opr *OraclePriceRecord) SetVersionEntryHash(versionEntryHash string) {
	opr.VersionEntryHash = versionEntryHash
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

func (opr *OraclePriceRecord) SetPegValues(assets PegAssets) {

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

	bNonce := base58.Decode(opr.Nonce)
	entryExtIDs := append([][]byte{}, bNonce)
	// The body Data is the marshal of the OPR
	bodyData, err := json.Marshal(opr)
	check(err)
	// Create the Entry struct
	assetEntry := factom.Entry{ChainID: chainID, ExtIDs: entryExtIDs, Content: bodyData}
	opr.Entry = &assetEntry
	return opr.Entry
}
