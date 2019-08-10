// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/FactomProject/btcutil/base58"
	"github.com/FactomProject/factom"
	"github.com/dustin/go-humanize"
	lxr "github.com/pegnet/LXRHash"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/polling"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

// TODO: Do not make this a global.
//		currently the OPR does the asset polling, this is bit backwards.
//		We should poll the asset prices, and set the OPR. Not create the OPR
//		and have it find it's own prices.
var PollingDataSource *polling.DataSources
var pollingDataSourceInitializer sync.Once

func InitDataSource(config *config.Config) {
	pollingDataSourceInitializer.Do(func() {
		PollingDataSource = polling.NewDataSources(config)
	})
}

// OraclePriceRecord is the data used and created by miners
type OraclePriceRecord struct {
	// These fields are not part of the OPR, but track values associated with the OPR.
	Protocol           string  `json:"-"` // The Protocol we are running on (PegNet)
	Network            string  `json:"-"` // The network we are running on (TestNet vs MainNet)
	Difficulty         uint64  `json:"-"` // The difficulty of the given nonce
	Grade              float64 `json:"-"` // The grade when OPR records are compared
	OPRHash            []byte  `json:"-"` // The hash of the OPR record (used by PegNet Mining)
	OPRChainID         string  `json:"-"` // [base58]  Chain ID of the chain used by the Oracle Miners
	CoinbasePNTAddress string  `json:"-"` // [base58]  PNT Address to pay PNT

	// This can be attached to an OPR, which indicates how low we should expect a mined
	// opr to be. Any OPRs mined below this are not worth submitting to the network.
	MinimumDifficulty uint64 `json:"-"`

	// Factom Entry data
	EntryHash              []byte `json:"-"` // Entry to record this record
	Nonce                  []byte `json:"-"` // Nonce used with OPR
	SelfReportedDifficulty []byte `json:"-"` // Miners self report their difficulty
	Version                uint8  `json:"-"`

	// These values define the context of the OPR, and they go into the PegNet OPR record, and are mined.
	CoinbaseAddress string     `json:"coinbase"` // [base58]  PNT Address to pay PNT
	Dbht            int32      `json:"dbht"`     //           The Directory Block Height of the OPR.
	WinPreviousOPR  [10]string `json:"winners"`  // First 8 bytes of the Entry Hashes of the previous winners
	FactomDigitalID string     `json:"minerid"`  // [unicode] Digital Identity of the miner

	// The Oracle values of the OPR, they are the meat of the OPR record, and are mined.
	Assets OraclePriceRecordAssetList `json:"assets"`
}

func NewOraclePriceRecord() *OraclePriceRecord {
	o := new(OraclePriceRecord)
	o.Assets = make(OraclePriceRecordAssetList)
	o.Version = OPRVersion

	return o
}

// CloneEntryData will clone the OPR data needed to make a factom entry.
//	This needs to be done because I need to marshal this into my factom entry.
func (c *OraclePriceRecord) CloneEntryData() *OraclePriceRecord {
	n := new(OraclePriceRecord)
	n.OPRChainID = c.OPRChainID
	n.Dbht = c.Dbht
	n.Version = c.Version
	copy(n.WinPreviousOPR[:], c.WinPreviousOPR[:])
	n.CoinbaseAddress = c.CoinbaseAddress
	n.CoinbasePNTAddress = c.CoinbasePNTAddress

	n.FactomDigitalID = c.FactomDigitalID
	n.Assets = make(OraclePriceRecordAssetList)
	for k, v := range c.Assets {
		n.Assets[k] = v
	}
	return n
}

// LX holds an instance of lxrhash
var LX lxr.LXRHash
var lxInitializer sync.Once

// The init function for LX is expensive. So we should explicitly call the init if we intend
// to use it. Make the init call idempotent
func InitLX() {
	lxInitializer.Do(func() {
		// This code will only be executed ONCE, no matter how often you call it
		LX.Verbose(true)
		if size, err := strconv.Atoi(os.Getenv("LXRBITSIZE")); err == nil && size >= 8 && size <= 30 {
			LX.Init(0xfafaececfafaecec, uint64(size), 256, 5)
		} else {
			LX.Init(0xfafaececfafaecec, 30, 256, 5)
		}

	})
}

// OPRChainID is the calculated chain id of the records chain
var OPRChainID string

