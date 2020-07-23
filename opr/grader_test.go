package opr_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/pegnet/pegnet/balances"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/database"
	. "github.com/pegnet/pegnet/opr"
	"github.com/pegnet/pegnet/polling"
	"github.com/pegnet/pegnet/testutils"
)

func TestOPRParse(t *testing.T) {
	makeList := func(a int) []*OraclePriceRecord {
		return make([]*OraclePriceRecord, a)
	}

	dbht := int32(10)

	test := func(t *testing.T) {
		common.SetTestingVersion(3)
		config := common.NewUnitTestConfig()
		net, _ := common.LoadConfigNetwork(config)

		alerts := make(chan *OPRs, 1)
		// Need to make some random winners
		v := common.OPRVersion(net, 10)
		a := 10
		if v == 2 {
			a = 25
		}

		list := new(OPRs)
		list.ToBePaid = makeList(a)
		for i := range list.ToBePaid {
			opr := RandomOPR()
			opr.Dbht = 1
			list.ToBePaid[i] = opr
		}
		alerts <- list

		PollingDataSource = testutils.AlwaysOnePolling()

		// An OPR made by our codebase should always be valid
		opr, err := NewOpr(context.Background(), 0, dbht, config, alerts)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		if !opr.Validate(config, int64(opr.Dbht)) {
			t.Errorf("Should be valid")
		}

	}

	// At low height
	t.Run("with winner", test)

	makeList = func(a int) []*OraclePriceRecord {
		return nil
	}
	t.Run("with no winners", test)

	// At high heights
	dbht = 999999999
	makeList = func(a int) []*OraclePriceRecord {
		return make([]*OraclePriceRecord, a)
	}
	t.Run("with winner", test)

	makeList = func(a int) []*OraclePriceRecord {
		return nil
	}
	t.Run("with no winners", test)

}

func RandomOPRBlock() *OPRBlockDatabaseObject {
	o := new(OPRBlockDatabaseObject)

	o.DblockHeight = rand.Int63n(1e6)
	o.GradedOprs = make([]*OraclePriceRecord, rand.Intn(100)+1)
	for i := range o.GradedOprs {
		tmp := RandomOPR()
		o.GradedOprs[i] = tmp
	}

	return o
}

func RandomOPR() *OraclePriceRecord {
	return RandomOPROfVersion(1)
}

func RandomOPROfVersion(version uint8) *OraclePriceRecord {
	tmp := new(OraclePriceRecord)
	tmp.OPRHash = make([]byte, 32)
	rand.Read(tmp.OPRHash)
	tmp.EntryHash = make([]byte, 32)
	rand.Read(tmp.EntryHash)
	tmp.Dbht = rand.Int31()
	tmp.FactomDigitalID = "test"
	tmp.Version = version

	min := 10
	if version > 1 {
		min = 25
	}
	tmp.WinPreviousOPR = make([]string, min, min)

	tmp.SelfReportedDifficulty = make([]byte, 8)
	tmp.Nonce = make([]byte, 8)
	rand.Read(tmp.Nonce)
	tmp.Difficulty = tmp.ComputeDifficulty(tmp.Nonce)
	binary.BigEndian.PutUint64(tmp.SelfReportedDifficulty, tmp.Difficulty)

	assets := common.AssetsV1
	if version > 1 {
		assets = common.AssetsV2
	}
	if version == 4 {
		assets = common.AssetsV4
	}
	if version == 5 {
		assets = common.AssetsV5
	}
	tmp.Assets = make(OraclePriceRecordAssetList)
	for _, asset := range assets {
		tmp.Assets.SetValue(asset, rand.Float64()*100)
	}

	tmp.CoinbaseAddress = common.ConvertRawToFCT(common.RandomByteSliceOfLen(32))

	// TODO: Add more fields to this
	return tmp
}

func SetOPRPriceClose(opr *OraclePriceRecord, center float64, std float64) {
	for asset := range opr.Assets {
		d := (rand.Float64() - .5) * 2 * std
		opr.Assets.SetValue(asset, polling.TruncateTo8(center+d))
	}
}

