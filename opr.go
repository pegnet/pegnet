package oprecord

// These

import (
	"encoding/binary"

	"errors"
	"fmt"
	"github.com/FactomProject/factom"
)

type OraclePriceRecord struct {
	Difficulty         uint64            // not part of OPR -The difficulty of the given nonce
	Grade              float64           // not part of OPR -The grade when OPR records are compared
	Nonce              [32]byte          // not part of OPR - nonce creacted by mining
	OPRHash            [32]byte          // not part of OPR - the hash of the OPR record
	EC                 *factom.ECAddress // not part of OPR - Entry Credit Address used by a miner
	ChainID            [32]byte
	VersionEntryHash   [32]byte
	WinningPreviousOPR [32]byte
	CoinbasePNTAddress [32]byte
	BlockReward        [8]byte
	FactomDigitalID    [32]byte
	PNT                [8]byte
	USD                [8]byte
	EUR                [8]byte
	JPY                [8]byte
	GBP                [8]byte
	CAD                [8]byte
	CHF                [8]byte
	INR                [8]byte
	SGD                [8]byte
	CNY                [8]byte
	HKD                [8]byte
	XAU                [8]byte
	XAG                [8]byte
	XPD                [8]byte
	XPT                [8]byte
	XBT                [8]byte
	ETH                [8]byte
	LTC                [8]byte
	XBC                [8]byte
	FCT                [8]byte
}

func (opr *OraclePriceRecord) GetTokens() (tokens []float64) {
	add := func(v [8]byte) {
		tokens = append(tokens, float64(binary.BigEndian.Uint64(v[:])))
	}
	add(opr.PNT)
	add(opr.USD)
	add(opr.EUR)
	add(opr.JPY)
	add(opr.GBP)
	add(opr.CAD)
	add(opr.CHF)
	add(opr.INR)
	add(opr.SGD)
	add(opr.CNY)
	add(opr.HKD)
	add(opr.XAU)
	add(opr.XAG)
	add(opr.XPD)
	add(opr.XPT)
	add(opr.XBT)
	add(opr.ETH)
	add(opr.LTC)
	add(opr.XBC)
	add(opr.FCT)
	return
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
	print32("ChainID", opr.ChainID[:])
	print32("VersionEntryHash", opr.VersionEntryHash[:])
	print32("WinningPreviousOPR", opr.WinningPreviousOPR[:])
	print32("CoinbasePNTAddress", opr.CoinbasePNTAddress[:])
	print32("BlockReward", opr.BlockReward[:])
	print32("FactomDigitalID", opr.FactomDigitalID[:])
	print32("PNT", opr.PNT[:])
	print32("USD", opr.USD[:])
	print32("EUR", opr.EUR[:])
	print32("JPY", opr.JPY[:])
	print32("GBP", opr.GBP[:])
	print32("CAD", opr.CAD[:])
	print32("CHF", opr.CHF[:])
	print32("INR", opr.INR[:])
	print32("SGD", opr.SGD[:])
	print32("CNY", opr.CNY[:])
	print32("HKD", opr.HKD[:])
	print32("XAU", opr.XAU[:])
	print32("XAG", opr.XAG[:])
	print32("XPD", opr.XPD[:])
	print32("XPT", opr.XPT[:])
	print32("XBT", opr.XBT[:])
	print32("ETH", opr.ETH[:])
	print32("LTC", opr.LTC[:])
	print32("XBC", opr.XBC[:])
	print32("FCT", opr.FCT[:])
	return str
}

func (opr *OraclePriceRecord) MarshalBinary() ([]byte, error) {
	record := []byte{}

	record = append(record, opr.ChainID[:]...)
	record = append(record, opr.VersionEntryHash[:]...)
	record = append(record, opr.WinningPreviousOPR[:]...)
	record = append(record, opr.CoinbasePNTAddress[:]...)
	record = append(record, opr.BlockReward[:]...)
	record = append(record, opr.FactomDigitalID[:]...)
	record = append(record, opr.PNT[:]...)
	record = append(record, opr.USD[:]...)
	record = append(record, opr.EUR[:]...)
	record = append(record, opr.JPY[:]...)
	record = append(record, opr.GBP[:]...)
	record = append(record, opr.CAD[:]...)
	record = append(record, opr.CHF[:]...)
	record = append(record, opr.INR[:]...)
	record = append(record, opr.SGD[:]...)
	record = append(record, opr.CNY[:]...)
	record = append(record, opr.HKD[:]...)
	record = append(record, opr.XAU[:]...)
	record = append(record, opr.XAG[:]...)
	record = append(record, opr.XPD[:]...)
	record = append(record, opr.XPT[:]...)
	record = append(record, opr.XBT[:]...)
	record = append(record, opr.ETH[:]...)
	record = append(record, opr.LTC[:]...)
	record = append(record, opr.XBC[:]...)
	record = append(record, opr.FCT[:]...)
	return record, nil
}

