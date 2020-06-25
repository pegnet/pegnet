package spr

import (
	"context"
	"github.com/zpatrick/go-config"
)

// StakingPriceRecord is the data used and created by staker
type StakingPriceRecord struct {
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
	Assets StakingPriceRecordAssetList `json:"assets"`
}

func NewOraclePriceRecord() *StakingPriceRecord {
	o := new(StakingPriceRecord)
	o.Assets = make(StakingPriceRecordAssetList)

	return o
}

// Token is a combination of currency Code and Value
type Token struct {
	Code  string
	Value float64
}

// NewSpr collects all the information unique to this staker and its configuration, and also
// goes and gets the oracle data.  Also collects the winners from the prior block and
// puts their entry hashes (base58) into this SPR
func NewSpr(ctx context.Context, dbht int32, c *config.Config) (spr *StakingPriceRecord, err error) {
	spr = NewOraclePriceRecord()
	return spr, nil
}
