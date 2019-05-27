package oprecord

// These

import (
	"encoding/binary"
	"github.com/pegnet/LXR256"

	"errors"
	"fmt"
	"github.com/FactomProject/factom"
	"github.com/zpatrick/go-config"
	"encoding/json"
)

type OraclePriceRecord struct {
	Config             *config.Config    // not part of OPR - The config of the miner using the record
	Difficulty         uint64            // not part of OPR -The difficulty of the given nonce
	Grade              float64           // not part of OPR -The grade when OPR records are compared
	Nonce              [32]byte          // not part of OPR - nonce creacted by mining
	EC                 *factom.ECAddress // not part of OPR - Entry Credit Address used by a miner
	Entry              *factom.Entry     // not part of OPR - Entry to record this record
	ChainID            [32]byte          `json:chainid`
	Dbht               int32 			 `json:dbht` // The Directory Block Height that this record is to contribute to.
	VersionEntryHash   [32]byte			 `json:version` // Entry hash for the PegNet version
	WinningPreviousOPR [10][32]byte		 `json:winners` // Winning OPR entries in the previous block
	CoinbasePNTAddress string			 `json:coinbase` // PNT Address to pay PNT
	FactomDigitalID    [32]byte          `did` // Digital Identity of the miner
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
	code string
	value float64
}

var lx lxr.LXRHash

func init() {
	lx.Init(0x123412341234, 32, 256, 5)
}



func (opr *OraclePriceRecord) GetTokens() []Token {
	tokens := []Token{}
	tokens = append(tokens, Token{"PNT", opr.PNT})
	return tokens
}

func (opr *OraclePriceRecord) GetHash() []byte {
	data, err := json.Marshal(opr)
	check(err)
	oprHash := lx.Hash(data)
	return oprHash
}

func (opr *OraclePriceRecord) GetNonceHash() []byte {
	no := append([]byte{}, opr.Nonce[:]...)
	oprHash := opr.GetHash()
	no = append(no, oprHash...)
	h := lx.Hash(no)
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
	print32 := func(label string, value []byte) {
		if len(value) == 8 {
			v := binary.BigEndian.Uint64(value)
			str = str + fmt.Sprintf("%32s %8d.%08d\n", label, v/100000000, v%100000000)
			return
		}
		str = str + fmt.Sprintf("%32s %x\n", label, value)
	}
	opr.ComputeDifficulty()
	print32("ChainID", opr.ChainID[:])
	str = str + fmt.Sprintf("%32s %v\n", "Difficulty", opr.Difficulty)
	str = str + fmt.Sprintf("%32s %v\n", "Directory Block Height", opr.Dbht)
	print32("VersionEntryHash", opr.VersionEntryHash[:])
	for _,v := range opr.WinningPreviousOPR {
		print32("  WinningPreviousOPR", v[:])
	}
	str = str + fmt.Sprintf("%32s %v\n",opr.CoinbasePNTAddress)
	print32("FactomDigitalID", opr.FactomDigitalID[:])

	str = fmt.Sprintf("%s%32s %v\n",str,"PNT", opr.PNT)
	str = fmt.Sprintf("%s%32s %v\n",str,"USD", opr.USD)
	str = fmt.Sprintf("%s%32s %v\n",str,"EUR", opr.EUR)
	str = fmt.Sprintf("%s%32s %v\n",str,"JPY", opr.JPY)
	str = fmt.Sprintf("%s%32s %v\n",str,"GBP", opr.GBP)
	str = fmt.Sprintf("%s%32s %v\n",str,"CAD", opr.CAD)
	str = fmt.Sprintf("%s%32s %v\n",str,"CHF", opr.CHF)
	str = fmt.Sprintf("%s%32s %v\n",str,"INR", opr.INR)
	str = fmt.Sprintf("%s%32s %v\n",str,"SGD", opr.SGD)
	str = fmt.Sprintf("%s%32s %v\n",str,"CNY", opr.CNY)
	str = fmt.Sprintf("%s%32s %v\n",str,"HKD", opr.HKD)
	str = fmt.Sprintf("%s%32s %v\n",str,"XAU", opr.XAU)
	str = fmt.Sprintf("%s%32s %v\n",str,"XAG", opr.XAG)
	str = fmt.Sprintf("%s%32s %v\n",str,"XPD", opr.XPD)
	str = fmt.Sprintf("%s%32s %v\n",str,"XPT", opr.XPT)
	str = fmt.Sprintf("%s%32s %v\n",str,"XBT", opr.XBT)
	str = fmt.Sprintf("%s%32s %v\n",str,"ETH", opr.ETH)
	str = fmt.Sprintf("%s%32s %v\n",str,"LTC", opr.LTC)
	str = fmt.Sprintf("%s%32s %v\n",str,"XBC", opr.XBC)
	str = fmt.Sprintf("%s%32s %v\n",str,"FCT", opr.FCT)
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

func (opr *OraclePriceRecord) SetChainID(chainID []byte) {
	copy(opr.ChainID[0:], chainID)
}

func (opr *OraclePriceRecord) SetVersionEntryHash(versionEntryHash []byte) {
	copy(opr.VersionEntryHash[0:], versionEntryHash)
}

func (opr *OraclePriceRecord) SetWinningPreviousOPR(winning []byte) {
	copy(opr.WinningPreviousOPR[0:], winning)
}

func (opr *OraclePriceRecord) SetCoinbasePNTAddress(coinbaseAddress []byte) {
	copy(opr.CoinbasePNTAddress[0:], coinbaseAddress)
}

func (opr *OraclePriceRecord) SetFactomDigitalID(factomDigitalID []byte) {
	copy(opr.FactomDigitalID[0:], factomDigitalID)

}

func (opr *OraclePriceRecord) SetPegValues(assets PegAssets) {
	b := make([]byte, 8)

	opr.PNT = assets.PNT.Value
	opr.USD =	assets.USD.Value
	opr.EUR=assets.EUR.Value
	opr.JPY=assets.JPY.Value
	opr.GBP=assets.GBP.Value
	opr.CAD=assets.CAD.Value
	opr.CHF=assets.CHF.Value
	opr.INR=assets.INR.Value
	opr.SGD=assets.SGD.Value
	opr.CNY=assets.CNY.Value
	opr.HKD=assets.HKD.Value
	opr.XAU=assets.XAU.Value
	opr.XAG=assets.XAG.Value
	opr.XPD=assets.XPD.Value
	opr.XPT=assets.XPT.Value
	opr.XBT=assets.XBT.Value
	opr.ETH=assets.ETH.Value
	opr.LTC=assets.LTC.Value
	opr.XBC=assets.XBC.Value
	opr.FCT=assets.FCT.Value

}

// GetEntry
// Given a particular chain to write this entry, compute a proper entry
// for this OraclePriceRecord
func (opr *OraclePriceRecord) GetEntry(chainID string) *factom.Entry {
	// An OPR record only has the nonce as an external ID
	entryExtIDs := [][]byte{opr.Nonce[:]}
	// The body Data is the marshal of the OPR
	bodyData, err := json.Marshal(opr)
	check(err)
	// Create the Entry struct
	assetEntry := factom.Entry{ChainID: chainID, ExtIDs: entryExtIDs, Content: bodyData}
	opr.Entry = &assetEntry
	return opr.Entry
}