// Token is a combination of currency code and value
type Token struct {
	code  string
	value float64
}

// Validate performs sanity checks of the structure and values of the OPR.
// It does not validate the winners of the previous block.
func (opr *OraclePriceRecord) Validate(c *config.Config, dbht int64) bool {

	// Validate there are no 0's
	for k, v := range opr.Assets {
		if v == 0 && k != "PNT" { // PNT is exception until we get a value for it
			return false
		}
	}

	if int64(opr.Dbht) != dbht {
		return false // DBHeight is not reported correctly
	}

	if opr.Version != OPRVersion {
		return false // We only support this version
	}

	// Validate all the Assets exists
	return opr.Assets.ContainsExactly(common.AllAssets)
}

// GetTokens creates an iterateable slice of Tokens containing all the currency values
func (opr *OraclePriceRecord) GetTokens() (tokens []Token) {
	return opr.Assets.List()
}

// GetHash returns the LXHash over the OPR's json representation
func (opr *OraclePriceRecord) GetHash() []byte {
	if len(opr.OPRHash) > 0 {
		return opr.OPRHash
	}
	data, err := json.Marshal(opr)
	common.CheckAndPanic(err)
	sha := sha256.Sum256(data)
	opr.OPRHash = sha[:]
	return opr.OPRHash
}

// ComputeDifficulty gets the difficulty by taking the hash of the OPRHash
// appended by the nonce. The difficulty is the highest 8 bytes of the hash
// taken as uint64 in Big Endian
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

