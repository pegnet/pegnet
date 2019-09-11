package opr

import "encoding/json"

type V1Content struct {
	CoinbaseAddress string   `json:"coinbase"`
	Dbht            int32    `json:"dbht"`
	WinPreviousOPR  []string `json:"winners"`
	FactomDigitalID string   `json:"minerid"`

	Assets V1AssetList `json:"assets"`
}

var _ OPR = (*V1Content)(nil)

func (c *V1Content) GetOrderedAssetsFloat() []AssetFloat {
	list := make([]AssetFloat, len(V1Assets))
	for i, name := range V1Assets {
		list[i] = AssetFloat{Name: name, Value: c.Assets[name]}
	}
	return list
}
func (c *V1Content) GetOrderedAssetsUint() []AssetUint {
	list := make([]AssetUint, len(V1Assets))
	for i, name := range V1Assets {
		list[i] = AssetUint{Name: name, Value: FloatToUint64(c.Assets[name])}
	}
	return list
}

func (c *V1Content) GetHeight() int32             { return c.Dbht }
func (c *V1Content) GetAddress() string           { return c.CoinbaseAddress }
func (c *V1Content) GetPreviousWinners() []string { return c.WinPreviousOPR }
func (c *V1Content) GetID() string                { return c.FactomDigitalID }

func (c *V1Content) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

func (c *V1Content) GetType() Type {
	return V1
}

func (c *V1Content) Clone() OPR {
	return &V1Content{
		CoinbaseAddress: c.CoinbaseAddress,
		Dbht:            c.Dbht,
		WinPreviousOPR:  append(c.WinPreviousOPR[:0:0], c.WinPreviousOPR...),
		FactomDigitalID: c.FactomDigitalID,
		Assets:          c.Assets.Clone(),
	}
}
