package spr

import (
	"fmt"
	"github.com/pegnet/pegnet/modules/opr"
)

var _ SPR = (*S1Content)(nil)

type S1Content struct {
	opr.V2Content
}

func (m *S1Content) GetOrderedAssetsFloat() []opr.AssetFloat {
	list := make([]opr.AssetFloat, len(m.Assets))
	for i, name := range opr.V5Assets {
		list[i] = opr.AssetFloat{Name: name, Value: Uint64ToFloat(m.Assets[i])}
	}
	return list
}

func (m *S1Content) GetOrderedAssetsUint() []opr.AssetUint {
	list := make([]opr.AssetUint, len(m.Assets))
	for i, name := range opr.V5Assets {
		list[i] = opr.AssetUint{Name: name, Value: m.Assets[i]}
	}
	return list
}

func (m *S1Content) GetType() Type {
	return V2
}

func (m *S1Content) GetPreviousWinners() []string {
	winners := make([]string, 0)
	for _, s := range m.Winners {
		winners = append(winners, fmt.Sprintf("%x", s))
	}
	return winners
}

func (m *S1Content) Clone() SPR {
	clone := new(S1Content)
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

// Uint64ToFloat converts a uint to a float and divides it by 1e8
func Uint64ToFloat(u uint64) float64 {
	return float64(u) / 1e8
}