func ComputeDifficulty(oprhash, nonce []byte) (difficulty uint64) {
	no := append(oprhash, nonce...)
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

// ShortString returns a human readable string with select data
func (opr *OraclePriceRecord) ShortString() string {
	str := fmt.Sprintf("DID %30x OPRHash %30x Nonce %33x Difficulty %15x Grade %20f",
		opr.FactomDigitalID,
		opr.OPRHash,
		opr.Nonce,
		opr.Difficulty,
		opr.Grade)
	return str
}

// String returns a human readable string for the Oracle Record
func (opr *OraclePriceRecord) String() (str string) {
	str = fmt.Sprintf("Nonce %x\n", opr.Nonce)
	str = str + fmt.Sprintf("%32s %v\n", "Difficulty", opr.Difficulty)
	str = str + fmt.Sprintf("%32s %v\n", "Directory Block Height", opr.Dbht)
	str = str + fmt.Sprintf("%32s %v\n", "WinningPreviousOPRs", "")
	for i, v := range opr.WinPreviousOPR {
		str = str + fmt.Sprintf("%32s %2d, %s\n", "", i+1, v)
	}
	str = str + fmt.Sprintf("%32s %s\n", "Coinbase PNT", opr.CoinbasePNTAddress)

	// Make a display string out of the Digital Identity.

	str = str + fmt.Sprintf("%32s %v\n", "FactomDigitalID", opr.FactomDigitalID)
	for _, asset := range opr.Assets.List() {
		str = str + fmt.Sprintf("%32s %v\n", "PNT", asset)
	}

	str = str + fmt.Sprintf("\nWinners\n\n")

	pwin := GetPreviousOPRs(opr.Dbht - 1)

	// If there were previous winners, we need to make sure this miner is running
	// the software to detect them, and that we agree with their conclusions.
	if pwin != nil {
		for i, v := range opr.WinPreviousOPR {
			balance := GetBalance(pwin[i].CoinbasePNTAddress)
			hbal := humanize.Comma(balance)
			str = str + fmt.Sprintf("   %16s %16x %30s %-56s = %10s\n",
				v,
				pwin[i].EntryHash[:8],
				pwin[i].FactomDigitalID,
				pwin[i].CoinbasePNTAddress,
				hbal,
			)
		}
	}
	return str
}

// LogFieldsShort returns a set of common fields to be included in logrus
func (opr *OraclePriceRecord) LogFieldsShort() log.Fields {
	return log.Fields{
		"did":        opr.FactomDigitalID,
		"opr_hash":   hex.EncodeToString(opr.OPRHash),
		"nonce":      hex.EncodeToString(opr.Nonce),
		"difficulty": opr.Difficulty,
		"grade":      opr.Grade,
	}
}

// SetPegValues assigns currency polling values to the OPR
func (opr *OraclePriceRecord) SetPegValues(assets polling.PegAssets) {
	for asset, v := range assets {
		opr.Assets[asset] = v.Value
	}
}

// NewOpr collects all the information unique to this miner and its configuration, and also
// goes and gets the oracle data.  Also collects the winners from the prior block and
// puts their entry hashes (base58) into this OPR
func NewOpr(ctx context.Context, minerNumber int, dbht int32, c *config.Config, alert chan *OPRs) (opr *OraclePriceRecord, err error) {
	opr = NewOraclePriceRecord()

	// Get the Identity Chain Specification
	if did, err := c.String("Miner.IdentityChain"); err != nil {
		return nil, errors.New("config file has no Miner.IdentityChain specified")
	} else {
		if minerNumber > 0 {
			did = fmt.Sprintf("%sminer%03d", did, minerNumber)
		}
		opr.FactomDigitalID = did
	}

	// Get the protocol chain to be used for pegnetMining records
	protocol, err1 := c.String("Miner.Protocol")
	network, err2 := common.LoadConfigNetwork(c)
	opr.Network = network
	opr.Protocol = protocol

	if err1 != nil {
		return nil, errors.New("config file has no Miner.Protocol specified")
	}
	if err2 != nil {
		return nil, errors.New("config file has no Miner.Network specified")
	}
	opr.OPRChainID = base58.Encode(common.ComputeChainIDFromStrings([]string{protocol, network, common.OPRChainTag}))

	opr.Dbht = dbht

	// If this is a test network, then give multiple miners their own tPNT address
	// because that is way more useful debugging than giving all miners the same
	// PNT address.  Otherwise, give all miners the same PNT address because most
	// users really doing mining will mostly be happen sending rewards to a single
	// address.
	if network == common.TestNetwork && minerNumber != 0 {
		fct := common.DebugFCTaddresses[minerNumber][1]
		opr.CoinbaseAddress = fct
	} else {
		if str, err := c.String("Miner.CoinbaseAddress"); err != nil {
			return nil, errors.New("config file has no Coinbase PNT Address")
		} else {
			opr.CoinbaseAddress = str
		}
	}

	opr.CoinbasePNTAddress, err = common.ConvertFCTtoPegNetAsset(network, "PNT", opr.CoinbaseAddress)
	if err != nil {
		log.Errorf("invalid fct address in config file: %v", err)
	}

	var winners *OPRs
	select {
	case winners = <-alert: // Wait for winner
	case <-ctx.Done(): // If we get cancelled
		return nil, context.Canceled
	}

	if winners.Error != nil {
		return nil, winners.Error
	}

	for i, w := range winners.ToBePaid {
		opr.WinPreviousOPR[i] = hex.EncodeToString(w.EntryHash[:8])
	}

	if len(winners.AllOPRs) > 0 {
		cutoff, _ := c.Int(common.ConfigSubmissionCutOff)
		if cutoff > 0 { // <= 0 disables it
			// This will calculate a minimum difficulty floor for our target cutoff.
			opr.MinimumDifficulty = CalculateMinimumDifficultyFromOPRs(winners.AllOPRs, cutoff)
		}
	}

	err = opr.GetOPRecord(c)
	if err != nil {
		return nil, err
	}

	return opr, nil
}

// GetOPRecord initializes the OPR with polling data and factom entry
func (opr *OraclePriceRecord) GetOPRecord(c *config.Config) error {
	InitDataSource(c) // Kinda odd to have this here.
	//get asset values
	Peg, err := PollingDataSource.PullAllPEGAssets()
	if err != nil {
		return err
	}
	opr.SetPegValues(Peg)

	data, err := json.Marshal(opr)
	if err != nil {
		panic(err)
	}
	sha := sha256.Sum256(data)
	opr.OPRHash = sha[:]
	return nil
}

// CreateOPREntry will create the entry from the EXISITING data.
// It will not set any fields like in `GetOPRecord`
func (opr *OraclePriceRecord) CreateOPREntry(nonce []byte, difficulty uint64) (*factom.Entry, error) {
	var err error
	e := new(factom.Entry)

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, difficulty)

	e.ChainID = hex.EncodeToString(base58.Decode(opr.OPRChainID))
	e.ExtIDs = [][]byte{nonce, buf, []byte{opr.Version}}
	e.Content, err = json.Marshal(opr)
	if err != nil {
		return nil, err
	}
	return e, nil
}
