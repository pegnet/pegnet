package staking

import (
	"context"
	"github.com/pegnet/pegnet/common"
	//"github.com/pegnet/pegnet/opr"
	"github.com/zpatrick/go-config"
)

// StakingCoordinator will poll data from exchange sources, make an SPR,
// get the SPR hash, and send it to staker.
type StakingCoordinator struct {
	config *config.Config

	// Factom blockchain related alerts
	FactomMonitor common.IMonitor
	//OPRGrader     opr.IGrader

	// Staker generates the opr hashes
	Staker *ControlledStaker

	// FactomEntryWriter writes the oprs to chain
	FactomEntryWriter IEntryWriter

	// Used when going over the network
	SPRMaker ISPRMaker
}

func NewStakingCoordinatorFromConfig(config *config.Config, monitor common.IMonitor) *StakingCoordinator {
	c := new(StakingCoordinator)
	c.config = config
	c.FactomMonitor = monitor
	k, err := config.Int("Staker.RecordsPerBlock")
	if err != nil {
		panic(err)
	}

	//c.OPRMaker = NewOPRMaker()

	c.FactomEntryWriter = NewEntryWriter(config, k)
	err = c.FactomEntryWriter.PopulateECAddress()
	if err != nil {
		panic(err)
	}

	return c
}

func (c *StakingCoordinator) InitStaker() error {
	c.Staker = c.NewStaker(0)
	return nil
}

func (c *StakingCoordinator) LaunchStaker(ctx context.Context) {

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

func (b *CommandBuilder) ResetRecords() *CommandBuilder {
	b.commands = append(b.commands, &StakerCommand{Command: ResetRecords, Data: nil})
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
