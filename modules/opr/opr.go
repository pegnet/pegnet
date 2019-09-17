package opr

import (
	"fmt"
	"math/rand"

	"github.com/pegnet/pegnet/common"
)

// Type is the format of the underlying data
type Type int

const (
	_ Type = iota
	// V1 is JSON
	V1
	// V2 is Protobuf
	V2
)

// OPR is a common interface for Oracle Price Records of various underlying types.
// The interface has getters for all data inside content
type OPR interface {
	GetHeight() int32
	GetAddress() string
	GetPreviousWinners() []string
	GetID() string
	GetOrderedAssetsFloat() []AssetFloat
	GetOrderedAssetsUint() []AssetUint
	Marshal() ([]byte, error)
	GetType() Type
	Clone() OPR
}

// RandomOPR is useful for unit testing
func RandomOPR(version uint8) (entryhash []byte, extids [][]byte, content []byte) {
	return RandomOPRWithFields(version, rand.Int31())
}

func RandomOPRWithFields(version uint8, dbht int32) (entryhash []byte, extids [][]byte, content []byte) {
	coinbase := common.ConvertRawToFCT(common.RandomByteSliceOfLen(32))
	id := make([]byte, 8)
	rand.Read(id)

	// Random Ehash
	entryhash = make([]byte, 32)
	rand.Read(entryhash)

	extids = make([][]byte, 3)

	// Nonce
	extids[0] = make([]byte, 8)
	rand.Read(extids[0])

	var io OPR

	// First random content
	switch version {
	case 1:
		o := new(V1Content)
		o.WinPreviousOPR = make([]string, 10)
		o.Dbht = dbht
		o.CoinbaseAddress = coinbase
		o.FactomDigitalID = fmt.Sprintf("%x", id)

		o.Assets = make(V1AssetList)
		for _, asset := range V1Assets {
			// Truncate to 4
			o.Assets[asset] = float64(int64(rand.Float64()*1e4)) / 1e4
			if o.Assets[asset] == 0 {
				o.Assets[asset] = 1
			}
		}
		extids[2] = []byte{1}

		io = o
	case 2:
		o := new(V2Content)
		o.Winners = make([][]byte, 25)
		o.Height = dbht
		o.Address = coinbase
		o.ID = fmt.Sprintf("%x", id)
		o.Assets = make([]uint64, len(V2Assets))
		for i := range V2Assets {
			o.Assets[i] = rand.Uint64() % 100000 * 1e8 // 100K max
			if o.Assets[i] == 0 {
				o.Assets[i] = 1e8
			}
		}
		extids[2] = []byte{2}

		io = o
	default:
		return nil, nil, nil
	}

	content, err := io.Marshal()
	if err != nil {
		return nil, nil, nil
	}

	// TODO: Set self report diffculty?
	extids[1] = make([]byte, 8)

	return entryhash, extids, content
}
