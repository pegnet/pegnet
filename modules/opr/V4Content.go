package opr

import "fmt"

var _ OPR = (*V4Content)(nil)

type V4Content struct {
	V2Content
}

func (m *V4Content) GetOrderedAssetsFloat() []AssetFloat {
	list := make([]AssetFloat, len(m.Assets))
	for i, name := range V4Assets {
		list[i] = AssetFloat{Name: name, Value: Uint64ToFloat(m.Assets[i])}
	}
	return list
}

func (m *V4Content) GetOrderedAssetsUint() []AssetUint {
	list := make([]AssetUint, len(m.Assets))
	for i, name := range V4Assets {
		list[i] = AssetUint{Name: name, Value: m.Assets[i]}
	}
	return list
}

func (m *V4Content) GetType() Type {
	return V2
}

func (m *V4Content) GetPreviousWinners() []string {
	winners := make([]string, 0)
	for _, s := range m.Winners {
		winners = append(winners, fmt.Sprintf("%x", s))
	}
	return winners
}

func (m *V4Content) Clone() OPR {
	clone := new(V4Content)
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
