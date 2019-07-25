// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package mining

import (
	"context"
	"fmt"

	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

const (
	_ = iota
	BatchCommand
	NewOPRHash
	ResetNonce
	ResetRecords
	MinimumAccept
	RecordsToKeep
	RecordAggregator
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

	// Where we will write our rankings too
	writeChannel chan<- *opr.NonceRanking

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
	select {
	// Wait for the first command to start
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
			p.waitForResume(ctx)
		}

		p.MiningState.NextNonce()

		diff := opr.ComputeDifficulty(p.MiningState.Nonce, p.MiningState.oprhash)
		if p.MiningState.rankings.AddNonce(p.MiningState.Nonce, diff, []string{}) {
			// Log?
			mineLog.WithFields(log.Fields{
				"oprhash": fmt.Sprintf("%x", p.MiningState.oprhash),
				"Nonce":   fmt.Sprintf("%x", p.MiningState.Nonce),
				"diff":    diff,
			}).Debugf("new Nonce")
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
	case ResetNonce:
		p.ResetNonce()
	case ResetRecords:
		p.MiningState.rankings = opr.NewNonceRanking(p.MiningState.keep)
	case MinimumAccept:
		p.MiningState.minimumDifficulty = c.Data.(uint64)
	case RecordsToKeep:
		p.MiningState.keep = c.Data.(int)
	case RecordAggregator:
		w := c.Data.(*EntryWriter)
		p.MiningState.writeChannel = w.AddMiner()
	case SubmitNonces:
		p.MiningState.writeChannel <- p.MiningState.rankings
	case PauseMining:
		// Pause until we get a new start
		p.paused = true
	}
}

func (p *PegnetMiner) waitForResume(ctx context.Context) {
	defer log.Debug("resumed")
	// Pause until we get a new start or are cancelled
	for {
		select {
		case <-ctx.Done(): // Mining cancelled
			return
		case c := <-p.commands:
			if c.Command == ResumeMining {
				p.paused = false
				return
			}
			// If nested in batch
			if c.Command == BatchCommand {
				for _, ci := range c.Data.([]*MinerCommand) {
					if ci.Command == ResumeMining {
						p.paused = false
						p.HandleCommand(c)
						return
					}
				}
			}
			p.HandleCommand(c)
		}
	}
}
