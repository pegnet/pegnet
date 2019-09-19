package chainsync

import (
	"crypto/sha256"
	"fmt"
	"math/rand"

	"github.com/pegnet/pegnet/modules/lxr30"

	"github.com/pegnet/pegnet/common"
	. "github.com/pegnet/pegnet/modules/opr"
)

// --------------------- OPR Stuff --------------------
// TODO: Move this

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
		for i := range o.WinPreviousOPR {
			o.WinPreviousOPR[i] = ""
		}
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
		for i := range o.Winners {
			o.Winners[i] = []byte{}
		}
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

	oprhash := sha256.Sum256(content)
	lxr := lxr30.Init() // TODO: Fix what lxr to use
	h := lxr.Hash(append(oprhash[:], extids[0]...))
	extids[1] = h[:8]

	return entryhash, extids, content
}
