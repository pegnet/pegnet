package networkMiner

import (
	"encoding/gob"
	"fmt"
	"net"

	"github.com/pegnet/pegnet/mining"

	"github.com/pegnet/pegnet/common"
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

	conn    net.Conn
	encoder *gob.Encoder
	decoder *gob.Decoder
}

func NewMiningClient(config *config.Config) *MiningClient {
	var err error
	s := new(MiningClient)
	s.config = config

	s.Host, err = config.String("Miner.MiningCoordinator")
	if err != nil {
		panic(err)
	}

	s.Monitor = common.NewFakeMonitor()
	s.Grader = opr.NewFakeGrader()
	s.OPRMaker = mining.NewBlockingOPRMaker()

	return s
}

func (c *MiningClient) Listeners() (common.IMonitor, opr.IGrader, mining.IOPRMaker) {
	return c.Monitor, c.Grader, c.OPRMaker
}

func (c *MiningClient) Connect() {
	log.Infof("Connected to %s", c.Host)
	conn, err := net.Dial("tcp", c.Host)
	if err != nil {
		panic(err)
	}
	c.conn = conn
	c.initCoders()
	fmt.Println("Connection established")
}

func (c *MiningClient) Listen() {
	fLog := log.WithField("func", "MiningClient.Listen()")
	for {
		var m NetworkMessage
		err := c.decoder.Decode(&m)
		if err != nil {
			panic(err)
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
			c.OPRMaker.RecOPR(&evt)
		case Ping:
			err := c.encoder.Encode(&NetworkMessage{NetworkCommand: Pong})
			if err != nil {
				fLog.WithField("evt", "ping").WithError(err).Error("failed to pong")
			}
		default:
			fLog.WithField("evt", "??").WithField("cmd", m.NetworkCommand).Warn("unrecognized message")
		}

	}
}

func (c *MiningClient) initCoders() {
	c.encoder = gob.NewEncoder(c.conn)
	c.decoder = gob.NewDecoder(c.conn)
}
