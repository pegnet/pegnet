// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package staking

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

const (
	_ = iota
	BatchCommand
	NewSPRHash
	RecordAggregator

	PauseStaking
	ResumeStaking
)

type StakerCommand struct {
	Command int
	Data    interface{}
}

// PegnetStaker stakes an SPRhash
type PegnetStaker struct {
	// ID is the staker number, starting with "1".
	ID     int            `json:"id"`
	Config *config.Config `json:"-"` //  The config of the staker using the record

	// Staker commands
	commands <-chan *StakerCommand

	// All the state variables PER sprhash.
	//	Typically want to update these all in parallel
	StakingState sprStakingState

	// Tells us we are paused
	paused bool
}

type sprStakingState struct {
	// Used to compute new hashes
	sprhash []byte
}

func NewPegnetStakerFromConfig(c *config.Config, id int, commands <-chan *StakerCommand) *PegnetStaker {
	p := new(PegnetStaker)
	p.Config = c
	p.ID = id
	p.commands = commands
	return p
}

func (p *PegnetStaker) Stake(ctx context.Context) {
	stakeLog := log.WithFields(log.Fields{"staker": p.ID})
	var _ = stakeLog
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
			return // Staking cancelled
		case c := <-p.commands:
			p.HandleCommand(c)
		default:
		}

		if p.paused {
			// Waiting on a resume command
			p.waitForResume(ctx)
			continue
		}
	}
}

func (p *PegnetStaker) HandleCommand(c *StakerCommand) {
	switch c.Command {
	case BatchCommand:
		commands := c.Data.([]*StakerCommand)
		for _, c := range commands {
			p.HandleCommand(c)
		}
	case NewSPRHash:
		p.StakingState.sprhash = c.Data.([]byte)
	case PauseStaking:
		// Pause until we get a new start
		p.paused = true
	case ResumeStaking:
		p.paused = false
	}
}

func (p *PegnetStaker) waitForResume(ctx context.Context) {
	// Pause until we get a new start or are cancelled
	for {
		select {
		case <-ctx.Done(): // Staking cancelled
			return
		case c := <-p.commands:
			p.HandleCommand(c)
			if !p.paused {
				return
			}
		}
	}
}
