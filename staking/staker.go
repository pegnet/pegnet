// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package staking

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/common"
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
	CheckStakingAddresses(c)
	p := new(PegnetStaker)
	p.Config = c
	p.ID = id
	p.commands = commands
	return p
}

func CheckStakingAddresses(config *config.Config) {
	fctList, err := config.String("Staker.CoinbaseAddress")
	if err != nil {
		panic(fmt.Sprintf("could not extract the CoinbaseAddress: %s ", err.Error()))
	}

	// Make sure no addresses are duplicted in the CoinbaseAddress list.
	fctSlice := strings.Split(fctList, ",")
	for i, fctAddress1 := range fctSlice {
		for j, fctAddress2 := range fctSlice {
			if i == j {
				continue
			}
			if fctAddress1 == fctAddress2 {
				panic(fmt.Sprintf("CoinbaseAddress %s is duplicated in position %d and %d in the config file for pegnet", fctAddress1, i+1, j+1))
			}
		}
	}

	// Check the validity of each address in CoinbaseAddress list
	for i, fctAddress := range fctSlice {
		_, err = common.ConvertFCTtoRaw(fctAddress)
		if err != nil {
			panic(fmt.Sprintf("coinbase address %d [%s] is invalid: %s", i+1, fctAddress, err.Error()))
		}
	}

	// Check the EC address and its balance.  We are failing at zero, but maybe we should require 144?
	// At least a day's worth of ECs?
	ecAddress, err := config.String("Staker.ECAddress")
	if err != nil {
		panic("entry credit address is invalid: " + err.Error())
	}
	bal, err := factom.GetECBalance(ecAddress)
	if err != nil {
		panic(fmt.Sprintf("entry credit address [%s] is invalid: %s", ecAddress, err.Error()))
	}
	if bal == 0 {
		panic("EC Balance is zero for " + ecAddress)
	}

	days := float64(bal) / 144
	eqs := "==============================================================\n"
	io.WriteString(os.Stderr, fmt.Sprintf("\n%s", eqs))
	io.WriteString(os.Stderr, fmt.Sprintf("%s\n", ecAddress))
	io.WriteString(os.Stderr, fmt.Sprintf("ECBalance is %d\n", bal))
	io.WriteString(os.Stderr, fmt.Sprintf("Staking can run with this balance for ~%7.3f days\n", days))
	io.WriteString(os.Stderr, fmt.Sprintf("%s\n", eqs))
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
			time.Sleep(100 * time.Millisecond)
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
