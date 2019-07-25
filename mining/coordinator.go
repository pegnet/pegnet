package mining

import (
	"context"

	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

type MiningCoordinator struct {
	config *config.Config

	// Factom blockchain related alerts
	FactomMonitor common.IMonitor
	OPRGrader     opr.IGrader

	// Identities holds all the identities we can mine with.
	//	The more identities we have, the more records we can submit
	//Identities []MiningIdentity

	Miners            []*ControlledMiner
	FactomEntryWriter *EntryWriter

	MinerSubmissions chan MinerSubmission

	// Who we submit our stats too
	StatTracker *GlobalStatTracker

	// To unique ID miners
	minerIDCounter int
}

type MinerSubmission struct {
	ID         int // Miner ID
	OPRHash    []byte
	Nonce      []byte
	Difficulty uint64
}

type MiningIdentity struct {
	Identity string
	Best     *opr.NonceRanking
}

func NewMiningCoordinatorFromConfig(config *config.Config, monitor common.IMonitor, grader opr.IGrader, s *GlobalStatTracker) *MiningCoordinator {
	c := new(MiningCoordinator)
	c.config = config
	c.FactomMonitor = monitor
	c.OPRGrader = grader
	c.StatTracker = s
	k, err := config.Int("Miner.RecordsPerBlock")
	if err != nil {
		panic(err)
	}

	c.FactomEntryWriter = NewEntryWriter(config, k)
	err = c.FactomEntryWriter.PopulateECAddress()
	if err != nil {
		panic(err)
	}

	c.MinerSubmissions = make(chan MinerSubmission, 100)

	return c
}

func (c *MiningCoordinator) InitMinters() error {
	numMiners, err := c.config.Int("Miner.NumberOfMiners")
	if err != nil {
		return err
	}

	c.Miners = make([]*ControlledMiner, numMiners)
	for i := range c.Miners {
		c.Miners[i] = c.NewMiner(i)
	}

	return nil
}

func (c *MiningCoordinator) LaunchMiners(ctx context.Context) {
	mineLog := log.WithFields(log.Fields{"miner": "coordinator"})

	// TODO: Also tell Factom Monitor we are done listening
	alert := c.FactomMonitor.NewListener()
	gAlert := c.OPRGrader.GetAlert("coordinator")
	// Tell OPR grader we are no longer listening
	defer c.OPRGrader.StopAlert("coordinator")

	var oprTemplate *opr.OraclePriceRecord
	var oprHash []byte
	var err error
	var statsAggregate chan *SingleMinerStats

	// Launch!
	for _, m := range c.Miners {
		go m.Miner.Mine(ctx)
	}

	mining := false
MiningLoop:
	for {
		var fds common.MonitorEvent
		select {
		case fds = <-alert:
		case <-ctx.Done(): // If cancelled
			return
		}

		mineLog.WithFields(log.Fields{
			"height": fds.Dbht,
			"minute": fds.Minute,
		}).Debug("Miner received alert")
		switch fds.Minute {
		case 1:
			if !mining {
				mining = true
				// Need to get an OPR record
				oprTemplate, err = opr.NewOpr(ctx, 0, fds.Dbht, c.config, gAlert)
				if err == context.Canceled {
					continue MiningLoop // OPR cancelled
				}

				// Get the OPRHash for miners to mine.
				oprHash = oprTemplate.GetHash()

				// The consolidator that will write to the blockchain
				c.FactomEntryWriter = c.FactomEntryWriter.NextBlockWriter()
				c.FactomEntryWriter.SetOPR(oprTemplate)
				statsAggregate = make(chan *SingleMinerStats, len(c.Miners))
				command := BuildCommand().
					Aggregator(c.FactomEntryWriter). // New aggregate per block. Writes the top X records
					StatsAggregator(statsAggregate). // Stat collection per block
					ResetRecords().                  // Reset the miner's stats/difficulty/etc
					NewOPRHash(oprHash).             // New OPR hash to mine
					MinimumDifficulty(0).            // TODO: Set this from the cfg?
					ResumeMining().                  // Start mining
					Build()
				mineLog.Debug("Mining started")

				// Need to send to our miners
				for _, m := range c.Miners {
					m.SendCommand(&MinerCommand{Command: ResumeMining})
					m.SendCommand(command)
				}

			}
		case 9:
			if mining {
				mining = false
				command := BuildCommand().
					SubmitNonces().
					PauseMining().
					Build()

				// Need to send to our miners
				for _, m := range c.Miners {
					m.SendCommand(command)
				}
				// Write to blockchain
				c.FactomEntryWriter.CollectAndWrite(false)

				groupStats := NewGroupMinerStats()
				groupStats.BlockHeight = int(fds.Dbht)
				// Collect stats
				cm := 0
				for s := range statsAggregate {
					groupStats.Miners[s.ID] = s
					cm++
					if cm == len(c.Miners) {
						break
					}
				}
				c.StatTracker.MiningStatsChannel <- groupStats

			}
		}
	}
}

type ControlledMiner struct {
	Miner          *PegnetMiner
	CommandChannel chan *MinerCommand
}

func (c *MiningCoordinator) NewMiner(id int) *ControlledMiner {
	m := new(ControlledMiner)
	channel := make(chan *MinerCommand, 10)
	m.Miner = NewPegnetMinerFromConfig(c.config, id, channel)
	m.CommandChannel = channel
	return m

}

func (c *ControlledMiner) SendCommand(command *MinerCommand) {
	c.CommandChannel <- command
}

type CommandBuilder struct {
	command  *MinerCommand
	commands []*MinerCommand
}

func BuildCommand() *CommandBuilder {
	c := new(CommandBuilder)
	c.command = new(MinerCommand)
	c.command.Command = BatchCommand
	return c
}

func (b *CommandBuilder) NewOPRHash(oprhash []byte) *CommandBuilder {
	b.commands = append(b.commands, &MinerCommand{Command: NewOPRHash, Data: oprhash})
	return b
}

func (b *CommandBuilder) ResetRecords() *CommandBuilder {
	b.commands = append(b.commands, &MinerCommand{Command: ResetRecords, Data: nil})
	return b
}

func (b *CommandBuilder) MinimumDifficulty(min uint64) *CommandBuilder {
	b.commands = append(b.commands, &MinerCommand{Command: MinimumAccept, Data: min})
	return b
}

func (b *CommandBuilder) SubmitNonces() *CommandBuilder {
	b.commands = append(b.commands, &MinerCommand{Command: SubmitNonces, Data: nil})
	return b
}

func (b *CommandBuilder) PauseMining() *CommandBuilder {
	b.commands = append(b.commands, &MinerCommand{Command: PauseMining, Data: nil})
	return b
}

func (b *CommandBuilder) ResumeMining() *CommandBuilder {
	b.commands = append(b.commands, &MinerCommand{Command: ResumeMining, Data: nil})
	return b
}

func (b *CommandBuilder) Aggregator(w *EntryWriter) *CommandBuilder {
	b.commands = append(b.commands, &MinerCommand{Command: RecordAggregator, Data: w})
	return b
}

func (b *CommandBuilder) StatsAggregator(w chan *SingleMinerStats) *CommandBuilder {
	b.commands = append(b.commands, &MinerCommand{Command: StatsAggregator, Data: w})
	return b
}

func (b *CommandBuilder) Build() *MinerCommand {
	b.command.Data = b.commands
	return b.command
}
