package oprecord

// These

import (
	"encoding/binary"

	"github.com/FactomProject/factomd/common/primitives/random"
)

type OraclePriceRecord struct {
	ChainID            [32]byte
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

func (opr *OraclePriceRecord) GetOPRecord() {

	//get asset values
	var Peg PegAssets
	Peg = PullPEGAssets()
	Peg.FillPriceBytes()

	opr.SetPegValues(Peg)
	opr.SetBlockReward()
	opr.SetCoinbasePNTAddress()
	opr.SetChainID()
	opr.SetFactomDigitalID()
	opr.SetWinningPreviousOPR()

}

func (opr *OraclePriceRecord) SetChainID() {
	// STUBBED !!!

	copy(opr.ChainID[0:], random.RandByteSliceOfLen(32))
}

func (opr *OraclePriceRecord) SetWinningPreviousOPR() {
	// STUBBED !!!
	copy(opr.WinningPreviousOPR[0:], random.RandByteSliceOfLen(32))
}

func (opr *OraclePriceRecord) SetCoinbasePNTAddress() {
	// STUBBED !!!

	copy(opr.CoinbasePNTAddress[0:], random.RandByteSliceOfLen(32))
}

func (opr *OraclePriceRecord) SetBlockReward() {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(20000))
	copy(opr.BlockReward[0:], b[0:8])
}

func (opr *OraclePriceRecord) SetFactomDigitalID() {
	// STUBBED !!!
	copy(opr.FactomDigitalID[0:], random.RandByteSliceOfLen(32))

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
