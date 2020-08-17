package spr

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
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
	// These fields are not part of the SPR, but track values associated with the SPR.
	Protocol           string `json:"-"` // The Protocol we are running on (PegNet)
	Network            string `json:"-"` // The network we are running on (TestNet vs MainNet)
	SPRHash            []byte `json:"-"` // The hash of the SPR record (used by PegNet Staking)
	SPRChainID         string `json:"-"` // [base58]  Chain ID of the chain used by the Oracle Stakers
	CoinbasePEGAddress string `json:"-"` // [base58]  Payout Address

	// Factom Entry data
	EntryHash []byte `json:"-"` // Entry to record this record
	Version   uint8  `json:"-"`

	// These values define the context of the SPR, and they go into the PegNet SPR record, and are staked.
	CoinbaseAddress string                      `json:"coinbase"` // [base58]  Payout Address
	Dbht            int32                       `json:"dbht"`     //           The Directory Block Height of the SPR.
	Assets          StakingPriceRecordAssetList `json:"assets"`   // The Oracle values of the SPR, they are the meat of the SPR record, and are staked.
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
	n.CoinbaseAddress = c.CoinbaseAddress
	n.CoinbasePEGAddress = c.CoinbasePEGAddress

	n.Assets = make(StakingPriceRecordAssetList)
	for k, v := range c.Assets {
		n.Assets[k] = v
	}
	return n
}

// Token is a combination of currency Code and Value
type Token struct {
	Code  string
	Value float64
}

// Validate performs sanity checks of the structure and values of the SPR.
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
	// Only enforce on version 2 and forward, checking valid FCT address
	if !ValidFCTAddress(spr.CoinbaseAddress) {
		return false
	}

	if int64(spr.Dbht) != dbht {
		return false // DBHeight is not reported correctly
	}

	// Validate all the Assets exists
	switch spr.Version {
	case 5:
		return spr.Assets.ContainsExactly(common.AssetsV5)
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
func (spr *StakingPriceRecord) GetTokens() (tokens []Token) {
	return spr.Assets.List(spr.Version)
}

// GetHash returns the LXHash over the SPR's json representation
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
	str := fmt.Sprintf("SPRHash %30x", opr.SPRHash)
	return str
}

// String returns a human readable string for the Oracle Record
func (spr *StakingPriceRecord) String() (str string) {
	str = fmt.Sprintf("%32s %v\n", "Directory Block Height", spr.Dbht)
	str = str + fmt.Sprintf("%32s %s\n", "Coinbase PEG", spr.CoinbasePEGAddress)
	for _, asset := range spr.Assets.List(spr.Version) {
		str = str + fmt.Sprintf("%32s %v\n", "PEG", asset)
	}
	return str
}

// LogFieldsShort returns a set of common fields to be included in logrus
func (spr *StakingPriceRecord) LogFieldsShort() log.Fields {
	return log.Fields{
		"spr_hash": hex.EncodeToString(spr.SPRHash),
	}
}

// SetPegValues assigns currency polling values to the SPR
func (spr *StakingPriceRecord) SetPegValues(assets polling.PegAssets) {
	for asset, v := range assets {
		spr.Assets.SetValue(asset, v.Value)
	}
}

