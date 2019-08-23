package opr_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/pegnet/pegnet/balances"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/database"
	. "github.com/pegnet/pegnet/opr"
	"github.com/pegnet/pegnet/testutils"
)

func TestOPRParse(t *testing.T) {
	makeList := func(a int) []*OraclePriceRecord {
		return make([]*OraclePriceRecord, a)
	}

	dbht := int32(10)

	test := func(t *testing.T) {
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
	tmp := new(OraclePriceRecord)
	tmp.OPRHash = make([]byte, 32)
	rand.Read(tmp.OPRHash)
	tmp.EntryHash = make([]byte, 32)
	rand.Read(tmp.EntryHash)
	tmp.Dbht = rand.Int31()
	tmp.FactomDigitalID = "test"
	// TODO: Add more fields to this
	return tmp
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
