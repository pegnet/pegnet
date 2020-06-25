package spr

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/FactomProject/btcutil/base58"
	"github.com/FactomProject/factom"
	"github.com/golang/protobuf/proto"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr/oprencoding"
	"github.com/pegnet/pegnet/polling"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

var PollingDataSource *polling.DataSources
var pollingDataSourceInitializer sync.Once

func InitDataSource(config *config.Config) {
	pollingDataSourceInitializer.Do(func() {
		if PollingDataSource == nil { // This can be inited from unit tests
			PollingDataSource = polling.NewDataSources(config)
		}
	})
}

// StakingPriceRecord is the data used and created by staker
type StakingPriceRecord struct {
	// These fields are not part of the OPR, but track values associated with the OPR.
	Protocol           string  `json:"-"` // The Protocol we are running on (PegNet)
	Network            string  `json:"-"` // The network we are running on (TestNet vs MainNet)
	Difficulty         uint64  `json:"-"` // The difficulty of the given nonce
	Grade              float64 `json:"-"` // The grade when OPR records are compared
	SPRHash            []byte  `json:"-"` // The hash of the OPR record (used by PegNet Mining)
	SPRChainID         string  `json:"-"` // [base58]  Chain ID of the chain used by the Oracle Miners
	CoinbasePEGAddress string  `json:"-"` // [base58]  PEG Address to pay PEG

	// This can be attached to an OPR, which indicates how low we should expect a mined
	// opr to be. Any OPRs mined below this are not worth submitting to the network.
	MinimumDifficulty uint64 `json:"-"`

	// Factom Entry data
	EntryHash              []byte `json:"-"` // Entry to record this record
	Nonce                  []byte `json:"-"` // Nonce used with OPR
	SelfReportedDifficulty []byte `json:"-"` // Miners self report their difficulty
	Version                uint8  `json:"-"`

	// These values define the context of the OPR, and they go into the PegNet OPR record, and are mined.
	CoinbaseAddress string   `json:"coinbase"` // [base58]  PEG Address to pay PEG
	Dbht            int32    `json:"dbht"`     //           The Directory Block Height of the OPR.
	WinPreviousOPR  []string `json:"winners"`  // First 8 bytes of the Entry Hashes of the previous winners
	FactomDigitalID string   `json:"minerid"`  // [unicode] Digital Identity of the miner

	// The Oracle values of the OPR, they are the meat of the OPR record, and are mined.
	Assets StakingPriceRecordAssetList `json:"assets"`
}

func NewStakingPriceRecord() *StakingPriceRecord {
	o := new(StakingPriceRecord)
	o.Assets = make(StakingPriceRecordAssetList)

	return o
}

// CloneEntryData will clone the SPR data needed to make a factom entry.
//	This needs to be done because I need to marshal this into my factom entry.
func (c *StakingPriceRecord) CloneEntryData() *StakingPriceRecord {
	n := NewStakingPriceRecord()
	n.SPRChainID = c.SPRChainID
	n.Dbht = c.Dbht
	n.Version = c.Version
	n.WinPreviousOPR = make([]string, len(c.WinPreviousOPR), len(c.WinPreviousOPR))
	copy(n.WinPreviousOPR[:], c.WinPreviousOPR[:])
	n.CoinbaseAddress = c.CoinbaseAddress
	n.CoinbasePEGAddress = c.CoinbasePEGAddress

	n.FactomDigitalID = c.FactomDigitalID
	n.Assets = make(StakingPriceRecordAssetList)
	for k, v := range c.Assets {
		n.Assets[k] = v
	}
	return n
}

// SPRChainID is the calculated chain id of the records chain
var SPRChainID string

// Token is a combination of currency Code and Value
type Token struct {
	Code  string
	Value float64
}

