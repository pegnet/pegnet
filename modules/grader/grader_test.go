package grader_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"sort"
	"testing"

	"github.com/zpatrick/go-config"

	"github.com/pegnet/pegnet/modules/lxr30"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/modules/grader"
	"github.com/pegnet/pegnet/modules/opr"
	opr2 "github.com/pegnet/pegnet/opr"
)

var LXR = lxr30.Init()

func TestGrading(t *testing.T) {
	for i := 0; i < 100; i++ {
		testGradingIsConsistent(t, 1)
	}
	for i := 0; i < 100; i++ {
		testGradingIsConsistent(t, 2)
	}
}

func testGradingIsConsistent(t *testing.T, version uint8) {
	dbht := int32(100)

	d := common.NewUnitTestConfigProvider()
	d.Data = `
[Miner]
  Network=unit-test
`
	con := config.NewConfig([]config.Provider{common.NewDefaultConfigOptionsProvider(), common.NewUnitTestConfigProvider(), d})
	g := opr2.NewQuickGrader(con, nil, nil)
	common.SetTestingVersion(version)
	g.Network = common.UnitTestNetwork

	// Get 50 random oprs, grade them
	mod, err := grader.NewGrader(version, dbht, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	oldOprs := make([]*opr2.OraclePriceRecord, 50)

	for i := range oldOprs {
		ehash, extids, content := opr.RandomOPRWithFields(version, dbht)
		// Self reported difficulty has to be set
		oprHash := sha256.Sum256(content)
		extids[1] = LXR.Hash(append(oprHash[:], extids[0]...))[:8]

		entry := &factom.Entry{ExtIDs: extids, Content: content}
		popr, err := g.ParseOPREntry(entry, int64(dbht))
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		if popr == nil {
			t.Error("nil opr")
			t.FailNow()
		}

		// We have to override
		popr.EntryHash = ehash

		oldOprs[i] = popr
		err = mod.AddOPR(ehash, extids, content)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}

	sort.SliceStable(oldOprs, func(i, j int) bool {
		return binary.BigEndian.Uint64(oldOprs[i].SelfReportedDifficulty) > binary.BigEndian.Uint64(oldOprs[j].SelfReportedDifficulty)
	})

	// Now grade both
	block := mod.Grade()
	newGraded := block.Graded()
	oldGraded := opr2.GradeMinimum(oldOprs, g.Network, int64(dbht))

	if len(oldGraded) != len(newGraded) {
		t.Error("diff length graded")
		t.FailNow()
	}

	for i := range oldGraded {
		if bytes.Compare(oldGraded[i].EntryHash, newGraded[i].EntryHash) != 0 {
			t.Error("Diff graded order")
			t.FailNow()
		}

		// Some grades are a lil different
		if oldGraded[i].Grade != newGraded[i].Grade {
			t.Errorf("Diff grade: %.8f, %.8f", oldGraded[i].Grade, newGraded[i].Grade)
			t.FailNow()
		}
	}

}
