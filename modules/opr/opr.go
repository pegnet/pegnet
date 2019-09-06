package opr

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/pegnet/pegnet/polling"

	"github.com/golang/protobuf/proto"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr/oprencoding"
)

// OPR
type OPR struct {
	// Extid related data
	EntryHash              []byte `json:"-"` // Entry to record this opr
	Nonce                  []byte `json:"-"` // Nonce in the extid for the pow of the OPR
	SelfReportedDifficulty []byte `json:"-"` // Miners self report their difficulty computed by lxrhash
	Version                uint8  `json:"-"` // OPR version for encoding rules

	// OPRHash is determined by the entry content. We should set this when we retrieve the OPR
	OPRHash []byte `json:"-"`
	// TODO: Should we also include the raw content? If we re-marshal, for v1 the content is not canonical

	// TODO: Should we include these, or find a way to exclude them?
	// Values computed during grading
	Grade      float64 `json:"-"`
	Difficulty uint64  `json:"-"`

	// ----- Marshaled Content into Factom Entry ----
	// These values define the context of the OPR, and they go into the PegNet OPR record, and are mined.
	CoinbaseAddress string   `json:"coinbase"` // [base58]  PEG Address to pay PEG
	Dbht            int32    `json:"dbht"`     //           The Directory Block Height of the OPR.
	WinPreviousOPR  []string `json:"winners"`  // First 8 bytes of the Entry Hashes of the previous winners
	FactomDigitalID string   `json:"minerid"`  // [unicode] Digital Identity of the miner

	// The Oracle values of the OPR, they are the meat of the OPR record, and are mined.
	Assets *OraclePriceRecordAssetList `json:"assets"`
}

// RandomOPR is useful for unit testing
//
//	Difficulty and grade is UNSET
func RandomOPR(version uint8) *OPR {
	o := new(OPR)

	o.Version = version

	o.EntryHash = make([]byte, 32)
	o.OPRHash = make([]byte, 32)
	o.SelfReportedDifficulty = make([]byte, 8)

	_, _ = rand.Read(o.EntryHash)
	_, _ = rand.Read(o.OPRHash)
	_, _ = rand.Read(o.SelfReportedDifficulty)

	o.CoinbaseAddress = common.ConvertRawToFCT(common.RandomByteSliceOfLen(32))
	o.Dbht = rand.Int31()
	o.WinPreviousOPR = make([]string, common.NumberOfWinners(o.Version), common.NumberOfWinners(o.Version))
	o.FactomDigitalID = hex.EncodeToString(o.SelfReportedDifficulty) // 16 hex characters

	assets := common.AssetsV1
	if version == 2 {
		assets = common.AssetsV2
	}
	o.Assets = NewOraclePriceRecordAssetList(o.Version)
	for _, asset := range assets {
		price := rand.Uint64()
		o.Assets.SetValueFromUint64(asset, price)
		if o.Version == 1 {
			// V1 is truncated to 4
			o.Assets.SetValue(asset, polling.TruncateTo4(o.Assets.Value(asset)))
		}
	}
	o.Assets.AssetList["PEG"] = 0 // PEG is 0

	return o
}

// Validate performs sanity checks of the structure and values of the OPR.
// It does not validate the winners of the previous block, and will validate according
// to the version in the OPR struct.
//
// Returns
//		err		Error with reason if invalid
func (opr *OPR) Validate(dbht int32) error {
	if opr.Assets == nil {
		return fmt.Errorf("assetlist is nil")
	}

	// Validate there are no 0's
	for k, v := range opr.Assets.AssetList {
		if v == 0 && k != "PEG" { // PEG is exception until we get a value for it
			return fmt.Errorf("%s has a value of 0", k)
		}
	}

	// Only enforce on version 2 and forward
	if err := common.ValidIdentity(opr.FactomDigitalID); opr.Version == 2 && err != nil {
		return fmt.Errorf("%s is an invalid identity", opr.FactomDigitalID)
	}

	if opr.Dbht != dbht {
		return fmt.Errorf("height expected %d, found %d", dbht, opr.Dbht) // DBHeight is not reported correctly
	}

	// Validate all the Assets exists and the number of winners is correct
	switch opr.Version {
	case 1:
		// V1 only accepts length 10 for winners
		if len(opr.WinPreviousOPR) != 10 {
			return fmt.Errorf("must have 10 winners, found %d", len(opr.WinPreviousOPR))
		}
		if !opr.Assets.ContainsExactly(common.AssetsV1) {
			return fmt.Errorf("asset list not correct")
		}
	case 2:
		// It can contain 10 winners when it is a transition record
		// So v2 can have 10 or 25 winners
		if !(len(opr.WinPreviousOPR) == 10 || len(opr.WinPreviousOPR) == 25) {
			return fmt.Errorf("must have 10 or 25 winners, found %d", len(opr.WinPreviousOPR))
		}
		if !opr.Assets.ContainsExactly(common.AssetsV2) {
			return fmt.Errorf("asset list not correct")
		}
	default:
		return fmt.Errorf("exp version 1 or 2, found %d", opr.Version)
	}
	return nil // All good
}

