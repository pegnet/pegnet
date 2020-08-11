// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package mining

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	lxr "github.com/pegnet/LXRHash"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

// LX holds an instance of lxrhash
var LX lxr.LXRHash
var lxInitializer sync.Once

// The init function for LX is expensive. So we should explicitly call the init if we intend
// to use it. Make the init call idempotent
func InitLX() {
	lxInitializer.Do(func() {
		// This code will only be executed ONCE, no matter how often you call it
		LX.Verbose(true)
		if size, err := strconv.Atoi(os.Getenv("LXRBITSIZE")); err == nil && size >= 8 && size <= 30 {
			LX.Init(0xfafaececfafaecec, uint64(size), 256, 5)
		} else {
			LX.Init(lxr.Seed, lxr.MapSizeBits, lxr.HashSize, lxr.Passes)
		}
	})
}

const (
	_ = iota
	BatchCommand
	NewOPRHash
	ResetRecords
	MinimumAccept
	RecordsToKeep
	RecordAggregator
	StatsAggregator
	SubmitNonces

	PauseMining
	ResumeMining
)

type MinerCommand struct {
	Command int
	Data    interface{}
}

type Winner struct {
	OPRHash string
	Nonce   string
	Target  string
}

// PegnetMiner mines an OPRhash
type PegnetMiner struct {
	// ID is the miner number, starting with "1". Every miner launched gets the next
	// sequential number.
	ID         int            `json:"id"`
	Config     *config.Config `json:"-"` //  The config of the miner using the record
	PersonalID uint32         // The miner thread id

	successes chan *Winner

	// Miner commands
	commands <-chan *MinerCommand

	// All the state variables PER oprhash.
	//	Typically want to update these all in parallel
	MiningState oprMiningState

	// Tells us we are paused
	paused bool
}

type oprMiningState struct {
	// Used to compute new hashes
	oprhash []byte
	static  []byte

	// Used to track noncing
	*NonceIncrementer
	start uint32 // For batch mining

	// Used to return hashes
	minimumDifficulty uint64

	// Where we keep the top X nonces to be written
	rankings *opr.NonceRanking
	stats    *SingleMinerStats // Miner stats are tied to the rankings

	// Where we will write our rankings too
	writeChannel chan<- *opr.NonceRanking
	statsChannel chan<- *SingleMinerStats

	keep int
}

// NonceIncrementer is just simple to increment nonces
type NonceIncrementer struct {
	Nonce         []byte
	lastNonceByte int
}

func NewNonceIncrementer(id int) *NonceIncrementer {
	n := new(NonceIncrementer)
	n.Nonce = []byte{byte(id), 0}
	n.lastNonceByte = 1
	return n
}

// NextNonce is just counting to get the next nonce. We preserve
// the first byte, as that is our ID and give us our nonce space
//	So []byte(ID, 255) -> []byte(ID, 1, 0) -> []byte(ID, 1, 1)
func (i *NonceIncrementer) NextNonce() {
	idx := len(i.Nonce) - 1
	for {
		i.Nonce[idx]++
		if i.Nonce[idx] == 0 {
			idx--
			if idx == 0 { // This is my prefix, don't touch it!
				rest := append([]byte{1}, i.Nonce[1:]...)
				i.Nonce = append([]byte{i.Nonce[0]}, rest...)
				break
			}
		} else {
			break
		}
	}

}

func (p *PegnetMiner) ResetNonce() {
	p.MiningState.NonceIncrementer = NewNonceIncrementer(p.ID)
}

func NewPegnetMinerFromConfig(c *config.Config, id int, commands <-chan *MinerCommand) *PegnetMiner {
	p := new(PegnetMiner)
	InitLX()
	p.Config = c
	p.ID = id
	p.commands = commands

	p.MiningState.keep, _ = p.Config.Int("Miner.RecordsPerBlock")

	// Some inits
	p.MiningState.NonceIncrementer = NewNonceIncrementer(p.ID)
	p.ResetNonce()
	p.MiningState.rankings = opr.NewNonceRanking(p.MiningState.keep)

	return p
}

func (p *PegnetMiner) Mine(ctx context.Context) {
	mineLog := log.WithFields(log.Fields{"miner": p.ID})
	var _ = mineLog
	select {
	// Wait for the first command to start
	// We start 'paused'. Any command will knock us out of this init phase
	case c := <-p.commands:
		p.HandleCommand(c)
	case <-ctx.Done():
		return // Cancelled
	}

	for {
		select {
		case <-ctx.Done():
			return // Mining cancelled
		case c := <-p.commands:
			p.HandleCommand(c)
		default:
		}

		if p.paused {
			// Waiting on a resume command
			p.waitForResume(ctx)
			continue
		}

		p.MiningState.NextNonce()

		p.MiningState.stats.TotalHashes++
		diff := opr.ComputeDifficulty(p.MiningState.oprhash, p.MiningState.Nonce)
		if diff > p.MiningState.minimumDifficulty && p.MiningState.rankings.AddNonce(p.MiningState.Nonce, diff) {
			p.MiningState.stats.NewDifficulty(diff)
			//mineLog.WithFields(log.Fields{
			//	"oprhash": fmt.Sprintf("%x", p.MiningState.oprhash),
			//	"Nonce":   fmt.Sprintf("%x", p.MiningState.Nonce),
			//	"diff":    diff,
			//}).Debugf("new Nonce")
		}
	}

}

