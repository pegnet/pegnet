package opr

type OPRType int

const (
	JSON OPRType = iota
	Protobuf
)

type OPR interface {
	GetHeight() int32
	GetAddress() string
	GetWinners() []string
	GetID() string
	GetOrderedAssets() []Asset
	Marshal() ([]byte, error)
	GetType() OPRType
}
