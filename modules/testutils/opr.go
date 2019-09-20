package testutils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"

	lxr "github.com/pegnet/LXRHash"
	"github.com/pegnet/pegnet/modules/factoidaddress"
	. "github.com/pegnet/pegnet/modules/opr"
)

var LXR *lxr.LXRHash

func SetTestLXR(l *lxr.LXRHash) {
	LXR = l
}

// RandomOPR is useful for unit testing
func RandomOPR(version uint8) (entryhash []byte, extids [][]byte, content []byte) {
	return RandomOPRWithHeight(version, rand.Int31())
}

func RandomOPRWithRandomWinners(version uint8, dbht int32) (entryhash []byte, extids [][]byte, content []byte) {
	return RandomOPRWithFields(version, dbht, RandomWinners(version))
}

func RandomOPRWithHeight(version uint8, dbht int32) (entryhash []byte, extids [][]byte, content []byte) {
	return RandomOPRWithFields(version, dbht, make([]string, amt(version)))
}

func RandomOPRWithFields(version uint8, dbht int32, prevWinners []string) (entryhash []byte, extids [][]byte, content []byte) {
	coinbase := factoidaddress.Random()
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
		o.WinPreviousOPR = prevWinners
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
		o.Winners = make([][]byte, len(prevWinners))
		for i := range o.Winners {
			o.Winners[i], _ = hex.DecodeString(prevWinners[i])
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
	h := LXR.Hash(append(oprhash[:], extids[0]...))
	extids[1] = h[:8]

	return entryhash, extids, content
}

func amt(version uint8) int {
	switch version {
	case 1:
		return 10
	case 2:
		return 25
	}
	return 0
}

func RandomWinners(version uint8) []string {
	winners := make([]string, amt(version))

	for i := range winners {
		b := make([]byte, 8, 8)
		rand.Read(b)
		winners[i] = hex.EncodeToString(b)
	}
	return winners
}

// PopulateRandomWinners adds random winners to the opr content
func PopulateRandomWinners(oI OPR) {
	if oI.GetType() == V1 {
		o := oI.(*V1Content)
		for i := range o.WinPreviousOPR {
			b := make([]byte, 8, 8)
			rand.Read(b)
			o.WinPreviousOPR[i] = hex.EncodeToString(b)
		}
	} else if oI.GetType() == V2 {
		o := oI.(*V2Content)
		for i := range o.Winners {
			b := make([]byte, 8, 8)
			rand.Read(b)
			o.Winners[i] = b
		}
	}
}

// PopulateWithWinners is a testutil call. It does not error check,
// so do not throw non hex strings in here
func PopulateWithWinners(oI OPR, winners []string) {
	if oI.GetType() == V1 {
		o := oI.(*V1Content)
		o.WinPreviousOPR = winners
	} else if oI.GetType() == V2 {
		o := oI.(*V2Content)
		o.Winners = make([][]byte, len(winners))
		for i, winner := range winners {
			o.Winners[i], _ = hex.DecodeString(winner)
		}
	}
}

// FlipVersion is helpful if you want the other version than you are using
func FlipVersion(version uint8) uint8 {
	// Invert and take the bottom 2 bits
	return ^version & 3
}