func TestOPRQuery(t *testing.T) {
	// Our fetch options were incorrect, using the OPRHash vs the entry hash. We use
	// the entry hash in the winner list. This will ensure we keep things as entry hashes
	t.Run("test fetch options", func(t *testing.T) {
		g := NewQuickGrader(common.NewUnitTestConfig(), database.NewMapDb(), balances.NewBalanceTracker())
		opr := RandomOPR()
		block := &OprBlock{OPRs: []*OraclePriceRecord{opr}, Dbht: int64(opr.Dbht)}
		for i := 0; i < 10; i++ {
			g.DEBUGAddOPRBlock(&OprBlock{})
		}

		// Make sure our block is hidden here
		g.DEBUGAddOPRBlock(block)

		for i := 0; i < 10; i++ {
			g.DEBUGAddOPRBlock(&OprBlock{})
		}

		// Can we find it?
		f := g.OprByShortHash(hex.EncodeToString(opr.EntryHash[:8]))
		if bytes.Compare(f.EntryHash, opr.EntryHash) != 0 {
			t.Errorf("not found")
		}

		f = g.OprByHash(hex.EncodeToString(opr.EntryHash))
		if bytes.Compare(f.EntryHash, opr.EntryHash) != 0 {
			t.Errorf("not found")
		}

		b := g.OprBlockByHeight(int64(opr.Dbht))
		if b == nil || len(b.OPRs) != 1 || bytes.Compare(b.OPRs[0].EntryHash, opr.EntryHash) != 0 {
			t.Errorf("not found")
		}

		l := g.OprsByDigitalID(opr.FactomDigitalID)
		if len(l) != 1 || bytes.Compare(l[0].EntryHash, opr.EntryHash) != 0 {
			t.Errorf("not found")
		}

		// Test not found (the error case)
		emptyG := NewQuickGrader(common.NewUnitTestConfig(), database.NewMapDb(), balances.NewBalanceTracker())
		f = emptyG.OprByShortHash(hex.EncodeToString(opr.EntryHash[:8]))
		if len(f.EntryHash) != 0 {
			t.Errorf("was found, but it should not be")
		}

		f = emptyG.OprByHash(hex.EncodeToString(opr.EntryHash))
		if len(f.EntryHash) != 0 {
			t.Errorf("was found, but it should not be")
		}

		b = emptyG.OprBlockByHeight(int64(opr.Dbht))
		if b != nil {
			t.Errorf("was found, but it should not be")
		}

		l = emptyG.OprsByDigitalID(opr.FactomDigitalID)
		if len(l) != 0 {
			t.Errorf("was found, but it should not be")
		}
	})
}

// TestGradingOrder for any different ordering and using floats
func TestGradingOrder(t *testing.T) {
	cycles := 25

	// Set the lists back to their original at the end of the unit test
	saveList := func(list *[]string) ([]string, *[]string) {
		orig := make([]string, len(*list))
		copy(orig, *list)
		return orig, list
	}

	returnList := func(orig []string, origP *[]string) {
		*origP = orig
	}

	defer returnList(saveList(&common.AllAssets))
	defer returnList(saveList(&common.AssetsV1))
	defer returnList(saveList(&common.AssetsV2))
	defer returnList(saveList(&common.AssetsV4))
	defer returnList(saveList(&common.AssetsV5))

	// Version 1
	t.Run("version 1", func(t *testing.T) {
		testGradingOrderVersion(t, 1, cycles)
	})
	t.Run("version 2", func(t *testing.T) {
		testGradingOrderVersion(t, 2, cycles)
	})
	t.Run("version 3", func(t *testing.T) {
		testGradingOrderVersion(t, 3, cycles)
	})
	t.Run("version 4", func(t *testing.T) {
		testGradingOrderVersion(t, 4, cycles)
	})
	t.Run("version 5", func(t *testing.T) {
		testGradingOrderVersion(t, 5, cycles)
	})
}

func testGradingOrderVersion(t *testing.T, version uint8, cycles int) {
	common.SetTestingVersion(version)
	for i := 0; i < cycles; i++ {
		set := make([]*OraclePriceRecord, 50)
		for i := range set {
			set[i] = RandomOPROfVersion(version)
			set[i].Network = common.UnitTestNetwork
			// All prices are super close. Within 0.001%
			SetOPRPriceClose(set[i], 1, 0.00001)
		}

		sort.SliceStable(set, func(i, j int) bool { return set[i].Difficulty > set[j].Difficulty })

		// Check graded order remains the same
		// mess with order
		graded := make([][]*OraclePriceRecord, 5)
		for i := 0; i < 5; i++ {
			rand.Shuffle(len(set), func(i, j int) {
				set[i], set[j] = set[j], set[i]
			})
			graded[i] = GradeMinimum(set, common.UnitTestNetwork, 0)
			shuffleList()
		}

		// Check for diffs
		o := order(graded[0])
		for i := 1; i < 5; i++ {
			nextO := order(graded[i])
			if bytes.Compare(o, nextO) != 0 {
				t.Errorf("%d has a different order", i)
			}
		}

	}
}

func sameAvgs(a []float64, b []float64) error {
	if len(a) != len(b) {
		return nil
	}
	for i := range a {
		if a[i] != b[i] {
			return fmt.Errorf("index %d is %.8f and %.8f", i, a[i], b[i])
		}
	}
	return nil
}

func order(list []*OraclePriceRecord) []byte {
	hash := sha256.New()
	for _, o := range list {
		_, err := hash.Write(o.OPRHash)
		if err != nil {
			panic(err)
		}
	}
	return hash.Sum(nil)
}

func shuffleList() {
	rand.Shuffle(len(common.AllAssets), func(i, j int) {
		common.AllAssets[i], common.AllAssets[j] = common.AllAssets[j], common.AllAssets[i]
	})
	rand.Shuffle(len(common.AssetsV1), func(i, j int) {
		common.AssetsV1[i], common.AssetsV1[j] = common.AssetsV1[j], common.AssetsV1[i]
	})
	// Version One, subtract 2 assets
	common.AssetsV2 = common.SubtractFromSet(common.AssetsV1, "XPD", "XPT")

	rand.Shuffle(len(common.AssetsV4), func(i, j int) {
		common.AssetsV4[i], common.AssetsV4[j] = common.AssetsV4[j], common.AssetsV4[i]
	})

	rand.Shuffle(len(common.AssetsV5), func(i, j int) {
		common.AssetsV5[i], common.AssetsV5[j] = common.AssetsV5[j], common.AssetsV5[i]
	})
}
