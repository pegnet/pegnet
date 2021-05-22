package networkMiner

import (
	"context"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"net"
	"time"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/balances"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/mining"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	config "github.com/zpatrick/go-config"
)

// MiningClient talks to a coordinator. It receives events and trys to maintain
// a connection
type MiningClient struct {
	config *config.Config

	Host            string // Coordinator Location
	FactomDigitalID string

	Monitor  *common.FakeMonitor
	Grader   *opr.FakeGrader
	OPRMaker *mining.BlockingOPRMaker

	entryChannel  chan *factom.Entry
	UpstreamStats chan *mining.GroupMinerStats

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
	b := balances.NewBalanceTracker()
	// The "Fakes" allow us to emit events
	s.Monitor = common.NewFakeMonitor()
	s.Grader = opr.NewFakeGrader(config, b)
	s.OPRMaker = mining.NewBlockingOPRMaker()

	// We need to put our data in it
	id, err := config.String("Miner.IdentityChain")
	if err != nil {
		panic(err)
	}
	s.FactomDigitalID = id

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
	log.Infof("Connected to %s", c.Host)
	c.conn = conn
	c.initCoders()

	// Send over our tags
	err = c.encoder.Encode(NewNetworkMessage(AddTag, Tag{
		Key:   "id",
		Value: c.FactomDigitalID,
	}))
	if err != nil {
		log.WithField("evt", "tag").WithError(err).Error("failed to send tag")
	} else {
		log.WithField("evt", "tag").Debugf("sent tag")
	}
	return nil
}

// ConnectionLost will put us into a holding pattern to reconnect
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

// Forwarder will forward our channels to the coordinator
func (c *MiningClient) Forwarder() {
	fLog := log.WithField("func", "MiningClient.Forwarder()")
	for {
		select {
		case ent := <-c.entryChannel:
			err := c.encoder.Encode(NewNetworkMessage(FactomEntry, GobbedEntry{
				ExtIDs:  ent.ExtIDs,
				ChainID: ent.ChainID,
				Content: ent.Content,
			}))
			if err != nil {
				fLog.WithField("evt", "entry").WithError(err).Error("failed to send entry")
			} else {
				fLog.WithField("evt", "entry").WithField("entry", fmt.Sprintf("%x", ent.Hash())).Debugf("sent entry")
			}
		case s := <-c.UpstreamStats:
			err := c.encoder.Encode(NewNetworkMessage(MiningStatistics, *s))
			if err != nil {
				fLog.WithField("evt", "entry").WithError(err).Error("failed to send stats")
			} else {
				fLog.WithField("evt", "entry").Debugf("sent entry")
			}
		}
	}
}

// Listen for server events
func (c *MiningClient) Listen(cancel context.CancelFunc) {
	fLog := log.WithField("func", "MiningClient.Listen()")
	for {
		var m NetworkMessage
		err := c.decoder.Decode(&m)
		if err != nil {
			c.ConnectionLost(fmt.Errorf("decode: %s", err.Error()))
		}

		switch m.NetworkCommand {
		case CoordinatorError:
			// Any non-regular error message we want to send to the clients comes here.
			evt, ok := m.Data.(ErrorMessage)
			if ok && evt.Error != "" {
				fLog.WithField("evt", "error").WithError(fmt.Errorf(evt.Error)).Error("error from coordinator")
			}
		case FactomEvent:
			evt := m.Data.(common.MonitorEvent)

			// Drain anything that was left over
			if evt.Minute == 1 {
				c.OPRMaker.Drain()
			}

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
			devt, ok := m.Data.(opr.OraclePriceRecord)
			if !ok {
				// An error has occurred
				c.OPRMaker.RecOPR(nil)
				continue
			}

			evt := &devt

			// We need to put our data in it
			id, _ := c.config.String("Miner.IdentityChain")
			evt.FactomDigitalID = id

			addr, _ := c.config.String(common.ConfigCoinbaseAddress)
			evt.CoinbaseAddress = addr
			evt.OPRHash = nil // Reset the oprhash since we changed some fields

			c.OPRMaker.RecOPR(evt)
		case Ping:
			err := c.encoder.Encode(NewNetworkMessage(Pong, nil))
			if err != nil {
				fLog.WithField("evt", "ping").WithError(err).Error("failed to pong")
			}
		case SecretChallenge:
			// Respond to the challenge with our secret
			challenge, ok := m.Data.(AuthenticationChallenge)
			if !ok {
				fLog.Errorf("server did not send a proper challenge")
				cancel() // Cancel mining
				return
			}

			secret, err := c.config.String(common.ConfigCoordinatorSecret)
			if err != nil {
				// Do not return here, let the empty secret fail the challenge.
				fLog.WithError(err).Errorf("client is missing coordinator secret")
			}

			// Challenge Resp is sha256(secret+challenge)
			resp := sha256.Sum256([]byte(secret + challenge.Challenge))
			challenge.Response = fmt.Sprintf("%x", resp)
			err = c.encoder.Encode(NewNetworkMessage(SecretChallenge, challenge))
			if err != nil {
				fLog.WithError(err).Errorf("failed to respond to challenge")
				cancel()
				return
			}
		case RejectedConnection:
			// This means the server rejected us. Probably due to a failed challenge
			fLog.Errorf("Our connection to the coordinator was rejected. This is most likely due to failing the " +
				"authentication challenge. Ensure this miner has the same secret as the coordinator, and try again.")
			cancel()
			return
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
