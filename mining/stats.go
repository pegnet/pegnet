package mining

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/dustin/go-humanize"

	log "github.com/sirupsen/logrus"
)

// GlobalStatTracker is the global tracker for the api's and whatnot
type GlobalStatTracker struct {
	// The sorted listed for block heights
	stats            sync.Mutex
	miningStatistics []*GroupMinerStats

	MiningStatsChannel chan *GroupMinerStats
}

func NewGlobalStatTracker() *GlobalStatTracker {
	g := new(GlobalStatTracker)
	g.MiningStatsChannel = make(chan *GroupMinerStats, 10)

	return g
}

func (t *GlobalStatTracker) Collect(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
		case g := <-t.MiningStatsChannel:
			t.InsertStats(g) // Does the locking
			// Log print the statistics
			log.WithFields(g.LogFields()).WithField("id", "statcollecter").Info("Update")
		}
	}
}

func (t *GlobalStatTracker) FetchAllStats() []*GroupMinerStats {
	t.stats.Lock()
	defer t.stats.Unlock()

	newStats := make([]*GroupMinerStats, len(t.miningStatistics))
	copy(newStats[:], t.miningStatistics[:])
	return newStats
}

func (t *GlobalStatTracker) FetchStats(height int) *GroupMinerStats {
	t.stats.Lock()
	defer t.stats.Unlock()
	return t.fetch(height)
}

func (t *GlobalStatTracker) InsertStats(g *GroupMinerStats) {
	t.stats.Lock()
	defer t.stats.Unlock()
	t.insert(g)
}

func (t *GlobalStatTracker) insert(g *GroupMinerStats) {
	t.miningStatistics = append(t.miningStatistics, g)
	sort.SliceStable(t.miningStatistics,
		func(i, j int) bool { return t.miningStatistics[i].BlockHeight > t.miningStatistics[j].BlockHeight })
}

func (t *GlobalStatTracker) fetch(height int) *GroupMinerStats {
	i := sort.Search(len(t.miningStatistics), func(i int) bool { return t.miningStatistics[i].BlockHeight <= height })
	if i < len(t.miningStatistics) && t.miningStatistics[i].BlockHeight == height {
		// height is present at data[i]
		return t.miningStatistics[i]
	}
	// height is not present in data,
	return nil
}

type GroupMinerStats struct {
	Miners      map[int]*SingleMinerStats
	BlockHeight int
}

func NewGroupMinerStats() *GroupMinerStats {
	g := new(GroupMinerStats)
	g.Miners = make(map[int]*SingleMinerStats)

	return g
}

func (g *GroupMinerStats) TotalHashPower() float64 {
	var totalDur time.Duration
	var acc float64
	// Weight by duration
	for _, m := range g.Miners {
		elapsed := m.Stop.Sub(m.Start)
		totalDur += elapsed
		acc += float64(m.TotalHashes) / elapsed.Seconds()
	}

	return acc
}

func (g *GroupMinerStats) AvgHashRatePerMiner() float64 {
	var totalDur time.Duration
	var acc float64
	// Weight by duration
	for _, m := range g.Miners {
		elapsed := m.Stop.Sub(m.Start)
		totalDur += elapsed
		acc += elapsed.Seconds() * (float64(m.TotalHashes) / elapsed.Seconds())
	}

	return acc / totalDur.Seconds()
}

func (g *GroupMinerStats) LogFields() log.Fields {
	return log.Fields{
		"dbht":           g.BlockHeight,
		"miners":         len(g.Miners),
		"miner_hashrate": fmt.Sprintf("%s/s", humanize.FormatFloat("", g.AvgHashRatePerMiner())),
		"total_hashrate": fmt.Sprintf("%s/s", humanize.FormatFloat("", g.TotalHashPower())),
	}
}

type SingleMinerStats struct {
	ID             int
	TotalHashes    int64
	BestDifficulty uint64
	Start          time.Time
	Stop           time.Time
}

func NewSingleMinerStats() *SingleMinerStats {
	s := new(SingleMinerStats)
	return s
}

func (s *SingleMinerStats) NewDifficulty(diff uint64) {
	if diff > s.BestDifficulty {
		s.BestDifficulty = diff
	}
}
