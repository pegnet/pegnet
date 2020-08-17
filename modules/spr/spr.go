package spr

import "github.com/pegnet/pegnet/modules/opr"

// Type is the format of the underlying data
type Type int

const (
	_ Type = iota
	// V1 is JSON
	V1
	// V2 is Protobuf encoding
	// V2 is for grading V2, V3, V4 & V5
	V2
)

// SPR is a common interface for Staking Price Records of various underlying types.
// The interface has getters for all data inside content
type SPR interface {
	GetHeight() int32
	GetAddress() string
	GetPreviousWinners() []string
	GetID() string
	GetOrderedAssetsFloat() []opr.AssetFloat
	GetOrderedAssetsUint() []opr.AssetUint
	Marshal() ([]byte, error)
	GetType() Type
	Clone() SPR
}