// NewSpr collects all the information unique to this staker and its configuration, and also
// goes and gets the oracle data.
func NewSpr(ctx context.Context, dbht int32, c *config.Config) (spr *StakingPriceRecord, err error) {
	spr = NewStakingPriceRecord()

	/**
	 *	Init SPR with configuration settings
	 */
	protocol, err1 := c.String("Staker.Protocol")
	network, err2 := common.LoadConfigStakerNetwork(c)
	spr.Network = network
	spr.Protocol = protocol

	if err1 != nil {
		return nil, errors.New("config file has no Staker.Protocol specified")
	}
	if err2 != nil {
		return nil, errors.New("config file has no Staker.Network specified")
	}
	spr.SPRChainID = base58.Encode(common.ComputeChainIDFromStrings([]string{protocol, network, common.SPRChainTag}))
	spr.Dbht = dbht
	spr.Version = common.SPRVersion(spr.Network, int64(spr.Dbht))

	if network == common.TestNetwork {
		fct := common.DebugFCTaddresses[0][1]
		spr.CoinbaseAddress = fct
	} else {
		if str, err := c.String("Staker.CoinbaseAddress"); err != nil {
			return nil, errors.New("config file has no Coinbase PEG Address")
		} else {
			spr.CoinbaseAddress = str
		}
	}

	spr.CoinbasePEGAddress, err = common.ConvertFCTtoPegNetAsset(network, "PEG", spr.CoinbaseAddress)
	if err != nil {
		log.Errorf("invalid fct address in config file: %v", err)
	}

	/**
	 *	Get SPR Record with Assets data (polling)
	 */
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
	//get asset values
	Peg, err := PollingDataSource.PullAllPEGAssets(uint8(spr.Version))
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

//  Staking Price Record (SPR)
//	ExtIDs:
//		version byte (byte, default 5)
//		RCD of the payout address
//		signature covering [ExtID]
//	Content: (protobuf)
//		Payout Address (string)
//		Height (int32)
//		Assets ([]uint64)
func (spr *StakingPriceRecord) CreateSPREntry() (*factom.Entry, error) {
	e := new(factom.Entry)
	e.ChainID = hex.EncodeToString(base58.Decode(spr.SPRChainID))
	rcd, err := common.ConvertFCTtoRaw(spr.CoinbaseAddress)
	if err != nil {
		return nil, err
	}
	e.ExtIDs = [][]byte{{spr.Version}, rcd, spr.GetHash()}
	e.Content, err = spr.SafeMarshal()
	if err != nil {
		return nil, err
	}
	return e, nil
}

// SafeMarshal will marshal the json depending on the opr version
func (spr *StakingPriceRecord) SafeMarshal() ([]byte, error) {
	// our opr version must be set before entering this
	if spr.Version < 5 {
		return nil, fmt.Errorf("spr version is not set")
	}

	if spr.Assets == nil {
		return nil, fmt.Errorf("assets is nil, cannot marshal")
	}

	if spr.Version == 5 {
		assetList := common.AssetsV5
		prices := make([]uint64, len(spr.Assets))

		for i, asset := range assetList {
			prices[i] = spr.Assets[asset]
		}

		pOpr := &oprencoding.ProtoOPR{
			Address: spr.CoinbaseAddress,
			ID:      "",
			Height:  spr.Dbht,
			Assets:  prices,
			Winners: nil,
		}

		return proto.Marshal(pOpr)
	}

	return nil, fmt.Errorf("spr version %d not supported", spr.Version)
}

// SafeMarshal will unmarshal the json
func (spr *StakingPriceRecord) SafeUnmarshal(data []byte) error {
	// our opr version must be set before entering this
	if spr.Version < 5 {
		return fmt.Errorf("opr version is 0")
	}

	if spr.Version == 5 {
		protoOPR := oprencoding.ProtoOPR{}
		err := proto.Unmarshal(data, &protoOPR)
		if err != nil {
			return err
		}

		assetList := common.AssetsV5
		spr.Assets = make(StakingPriceRecordAssetList)
		// Populate the original opr
		spr.CoinbaseAddress = protoOPR.Address
		spr.Dbht = protoOPR.Height

		if len(protoOPR.Assets) != len(assetList) {
			return fmt.Errorf("found %d assets, expected %d", len(protoOPR.Assets), len(assetList))
		}

		// Hard coded list of assets
		for i, asset := range assetList {
			spr.Assets[asset] = protoOPR.Assets[i]
		}

		return nil
	}

	return fmt.Errorf("spr version %d not supported", spr.Version)
}
