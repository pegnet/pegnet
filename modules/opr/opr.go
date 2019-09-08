package opr

type OPR interface {
	GetHeight() int32
	GetAddress() string
	GetWinners() []string
	GetID() string
	GetOrderedAssets() []Asset
}