// Validate performs sanity checks of the structure and values of the SPR.
// It does not validate the winners of the previous block.
func (spr *StakingPriceRecord) Validate(c *config.Config, dbht int64) bool {
	net, _ := common.LoadConfigNetwork(c)
	if !common.NetworkActive(net, dbht) {
		return false
	}

	// Validate there are no 0's
	for k, v := range spr.Assets {
		if v == 0 {
			fmt.Println("[error]:", k, v)
			return false
		}
	}

	//// Only enforce on version 2 and forward
	//if err := common.ValidIdentity(spr.FactomDigitalID); err != nil {
	//	return false
	//}
	//
	//// Only enforce on version 2 and forward, checking valid FCT address
	//if !ValidFCTAddress(spr.CoinbaseAddress) {
	//	return false
	//}

	//if int64(spr.Dbht) != dbht {
	//	return false // DBHeight is not reported correctly
	//}

	// Validate all the Assets exists
	return spr.Assets.ContainsExactly(common.AssetsV4)
}

// ValidFCTAddress will be removed in the grading module refactor. This is just temporary to get this
// functionality, and be easily unit testable.
func ValidFCTAddress(addr string) bool {
	return len(addr) > 2 && addr[:2] == "FA" && factom.IsValidAddress(addr)
}

// GetTokens creates an iterateable slice of Tokens containing all the currency values
func (spr *StakingPriceRecord) GetTokens() (tokens []Token) {
	return spr.Assets.List(spr.Version)
}

// GetHash returns the LXHash over the OPR's json representation
func (spr *StakingPriceRecord) GetHash() []byte {
	if len(spr.SPRHash) > 0 {
		return spr.SPRHash
	}

	// SafeMarshal handles the PNT/PEG issue
	data, err := spr.SafeMarshal()
	common.CheckAndPanic(err)
	sha := sha256.Sum256(data)
	spr.SPRHash = sha[:]
	return spr.SPRHash
}

// ShortString returns a human readable string with select data
func (opr *StakingPriceRecord) ShortString() string {
	str := fmt.Sprintf("DID %30x SPRHash %30x Nonce %33x Difficulty %15x Grade %20f",
		opr.FactomDigitalID,
		opr.SPRHash,
		opr.Nonce,
		opr.Difficulty,
		opr.Grade)
	return str
}

// String returns a human readable string for the Oracle Record
func (spr *StakingPriceRecord) String() (str string) {
	str = fmt.Sprintf("Nonce %x\n", spr.Nonce)
	str = str + fmt.Sprintf("%32s %v\n", "Difficulty", spr.Difficulty)
	str = str + fmt.Sprintf("%32s %v\n", "Directory Block Height", spr.Dbht)
	str = str + fmt.Sprintf("%32s %v\n", "WinningPreviousOPRs", "")
	for i, v := range spr.WinPreviousOPR {
		str = str + fmt.Sprintf("%32s %2d, %s\n", "", i+1, v)
	}
	str = str + fmt.Sprintf("%32s %s\n", "Coinbase PEG", spr.CoinbasePEGAddress)

	// Make a display string out of the Digital Identity.

	str = str + fmt.Sprintf("%32s %v\n", "FactomDigitalID", spr.FactomDigitalID)
	for _, asset := range spr.Assets.List(spr.Version) {
		str = str + fmt.Sprintf("%32s %v\n", "PEG", asset)
	}

	str = str + fmt.Sprintf("\nWinners\n\n")

	// If there were previous winners, we need to make sure this miner is running
	// the software to detect them, and that we agree with their conclusions.
	for i, v := range spr.WinPreviousOPR {
		str = str + fmt.Sprintf("   %2d\t%16s\n",
			i,
			v,
		)
	}
	return str
}

// LogFieldsShort returns a set of common fields to be included in logrus
func (spr *StakingPriceRecord) LogFieldsShort() log.Fields {
	return log.Fields{
		"did":        spr.FactomDigitalID,
		"spr_hash":   hex.EncodeToString(spr.SPRHash),
		"nonce":      hex.EncodeToString(spr.Nonce),
		"difficulty": spr.Difficulty,
		"grade":      spr.Grade,
	}
}