func (opr *OraclePriceRecord) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Data source does not have enough data")
			return
		}
	}()
	copy(opr.ChainID[:], data[:32])
	data = data[32:]
	copy(opr.VersionEntryHash[:], data[:32])
	data = data[32:]
	copy(opr.WinningPreviousOPR[:], data[:32])
	data = data[32:]
	copy(opr.CoinbasePNTAddress[:], data[:32])
	data = data[32:]
	copy(opr.BlockReward[:], data[:8])
	data = data[8:]
	copy(opr.FactomDigitalID[:], data[:32])
	data = data[32:]
	copy(opr.PNT[:], data[:8])
	data = data[8:]
	copy(opr.USD[:], data[:8])
	data = data[8:]
	copy(opr.EUR[:], data[:8])
	data = data[8:]
	copy(opr.JPY[:], data[:8])
	data = data[8:]
	copy(opr.GBP[:], data[:8])
	data = data[8:]
	copy(opr.CAD[:], data[:8])
	data = data[8:]
	copy(opr.CHF[:], data[:8])
	data = data[8:]
	copy(opr.INR[:], data[:8])
	data = data[8:]
	copy(opr.SGD[:], data[:8])
	data = data[8:]
	copy(opr.CNY[:], data[:8])
	data = data[8:]
	copy(opr.HKD[:], data[:8])
	data = data[8:]
	copy(opr.XAU[:], data[:8])
	data = data[8:]
	copy(opr.XAG[:], data[:8])
	data = data[8:]
	copy(opr.XPD[:], data[:8])
	data = data[8:]
	copy(opr.XPT[:], data[:8])
	data = data[8:]
	copy(opr.XBT[:], data[:8])
	data = data[8:]
	copy(opr.ETH[:], data[:8])
	data = data[8:]
	copy(opr.LTC[:], data[:8])
	data = data[8:]
	copy(opr.XBC[:], data[:8])
	data = data[8:]
	copy(opr.FCT[:], data[:8])
	data = data[8:]
	if len(data) > 0 {
		err = errors.New("data source is too long for an OPR")
	}
	return
}

func (opr *OraclePriceRecord) GetOPRecord() {

	//get asset values
	var Peg PegAssets
	Peg = PullPEGAssets()
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

func (opr *OraclePriceRecord) SetBlockReward(blockreward uint64) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, blockreward)
	copy(opr.BlockReward[0:], b[0:8])
}

func (opr *OraclePriceRecord) SetFactomDigitalID(factomDigitalID []byte) {
	copy(opr.FactomDigitalID[0:], factomDigitalID)

}

func (opr *OraclePriceRecord) SetPegValues(assets PegAssets) {
	b := make([]byte, 8)

	binary.BigEndian.PutUint64(b, uint64(assets.PNT.Value))
	copy(opr.PNT[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.USD.Value))
	copy(opr.USD[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.EUR.Value))
	copy(opr.EUR[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.JPY.Value))
	copy(opr.JPY[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.GBP.Value))
	copy(opr.GBP[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.CAD.Value))
	copy(opr.CAD[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.CHF.Value))
	copy(opr.CHF[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.INR.Value))
	copy(opr.INR[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.SGD.Value))
	copy(opr.SGD[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.CNY.Value))
	copy(opr.CNY[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.HKD.Value))
	copy(opr.HKD[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.XAU.Value))
	copy(opr.XAU[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.XAG.Value))
	copy(opr.XAG[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.XPD.Value))
	copy(opr.XPD[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.XPT.Value))
	copy(opr.XPT[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.XBT.Value))
	copy(opr.XBT[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.ETH.Value))
	copy(opr.ETH[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.LTC.Value))
	copy(opr.LTC[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.XBC.Value))
	copy(opr.XBC[0:], b[:])
	binary.BigEndian.PutUint64(b, uint64(assets.FCT.Value))
	copy(opr.FCT[0:], b[:])

}