func (opr *OPR) ShortEntryHash() string {
	if opr.EntryHash == nil {
		return ""
	}
	return fmt.Sprintf("%x", opr.EntryHash[:8])
}

// GetTokens creates an iterateable slice of Tokens containing all the currency values
func (opr *OPR) GetTokens() (tokens []Token) {
	return opr.Assets.List()
}

func (opr *OPR) ExtIDs() [][]byte {
	return [][]byte{
		opr.Nonce,
		opr.SelfReportedDifficulty,
		[]byte{opr.Version},
	}
}

// SafeMarshal will marshal the json depending on the opr version
func (opr *OPR) SafeMarshal() ([]byte, error) {
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
	if _, ok := opr.Assets.AssetList["PNT"]; ok {
		return nil, fmt.Errorf("this opr has asset 'PNT', it should have 'PEG'")
	}

	// TODO: Do we need to set this here?
	opr.Assets.Version = opr.Version

	// Version 1 we json marshal
	if opr.Version == 1 {
		data, err := json.Marshal(opr)
		return data, err
	} else if opr.Version == 2 {
		// Version 2 uses Protobufs for encoding
		pOpr := &oprencoding.ProtoOPRMin{
			Address: opr.CoinbaseAddress,
			Id:      opr.FactomDigitalID,
			Height:  opr.Dbht,
			Winners: opr.WinPreviousOPR,
			// Hardcoded list order.
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
		}
		data, err := proto.Marshal(pOpr)
		return data, err
	}

	return nil, fmt.Errorf("opr version %d not supported", opr.Version)
}

// SafeMarshal will unmarshal the json depending on the opr version
func (opr *OPR) SafeUnmarshal(data []byte) error {
	// our opr version must be set before entering this
	if opr.Version == 0 {
		return fmt.Errorf("opr version is 0")
	}

	if data == nil {
		return fmt.Errorf("nil data provided to marshal")
	}

	// Set this for the unmarshal functions
	opr.Assets = NewOraclePriceRecordAssetList(opr.Version)

	// If version 1, we need to json unmarshal and swap PNT and PEG
	if opr.Version == 1 {
		err := json.Unmarshal(data, opr)
		if err != nil {
			return err
		}
		return nil
	} else if opr.Version == 2 {
		oprMin := oprencoding.ProtoOPRMin{}
		err := proto.Unmarshal(data, &oprMin)
		if err != nil {
			return err
		}

		// Populate the original opr
		opr.CoinbaseAddress = oprMin.Address
		opr.FactomDigitalID = oprMin.Id
		opr.Dbht = oprMin.Height
		opr.WinPreviousOPR = oprMin.Winners
		// Hard coded list of assets
		opr.Assets.SetValueFromUint64("PEG", oprMin.PEG)
		opr.Assets.SetValueFromUint64("USD", oprMin.PUSD)
		opr.Assets.SetValueFromUint64("EUR", oprMin.PEUR)
		opr.Assets.SetValueFromUint64("JPY", oprMin.PJPY)
		opr.Assets.SetValueFromUint64("GBP", oprMin.PGBP)
		opr.Assets.SetValueFromUint64("CAD", oprMin.PCAD)
		opr.Assets.SetValueFromUint64("CHF", oprMin.PCHF)
		opr.Assets.SetValueFromUint64("INR", oprMin.PINR)
		opr.Assets.SetValueFromUint64("SGD", oprMin.PSGD)
		opr.Assets.SetValueFromUint64("CNY", oprMin.PCNY)
		opr.Assets.SetValueFromUint64("HKD", oprMin.PHKD)
		opr.Assets.SetValueFromUint64("KRW", oprMin.PKRW)
		opr.Assets.SetValueFromUint64("BRL", oprMin.PBRL)
		opr.Assets.SetValueFromUint64("PHP", oprMin.PPHP)
		opr.Assets.SetValueFromUint64("MXN", oprMin.PMXN)
		opr.Assets.SetValueFromUint64("XAU", oprMin.PXAU)
		opr.Assets.SetValueFromUint64("XAG", oprMin.PXAG)
		opr.Assets.SetValueFromUint64("XBT", oprMin.PXBT)
		opr.Assets.SetValueFromUint64("ETH", oprMin.PETH)
		opr.Assets.SetValueFromUint64("LTC", oprMin.PLTC)
		opr.Assets.SetValueFromUint64("RVN", oprMin.PRVN)
		opr.Assets.SetValueFromUint64("XBC", oprMin.PXBC)
		opr.Assets.SetValueFromUint64("FCT", oprMin.PFCT)
		opr.Assets.SetValueFromUint64("BNB", oprMin.PBNB)
		opr.Assets.SetValueFromUint64("XLM", oprMin.PXLM)
		opr.Assets.SetValueFromUint64("ADA", oprMin.PADA)
		opr.Assets.SetValueFromUint64("XMR", oprMin.PXMR)
		opr.Assets.SetValueFromUint64("DASH", oprMin.PDASH)
		opr.Assets.SetValueFromUint64("ZEC", oprMin.PZEC)
		opr.Assets.SetValueFromUint64("DCR", oprMin.PDCR)
		return nil
	}

	return fmt.Errorf("opr version %d not supported", opr.Version)
}
