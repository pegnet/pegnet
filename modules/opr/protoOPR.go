package opr

var _ OPR = (*ProtoOPR)(nil)

func (m *ProtoOPR) GetOrderedAssets() []Asset {
	list := make([]Asset, len(m.Assets))
	for i, name := range V2Assets {
		list[i] = Asset{Name: name, Value: m.Assets[i]}
	}
	return list
}

func (m *ProtoOPR) GetType() OPRType {
	return Protobuf
}
