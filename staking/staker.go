// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package staking

import (
	"github.com/zpatrick/go-config"
)

const (
	_ = iota
)

type StakerCommand struct {
	Command int
	Data    interface{}
}

// PegnetStaker mines an SPRhash
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

	keep int
}

func NewPegnetStakerFromConfig(c *config.Config, id int, commands <-chan *StakerCommand) *PegnetStaker {
	p := new(PegnetStaker)
	p.Config = c
	p.ID = id
	p.commands = commands

	p.StakingState.keep, _ = p.Config.Int("Staker.RecordsPerBlock")

	return p
}

