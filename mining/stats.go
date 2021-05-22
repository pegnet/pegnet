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

const (
	// MaxGlobalStatsBuckets tells us when to garbage collect
	MaxGlobalStatsBuckets = 250
)

// GlobalStatTracker is the global tracker for the api's and whatnot
//	It has threadsafe queryable stats for the miners and their blockheights.
type GlobalStatTracker struct {
	// The sorted listed for block heights
	stats            sync.Mutex
	miningStatistics []*StatisticBucket

	MiningStatsChannel chan *GroupMinerStats

	// People who also want the stats
	upstreams     map[string]chan *GroupMinerStats
	upstreamMutex sync.Mutex // Maps are not thread safe

}

type StatisticBucket struct {
	// A statistic collection of each group
	GroupStats  map[string]*GroupMinerStats `json:"allgroupstats"`
	BlockHeight int                         `json:"blockheight"`
}

func NewGlobalStatTracker() *GlobalStatTracker {
	g := new(GlobalStatTracker)
	g.MiningStatsChannel = make(chan *GroupMinerStats, 10)
	g.upstreams = make(map[string]chan *GroupMinerStats)

	return g
}

func (g *GlobalStatTracker) GetUpstream(id string) (upstream chan *GroupMinerStats) {
	g.upstreamMutex.Lock()
	defer g.upstreamMutex.Unlock()

	// If the upstream already exists for the id, close it.
	// We only want 1 upstream per id
	upstream, ok := g.upstreams[id]
	if ok {
		close(upstream)
	}

	upstream = make(chan *GroupMinerStats, 10)
	g.upstreams[id] = upstream
	return g.upstreams[id]
}

func (g *GlobalStatTracker) StopUpstream(id string) {
	g.upstreamMutex.Lock()
	defer g.upstreamMutex.Unlock()

	alert, ok := g.upstreams[id]
	if ok {
		close(alert)
	}
	delete(g.upstreams, id)
}

// Collect listens for new stats, and manages them
//	ctx can be cancelled
func (t *GlobalStatTracker) Collect(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
		case g := <-t.MiningStatsChannel:
			t.InsertStats(g) // Does the locking
			// Log print the statistics
			log.WithFields(g.LogFields()).WithField("id", g.ID).WithField("height", g.BlockHeight).Info("mining statistics")
			for _, up := range t.upstreams {
				select {
				case up <- g:
				default:
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}
}

// FetchAllStats is really for unit tests
func (t *GlobalStatTracker) FetchAllStats() []*StatisticBucket {
	t.stats.Lock()
	defer t.stats.Unlock()

	newStats := make([]*StatisticBucket, len(t.miningStatistics))
	copy(newStats[:], t.miningStatistics[:])
	return newStats
}

func (t *GlobalStatTracker) FetchStats(height int) *StatisticBucket {
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
	bucket := t.fetch(g.BlockHeight)
	if bucket != nil {
		bucket.GroupStats[g.ID] = g
	} else {
		bucket = new(StatisticBucket)
		bucket.BlockHeight = g.BlockHeight
		bucket.GroupStats = make(map[string]*GroupMinerStats)
		bucket.GroupStats[g.ID] = g

		t.miningStatistics = append(t.miningStatistics, bucket)
		sort.SliceStable(t.miningStatistics,
			func(i, j int) bool { return t.miningStatistics[i].BlockHeight > t.miningStatistics[j].BlockHeight })

		// TODO: Optimize this a bit better. Maybe used a fixed slice?
		//		Currently it is not that huge of an issue to do.
		if len(t.miningStatistics) > MaxGlobalStatsBuckets {
			tmp := make([]*StatisticBucket, MaxGlobalStatsBuckets)
			copy(tmp, t.miningStatistics[:MaxGlobalStatsBuckets])
			t.miningStatistics = tmp
		}
	}

}

func (t *GlobalStatTracker) fetch(height int) *StatisticBucket {
	i := sort.Search(len(t.miningStatistics), func(i int) bool { return t.miningStatistics[i].BlockHeight <= height })
	if i < len(t.miningStatistics) && t.miningStatistics[i].BlockHeight == height {
		// height is present at data[i]
		return t.miningStatistics[i]
	}
	// height is not present in data,
	return nil
}

// GroupMinerStats has the stats for all miners running from a
// coordinator. It will do aggregation for simple global stats
type GroupMinerStats struct {
	Miners      map[int]*SingleMinerStats `json:"miners"`
	BlockHeight int                       `json:"blockheight"`
	ID          string                    `json:"id"`

	Tags map[string]string `json:"tags"`
}

func NewGroupMinerStats(id string, height int) *GroupMinerStats {
	g := new(GroupMinerStats)
	g.Miners = make(map[int]*SingleMinerStats)
	g.ID = id
	g.BlockHeight = height
	g.Tags = make(map[string]string)

	return g
}

// TotalHashPower is the sum of all miner's hashpower
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

// AvgDurationPerMiner is the average duration of mining across all miners.
func (g *GroupMinerStats) AvgDurationPerMiner() time.Duration {
	var totalDur time.Duration
	// Weight by duration
	for _, m := range g.Miners {
		elapsed := m.Stop.Sub(m.Start)
		totalDur += elapsed
	}

	return totalDur / time.Duration(len(g.Miners))
}

func (g *GroupMinerStats) LogFields() log.Fields {
	f := log.Fields{
		"dbht":           g.BlockHeight,
		"miners":         len(g.Miners),
		"miner_hashrate": fmt.Sprintf("%s/s", humanize.FormatFloat("", g.AvgHashRatePerMiner())),
		"total_hashrate": fmt.Sprintf("%s/s", humanize.FormatFloat("", g.TotalHashPower())),
		"avg_duration":   fmt.Sprintf("%s", g.AvgDurationPerMiner()),
	}

	for k, v := range g.Tags {
		f[k] = v
	}
	return f
}

// SingleMinerStats is the stats of a single miner
type SingleMinerStats struct {
	ID             int       `json:"id"`
	TotalHashes    int64     `json:"totalhashes"`
	BestDifficulty uint64    `json:"bestdifficulty"`
	Start          time.Time `json:"start"`
	Stop           time.Time `json:"stop"`
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