func (p *PegnetMiner) MineBatch(ctx context.Context, batchsize int) {
	limit := uint32(math.MaxUint32) - uint32(batchsize)
	mineLog := log.WithFields(log.Fields{"miner": p.ID,
		"pid": p.PersonalID})

	select {
	// Wait for the first command to start
	// We start 'paused'. Any command will knock us out of this init phase
	case c := <-p.commands:
		p.HandleCommand(c)
	case <-ctx.Done():
		mineLog.Debugf("Mining init cancelled for miner %d\n", p.ID)
		return // Cancelled
	}

	for {
		select {
		case <-ctx.Done():
			mineLog.Debugf("Mining cancelled for miner: %d\n", p.ID)
			return // Mining cancelled
		case c := <-p.commands:
			p.HandleCommand(c)
		default:
		}

		if len(p.MiningState.oprhash) == 0 {
			p.paused = true
		}

		if p.paused {
			// Waiting on a resume command
			p.waitForResume(ctx)
			continue
		}

		batch := make([][]byte, batchsize)

		for i := range batch {
			batch[i] = make([]byte, 4)
			binary.BigEndian.PutUint32(batch[i], p.MiningState.start+uint32(i))
		}
		p.MiningState.start += uint32(batchsize)
		if p.MiningState.start > limit {
			mineLog.Warnf("repeating nonces, hit the cycle's limit")
		}

		var results [][]byte
		results = LX.HashParallel(p.MiningState.static, batch)
		for i := range results {
			// do something with the result here
			// nonce = batch[i]
			// input = append(base, batch[i]...)
			// hash = results[i]
			h := results[i]

			diff := ComputeHashDifficulty(h)
			p.MiningState.stats.NewDifficulty(diff)
			p.MiningState.stats.TotalHashes++

			if diff > p.MiningState.minimumDifficulty {
				success := &Winner{
					OPRHash: hex.EncodeToString(p.MiningState.oprhash),
					Nonce:   hex.EncodeToString(append(p.MiningState.static[32:], batch[i]...)),
					Target:  fmt.Sprintf("%x", diff),
				}
				p.MiningState.stats.TotalSubmissions++
				select {
				case p.successes <- success:
					mineLog.WithFields(log.Fields{
						"nonce":        batch[i],
						"id":           p.ID,
						"staticPrefix": p.MiningState.static[32:],
						"target":       success.Target,
					}).Trace("Submitted share")
				default:
					mineLog.WithField("channel", fmt.Sprintf("%p", p.successes)).Errorf("failed to submit, %d/%d", len(p.successes), cap(p.successes))
				}
			}
		}
	}
}

func (p *PegnetMiner) HandleCommand(c *MinerCommand) {
	switch c.Command {
	case BatchCommand:
		commands := c.Data.([]*MinerCommand)
		for _, c := range commands {
			p.HandleCommand(c)
		}
	case NewOPRHash:
		p.MiningState.oprhash = c.Data.([]byte)
	case ResetRecords:
		p.ResetNonce()
		p.MiningState.rankings = opr.NewNonceRanking(p.MiningState.keep)
		p.MiningState.stats = NewSingleMinerStats()
		p.MiningState.stats.ID = p.ID
		p.MiningState.stats.Start = time.Now()
	case MinimumAccept:
		p.MiningState.minimumDifficulty = c.Data.(uint64)
	case RecordsToKeep:
		p.MiningState.keep = c.Data.(int)
	case RecordAggregator:
		w := c.Data.(IEntryWriter)
		p.MiningState.writeChannel = w.AddMiner()
	case StatsAggregator:
		w := c.Data.(chan *SingleMinerStats)
		p.MiningState.statsChannel = w
	case SubmitNonces:
		p.MiningState.stats.Stop = time.Now()
		p.MiningState.writeChannel <- p.MiningState.rankings
		if p.MiningState.statsChannel != nil {
			p.MiningState.statsChannel <- p.MiningState.stats
		}
	case PauseMining:
		// Pause until we get a new start
		p.paused = true
	case ResumeMining:
		p.paused = false
	}
}

func (p *PegnetMiner) waitForResume(ctx context.Context) {
	// Pause until we get a new start or are cancelled
	for {
		select {
		case <-ctx.Done(): // Mining cancelled
			return
		case c := <-p.commands:
			p.HandleCommand(c)
			if !p.paused {
				return
			}
		}
	}
}

func ComputeHashDifficulty(b []byte) (difficulty uint64) {
	// The high eight bytes of the hash(hash(entry.Content) + nonce) is the difficulty.
	// Because we don't have a difficulty bar, we can define difficulty as the greatest
	// value, rather than the minimum value.  Our bar is the greatest difficulty found
	// within a 10 minute period.  We compute difficulty as Big Endian.
	return uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
		uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
}
