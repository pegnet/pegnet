package opr

// Type is the format of the underlying data
type Type int

const (
	// JSON is used in v1
	V1 Type = iota
	// Protobuf is used in v2
	V2
)

// OPR is a common interface for Oracle Price Records of various underlying types.
// The interface has getters for all data inside content
type OPR interface {
	GetHeight() int32
	GetAddress() string
	GetPreviousWinners() []string
	GetID() string
	GetOrderedAssets() []Asset
	Marshal() ([]byte, error)
	GetType() Type
	Clone() OPR
}
