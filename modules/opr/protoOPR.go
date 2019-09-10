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

func (m *ProtoOPR) Clone() OPR {
	clone := new(ProtoOPR)
	clone.Address = m.Address
	clone.Height = m.Height
	clone.ID = m.ID
	clone.Assets = append(m.Assets[:0:0], m.Assets...)

	cloneWinners := make([][]byte, 0)
	for _, w := range m.Winners {
		cloneWinners = append(cloneWinners, append(w[:0:0], w...))
	}
	clone.Winners = cloneWinners

	return clone
}