// SetPegValues assigns currency polling values to the SPR
func (spr *StakingPriceRecord) SetPegValues(assets polling.PegAssets) {
	for asset, v := range assets {
		spr.Assets.SetValue(asset, v.Value)
	}
}

// NewSpr collects all the information unique to this staker and its configuration, and also
// goes and gets the oracle data.  Also collects the winners from the prior block and
// puts their entry hashes (base58) into this SPR
func NewSpr(ctx context.Context, dbht int32, c *config.Config) (spr *StakingPriceRecord, err error) {
	spr = NewStakingPriceRecord()
	spr.SPRChainID = base58.Encode(common.ComputeChainIDFromStrings([]string{"PegNet", "MainNet", common.SPRChainTag}))

	err = spr.GetSPRecord(c)
	if err != nil {
		return nil, err
	}

	if !spr.Validate(c, int64(dbht)) {
		if !common.NetworkActive(spr.Network, int64(dbht)) {
			return nil, fmt.Errorf("Waiting for activation height")
		}
		return nil, fmt.Errorf("spr invalid")
	}
	return spr, nil
}

// GetSPRecord initializes the SPR with polling data and factom entry
func (spr *StakingPriceRecord) GetSPRecord(c *config.Config) error {
	InitDataSource(c) // Kinda odd to have this here.

	oprVersion := 4
	//get asset values
	Peg, err := PollingDataSource.PullAllPEGAssets(uint8(oprVersion))
	if err != nil {
		return err
	}
	spr.SetPegValues(Peg)

	data, err := spr.SafeMarshal()
	if err != nil {
		panic(err)
	}
	sha := sha256.Sum256(data)
	spr.SPRHash = sha[:]
	return nil
}

// CreateSPREntry will create the entry from the EXISITING data.
// It will not set any fields like in `GetSPRecord`
func (spr *StakingPriceRecord) CreateSPREntry() (*factom.Entry, error) {
	var err error
	e := new(factom.Entry)

	e.ChainID = hex.EncodeToString(base58.Decode(spr.SPRChainID))
	fmt.Println("e.ChainID:", e.ChainID)

	//ExtIDs:
	//	1) version byte (byte, default 0)
	//	2) RCD of the payout address
	//	3) signature covering [ExtID]
	e.ExtIDs = [][]byte{{}, {}, {spr.Version}}

	e.Content, err = spr.SafeMarshal()
	if err != nil {
		fmt.Println("error:", err)
		return nil, err
	}
	return e, nil
}

// SafeMarshal will marshal the json depending on the opr version
func (spr *StakingPriceRecord) SafeMarshal() ([]byte, error) {
	if spr.Assets == nil {
		return nil, fmt.Errorf("assets is nil, cannot marshal")
	}

	assetList := common.AssetsV4
	prices := make([]uint64, len(spr.Assets))

	for i, asset := range assetList {
		prices[i] = spr.Assets[asset]
	}

	pOpr := &oprencoding.ProtoOPR{
		Address: spr.CoinbaseAddress,
		ID:      spr.FactomDigitalID,
		Height:  spr.Dbht,
		Assets:  prices,
		Winners: nil,
	}

	return proto.Marshal(pOpr)
}

// SafeMarshal will unmarshal the json
func (spr *StakingPriceRecord) SafeUnmarshal(data []byte) error {
	protoOPR := oprencoding.ProtoOPR{}
	err := proto.Unmarshal(data, &protoOPR)
	if err != nil {
		return err
	}

	assetList := common.AssetsV4
	spr.Assets = make(StakingPriceRecordAssetList)
	// Populate the original opr
	spr.CoinbaseAddress = protoOPR.Address
	spr.FactomDigitalID = protoOPR.ID
	spr.Dbht = protoOPR.Height

	if len(protoOPR.Assets) != len(assetList) {
		return fmt.Errorf("found %d assets, expected %d", len(protoOPR.Assets), len(assetList))
	}

	// Hard coded list of assets
	for i, asset := range assetList {
		spr.Assets[asset] = protoOPR.Assets[i]
	}

	// Decode winners
	spr.WinPreviousOPR = nil

	return nil
}
