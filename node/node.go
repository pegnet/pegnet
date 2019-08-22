package node

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/node/database"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

// PegnetNode is the pegnet full node implementation. This will support the expanded apis
// and support better databasing and indexing.
type PegnetNode struct {
	FactomMonitor common.IMonitor
	PegnetGrader  *opr.QuickGrader

	config       *config.Config
	NodeDatabase *database.PegnetNodeDatabase

	// We need to embed this so we have the same functionality as the existing one
	opr.IOPRBlockStore
}

func NewPegnetNode(config *config.Config, monitor common.IMonitor, grader *opr.QuickGrader) (*PegnetNode, error) {
	n := new(PegnetNode)
	n.config = config
	n.FactomMonitor = monitor
	n.PegnetGrader = grader

	var err error
	n.NodeDatabase, err = database.NewPegnetNodeDatabase(config)
	if err != nil {
		return nil, err
	}

	// Migrating updates the tables with anything new
	n.NodeDatabase.Migrate()

	// Overwrite the blockstore with our node, so we capture every new opr synced.
	n.IOPRBlockStore = grader.BlockStore
	grader.BlockStore = n

	return n, nil
}

//  Run will run the sync everytime we hit a new block
func (n *PegnetNode) Run(ctx context.Context) {
	fLog := n.logger()
	fLog.Info("Running initial sync")
	opr.InitLX() // We intend to use the LX hash
	g := n.PegnetGrader
	for { // We need to first sync our grader before we start syncing new blocks
		select { // If we get stuck in the sync loop, this is how can cancel it
		case <-ctx.Done():
			return // Grader stopped
		default:
		}
		err := g.Sync() // Might want to pass a context down?
		if err != nil { // We will try again in a little bit
			fLog.WithError(err).Errorf("failed to sync")
			time.Sleep(2 * time.Second)
			continue
		}
		break // Initial sync done
	}

	fdAlert := n.FactomMonitor.NewListener()
	for {
		var fds common.MonitorEvent
		select {
		case fds = <-fdAlert:
		case <-ctx.Done():
			return // Grader stopped
		}
		fLog := fLog.WithFields(log.Fields{"minute": fds.Minute, "dbht": fds.Dbht})
		if fds.Minute == 1 {
			var err error
			tries := 0
			// Try 3 times to sync the grader
			for tries = 0; tries < 3; tries++ {
				err = nil
				err = g.Sync()
				if err == nil {
					break
				}
				if err != nil {
					// If this fails, we probably can't recover this block.
					// Can't hurt to try though
					time.Sleep(200 * time.Millisecond)
				}
			}

			if err != nil {
				fLog.WithError(err).WithField("tries", tries).Errorf("Grader failed to grade blocks. Sitting out this block")
				continue
			}
		}

	}
}

// WriteOPRBlock will hijack the opr write to also write to our sqldb
func (n *PegnetNode) WriteOPRBlock(block *opr.OprBlock) error {
	// This will write 1 entry for each time series if the block is valid (meaning has > 10 oprs)
	if !block.EmptyOPRBlock {
		// Only keep the graded ones
		sorted := make([]*opr.OraclePriceRecord, len(block.GradedOPRs))
		copy(sorted, block.GradedOPRs)
		sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Difficulty > sorted[j].Difficulty })

		tx := n.NodeDatabase.DB.Begin()
		t, err := database.TimeSeriesFromOPRBlock(block)
		if err != nil {
			return common.DetailError(err)
		}

		hr := database.NetworkHashrateTimeSeriesFromOPRBlock(sorted, t)
		err = database.InsertTimeSeries(tx, &hr)
		if err != nil {
			tx.Rollback()
			return common.DetailError(err)
		}

		df := database.DifficultyTimeSeriesTimeSeriesFromOPRBlock(sorted, t)
		err = database.InsertTimeSeries(tx, &df)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("%d %s", block.Dbht, common.DetailError(err))
		}

		recs := database.NumberOPRRecordsTimeSeries{
			TimeSeries:       t,
			NumberOfOPRs:     len(block.OPRs),
			NumberGradedOPRs: len(block.GradedOPRs)}
		err = database.InsertTimeSeries(tx, &recs)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("%d %s", block.Dbht, common.DetailError(err))
		}

		// Asset pricing
		winner := block.GradedOPRs[0]
		for asset, price := range winner.Assets {
			at := database.AssetPricingTimeSeries{
				TimeSeries: t,
				Asset:      asset,
				Price:      price,
			}
			err = database.InsertAssetTimeSeries(tx, &at)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("%d %s", block.Dbht, common.DetailError(err))
			}
		}

		// Unique coinbases
		uniqueGraded := make(map[string]int)
		uniqueWining := make(map[string]int)
		for i, r := range block.GradedOPRs {
			if opr.GetRewardFromPlace(i) > 0 {
				uniqueWining[r.CoinbaseAddress] += 1
			}
			uniqueGraded[r.CoinbaseAddress] += 1
		}

		largestMiner := 0
		for _, v := range uniqueGraded {
			if v > largestMiner {
				largestMiner = v
			}
		}

		uc := database.UniqueGradedCoinbasesTimeSeries{
			TimeSeries:             t,
			BiggestMiner:           largestMiner,
			UniqueGradedCoinbases:  len(uniqueGraded),
			UniqueWinningCoinbases: len(uniqueWining),
		}

		err = database.InsertTimeSeries(tx, &uc)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("%d %s", block.Dbht, common.DetailError(err))
		}

		dberr := tx.Commit()
		if dberr.Error != nil {
			return dberr.Error
		}
	}

	// Write to the regular opr store, we will also write the data to our sqldb
	err := n.IOPRBlockStore.WriteOPRBlock(block)
	if err != nil {
		return common.DetailError(err)
	}

	return nil
}

func (n *PegnetNode) APIMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/node/v1", n.NodeAPI)

	return mux
}

func (n *PegnetNode) logger() *log.Entry {
	return log.WithField("id", "pegnetnode")
}
