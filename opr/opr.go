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
	"github.com/golang/protobuf/proto"
	lxr "github.com/pegnet/LXRHash"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr/oprencoding"
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
		if PollingDataSource == nil { // This can be inited from unit tests
			PollingDataSource = polling.NewDataSources(config)
		}
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
	Assets OraclePriceRecordAssetList `json:"assets"`
}

func NewOraclePriceRecord() *OraclePriceRecord {
	o := new(OraclePriceRecord)
	o.Assets = make(OraclePriceRecordAssetList)

	return o
}

// CloneEntryData will clone the OPR data needed to make a factom entry.
//	This needs to be done because I need to marshal this into my factom entry.
func (c *OraclePriceRecord) CloneEntryData() *OraclePriceRecord {
	n := NewOraclePriceRecord()
	n.OPRChainID = c.OPRChainID
	n.Dbht = c.Dbht
	n.Version = c.Version
	n.WinPreviousOPR = make([]string, len(c.WinPreviousOPR), len(c.WinPreviousOPR))
	copy(n.WinPreviousOPR[:], c.WinPreviousOPR[:])
	n.CoinbaseAddress = c.CoinbaseAddress
	n.CoinbasePEGAddress = c.CoinbasePEGAddress

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

// Token is a combination of currency Code and Value
type Token struct {
	Code  string
	Value float64
}

// Validate performs sanity checks of the structure and values of the OPR.
// It does not validate the winners of the previous block.
func (opr *OraclePriceRecord) Validate(c *config.Config, dbht int64) bool {
	net, _ := common.LoadConfigNetwork(c)
	if !common.NetworkActive(net, dbht) {
		return false
	}

	// Validate there are no 0's
	for k, v := range opr.Assets {
		if v == 0 && k != "PEG" { // PEG is exception until we get a value for it
			return false
		}
	}

	// Only enforce on version 2 and forward
	if err := common.ValidIdentity(opr.FactomDigitalID); opr.Version == 2 && err != nil {
		return false
	}

	// Only enforce on version 2 and forward, checking valid FCT address
	if opr.Version == 2 && !ValidFCTAddress(opr.CoinbaseAddress) {
		return false
	}

	if int64(opr.Dbht) != dbht {
		return false // DBHeight is not reported correctly
	}

	if opr.Version != common.OPRVersion(net, int64(dbht)) {
		return false // We only support this version
	}

	// Validate all the Assets exists
	switch opr.Version {
	case 1:
		if len(opr.WinPreviousOPR) != 10 {
			return false
		}
		return opr.Assets.ContainsExactly(common.AssetsV1)
	case 2:
		// It can contain 10 winners when it is a transition record
		return opr.Assets.ContainsExactly(common.AssetsV2)
	default:
		return false
	}
}

// ValidFCTAddress will be removed in the grading module refactor. This is just temporary to get this
// functionality, and be easily unit testable.
func ValidFCTAddress(addr string) bool {
	return len(addr) > 2 && addr[:2] == "FA" && factom.IsValidAddress(addr)
}

// GetTokens creates an iterateable slice of Tokens containing all the currency values
func (opr *OraclePriceRecord) GetTokens() (tokens []Token) {
	return opr.Assets.List(opr.Version)
}

// GetHash returns the LXHash over the OPR's json representation
func (opr *OraclePriceRecord) GetHash() []byte {
	if len(opr.OPRHash) > 0 {
		return opr.OPRHash
	}

	// SafeMarshal handles the PNT/PEG issue
	data, err := opr.SafeMarshal()
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
	str = str + fmt.Sprintf("%32s %s\n", "Coinbase PEG", opr.CoinbasePEGAddress)

	// Make a display string out of the Digital Identity.

	str = str + fmt.Sprintf("%32s %v\n", "FactomDigitalID", opr.FactomDigitalID)
	for _, asset := range opr.Assets.List(opr.Version) {
		str = str + fmt.Sprintf("%32s %v\n", "PEG", asset)
	}

	str = str + fmt.Sprintf("\nWinners\n\n")

	// If there were previous winners, we need to make sure this miner is running
	// the software to detect them, and that we agree with their conclusions.
	for i, v := range opr.WinPreviousOPR {
		str = str + fmt.Sprintf("   %2d\t%16s\n",
			i,
			v,
		)
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
	// TODO: Remove when version 2 is activated
	switch common.OPRVersion(opr.Network, int64(opr.Dbht)) {
	case 1:
		for asset, v := range assets {
			opr.Assets.SetValue(asset, v.Value)
		}
	case 2:
		for asset, v := range assets {
			// Skip XPT and XPD
			if asset == "XPT" || asset == "XPD" {
				continue
			}
			opr.Assets.SetValue(asset, v.Value)
		}
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
	opr.Version = common.OPRVersion(opr.Network, int64(opr.Dbht))
	OPRVersion = opr.Version // Update this for the polling to know

	// If this is a test network, then give multiple miners their own tPEG address
	// because that is way more useful debugging than giving all miners the same
	// PEG address.  Otherwise, give all miners the same PEG address because most
	// users really doing mining will mostly be happen sending rewards to a single
	// address.
	if network == common.TestNetwork && minerNumber != 0 {
		fct := common.DebugFCTaddresses[minerNumber][1]
		opr.CoinbaseAddress = fct
	} else {
		if str, err := c.String("Miner.CoinbaseAddress"); err != nil {
			return nil, errors.New("config file has no Coinbase PEG Address")
		} else {
			opr.CoinbaseAddress = str
		}
	}

	opr.CoinbasePEGAddress, err = common.ConvertFCTtoPegNetAsset(network, "PEG", opr.CoinbaseAddress)
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

	// For the transition, we need to support a 10 winner opr.
	// The winner's should be correct from our grader, so we will accept it
	if len(winners.ToBePaid) > 0 {
		opr.WinPreviousOPR = make([]string, len(winners.ToBePaid), len(winners.ToBePaid))
		for i, w := range winners.ToBePaid {
			opr.WinPreviousOPR[i] = hex.EncodeToString(w.EntryHash[:8])
		}
	} else {
		// If there are no previous winners, this is a bootstrap record
		min := 0
		switch common.OPRVersion(network, int64(dbht)) {
		case 1:
			min = 10
		case 2:
			min = 25
		}
		opr.WinPreviousOPR = make([]string, min, min)
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

	if !opr.Validate(c, int64(dbht)) {
		// TODO: Remove this custom error handle once the network is live.
		//		This is just to give a better error when are waiting for activation.
		if !common.NetworkActive(opr.Network, int64(dbht)) {
			return nil, fmt.Errorf("Waiting for activation height")
		}
		return nil, fmt.Errorf("opr invalid")
	}

	return opr, nil
}

// GetOPRecord initializes the OPR with polling data and factom entry
func (opr *OraclePriceRecord) GetOPRecord(c *config.Config) error {
	InitDataSource(c) // Kinda odd to have this here.
	//get asset values
	Peg, err := PollingDataSource.PullAllPEGAssets(opr.Version)
	if err != nil {
		return err
	}
	opr.SetPegValues(Peg)

	data, err := opr.SafeMarshal()
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
	e.ExtIDs = [][]byte{nonce, buf, {opr.Version}}
	e.Content, err = opr.SafeMarshal()
	if err != nil {
		return nil, err
	}
	return e, nil
}

// SafeMarshal will marshal the json depending on the opr version
func (opr *OraclePriceRecord) SafeMarshal() ([]byte, error) {
	// our opr version must be set before entering this
	if opr.Version == 0 {
		return nil, fmt.Errorf("opr version is 0")
	}

	// This function relies on the assets, so check up front
	if opr.Assets == nil {
		return nil, fmt.Errorf("assets is nil, cannot marshal")
	}

	// When we marshal a version 1 opr, we need to change PEG -> PNT
	// No opr in the code should ever have 'PNT'. We only use PNT in the marshal
	// function, no where else.
	if _, ok := opr.Assets["PNT"]; ok {
		return nil, fmt.Errorf("this opr has asset 'PNT', it should have 'PEG'")
	}

	// Version 1 we json marshal and
	// do the swap of PEG -> PNT
	if opr.Version == 1 {
		opr.Assets["PNT"] = opr.Assets["PEG"]
		delete(opr.Assets, "PEG")

		// This is a known key that will be removed by the marshal json function. It indicates
		// to the marshaler that it was called from a safe path. This is not the cleanest method,
		// but to override the json function, and still use the default, it would require an odd
		// structure nesting and a lot of code changes
		opr.Assets["version"] = uint64(opr.Version)
		data, err := json.Marshal(opr)
		delete(opr.Assets, "version") // Should be deleted by the json.Marshal, but that can error out

		// Revert the swap
		opr.Assets["PEG"] = opr.Assets["PNT"]
		delete(opr.Assets, "PNT")
		return data, err
	} else if opr.Version == 2 {
		// Version 2 uses Protobufs for encoding
		pOpr := &oprencoding.ProtoOPR{
			Address: opr.CoinbaseAddress,
			Id:      opr.FactomDigitalID,
			Height:  opr.Dbht,
			// Hardcoded list order.
			Assets: &oprencoding.Assets{
				PEG:   opr.Assets.Uint64Value("PEG"),
				PUSD:  opr.Assets.Uint64Value("USD"),
				PEUR:  opr.Assets.Uint64Value("EUR"),
				PJPY:  opr.Assets.Uint64Value("JPY"),
				PGBP:  opr.Assets.Uint64Value("GBP"),
				PCAD:  opr.Assets.Uint64Value("CAD"),
				PCHF:  opr.Assets.Uint64Value("CHF"),
				PINR:  opr.Assets.Uint64Value("INR"),
				PSGD:  opr.Assets.Uint64Value("SGD"),
				PCNY:  opr.Assets.Uint64Value("CNY"),
				PHKD:  opr.Assets.Uint64Value("HKD"),
				PKRW:  opr.Assets.Uint64Value("KRW"),
				PBRL:  opr.Assets.Uint64Value("BRL"),
				PPHP:  opr.Assets.Uint64Value("PHP"),
				PMXN:  opr.Assets.Uint64Value("MXN"),
				PXAU:  opr.Assets.Uint64Value("XAU"),
				PXAG:  opr.Assets.Uint64Value("XAG"),
				PXBT:  opr.Assets.Uint64Value("XBT"),
				PETH:  opr.Assets.Uint64Value("ETH"),
				PLTC:  opr.Assets.Uint64Value("LTC"),
				PRVN:  opr.Assets.Uint64Value("RVN"),
				PXBC:  opr.Assets.Uint64Value("XBC"),
				PFCT:  opr.Assets.Uint64Value("FCT"),
				PBNB:  opr.Assets.Uint64Value("BNB"),
				PXLM:  opr.Assets.Uint64Value("XLM"),
				PADA:  opr.Assets.Uint64Value("ADA"),
				PXMR:  opr.Assets.Uint64Value("XMR"),
				PDASH: opr.Assets.Uint64Value("DASH"),
				PZEC:  opr.Assets.Uint64Value("ZEC"),
				PDCR:  opr.Assets.Uint64Value("DCR"),
			},
		}

		// Decode winners into strings
		var err error
		pOpr.Winners = make([][]byte, len(opr.WinPreviousOPR), len(opr.WinPreviousOPR))
		for i, winner := range opr.WinPreviousOPR {
			pOpr.Winners[i], err = hex.DecodeString(winner)
			if err != nil {
				return nil, err
			}
		}

		data, err := proto.Marshal(pOpr)
		return data, err
	}

	return nil, fmt.Errorf("opr version %d not supported", opr.Version)
}

// SafeMarshal will unmarshal the json depending on the opr version
func (opr *OraclePriceRecord) SafeUnmarshal(data []byte) error {
	// our opr version must be set before entering this
	if opr.Version == 0 {
		return fmt.Errorf("opr version is 0")
	}

	// If version 1, we need to json unmarshal and swap PNT and PEG
	if opr.Version == 1 {
		err := json.Unmarshal(data, opr)
		if err != nil {
			return err
		}

		if v, ok := opr.Assets["PNT"]; ok {
			opr.Assets["PEG"] = v
			delete(opr.Assets, "PNT")
		} else {
			return fmt.Errorf("exp version 1 to have 'PNT', but it did not")
		}
		return nil
	} else if opr.Version == 2 {
		oprMin := oprencoding.ProtoOPR{}
		err := proto.Unmarshal(data, &oprMin)
		if err != nil {
			return err
		}

		opr.Assets = make(OraclePriceRecordAssetList)
		// Populate the original opr
		opr.CoinbaseAddress = oprMin.Address
		opr.FactomDigitalID = oprMin.Id
		opr.Dbht = oprMin.Height
		//opr.WinPreviousOPR = oprMin.Winners
		// Hard coded list of assets
		opr.Assets.SetValueFromUint64("PEG", oprMin.Assets.PEG)
		opr.Assets.SetValueFromUint64("USD", oprMin.Assets.PUSD)
		opr.Assets.SetValueFromUint64("EUR", oprMin.Assets.PEUR)
		opr.Assets.SetValueFromUint64("JPY", oprMin.Assets.PJPY)
		opr.Assets.SetValueFromUint64("GBP", oprMin.Assets.PGBP)
		opr.Assets.SetValueFromUint64("CAD", oprMin.Assets.PCAD)
		opr.Assets.SetValueFromUint64("CHF", oprMin.Assets.PCHF)
		opr.Assets.SetValueFromUint64("INR", oprMin.Assets.PINR)
		opr.Assets.SetValueFromUint64("SGD", oprMin.Assets.PSGD)
		opr.Assets.SetValueFromUint64("CNY", oprMin.Assets.PCNY)
		opr.Assets.SetValueFromUint64("HKD", oprMin.Assets.PHKD)
		opr.Assets.SetValueFromUint64("KRW", oprMin.Assets.PKRW)
		opr.Assets.SetValueFromUint64("BRL", oprMin.Assets.PBRL)
		opr.Assets.SetValueFromUint64("PHP", oprMin.Assets.PPHP)
		opr.Assets.SetValueFromUint64("MXN", oprMin.Assets.PMXN)
		opr.Assets.SetValueFromUint64("XAU", oprMin.Assets.PXAU)
		opr.Assets.SetValueFromUint64("XAG", oprMin.Assets.PXAG)
		opr.Assets.SetValueFromUint64("XBT", oprMin.Assets.PXBT)
		opr.Assets.SetValueFromUint64("ETH", oprMin.Assets.PETH)
		opr.Assets.SetValueFromUint64("LTC", oprMin.Assets.PLTC)
		opr.Assets.SetValueFromUint64("RVN", oprMin.Assets.PRVN)
		opr.Assets.SetValueFromUint64("XBC", oprMin.Assets.PXBC)
		opr.Assets.SetValueFromUint64("FCT", oprMin.Assets.PFCT)
		opr.Assets.SetValueFromUint64("BNB", oprMin.Assets.PBNB)
		opr.Assets.SetValueFromUint64("XLM", oprMin.Assets.PXLM)
		opr.Assets.SetValueFromUint64("ADA", oprMin.Assets.PADA)
		opr.Assets.SetValueFromUint64("XMR", oprMin.Assets.PXMR)
		opr.Assets.SetValueFromUint64("DASH", oprMin.Assets.PDASH)
		opr.Assets.SetValueFromUint64("ZEC", oprMin.Assets.PZEC)
		opr.Assets.SetValueFromUint64("DCR", oprMin.Assets.PDCR)

		// Decode winners
		opr.WinPreviousOPR = make([]string, len(oprMin.Winners), len(oprMin.Winners))
		for i, winner := range oprMin.Winners {
			opr.WinPreviousOPR[i] = hex.EncodeToString(winner)
		}

		return nil
	}

	return fmt.Errorf("opr version %d not supported", opr.Version)
}
