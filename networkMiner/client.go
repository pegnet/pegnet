package networkMiner

import (
	"encoding/gob"
	"fmt"
	"net"
	"time"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/mining"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

type MiningClient struct {
	config *config.Config

	Host string

	Monitor  *common.FakeMonitor
	Grader   *opr.FakeGrader
	OPRMaker *mining.BlockingOPRMaker

	entryChannel chan *factom.Entry

	conn    net.Conn
	encoder *gob.Encoder
	decoder *gob.Decoder
}

func NewMiningClient(config *config.Config) *MiningClient {
	var err error
	s := new(MiningClient)
	s.config = config

	s.Host, err = config.String(common.ConfigCoordinatorLocation)
	if err != nil {
		panic(err)
	}

	s.entryChannel = make(chan *factom.Entry, 25)
	s.Monitor = common.NewFakeMonitor()
	s.Grader = opr.NewFakeGrader()
	s.OPRMaker = mining.NewBlockingOPRMaker()

	return s
}

func (c *MiningClient) Listeners() (common.IMonitor, opr.IGrader, mining.IOPRMaker) {
	return c.Monitor, c.Grader, c.OPRMaker
}

func (c *MiningClient) Connect() error {
	conn, err := net.Dial("tcp", c.Host)
	if err != nil {
		return err
	}
	log.WithTime(time.Now()).Infof("Connected to %s", c.Host)
	c.conn = conn
	c.initCoders()
	if err != nil {
		panic(err)
	}
	return nil
}

func (c *MiningClient) ConnectionLost(err error) {
	log.WithTime(time.Now()).WithFields(log.Fields{"host": c.Host, "time": time.Now().Format("15:04:05")}).WithError(err).Errorf("lost connection to host, retrying...")

	// Endless try to reconnect
	for {
		time.Sleep(1 * time.Second)
		err := c.Connect()
		if err != nil {
			log.WithFields(log.Fields{"host": c.Host, "time": time.Now().Format("15:04:05")}).WithError(err).Errorf("failed to reconnect, retrying...")
			time.Sleep(5 * time.Second)
			continue
		}
		break

	}
}

func (c *MiningClient) RunForwardEntries() {
	fLog := log.WithField("func", "MiningClient.RunForwardEntries()")
	for {
		select {
		case ent := <-c.entryChannel:
			err := c.encoder.Encode(&NetworkMessage{NetworkCommand: FactomEntry, Data: GobbedEntry{
				ExtIDs:  ent.ExtIDs,
				ChainID: ent.ChainID,
				Content: ent.Content,
			}})
			if err != nil {
				fLog.WithField("evt", "entry").WithError(err).Error("failed to send entry")
			} else {
				fLog.WithField("evt", "entry").WithField("entry", fmt.Sprintf("%x", ent.Hash())).Debugf("sent entry")
			}
		}
	}
}

func (c *MiningClient) Listen() {
	fLog := log.WithField("func", "MiningClient.Listen()")
	for {
		var m NetworkMessage
		err := c.decoder.Decode(&m)
		if err != nil {
			c.ConnectionLost(err)
		}

		switch m.NetworkCommand {
		case FactomEvent:
			evt := m.Data.(common.MonitorEvent)
			c.Monitor.FakeNotifyEvt(evt)
			fLog.WithField("evt", "factom").
				WithFields(log.Fields{
					"height": evt.Dbht,
					"minute": evt.Minute,
				}).Debug("network received alert")
		case GraderEvent:
			evt := m.Data.(opr.OPRs)
			c.Grader.EmitFakeEvent(evt)
		case ConstructedOPR:
			evt := m.Data.(opr.OraclePriceRecord)
			// We need to put our data in it
			id, _ := c.config.String("Miner.IdentityChain")
			evt.FactomDigitalID = id

			addr, _ := c.config.String(common.ConfigCoinbaseAddress)
			evt.CoinbasePNTAddress = addr

			c.OPRMaker.RecOPR(&evt)
		case Ping:
			err := c.encoder.Encode(&NetworkMessage{NetworkCommand: Pong})
			if err != nil {
				fLog.WithField("evt", "ping").WithError(err).Error("failed to pong")
			}
		case Pong:
			// Do nothing
		default:
			fLog.WithField("evt", "??").WithField("cmd", m.NetworkCommand).Warn("unrecognized message")
		}

	}
}

func (c *MiningClient) NewEntryForwarder() *mining.EntryForwarder {
	k, err := c.config.Int("Miner.RecordsPerBlock")
	if err != nil {
		panic(err)
	}
	f := mining.NewEntryForwarder(c.config, k, c.entryChannel)
	return f
}

func (c *MiningClient) initCoders() {
	c.encoder = gob.NewEncoder(c.conn)
	c.decoder = gob.NewDecoder(c.conn)
}
