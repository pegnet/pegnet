package opr

import "fmt"

var _ OPR = (*ProtoOPR)(nil)

func (m *ProtoOPR) GetOrderedAssets() []Asset {
	list := make([]Asset, len(m.Assets))
	for i, name := range V2Assets {
		list[i] = Asset{Name: name, Value: m.Assets[i]}
	}
	return list
}

func (m *ProtoOPR) GetType() Type {
	return Protobuf
}

func (m *ProtoOPR) GetPreviousWinners() []string {
	winners := make([]string, 0)
	for _, s := range m.Winners {
		winners = append(winners, fmt.Sprintf("%x", s))
	}
	return winners
}
