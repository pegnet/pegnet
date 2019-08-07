// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package mining

import (
	"context"
	"time"

	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

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

// PegnetMiner mines an OPRhash
type PegnetMiner struct {
	// ID is the miner number, starting with "1". Every miner launched gets the next
	// sequential number.
	ID     int            `json:"id"`
	Config *config.Config `json:"-"` //  The config of the miner using the record

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

	// Used to track noncing
	*NonceIncrementer

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
