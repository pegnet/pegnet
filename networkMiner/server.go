package networkMiner

import (
	"context"
	"encoding/gob"
	"errors"
	"sync"

	"github.com/FactomProject/factom"
	"github.com/cenkalti/backoff"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

const (
	_ = iota
	Ping
	Pong
	FactomEvent
	GraderEvent
	ConstructedOPR
	FactomEntry
)

// Idk why the factom.entry does not work
type GobbedEntry struct {
	ChainID string   `json:"chainid"`
	ExtIDs  [][]byte `json:"extids"`
	Content []byte   `json:"content"`
}

func init() {
	gob.Register(common.MonitorEvent{})
	gob.Register(opr.OPRs{})
	gob.Register(GobbedEntry{})
	gob.Register([][]byte{})
	gob.Register(opr.OraclePriceRecord{})
}

type MiningServer struct {
	config *config.Config

	FactomMonitor common.IMonitor
	OPRGrader     opr.IGrader
	Host          string

	Server *TCPServer
	EC     *factom.ECAddress

	clientsLock sync.Mutex
	clients     map[int]*TCPClient
}

func NewMiningServer(config *config.Config, monitor common.IMonitor, grader opr.IGrader) *MiningServer {
	var err error
	s := new(MiningServer)
	s.config = config

	s.clients = make(map[int]*TCPClient)
	s.FactomMonitor = monitor
	s.OPRGrader = grader

	s.Host, err = config.String(common.ConfigCoordinatorListen)
	if err != nil {
		panic(err)
	}

	if ecadrStr, err := config.String("Miner.ECAddress"); err != nil {
		log.WithError(err).Fatalf("missing ec addr in config")
	} else {
		ecAdr, err := factom.FetchECAddress(ecadrStr)
		if err != nil {
			log.WithError(err).Fatalf("could not fetch ec addr")
		}
		s.EC = ecAdr
	}

	s.Server = NewTCPServer(s.Host)
	s.Server.onNewClientCallback = s.onNewClient
	s.Server.onNewMessage = s.onNewMessage
	s.Server.onClientConnectionClosed = s.onClientConnectionClosed

	return s
}

func (s *MiningServer) Listen() {
	s.Server.Listen()
}

func (c *MiningServer) ForwardMonitorEvents() {
	fLog := log.WithFields(log.Fields{"func": "ForwardMonitorEvents"})
	alert := c.FactomMonitor.NewListener()
	gAlerts := c.OPRGrader.GetAlert("evt-forwarder")
	var last common.MonitorEvent
	for {
		select {
		case fds := <-alert:
			m := new(NetworkMessage)
			m.NetworkCommand = FactomEvent
			m.Data = fds
			last = fds

			c.clientsLock.Lock()
			for _, c := range c.clients {
				err := c.SendNetworkCommand(m)
				if err != nil {
					fLog.WithField("evt", "factom").WithError(err).Error("failed to send")
				}
			}
			c.clientsLock.Unlock()
			fLog.WithFields(log.Fields{
				"height": fds.Dbht,
				"minute": fds.Minute,
			}).Debug("server sent alert")
		case g := <-gAlerts:
			//m := new(NetworkMessage)
			//m.NetworkCommand = GraderEvent
			//m.Data = *g
			//
			//c.clientsLock.Lock()
			//for _, c := range c.clients {
			//	err := c.SendNetworkCommand(m)
			//	if err != nil {
			//		fLog.WithField("evt", "grader").WithError(err).Error("failed to send")
			//	}
			//}
			//c.clientsLock.Unlock()

			oprobject, err := opr.NewOprFromWinners(context.Background(), 0, last.Dbht, c.config, g)
			if err != nil {
				fLog.WithField("evt", "grader").WithError(err).Error("failed to make opr")
			}

			m := new(NetworkMessage)
			m.NetworkCommand = ConstructedOPR
			m.Data = *oprobject
			c.clientsLock.Lock()
			for _, c := range c.clients {
				err := c.SendNetworkCommand(m)
				if err != nil {
					fLog.WithField("evt", "opr").WithError(err).Error("failed to send")
				}
			}
			c.clientsLock.Unlock()

			fLog.WithFields(c.Fields()).Info("sent opr to miners")

		}
	}
}

// onNewMessage is when the client messages us.
func (n *MiningServer) onNewMessage(c *TCPClient, message *NetworkMessage) {
	switch message.NetworkCommand {
	case Pong:
	case Ping:
		err := c.SendNetworkCommand(&NetworkMessage{NetworkCommand: Pong})
		if err != nil {
			log.WithFields(n.Fields()).WithError(err).Errorf("failed to pong")
		}
	case FactomEntry:
		b, ok := message.Data.(GobbedEntry)
		if !ok {
			log.WithFields(n.Fields()).Errorf("client did not send a proper entry")
			return
		}

		e := new(factom.Entry)
		e.ExtIDs = b.ExtIDs
		e.Content = b.Content
		e.ChainID = b.ChainID

		// Thread the write
		go func() {
			err := n.WriteEntry(e)
			if err != nil {
				log.WithFields(n.Fields()).WithError(err).Errorf("failed to submit entry from client")
			} else {
				log.WithFields(n.Fields()).WithField("client", c.id).Debugf("submitted entry %x", e.Hash())
			}
		}()
	default:
		log.WithFields(n.Fields()).WithField("cmd", message.NetworkCommand).Warn("command not recognized from client")
	}
}

func (s *MiningServer) onClientConnectionClosed(c *TCPClient, err error) {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	delete(s.clients, c.id)
	log.WithFields(s.Fields()).Info("Client disconnected")
}

func (s *MiningServer) onNewClient(c *TCPClient) {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	s.clients[c.id] = c
	log.WithFields(s.Fields()).WithField("id", c.id).Info("Client connected")

	err := c.SendNetworkCommand(&NetworkMessage{NetworkCommand: Ping})
	if err != nil {
		log.WithFields(s.Fields()).WithError(err).WithField("func", "onNewClient").Error("ping failed")
	}
}

func (s *MiningServer) WriteEntry(entry *factom.Entry) error {
	operation := func() error {
		_, err1 := factom.CommitEntry(entry, s.EC)
		_, err2 := factom.RevealEntry(entry)
		if err1 == nil && err2 == nil {
			return nil
		}

		return errors.New("Unable to commit entry to factom")
	}

	err := backoff.Retry(operation, common.PegExponentialBackOff())
	return err
}

func (s *MiningServer) Fields() log.Fields {
	// TODO: Is this threasafe?
	return log.Fields{"clients": len(s.clients)}
}
