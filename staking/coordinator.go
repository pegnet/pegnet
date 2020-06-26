package staking

import (
	"context"
	"fmt"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/spr"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

// StakingCoordinator will poll data from exchange sources, make an SPR,
// get the SPR hash, and send it to staker.
type StakingCoordinator struct {
	config *config.Config

	// Factom blockchain related alerts
	FactomMonitor common.IMonitor

	// Staker generates the spr hashes
	Staker *ControlledStaker

	// FactomEntryWriter writes the sprs to chain
	FactomEntryWriter IEntryWriter

	// Used when going over the network
	SPRMaker ISPRMaker
}

func NewStakingCoordinatorFromConfig(config *config.Config, monitor common.IMonitor) *StakingCoordinator {
	c := new(StakingCoordinator)
	c.config = config
	c.FactomMonitor = monitor
	c.SPRMaker = NewSPRMaker()
	c.FactomEntryWriter = NewEntryWriter(config)

	err := c.FactomEntryWriter.PopulateECAddress()
	if err != nil {
		panic(err)
	}

	return c
}

func (c *StakingCoordinator) InitStaker() error {
	c.Staker = c.NewStaker(1)
	return nil
}

func (c *StakingCoordinator) LaunchStaker(ctx context.Context) {
	stakeLog := log.WithFields(log.Fields{"id": "coordinator"})

	alert := c.FactomMonitor.NewListener()

	var sprTemplate *spr.StakingPriceRecord
	var sprHash []byte

	// Launch
	go c.Staker.Staker.Stake(ctx)

	first := false
	stakeLog.Info("Staker launched. Waiting for minute 1 to start staking...")
	staking := false
StakingLoop:
	for {
		var fds common.MonitorEvent
		select {
		case fds = <-alert:
		case <-ctx.Done(): // If cancelled
			return
		}

		hLog := stakeLog.WithFields(log.Fields{
			"height": fds.Dbht,
			"minute": fds.Minute,
		})
		if !first {
			// On the first minute log how far away to staking
			hLog.Infof("On minute %d. %d minutes until minute 1 before staking starts.", fds.Minute, common.Abs(int(fds.Minute)-11)%10)
			first = true
		}

		hLog.Debug("Staker received alert")
		switch fds.Minute {
		case 1:
			// First check if we have the funds to stake
			bal, err := c.FactomEntryWriter.ECBalance()
			if err != nil {
				hLog.WithError(err).WithField("action", "balance-query").Error("failed to stake this block")
				continue StakingLoop // SPR cancelled
			}
			if bal == 0 {
				hLog.WithError(fmt.Errorf("entry credit balance is 0")).WithField("action", "balance-query").Error("will not stake, out of entry credits")
				continue StakingLoop // SPR cancelled
			}

			if !staking {
				staking = true
				hLog.Debug("Minute 1 for Staker")

				// Need to get an SPR record
				sprTemplate, err = c.SPRMaker.NewSPR(ctx, fds.Dbht, c.config)
				if err == context.Canceled {
					staking = false
					continue StakingLoop // SPR cancelled
				}
				if err != nil {
					hLog.WithError(err).Error("failed to stake this block")
					staking = false
					continue StakingLoop // SPR cancelled
				}

				// Get the SPRHash for stakers to stake.
				sprHash = sprTemplate.GetHash()

				// The consolidator that will write to the blockchain
				c.FactomEntryWriter = c.FactomEntryWriter.NextBlockWriter()
				c.FactomEntryWriter.SetSPR(sprTemplate)

				command := BuildCommand().
					Aggregator(c.FactomEntryWriter). // New aggregate per block. Writes the top X records
					NewSPRHash(sprHash).             // New SPR hash to stake
					ResumeStaking().                 // Start staking
					Build()
				c.Staker.SendCommand(command)

				hLog.Debug("Begin staking new SPR")
			}
		case 8:
			if staking {
				staking = false
				hLog.Debug("Minute 8 for Staker")

				command := BuildCommand().
					PauseStaking(). // Pause staking until further notice
					Build()

				// Need to send to staker
				c.Staker.SendCommand(command)

				// Write to blockchain (this is non blocking)
				c.FactomEntryWriter.CollectAndWrite(false)
			}
		}
	}
}

type ControlledStaker struct {
	Staker         *PegnetStaker
	CommandChannel chan *StakerCommand
}

func (c *StakingCoordinator) NewStaker(id int) *ControlledStaker {
	m := new(ControlledStaker)
	channel := make(chan *StakerCommand, 10)
	m.Staker = NewPegnetStakerFromConfig(c.config, id, channel)
	m.CommandChannel = channel
	return m
}

func (c *ControlledStaker) SendCommand(command *StakerCommand) {
	c.CommandChannel <- command
}

// CommandBuilder just let's me use building syntax to build commands
type CommandBuilder struct {
	command  *StakerCommand
	commands []*StakerCommand
}

func BuildCommand() *CommandBuilder {
	c := new(CommandBuilder)
	c.command = new(StakerCommand)
	c.command.Command = BatchCommand
	return c
}

func (b *CommandBuilder) NewSPRHash(sprhash []byte) *CommandBuilder {
	b.commands = append(b.commands, &StakerCommand{Command: NewSPRHash, Data: sprhash})
	return b
}

func (b *CommandBuilder) PauseStaking() *CommandBuilder {
	b.commands = append(b.commands, &StakerCommand{Command: PauseStaking, Data: nil})
	return b
}

func (b *CommandBuilder) ResumeStaking() *CommandBuilder {
	b.commands = append(b.commands, &StakerCommand{Command: ResumeStaking, Data: nil})
	return b
}

func (b *CommandBuilder) Aggregator(w IEntryWriter) *CommandBuilder {
	b.commands = append(b.commands, &StakerCommand{Command: RecordAggregator, Data: w})
	return b
}

func (b *CommandBuilder) Build() *StakerCommand {
	b.command.Data = b.commands
	return b.command
}
