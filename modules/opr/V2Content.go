package opr

import "fmt"

var _ OPR = (*V2Content)(nil)

func (m *V2Content) GetOrderedAssetsFloat() []AssetFloat {
	list := make([]AssetFloat, len(m.Assets))
	for i, name := range V2Assets {
		list[i] = AssetFloat{Name: name, Value: Uint64ToFloat(m.Assets[i])}
	}
	return list
}

func (m *V2Content) GetOrderedAssetsUint() []AssetUint {
	list := make([]AssetUint, len(m.Assets))
	for i, name := range V2Assets {
		list[i] = AssetUint{Name: name, Value: m.Assets[i]}
	}
	return list
}

func (m *V2Content) GetType() Type {
	return V2
}

func (m *V2Content) GetPreviousWinners() []string {
	winners := make([]string, 0)
	for _, s := range m.Winners {
		winners = append(winners, fmt.Sprintf("%x", s))
	}
	return winners
}

func (m *V2Content) Clone() OPR {
	clone := new(V2Content)
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
