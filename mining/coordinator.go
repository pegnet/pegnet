package mining

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

// MiningCoordinator can coordinate multiple miners. This object will
// poll data from exchange sources, make an OPR, get the OPR hash, and send
// it to miners for them to work on. Once the miners get a top X records, we
// will aggregate and submit.
//	TODO: Make the coordinator look at the difficulties in the last block, and determine
//			a minimum based on that.
type MiningCoordinator struct {
	config *config.Config

	// Factom blockchain related alerts
	FactomMonitor common.IMonitor
	OPRGrader     opr.IGrader

	// Miners mine the opr hashes
	Miners []*ControlledMiner
	// FactomEntryWriter writes the oprs to chain
	FactomEntryWriter IEntryWriter

	// Who we submit our stats too
	StatTracker *GlobalStatTracker

	// Used when going over the network
	OPRMaker IOPRMaker

	// To give miners unique IDs
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

func NewNetworkedMiningCoordinatorFromConfig(config *config.Config, monitor common.IMonitor, grader opr.IGrader, s *GlobalStatTracker) *MiningCoordinator {
	c := new(MiningCoordinator)
	c.config = config
	c.FactomMonitor = monitor
	c.OPRGrader = grader
	c.StatTracker = s

	// OPRMaker and writer set by client

	return c
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

	c.OPRMaker = NewOPRMaker()

	c.FactomEntryWriter = NewEntryWriter(config, k)
	err = c.FactomEntryWriter.PopulateECAddress()
	if err != nil {
		panic(err)
	}

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
	opr.InitLX()
	mineLog := log.WithFields(log.Fields{"id": "coordinator"})

	// TODO: Also tell Factom Monitor we are done listening
	alert := c.FactomMonitor.NewListener()
	gAlert := c.OPRGrader.GetAlert("coordinator")
	// Tell OPR grader we are no longer listening
	defer c.OPRGrader.StopAlert("coordinator")

	var oprTemplate *opr.OraclePriceRecord
	var oprHash []byte
	var statsAggregate chan *SingleMinerStats

	// Launch!
	for _, m := range c.Miners {
		go m.Miner.Mine(ctx)
	}

	first := false
	mineLog.Info("Miners launched. Waiting for minute 1 to start mining...")
	mining := false
MiningLoop:
	for {
		var fds common.MonitorEvent
		select {
		case fds = <-alert:
		case <-ctx.Done(): // If cancelled
			return
		}

		hLog := mineLog.WithFields(log.Fields{
			"height": fds.Dbht,
			"minute": fds.Minute,
		})
		if !first {
			// On the first minute log how far away to mining
			hLog.Infof("On minute %d. %d minutes until minute 1 before mining starts.", fds.Minute, common.Abs(int(fds.Minute)-11)%10)
			first = true
		}

		hLog.Debug("Miner received alert")
		switch fds.Minute {
		case 1:
			// First check if we have the funds to mine
			bal, err := c.FactomEntryWriter.ECBalance()
			if err != nil {
				hLog.WithError(err).WithField("action", "balance-query").Error("failed to mine this block")
				continue MiningLoop // OPR cancelled
			}
			if bal == 0 {
				hLog.WithError(fmt.Errorf("entry credit balance is 0")).WithField("action", "balance-query").Error("will not mine, out of entry credits")
				continue MiningLoop // OPR cancelled
			}

			if !mining {
				mining = true
				// Need to get an OPR record
				oprTemplate, err = c.OPRMaker.NewOPR(ctx, 0, fds.Dbht, c.config, gAlert)
				if err == context.Canceled {
					mining = false
					continue MiningLoop // OPR cancelled
				}
				if err != nil {
					hLog.WithError(err).Error("failed to mine this block")
					mining = false
					continue MiningLoop // OPR cancelled
				}

				// Get the OPRHash for miners to mine.
				oprHash = oprTemplate.GetHash()

				// The consolidator that will write to the blockchain
				c.FactomEntryWriter = c.FactomEntryWriter.NextBlockWriter()
				c.FactomEntryWriter.SetOPR(oprTemplate)

				// We aggregate mining stats per block
				statsAggregate = make(chan *SingleMinerStats, len(c.Miners))

				command := BuildCommand().
					Aggregator(c.FactomEntryWriter).                  // New aggregate per block. Writes the top X records
					StatsAggregator(statsAggregate).                  // Stat collection per block
					ResetRecords().                                   // Reset the miner's stats/difficulty/etc
					NewOPRHash(oprHash).                              // New OPR hash to mine
					MinimumDifficulty(oprTemplate.MinimumDifficulty). // Floor difficulty to use
					ResumeMining().                                   // Start mining
					Build()

				// Need to send to our miners
				for _, m := range c.Miners {
					m.SendCommand(command)
				}

				buf := make([]byte, 8)
				binary.BigEndian.PutUint64(buf, oprTemplate.MinimumDifficulty)
				hLog.WithField("mindiff", fmt.Sprintf("%x", buf)).Info("Begin mining new OPR")

			}
		case 9:
			if mining {
				mining = false
				command := BuildCommand().
					SubmitNonces(). // Submit nonces to aggregator
					PauseMining().  // Pause mining until further notice
					Build()

				// Need to send to our miners
				for _, m := range c.Miners {
					m.SendCommand(command)
				}

				// Write to blockchain (this is non blocking)
				c.FactomEntryWriter.CollectAndWrite(false)

				groupStats := NewGroupMinerStats("main", int(fds.Dbht))
				// Collect stats
				cm := 0
				for s := range statsAggregate {
					groupStats.Miners[s.ID] = s
					cm++
					if cm == len(c.Miners) {
						break
					}
				}

				// groupStats is the stats for all the miners for this block
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

// CommandBuilder just let's me use building syntax to build commands
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

func (b *CommandBuilder) Aggregator(w IEntryWriter) *CommandBuilder {
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
