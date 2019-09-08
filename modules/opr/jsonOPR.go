package opr

type JSONOPR struct {
	CoinbaseAddress string   `json:"coinbase"`
	Dbht            int32    `json:"dbht"`
	WinPreviousOPR  []string `json:"winners"`
	FactomDigitalID string   `json:"minerid"`

	Assets AssetList `json:"assets"`
}

var _ OPR = (*JSONOPR)(nil)

func (j *JSONOPR) GetOrderedAssets() []Asset {
	list := make([]Asset, len(V1Assets))
	for i, name := range V1Assets {
		list[i] = Asset{Name: name, Value: j.Assets[name]}
	}
	return list
}

func (j *JSONOPR) GetHeight() int32     { return j.Dbht }
func (j *JSONOPR) GetAddress() string   { return j.CoinbaseAddress }
func (j *JSONOPR) GetWinners() []string { return j.WinPreviousOPR }
func (j *JSONOPR) GetID() string        { return j.FactomDigitalID }
